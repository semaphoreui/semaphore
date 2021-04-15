package tasks

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"

	log "github.com/Sirupsen/logrus"

	"github.com/ansible-semaphore/semaphore/util"
)

const (
	taskFailStatus = "error"
	taskTypeID     = "task"
)

type task struct {
	store       db.Store
	task        db.Task
	template    db.Template
	sshKey      db.AccessKey
	inventory   db.Inventory
	repository  db.Repository
	environment db.Environment
	users       []int
	projectID   int
	hosts       []string
	alertChat   string
	alert       bool
	prepared    bool
}

func (t *task) fail() {
	t.task.Status = taskFailStatus
	t.updateStatus()
	t.sendMailAlert()
	t.sendTelegramAlert()
}

func (t *task) prepareRun() {
	t.prepared = false

	defer func() {
		log.Info("Stopped preparing task " + strconv.Itoa(t.task.ID))
		log.Info("Release resource locker with task " + strconv.Itoa(t.task.ID))
		resourceLocker <- &resourceLock{lock: false, holder: t}

		objType := taskTypeID
		desc := "Task ID " + strconv.Itoa(t.task.ID) + " (" + t.template.Alias + ")" + " finished - " + strings.ToUpper(t.task.Status)

		_, err := t.store.CreateEvent(db.Event{
			ProjectID:   &t.projectID,
			ObjectType:  &objType,
			ObjectID:    &t.task.ID,
			Description: &desc,
		})

		if err != nil {
			t.panicOnError(err, "Fatal error inserting an event")
		}
	}()

	t.log("Preparing: " + strconv.Itoa(t.task.ID))

	err := checkTmpDir(util.Config.TmpPath)
	if err != nil {
		t.log("Creating tmp dir failed: " + err.Error())
		t.fail()
		return
	}

	if err := t.populateDetails(); err != nil {
		t.log("Error: " + err.Error())
		t.fail()
		return
	}

	objType := taskTypeID
	desc := "Task ID " + strconv.Itoa(t.task.ID) + " (" + t.template.Alias + ")" + " is preparing"
	_, err = t.store.CreateEvent(db.Event{
		ProjectID:   &t.projectID,
		ObjectType:  &objType,
		ObjectID:    &t.task.ID,
		Description: &desc,
	})

	if err != nil {
		t.log("Fatal error inserting an event")
		panic(err)
	}

	t.log("Prepare task with template: " + t.template.Alias + "\n")

	if err := t.installKey(t.repository.SSHKey); err != nil {
		t.log("Failed installing ssh key for repository access: " + err.Error())
		t.fail()
		return
	}

	if err := t.updateRepository(); err != nil {
		t.log("Failed updating repository: " + err.Error())
		t.fail()
		return
	}

	if err := t.installInventory(); err != nil {
		t.log("Failed to install inventory: " + err.Error())
		t.fail()
		return
	}

	if err := t.installRequirements(); err != nil {
		t.log("Running galaxy failed: " + err.Error())
		t.fail()
		return
	}

	// todo: write environment

	if stderr, err := t.listPlaybookHosts(); err != nil {
		t.log("Listing playbook hosts failed: " + err.Error() + "\n" + stderr)
		t.fail()
		return
	}

	t.prepared = true
}

func (t *task) run() {
	defer func() {
		log.Info("Stopped running task " + strconv.Itoa(t.task.ID))
		log.Info("Release resource locker with task " + strconv.Itoa(t.task.ID))
		resourceLocker <- &resourceLock{lock: false, holder: t}

		now := time.Now()
		t.task.End = &now
		t.updateStatus()

		objType := taskTypeID
		desc := "Task ID " + strconv.Itoa(t.task.ID) + " (" + t.template.Alias + ")" + " finished - " + strings.ToUpper(t.task.Status)

		_, err := t.store.CreateEvent(db.Event{
			ProjectID:   &t.projectID,
			ObjectType:  &objType,
			ObjectID:    &t.task.ID,
			Description: &desc,
		})

		if err != nil {
			t.log("Fatal error inserting an event")
			panic(err)
		}
	}()

	{
		now := time.Now()
		t.task.Status = "running"
		t.task.Start = &now

		t.updateStatus()
	}

	objType := taskTypeID
	desc := "Task ID " + strconv.Itoa(t.task.ID) + " (" + t.template.Alias + ")" + " is running"


	_, err := t.store.CreateEvent(db.Event{
		ProjectID:   &t.projectID,
		ObjectType:  &objType,
		ObjectID:    &t.task.ID,
		Description: &desc,
	})

	if err != nil {
		t.log("Fatal error inserting an event")
		panic(err)
	}

	t.log("Started: " + strconv.Itoa(t.task.ID))
	t.log("Run task with template: " + t.template.Alias + "\n")

	if err := t.runPlaybook(); err != nil {
		t.log("Running playbook failed: " + err.Error())
		t.fail()
		return
	}

	t.task.Status = "success"
	t.updateStatus()
}

func (t *task) prepareError(err error, errMsg string) error {
	if err == sql.ErrNoRows {
		t.log(errMsg)
		return err
	}

	if err != nil {
		t.fail()
		panic(err)
	}

	return nil
}

func (t *task) fetch(errMsg string, ptr interface{}, query string, args ...interface{}) error {
	err := t.store.Sql().SelectOne(ptr, query, args...)
	if err == sql.ErrNoRows {
		t.log(errMsg)
		return err
	}

	if err != nil {
		t.fail()
		panic(err)
	}

	return nil
}

//nolint: gocyclo
func (t *task) populateDetails() error {
	// get template
	var err error

	t.template, err = t.store.GetTemplate(t.projectID, t.task.TemplateID)
	if err != nil {
		return t.prepareError(err, "Template not found!")
	}

	// get project alert setting
	project, err := t.store.GetProject(t.template.ProjectID)
	if err != nil {
		return t.prepareError(err, "Project not found!")
	}

	t.alert = project.Alert
	t.alertChat = project.AlertChat

	// get project users
	users, err := t.store.GetProjectUsers(t.template.ProjectID, db.RetrieveQueryParams{})
	if err != nil {
		return t.prepareError(err, "Users not found!")
	}

	t.users = []int{}
	for _, user := range users {
		t.users = append(t.users, user.ID)
	}

	// get access key
	t.sshKey, err = t.store.GetAccessKey(t.template.ProjectID, t.template.SSHKeyID)
	if err != nil {
		return t.prepareError(err, "Template AccessKey not found!")
	}

	if t.sshKey.Type != "ssh" {
		t.log("Non ssh-type keys are currently not supported: " + t.sshKey.Type)
		return errors.New("unsupported SSH Key")
	}

	// get inventory
	t.inventory, err = t.store.GetInventory(t.template.ProjectID, t.template.InventoryID)
	if err != nil {
		return t.prepareError(err, "Template Inventory not found!")
	}

	// get inventory services key

	if t.inventory.KeyID != nil {
		if err := t.fetch("Inventory AccessKey not found!", &t.inventory.Key, "select * from access_key where id=?", *t.inventory.KeyID); err != nil {
			return err
		}
	}

	// get inventory ssh key
	if t.inventory.SSHKeyID != nil {
		if err := t.fetch("Inventory Ssh Key not found!", &t.inventory.SSHKey, "select * from access_key where id=?", *t.inventory.SSHKeyID); err != nil {
			return err
		}
	}

	// get repository
	if err := t.fetch("Repository not found!", &t.repository, "select * from project__repository where id=?", t.template.RepositoryID); err != nil {
		return err
	}

	// get repository access key
	if err := t.fetch("Repository Access Key not found!", &t.repository.SSHKey, "select * from access_key where id=?", t.repository.SSHKeyID); err != nil {
		return err
	}
	if t.repository.SSHKey.Type != "ssh" {
		t.log("Repository Access Key is not 'SSH': " + t.repository.SSHKey.Type)
		return errors.New("unsupported SSH Key")
	}

	// get environment
	if len(t.task.Environment) == 0 && t.template.EnvironmentID != nil {
		err := t.fetch("Environment not found", &t.environment, "select * from project__environment where id=?", *t.template.EnvironmentID)
		if err != nil {
			return err
		}
	} else if len(t.task.Environment) > 0 {
		t.environment.JSON = t.task.Environment
	}

	return nil
}

func (t *task) installKey(key db.AccessKey) error {
	t.log("access key " + key.Name + " installed")

	path := key.GetPath()
	if key.Key != nil {
		if err := ioutil.WriteFile(path+"-cert.pub", []byte(*key.Key), 0600); err != nil {
			return err
		}
	}

	return ioutil.WriteFile(path, []byte(*key.Secret), 0600)
}

func (t *task) updateRepository() error {
	repoName := "repository_" + strconv.Itoa(t.repository.ID)
	_, err := os.Stat(util.Config.TmpPath + "/" + repoName)

	cmd := exec.Command("git") //nolint: gas
	cmd.Dir = util.Config.TmpPath

	gitSSHCommand := "ssh -o StrictHostKeyChecking=no -i " + t.repository.SSHKey.GetPath()
	cmd.Env = t.envVars(util.Config.TmpPath, util.Config.TmpPath, &gitSSHCommand)

	repoURL, repoTag := t.repository.GitURL, "master"
	if split := strings.Split(repoURL, "#"); len(split) > 1 {
		repoURL, repoTag = split[0], split[1]
	}

	if err != nil && os.IsNotExist(err) {
		t.log("Cloning repository " + repoURL)
		cmd.Args = append(cmd.Args, "clone", "--recursive", "--branch", repoTag, repoURL, repoName)
	} else if err != nil {
		return err
	} else {
		t.log("Updating repository " + repoURL)
		cmd.Dir += "/" + repoName
		cmd.Args = append(cmd.Args, "pull", "origin", repoTag)
	}

	t.logCmd(cmd)
	return cmd.Run()
}

func (t *task) installRequirements() error {
	requirementsFilePath := fmt.Sprintf("%s/repository_%d/roles/requirements.yml", util.Config.TmpPath, t.repository.ID)
	requirementsHashFilePath := fmt.Sprintf("%s/repository_%d/requirements.md5", util.Config.TmpPath, t.repository.ID)

	if _, err := os.Stat(requirementsFilePath); err != nil {
		t.log("No roles/requirements.yml file found. Skip galaxy install process.\n")
		return nil
	}

	if hasRequirementsChanges(requirementsFilePath, requirementsHashFilePath) {
		if err := t.runGalaxy([]string{
			"install",
			"-r",
			"roles/requirements.yml",
			"--force",
		}); err != nil {
			return err
		}
		if err := writeMD5Hash(requirementsFilePath, requirementsHashFilePath); err != nil {
			return err
		}
	} else {
		t.log("roles/requirements.yml has no changes. Skip galaxy install process.\n")
	}

	return nil
}

func (t *task) runGalaxy(args []string) error {
	cmd := exec.Command("ansible-galaxy", args...) //nolint: gas
	cmd.Dir = util.Config.TmpPath + "/repository_" + strconv.Itoa(t.repository.ID)

	gitSSHCommand := "ssh -o StrictHostKeyChecking=no -i " + t.repository.SSHKey.GetPath()
	cmd.Env = t.envVars(util.Config.TmpPath, cmd.Dir, &gitSSHCommand)

	t.logCmd(cmd)
	return cmd.Run()
}

func (t *task) listPlaybookHosts() (string, error) {

	if util.Config.ConcurrencyMode == "project" {
		return "", nil
	}

	args, err := t.getPlaybookArgs()
	if err != nil {
		return "", err
	}
	args = append(args, "--list-hosts")

	cmd := exec.Command("ansible-playbook", args...) //nolint: gas
	cmd.Dir = util.Config.TmpPath + "/repository_" + strconv.Itoa(t.repository.ID)
	cmd.Env = t.envVars(util.Config.TmpPath, cmd.Dir, nil)

	var errb bytes.Buffer
	cmd.Stderr = &errb

	out, err := cmd.Output()

	re := regexp.MustCompile(`(?m)^\\s{6}(.*)$`)
	matches := re.FindAllSubmatch(out, 20)
	hosts := make([]string, len(matches))
	for i := range matches {
		hosts[i] = string(matches[i][1])
	}
	t.hosts = hosts
	return errb.String(), err
}

func (t *task) runPlaybook() error {
	args, err := t.getPlaybookArgs()
	if err != nil {
		return err
	}
	cmd := exec.Command("ansible-playbook", args...) //nolint: gas
	cmd.Dir = util.Config.TmpPath + "/repository_" + strconv.Itoa(t.repository.ID)
	cmd.Env = t.envVars(util.Config.TmpPath, cmd.Dir, nil)

	t.logCmd(cmd)
	cmd.Stdin = strings.NewReader("")
	return cmd.Run()
}

//nolint: gocyclo
func (t *task) getPlaybookArgs() ([]string, error) {
	playbookName := t.task.Playbook
	if len(playbookName) == 0 {
		playbookName = t.template.Playbook
	}

	var inventory string
	switch t.inventory.Type {
	case "file":
		inventory = t.inventory.Inventory
	default:
		inventory = util.Config.TmpPath + "/inventory_" + strconv.Itoa(t.task.ID)
	}

	args := []string{
		"-i", inventory,
	}

	if t.inventory.SSHKeyID != nil {
		args = append(args, "--private-key="+t.inventory.SSHKey.GetPath())
	}

	if t.task.Debug {
		args = append(args, "-vvvv")
	}

	if t.task.DryRun {
		args = append(args, "--check")
	}

	if len(t.environment.JSON) > 0 {
		var js map[string]interface{}
		err := json.Unmarshal([]byte(t.environment.JSON), &js)
		if err != nil {
			t.log("JSON is not valid")
			return nil, err
		}

		extraVar, err := removeCommandEnvironment(t.environment.JSON, js)
		if err != nil {
			t.log("Could not remove command environment, if existant it will be passed to --extra-vars. This is not fatal but be aware of side effects")
		}

		args = append(args, "--extra-vars", extraVar)
	}

	var templateExtraArgs []string
	if t.template.Arguments != nil {
		err := json.Unmarshal([]byte(*t.template.Arguments), &templateExtraArgs)
		if err != nil {
			t.log("Could not unmarshal arguments to []string")
			return nil, err
		}
	}

	var taskExtraArgs []string
	if t.task.Arguments != nil {
		err := json.Unmarshal([]byte(*t.task.Arguments), &taskExtraArgs)
		if err != nil {
			t.log("Could not unmarshal arguments to []string")
			return nil, err
		}
	}

	if t.template.OverrideArguments {
		args = templateExtraArgs
	} else {
		args = append(args, templateExtraArgs...)
		args = append(args, taskExtraArgs...)
		args = append(args, playbookName)
	}
	return args, nil
}

func (t *task) envVars(home string, pwd string, gitSSHCommand *string) []string {
	env := os.Environ()
	env = append(env, fmt.Sprintf("HOME=%s", home))
	env = append(env, fmt.Sprintf("PWD=%s", pwd))
	env = append(env, fmt.Sprintln("PYTHONUNBUFFERED=1"))
	//env = append(env, fmt.Sprintln("GIT_FLUSH=1"))

	env = append(env, extractCommandEnvironment(t.environment.JSON)...)

	if gitSSHCommand != nil {
		env = append(env, fmt.Sprintf("GIT_SSH_COMMAND=%s", *gitSSHCommand))
	}

	return env
}

func hasRequirementsChanges(requirementsFilePath string, requirementsHashFilePath string) bool {
	oldFileMD5HashBytes, err := ioutil.ReadFile(requirementsHashFilePath)
	if err != nil {
		return true
	}

	newFileMD5Hash, err := helpers.GetMD5Hash(requirementsFilePath)
	if err != nil {
		return true
	}

	return string(oldFileMD5HashBytes) != newFileMD5Hash
}

func writeMD5Hash(requirementsFile string, requirementsHashFile string) error {
	newFileMD5Hash, err := helpers.GetMD5Hash(requirementsFile)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(requirementsHashFile, []byte(newFileMD5Hash), 0644)
}

// extractCommandEnvironment unmarshalls a json string, extracts the ENV key from it and returns it as
// []string where strings are in key=value format
func extractCommandEnvironment(envJSON string) []string {
	env := make([]string, 0)
	var js map[string]interface{}
	err := json.Unmarshal([]byte(envJSON), &js)
	if err == nil {
		if cfg, ok := js["ENV"]; ok {
			switch v := cfg.(type) {
			case map[string]interface{}:
				for key, val := range v {
					env = append(env, fmt.Sprintf("%s=%s", key, val))
				}
			}
		}
	}
	return env
}

// removeCommandEnvironment removes the ENV key from task environments and returns the resultant json encoded string
// which can be passed as the --extra-vars flag values
func removeCommandEnvironment(envJSON string, envJs map[string]interface{}) (string, error) {
	if _, ok := envJs["ENV"]; ok {
		delete(envJs, "ENV")
		ev, err := json.Marshal(envJs)
		if err != nil {
			return envJSON, err
		}
		envJSON = string(ev)
	}

	return envJSON, nil

}

// checkTmpDir checks to see if the temporary directory exists
// and if it does not attempts to create it
func checkTmpDir(path string) error {
	var err error
	if _, err = os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(path, 0700)
		}
	}
	return err
}