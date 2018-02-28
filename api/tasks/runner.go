package tasks

import (
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

	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
)

type task struct {
	task        db.Task
	template    db.Template
	sshKey      db.AccessKey
	inventory   db.Inventory
	repository  db.Repository
	environment db.Environment
	users       []int
	projectID   int
	alert       bool
	hosts       []string
	prepared    bool
	alert_chat  string
}

func (t *task) fail() {
	t.task.Status = "error"
	t.updateStatus()
	t.sendMailAlert()
	t.sendTelegramAlert()
}

func (t *task) prepareRun() {
	t.prepared = false

	defer func() {
		fmt.Println("Stopped preparing task")

		objType := "task"
		desc := "Task ID " + strconv.Itoa(t.task.ID) + " (" + t.template.Alias + ")" + " finished - " + strings.ToUpper(t.task.Status)
		if err := (db.Event{
			ProjectID:   &t.projectID,
			ObjectType:  &objType,
			ObjectID:    &t.task.ID,
			Description: &desc,
		}.Insert()); err != nil {
			t.log("Fatal error inserting an event")
			panic(err)
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

	objType := "task"
	desc := "Task ID " + strconv.Itoa(t.task.ID) + " (" + t.template.Alias + ")" + " is preparing"
	if err := (db.Event{
		ProjectID:   &t.projectID,
		ObjectType:  &objType,
		ObjectID:    &t.task.ID,
		Description: &desc,
	}.Insert()); err != nil {
		t.log("Fatal error inserting an event")
		panic(err)
	}

	t.log("Prepare task with template: " + t.template.Alias + "\n")

	if err := t.installKey(t.repository.SshKey); err != nil {
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

	if err := t.runGalaxy(); err != nil {
		t.log("Running galaxy failed: " + err.Error())
		t.fail()
		return
	}

	// todo: write environment

	if err := t.listPlaybookHosts(); err != nil {
		t.log("Listing playbook hosts failed: " + err.Error())
		t.fail()
		return
	}

	t.prepared = true
}

func (t *task) run() {
	defer func() {
		fmt.Println("Stopped running tasks")
		resourceLocker <- &resourceLock{lock: false, holder: t}

		now := time.Now()
		t.task.End = &now
		t.updateStatus()

		objType := "task"
		desc := "Task ID " + strconv.Itoa(t.task.ID) + " (" + t.template.Alias + ")" + " finished - " + strings.ToUpper(t.task.Status)
		if err := (db.Event{
			ProjectID:   &t.projectID,
			ObjectType:  &objType,
			ObjectID:    &t.task.ID,
			Description: &desc,
		}.Insert()); err != nil {
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

	objType := "task"
	desc := "Task ID " + strconv.Itoa(t.task.ID) + " (" + t.template.Alias + ")" + " is running"
	if err := (db.Event{
		ProjectID:   &t.projectID,
		ObjectType:  &objType,
		ObjectID:    &t.task.ID,
		Description: &desc,
	}.Insert()); err != nil {
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

func (t *task) fetch(errMsg string, ptr interface{}, query string, args ...interface{}) error {
	err := db.Mysql.SelectOne(ptr, query, args...)
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

func (t *task) populateDetails() error {
	// get template
	if err := t.fetch("Template not found!", &t.template, "select * from project__template where id=?", t.task.TemplateID); err != nil {
		return err
	}

	type AlertSettings struct {
		Alert     bool   `db:"alert"`
		AlertChat string `db:"alert_chat"`
	}

	var project db.Project
	// get project alert setting
	if err := t.fetch("Alert setting not found!", &project, "select alert, alert_chat from project where id=?", t.template.ProjectID); err != nil {
		return err
	}
	t.alert = project.Alert
	t.alert_chat = project.AlertChat

	// get project users
	var users []struct {
		ID int `db:"id"`
	}
	if _, err := db.Mysql.Select(&users, "select user_id as id from project__user where project_id=?", t.template.ProjectID); err != nil {
		return err
	}

	t.users = []int{}
	for _, user := range users {
		t.users = append(t.users, user.ID)
	}

	// get access key
	if err := t.fetch("Template Access Key not found!", &t.sshKey, "select * from access_key where id=?", t.template.SshKeyID); err != nil {
		return err
	}

	if t.sshKey.Type != "ssh" {
		t.log("Non ssh-type keys are currently not supported: " + t.sshKey.Type)
		return errors.New("Unsupported SSH Key")
	}

	// get inventory
	if err := t.fetch("Template Inventory not found!", &t.inventory, "select * from project__inventory where id=?", t.template.InventoryID); err != nil {
		return err
	}

	// get inventory services key
	if t.inventory.KeyID != nil {
		if err := t.fetch("Inventory AccessKey not found!", &t.inventory.Key, "select * from access_key where id=?", *t.inventory.KeyID); err != nil {
			return err
		}
	}

	// get inventory ssh key
	if t.inventory.SshKeyID != nil {
		if err := t.fetch("Inventory Ssh Key not found!", &t.inventory.SshKey, "select * from access_key where id=?", *t.inventory.SshKeyID); err != nil {
			return err
		}
	}

	// get repository
	if err := t.fetch("Repository not found!", &t.repository, "select * from project__repository where id=?", t.template.RepositoryID); err != nil {
		return err
	}

	// get repository access key
	if err := t.fetch("Repository Access Key not found!", &t.repository.SshKey, "select * from access_key where id=?", t.repository.SshKeyID); err != nil {
		return err
	}
	if t.repository.SshKey.Type != "ssh" {
		t.log("Repository Access Key is not 'SSH': " + t.repository.SshKey.Type)
		return errors.New("Unsupported SSH Key")
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

	cmd := exec.Command("git")
	cmd.Dir = util.Config.TmpPath

	gitSSHCommand := "ssh -o StrictHostKeyChecking=no -i " + t.repository.SshKey.GetPath()
	cmd.Env = t.envVars(util.Config.TmpPath, util.Config.TmpPath, &gitSSHCommand)

	repoURL, repoTag := t.repository.GitUrl, "master"
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

func (t *task) runGalaxy() error {
	args := []string{
		"install",
		"-r",
		"roles/requirements.yml",
		"-p",
		"./roles/",
		"--force",
	}

	cmd := exec.Command("ansible-galaxy", args...)
	cmd.Dir = util.Config.TmpPath + "/repository_" + strconv.Itoa(t.repository.ID)

	gitSSHCommand := "ssh -o StrictHostKeyChecking=no -i " + t.repository.SshKey.GetPath()
	cmd.Env = t.envVars(util.Config.TmpPath, cmd.Dir, &gitSSHCommand)

	if _, err := os.Stat(cmd.Dir + "/roles/requirements.yml"); err != nil {
		return nil
	}

	t.logCmd(cmd)
	return cmd.Run()
}

func (t *task) listPlaybookHosts() error {
	args, err := t.getPlaybookArgs()
	if err != nil {
		return err
	}
	args = append(args, "--list-hosts")

	cmd := exec.Command("ansible-playbook", args...)
	cmd.Dir = util.Config.TmpPath + "/repository_" + strconv.Itoa(t.repository.ID)
	cmd.Env = t.envVars(util.Config.TmpPath, cmd.Dir, nil)

	out, err := cmd.Output()
	re := regexp.MustCompile("(?m)^\\s{6}(.*)$")
	matches := re.FindAllSubmatch(out, 20)
	hosts := make([]string, len(matches))
	for i, _ := range matches {
		hosts[i] = string(matches[i][1])
	}
	t.hosts = hosts
	return err
}

func (t *task) runPlaybook() error {
	args, err := t.getPlaybookArgs()
	if err != nil {
		return err
	}
	cmd := exec.Command("ansible-playbook", args...)
	cmd.Dir = util.Config.TmpPath + "/repository_" + strconv.Itoa(t.repository.ID)
	cmd.Env = t.envVars(util.Config.TmpPath, cmd.Dir, nil)

	t.logCmd(cmd)
	cmd.Stdin = strings.NewReader("")
	return cmd.Run()
}

func (t *task) getPlaybookArgs() ([]string, error) {
	playbookName := t.task.Playbook
	if len(playbookName) == 0 {
		playbookName = t.template.Playbook
	}

	args := []string{
		"-i", util.Config.TmpPath + "/inventory_" + strconv.Itoa(t.task.ID),
	}

	if t.inventory.SshKeyID != nil {
		args = append(args, "--private-key="+t.inventory.SshKey.GetPath())
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

		args = append(args, "--extra-vars", t.environment.JSON)
	}

	var extraArgs []string
	if t.template.Arguments != nil {
		err := json.Unmarshal([]byte(*t.template.Arguments), &extraArgs)
		if err != nil {
			t.log("Could not unmarshal arguments to []string")
			return nil, err
		}
	}

	if t.template.OverrideArguments {
		args = extraArgs
	} else {
		args = append(args, extraArgs...)
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

	if gitSSHCommand != nil {
		env = append(env, fmt.Sprintf("GIT_SSH_COMMAND=%s", *gitSSHCommand))
	}

	return env
}

// checkTmpDir checks to see if the temporary directory exists
// and if it does not attempts to create it
func checkTmpDir(path string) error {
	var err error = nil
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(path, 1777)
		}
	}
	return err
}
