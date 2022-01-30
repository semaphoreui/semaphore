package tasks

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/ansible-semaphore/semaphore/lib"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/api/sockets"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
)

const (
	gitURLFilePrefix = "file://"
)

type TaskRunner struct {
	task        db.Task
	template    db.Template
	inventory   db.Inventory
	repository  db.Repository
	environment db.Environment

	users     []int
	hosts     []string
	alertChat string
	alert     bool
	prepared  bool
	process   *os.Process
	pool      *TaskPool
}

//func (t *TaskRunner) validate() error {
//	if t.task.ProjectID != t.template.ProjectID ||
//		t.task.ProjectID != t.inventory.ProjectID ||
//		t.task.ProjectID != t.repository.ProjectID ||
//		t.task.ProjectID != t.environment.ProjectID {
//		return fmt.Errorf("invalid project id")
//	}
//	return nil
//}

func getMD5Hash(filepath string) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func (t *TaskRunner) getRepoPath() string {
	repo := lib.GitRepository{
		Logger:     t,
		TemplateID: t.template.ID,
		Repository: t.repository,
	}

	return repo.GetFullPath()
}

func (t *TaskRunner) setStatus(status string) {
	if t.task.Status == db.TaskStoppingStatus {
		switch status {
		case db.TaskFailStatus:
			status = db.TaskStoppedStatus
		case db.TaskStoppedStatus:
		default:
			panic("stopping TaskRunner cannot be " + status)
		}
	}

	t.task.Status = status

	t.updateStatus()

	if status == db.TaskFailStatus {
		t.sendMailAlert()
	}

	if status == db.TaskSuccessStatus || status == db.TaskFailStatus {
		t.sendTelegramAlert()
	}
}

func (t *TaskRunner) updateStatus() {
	for _, user := range t.users {
		b, err := json.Marshal(&map[string]interface{}{
			"type":        "update",
			"start":       t.task.Start,
			"end":         t.task.End,
			"status":      t.task.Status,
			"task_id":     t.task.ID,
			"template_id": t.task.TemplateID,
			"project_id":  t.task.ProjectID,
			"version":     t.task.Version,
		})

		util.LogPanic(err)

		sockets.Message(user, b)
	}

	if err := t.pool.store.UpdateTask(t.task); err != nil {
		t.panicOnError(err, "Failed to update TaskRunner status")
	}
}

func (t *TaskRunner) fail() {
	t.setStatus(db.TaskFailStatus)
}

func (t *TaskRunner) destroyKeys() {
	err := t.repository.SSHKey.Destroy()
	if err != nil {
		t.Log("Can't destroy repository key, error: " + err.Error())
	}

	err = t.inventory.SSHKey.Destroy()
	if err != nil {
		t.Log("Can't destroy inventory user key, error: " + err.Error())
	}

	err = t.inventory.BecomeKey.Destroy()
	if err != nil {
		t.Log("Can't destroy inventory become user key, error: " + err.Error())
	}

	err = t.template.VaultKey.Destroy()
	if err != nil {
		t.Log("Can't destroy inventory vault password file, error: " + err.Error())
	}
}

func (t *TaskRunner) createTaskEvent() {
	objType := db.EventTask
	desc := "Task ID " + strconv.Itoa(t.task.ID) + " (" + t.template.Alias + ")" + " finished - " + strings.ToUpper(t.task.Status)

	_, err := t.pool.store.CreateEvent(db.Event{
		UserID:      t.task.UserID,
		ProjectID:   &t.task.ProjectID,
		ObjectType:  &objType,
		ObjectID:    &t.task.ID,
		Description: &desc,
	})

	if err != nil {
		t.panicOnError(err, "Fatal error inserting an event")
	}
}

func (t *TaskRunner) prepareRun() {
	t.prepared = false

	defer func() {
		log.Info("Stopped preparing TaskRunner " + strconv.Itoa(t.task.ID))
		log.Info("Release resource locker with TaskRunner " + strconv.Itoa(t.task.ID))
		resourceLocker <- &resourceLock{lock: false, holder: t}

		t.createTaskEvent()
	}()

	t.Log("Preparing: " + strconv.Itoa(t.task.ID))

	err := checkTmpDir(util.Config.TmpPath)
	if err != nil {
		t.Log("Creating tmp dir failed: " + err.Error())
		t.fail()
		return
	}

	err = t.populateDetails()
	if err != nil {
		t.Log("Error: " + err.Error())
		t.fail()
		return
	}

	objType := db.EventTask
	desc := "Task ID " + strconv.Itoa(t.task.ID) + " (" + t.template.Alias + ")" + " is preparing"
	_, err = t.pool.store.CreateEvent(db.Event{
		UserID:      t.task.UserID,
		ProjectID:   &t.task.ProjectID,
		ObjectType:  &objType,
		ObjectID:    &t.task.ID,
		Description: &desc,
	})

	if err != nil {
		t.Log("Fatal error inserting an event")
		panic(err)
	}

	t.Log("Prepare TaskRunner with template: " + t.template.Alias + "\n")

	t.updateStatus()

	//if err := t.installKey(t.repository.SSHKey, db.AccessKeyUsagePrivateKey); err != nil {
	if err := t.repository.SSHKey.Install(db.AccessKeyUsagePrivateKey); err != nil {
		t.Log("Failed installing ssh key for repository access: " + err.Error())
		t.fail()
		return
	}

	if strings.HasPrefix(t.repository.GitURL, gitURLFilePrefix) {
		repositoryPath := strings.TrimPrefix(t.repository.GitURL, gitURLFilePrefix)
		if _, err := os.Stat(repositoryPath); err != nil {
			t.Log("Failed in finding static repository at " + repositoryPath + ": " + err.Error())
			t.fail()
			return
		}
	} else {
		if err := t.updateRepository(); err != nil {
			t.Log("Failed updating repository: " + err.Error())
			t.fail()
			return
		}
	}

	if err := t.checkoutRepository(); err != nil {
		t.Log("Failed to checkout repository to required commit: " + err.Error())
		t.fail()
		return
	}

	if err := t.installInventory(); err != nil {
		t.Log("Failed to install inventory: " + err.Error())
		t.fail()
		return
	}

	if err := t.installRequirements(); err != nil {
		t.Log("Running galaxy failed: " + err.Error())
		t.fail()
		return
	}

	if err := t.installVaultKeyFile(); err != nil {
		t.Log("Failed to install vault password file: " + err.Error())
		t.fail()
		return
	}

	// todo: write environment

	if stderr, err := t.listPlaybookHosts(); err != nil {
		t.Log("Listing playbook hosts failed: " + err.Error() + "\n" + stderr)
		t.fail()
		return
	}

	t.prepared = true
}

func (t *TaskRunner) run() {
	defer func() {
		log.Info("Stopped running TaskRunner " + strconv.Itoa(t.task.ID))
		log.Info("Release resource locker with TaskRunner " + strconv.Itoa(t.task.ID))
		resourceLocker <- &resourceLock{lock: false, holder: t}

		now := time.Now()
		t.task.End = &now
		t.updateStatus()
		t.createTaskEvent()
		t.destroyKeys()
	}()

	if t.task.Status == db.TaskStoppingStatus {
		t.setStatus(db.TaskStoppedStatus)
		return
	}

	now := time.Now()
	t.task.Start = &now
	t.setStatus(db.TaskRunningStatus)

	objType := db.EventTask
	desc := "Task ID " + strconv.Itoa(t.task.ID) + " (" + t.template.Alias + ")" + " is running"

	_, err := t.pool.store.CreateEvent(db.Event{
		UserID:      t.task.UserID,
		ProjectID:   &t.task.ProjectID,
		ObjectType:  &objType,
		ObjectID:    &t.task.ID,
		Description: &desc,
	})

	if err != nil {
		t.Log("Fatal error inserting an event")
		panic(err)
	}

	t.Log("Started: " + strconv.Itoa(t.task.ID))
	t.Log("Run TaskRunner with template: " + t.template.Alias + "\n")

	if t.task.Status == db.TaskStoppingStatus {
		t.setStatus(db.TaskStoppedStatus)
		return
	}

	err = t.runPlaybook()
	if err != nil {
		t.Log("Running playbook failed: " + err.Error())
		t.fail()
		return
	}

	t.setStatus(db.TaskSuccessStatus)

	templates, err := t.pool.store.GetTemplates(t.task.ProjectID, db.TemplateFilter{
		BuildTemplateID: &t.task.TemplateID,
		AutorunOnly:     true,
	}, db.RetrieveQueryParams{})
	if err != nil {
		t.Log("Running playbook failed: " + err.Error())
		return
	}

	for _, tpl := range templates {
		_, err = t.pool.AddTask(db.Task{
			TemplateID:  tpl.ID,
			ProjectID:   tpl.ProjectID,
			BuildTaskID: &t.task.ID,
		}, nil, tpl.ProjectID)
		if err != nil {
			t.Log("Running playbook failed: " + err.Error())
			continue
		}
	}
}

func (t *TaskRunner) prepareError(err error, errMsg string) error {
	if err == db.ErrNotFound {
		t.Log(errMsg)
		return err
	}

	if err != nil {
		t.fail()
		panic(err)
	}

	return nil
}

//nolint: gocyclo
func (t *TaskRunner) populateDetails() error {
	// get template
	var err error

	t.template, err = t.pool.store.GetTemplate(t.task.ProjectID, t.task.TemplateID)
	if err != nil {
		return t.prepareError(err, "Template not found!")
	}

	// get project alert setting
	project, err := t.pool.store.GetProject(t.template.ProjectID)
	if err != nil {
		return t.prepareError(err, "Project not found!")
	}

	t.alert = project.Alert
	t.alertChat = project.AlertChat

	// get project users
	users, err := t.pool.store.GetProjectUsers(t.template.ProjectID, db.RetrieveQueryParams{})
	if err != nil {
		return t.prepareError(err, "Users not found!")
	}

	t.users = []int{}
	for _, user := range users {
		t.users = append(t.users, user.ID)
	}

	// get inventory
	t.inventory, err = t.pool.store.GetInventory(t.template.ProjectID, t.template.InventoryID)
	if err != nil {
		return t.prepareError(err, "Template Inventory not found!")
	}

	// get repository
	t.repository, err = t.pool.store.GetRepository(t.template.ProjectID, t.template.RepositoryID)

	if err != nil {
		return err
	}

	// get environment
	if t.template.EnvironmentID != nil {
		t.environment, err = t.pool.store.GetEnvironment(t.template.ProjectID, *t.template.EnvironmentID)
		if err != nil {
			return err
		}
	}

	if t.task.Environment != "" {
		environment := make(map[string]interface{})
		if t.environment.JSON != "" {
			err = json.Unmarshal([]byte(t.task.Environment), &environment)
			if err != nil {
				return err
			}
		}

		taskEnvironment := make(map[string]interface{})
		err = json.Unmarshal([]byte(t.environment.JSON), &taskEnvironment)
		if err != nil {
			return err
		}

		for k, v := range taskEnvironment {
			environment[k] = v
		}

		var ev []byte
		ev, err = json.Marshal(environment)
		if err != nil {
			return err
		}

		t.environment.JSON = string(ev)
	}

	return nil
}

func (t *TaskRunner) installVaultKeyFile() error {
	if t.template.VaultKeyID == nil {
		return nil
	}

	return t.template.VaultKey.Install(db.AccessKeyUsageVault)
}

func (t *TaskRunner) checkoutRepository() error {
	repo := lib.GitRepository{
		Logger:     t,
		TemplateID: t.template.ID,
		Repository: t.repository,
	}

	err := repo.ValidateRepo()

	if err != nil {
		return err
	}

	if t.task.CommitHash != nil {
		// checkout to commit if it is provided for TaskRunner
		return repo.Checkout(*t.task.CommitHash)
	}

	// store commit to TaskRunner table

	commitHash, err := repo.GetLastCommitHash()

	if err != nil {
		return err
	}

	commitMessage, _ := repo.GetLastCommitMessage()

	t.task.CommitHash = &commitHash
	t.task.CommitMessage = commitMessage

	return t.pool.store.UpdateTask(t.task)
}

func (t *TaskRunner) updateRepository() error {
	repo := lib.GitRepository{
		Logger:     t,
		TemplateID: t.template.ID,
		Repository: t.repository,
	}

	err := repo.ValidateRepo()

	if err != nil {
		if !os.IsNotExist(err) {
			err = os.RemoveAll(repo.GetFullPath())
			if err != nil {
				return err
			}
		}
		return repo.Clone()
	}

	if repo.CanBePulled() {
		err = repo.Pull()
		if err == nil {
			return nil
		}
	}

	err = os.RemoveAll(repo.GetFullPath())
	if err != nil {
		return err
	}

	return repo.Clone()
}

func (t *TaskRunner) installRequirements() error {
	requirementsFilePath := fmt.Sprintf("%s/roles/requirements.yml", t.getRepoPath())
	requirementsHashFilePath := fmt.Sprintf("%s/requirements.md5", t.getRepoPath())

	if _, err := os.Stat(requirementsFilePath); err != nil {
		t.Log("No roles/requirements.yml file found. Skip galaxy install process.\n")
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
		t.Log("roles/requirements.yml has no changes. Skip galaxy install process.\n")
	}

	return nil
}

func (t *TaskRunner) runGalaxy(args []string) error {
	cmd := exec.Command("ansible-galaxy", args...) //nolint: gas
	cmd.Dir = t.getRepoPath()

	t.setCmdEnvironment(cmd, t.repository.SSHKey.GetSshCommand())

	t.LogCmd(cmd)
	return cmd.Run()
}

func (t *TaskRunner) listPlaybookHosts() (string, error) {

	if util.Config.ConcurrencyMode == "project" {
		return "", nil
	}

	args, err := t.getPlaybookArgs()
	if err != nil {
		return "", err
	}
	args = append(args, "--list-hosts")

	cmd := exec.Command("ansible-playbook", args...) //nolint: gas
	cmd.Dir = t.getRepoPath()
	t.setCmdEnvironment(cmd, "")

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

func (t *TaskRunner) runPlaybook() (err error) {
	args, err := t.getPlaybookArgs()
	if err != nil {
		return
	}
	cmd := exec.Command("ansible-playbook", args...) //nolint: gas
	cmd.Dir = t.getRepoPath()
	t.setCmdEnvironment(cmd, "")

	t.LogCmd(cmd)
	cmd.Stdin = strings.NewReader("")
	err = cmd.Start()
	if err != nil {
		return
	}
	t.process = cmd.Process
	err = cmd.Wait()
	return
}

func (t *TaskRunner) getExtraVars() (str string, err error) {
	extraVars := make(map[string]interface{})

	if t.environment.JSON != "" {
		err = json.Unmarshal([]byte(t.environment.JSON), &extraVars)
		if err != nil {
			return
		}
	}

	delete(extraVars, "ENV")

	taskDetails := make(map[string]interface{})

	if t.task.Message != "" {
		taskDetails["message"] = t.task.Message
	}

	if t.task.UserID != nil {
		var user db.User
		user, err = t.pool.store.GetUser(*t.task.UserID)
		if err == nil {
			taskDetails["username"] = user.Username
		}
	}

	if t.template.Type != db.TemplateTask {
		taskDetails["type"] = t.template.Type
		incomingVersion := t.task.GetIncomingVersion(t.pool.store)
		if incomingVersion != nil {
			taskDetails["incoming_version"] = incomingVersion
		}
		if t.template.Type == db.TemplateBuild {
			taskDetails["target_version"] = t.task.Version
		}
	}

	vars := make(map[string]interface{})
	vars["task_details"] = taskDetails
	extraVars["semaphore_vars"] = vars

	ev, err := json.Marshal(extraVars)
	if err != nil {
		return
	}

	str = string(ev)

	return
}

//nolint: gocyclo
func (t *TaskRunner) getPlaybookArgs() (args []string, err error) {
	playbookName := t.task.Playbook
	if playbookName == "" {
		playbookName = t.template.Playbook
	}

	var inventory string
	switch t.inventory.Type {
	case db.InventoryFile:
		inventory = t.inventory.Inventory
	default:
		inventory = util.Config.TmpPath + "/inventory_" + strconv.Itoa(t.task.ID)
	}

	args = []string{
		"-i", inventory,
	}

	if t.inventory.SSHKeyID != nil {
		switch t.inventory.SSHKey.Type {
		case db.AccessKeySSH:
			args = append(args, "--private-key="+t.inventory.SSHKey.GetPath())
		case db.AccessKeyLoginPassword:
			args = append(args, "--extra-vars=@"+t.inventory.SSHKey.GetPath())
		case db.AccessKeyNone:
		default:
			err = fmt.Errorf("access key does not suite for inventory's User Access Key")
			return
		}
	}

	if t.inventory.BecomeKeyID != nil {
		switch t.inventory.BecomeKey.Type {
		case db.AccessKeyLoginPassword:
			args = append(args, "--extra-vars=@"+t.inventory.BecomeKey.GetPath())
		case db.AccessKeyNone:
		default:
			err = fmt.Errorf("access key does not suite for inventory's Become User Access Key")
			return
		}
	}

	if t.task.Debug {
		args = append(args, "-vvvv")
	}

	if t.task.DryRun {
		args = append(args, "--check")
	}

	if t.template.VaultKeyID != nil {
		args = append(args, "--vault-password-file", t.template.VaultKey.GetPath())
	}

	extraVars, err := t.getExtraVars()
	if err != nil {
		t.Log(err.Error())
		t.Log("Could not remove command environment, if existant it will be passed to --extra-vars. This is not fatal but be aware of side effects")
	} else if extraVars != "" {
		args = append(args, "--extra-vars", extraVars)
	}

	var templateExtraArgs []string
	if t.template.Arguments != nil {
		err = json.Unmarshal([]byte(*t.template.Arguments), &templateExtraArgs)
		if err != nil {
			t.Log("Invalid format of the template extra arguments, must be valid JSON")
			return
		}
	}

	var taskExtraArgs []string

	if t.template.AllowOverrideArgsInTask && t.task.Arguments != nil {
		err = json.Unmarshal([]byte(*t.task.Arguments), &taskExtraArgs)
		if err != nil {
			t.Log("Invalid format of the TaskRunner extra arguments, must be valid JSON")
			return
		}
	}

	args = append(args, templateExtraArgs...)
	args = append(args, taskExtraArgs...)
	args = append(args, playbookName)

	return
}

func (t *TaskRunner) setCmdEnvironment(cmd *exec.Cmd, gitSSHCommand string) {
	env := os.Environ()
	env = append(env, fmt.Sprintf("HOME=%s", util.Config.TmpPath))
	env = append(env, fmt.Sprintf("PWD=%s", cmd.Dir))
	env = append(env, fmt.Sprintln("PYTHONUNBUFFERED=1"))
	env = append(env, fmt.Sprintln("GIT_TERMINAL_PROMPT=0"))
	env = append(env, extractCommandEnvironment(t.environment.JSON)...)

	if gitSSHCommand != "" {
		env = append(env, fmt.Sprintf("GIT_SSH_COMMAND=%s", gitSSHCommand))
	}
	cmd.Env = env
}

func hasRequirementsChanges(requirementsFilePath string, requirementsHashFilePath string) bool {
	oldFileMD5HashBytes, err := ioutil.ReadFile(requirementsHashFilePath)
	if err != nil {
		return true
	}

	newFileMD5Hash, err := getMD5Hash(requirementsFilePath)
	if err != nil {
		return true
	}

	return string(oldFileMD5HashBytes) != newFileMD5Hash
}

func writeMD5Hash(requirementsFile string, requirementsHashFile string) error {
	newFileMD5Hash, err := getMD5Hash(requirementsFile)
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
