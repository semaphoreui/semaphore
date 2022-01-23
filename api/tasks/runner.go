package tasks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/api/sockets"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
)

const (
	taskRunningStatus  = "running"
	taskWaitingStatus  = "waiting"
	taskStoppingStatus = "stopping"
	taskStoppedStatus  = "stopped"
	taskSuccessStatus  = "success"
	taskFailStatus     = "error"
	gitURLFilePrefix   = "file://"
)

type task struct {
	store       db.Store
	task        db.Task
	template    db.Template
	inventory   db.Inventory
	repository  db.Repository
	environment db.Environment
	users       []int
	projectID   int
	hosts       []string
	alertChat   string
	alert       bool
	prepared    bool
	process     *os.Process
}

func (t *task) getRepoName() string {
	return "repository_" + strconv.Itoa(t.repository.ID) + "_" + strconv.Itoa(t.template.ID)
}

func (t *task) getRepoPath() string {
	return path.Join(util.Config.TmpPath, t.getRepoName())
}

func (t *task) validateRepo() error {
	_, err := os.Stat(t.getRepoPath())
	return err
}

func (t *task) setStatus(status string) {
	if t.task.Status == taskStoppingStatus {
		switch status {
		case taskFailStatus:
			status = taskStoppedStatus
		case taskStoppedStatus:
		default:
			panic("stopping task cannot be " + status)
		}
	}

	t.task.Status = status

	t.updateStatus()

	if status == taskFailStatus {
		t.sendMailAlert()
	}

	if status == taskSuccessStatus || status == taskFailStatus {
		t.sendTelegramAlert()
	}
}

func (t *task) updateStatus() {
	for _, user := range t.users {
		b, err := json.Marshal(&map[string]interface{}{
			"type":        "update",
			"start":       t.task.Start,
			"end":         t.task.End,
			"status":      t.task.Status,
			"task_id":     t.task.ID,
			"template_id": t.task.TemplateID,
			"project_id":  t.projectID,
			"version":     t.task.Version,
		})

		util.LogPanic(err)

		sockets.Message(user, b)
	}

	if err := t.store.UpdateTask(t.task); err != nil {
		t.panicOnError(err, "Failed to update task status")
	}
}

func (t *task) fail() {
	t.setStatus(taskFailStatus)
}

func (t *task) destroyKeys() {
	err := t.repository.SSHKey.Destroy()
	if err != nil {
		t.log("Can't destroy repository key, error: " + err.Error())
	}

	err = t.inventory.SSHKey.Destroy()
	if err != nil {
		t.log("Can't destroy inventory user key, error: " + err.Error())
	}

	err = t.inventory.BecomeKey.Destroy()
	if err != nil {
		t.log("Can't destroy inventory become user key, error: " + err.Error())
	}

	err = t.template.VaultKey.Destroy()
	if err != nil {
		t.log("Can't destroy inventory vault password file, error: " + err.Error())
	}
}

func (t *task) createTaskEvent() {
	objType := db.EventTask
	desc := "Task ID " + strconv.Itoa(t.task.ID) + " (" + t.template.Alias + ")" + " finished - " + strings.ToUpper(t.task.Status)

	_, err := t.store.CreateEvent(db.Event{
		UserID:      t.task.UserID,
		ProjectID:   &t.projectID,
		ObjectType:  &objType,
		ObjectID:    &t.task.ID,
		Description: &desc,
	})

	if err != nil {
		t.panicOnError(err, "Fatal error inserting an event")
	}
}

func (t *task) prepareRun() {
	t.prepared = false

	defer func() {
		log.Info("Stopped preparing task " + strconv.Itoa(t.task.ID))
		log.Info("Release resource locker with task " + strconv.Itoa(t.task.ID))
		resourceLocker <- &resourceLock{lock: false, holder: t}

		t.createTaskEvent()
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

	objType := db.EventTask
	desc := "Task ID " + strconv.Itoa(t.task.ID) + " (" + t.template.Alias + ")" + " is preparing"
	_, err = t.store.CreateEvent(db.Event{
		UserID:      t.task.UserID,
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

	t.updateStatus()

	//if err := t.installKey(t.repository.SSHKey, db.AccessKeyUsagePrivateKey); err != nil {
	if err := t.repository.SSHKey.Install(db.AccessKeyUsagePrivateKey); err != nil {
		t.log("Failed installing ssh key for repository access: " + err.Error())
		t.fail()
		return
	}

	if strings.HasPrefix(t.repository.GitURL, gitURLFilePrefix) {
		repositoryPath := strings.TrimPrefix(t.repository.GitURL, gitURLFilePrefix)
		if _, err := os.Stat(repositoryPath); err != nil {
			t.log("Failed in finding static repository at " + repositoryPath + ": " + err.Error())
			t.fail()
			return
		}
	} else {
		if err := t.updateRepository(); err != nil {
			t.log("Failed updating repository: " + err.Error())
			t.fail()
			return
		}
	}

	if err := t.checkoutRepository(); err != nil {
		t.log("Failed to checkout repository to required commit: " + err.Error())
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

	if err := t.installVaultKeyFile(); err != nil {
		t.log("Failed to install vault password file: " + err.Error())
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
		t.createTaskEvent()
		t.destroyKeys()
	}()

	if t.task.Status == taskStoppingStatus {
		t.setStatus(taskStoppedStatus)
		return
	}

	now := time.Now()
	t.task.Start = &now
	t.setStatus(taskRunningStatus)

	objType := db.EventTask
	desc := "Task ID " + strconv.Itoa(t.task.ID) + " (" + t.template.Alias + ")" + " is running"

	_, err := t.store.CreateEvent(db.Event{
		UserID:      t.task.UserID,
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

	if t.task.Status == taskStoppingStatus {
		t.setStatus(taskStoppedStatus)
		return
	}

	err = t.runPlaybook()
	if err != nil {
		t.log("Running playbook failed: " + err.Error())
		t.fail()
		return
	}

	t.setStatus(taskSuccessStatus)

	templates, err := t.store.GetTemplates(t.task.ProjectID, db.TemplateFilter{
		BuildTemplateID: &t.task.TemplateID,
		AutorunOnly:     true,
	}, db.RetrieveQueryParams{})
	if err != nil {
		t.log("Running playbook failed: " + err.Error())
		return
	}

	for _, tpl := range templates {
		_, err = AddTaskToPool(t.store, db.Task{
			TemplateID:  tpl.ID,
			ProjectID:   tpl.ProjectID,
			BuildTaskID: &t.task.ID,
		}, nil, tpl.ProjectID)
		if err != nil {
			t.log("Running playbook failed: " + err.Error())
		}
	}
}

func (t *task) prepareError(err error, errMsg string) error {
	if err == db.ErrNotFound {
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

	// get inventory
	t.inventory, err = t.store.GetInventory(t.template.ProjectID, t.template.InventoryID)
	if err != nil {
		return t.prepareError(err, "Template Inventory not found!")
	}

	// get repository
	t.repository, err = t.store.GetRepository(t.template.ProjectID, t.template.RepositoryID)

	if err != nil {
		return err
	}

	// get environment
	if t.template.EnvironmentID != nil {
		t.environment, err = t.store.GetEnvironment(t.template.ProjectID, *t.template.EnvironmentID)
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

func (t *task) installVaultKeyFile() error {
	if t.template.VaultKeyID == nil {
		return nil
	}

	return t.template.VaultKey.Install(db.AccessKeyUsageVault)
}

func (t *task) checkoutRepository() error {
	if t.task.CommitHash != nil { // checkout to commit if it is provided for task
		err := t.validateRepo()
		if err != nil {
			return err
		}

		cmd := exec.Command("git")
		cmd.Dir = t.getRepoPath()
		t.log("Checkout repository to commit " + *t.task.CommitHash)
		cmd.Args = append(cmd.Args, "checkout", *t.task.CommitHash)
		t.logCmd(cmd)
		return cmd.Run()
	}

	// store commit to task table

	commitHash, err := t.getCommitHash()
	if err != nil {
		return err
	}
	commitMessage, _ := t.getCommitMessage()
	t.task.CommitHash = &commitHash
	t.task.CommitMessage = commitMessage

	return t.store.UpdateTask(t.task)
}

// getCommitHash retrieves current commit hash from task repository
func (t *task) getCommitHash() (res string, err error) {
	err = t.validateRepo()
	if err != nil {
		return
	}

	cmd := exec.Command("git")
	cmd.Dir = t.getRepoPath()
	t.log("Get current commit hash")
	cmd.Args = append(cmd.Args, "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return
	}
	res = strings.Trim(string(out), " \n")
	return
}

// getCommitMessage retrieves current commit message from task repository
func (t *task) getCommitMessage() (res string, err error) {
	err = t.validateRepo()
	if err != nil {
		return
	}

	cmd := exec.Command("git")
	cmd.Dir = t.getRepoPath()
	t.log("Get current commit message")
	cmd.Args = append(cmd.Args, "show-branch", "--no-name", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return
	}
	res = strings.Trim(string(out), " \n")

	if len(res) > 100 {
		res = res[0:100]
	}

	return
}

func (t *task) makeGitCommand() *exec.Cmd {
	var gitSSHCommand string
	if t.repository.SSHKey.Type == db.AccessKeySSH {
		gitSSHCommand = t.repository.SSHKey.GetSshCommand()
	}

	cmd := exec.Command("git") //nolint: gas
	cmd.Dir = util.Config.TmpPath
	t.setCmdEnvironment(cmd, gitSSHCommand)

	return cmd
}

func (t *task) canRepositoryBePulled() bool {
	fetchCmd := t.makeGitCommand()
	fetchCmd.Args = append(fetchCmd.Args, "fetch")
	t.logCmd(fetchCmd)
	err := fetchCmd.Run()
	if err != nil {
		return false
	}

	checkCmd := t.makeGitCommand()
	checkCmd.Args = append(checkCmd.Args, "merge-base", "--is-ancestor", "HEAD", "origin/"+t.repository.GitBranch)
	t.logCmd(checkCmd)
	err = checkCmd.Run()
	return err != nil
}

func (t *task) cloneRepository() error {
	cmd := t.makeGitCommand()
	t.log("Cloning repository " + t.repository.GitURL)
	cmd.Args = append(cmd.Args, "clone", "--recursive", "--branch", t.repository.GitURL, t.repository.GitBranch, t.getRepoName())
	t.logCmd(cmd)
	return cmd.Run()
}

func (t *task) pullRepository() error {
	cmd := t.makeGitCommand()
	t.log("Updating repository " + t.repository.GitURL)
	cmd.Dir = t.getRepoPath()
	cmd.Args = append(cmd.Args, "pull", "origin", t.repository.GitBranch)
	t.logCmd(cmd)
	return cmd.Run()
}

func (t *task) updateRepository() error {
	err := t.validateRepo()

	if err != nil {
		if !os.IsNotExist(err) {
			err = os.RemoveAll(t.getRepoPath())
			if err != nil {
				return err
			}
		}
		return t.cloneRepository()
	}

	if t.canRepositoryBePulled() {
		err = t.pullRepository()
		if err == nil {
			return nil
		}
	}

	err = os.RemoveAll(t.getRepoPath())
	if err != nil {
		return err
	}

	return t.cloneRepository()
}

func (t *task) installRequirements() error {
	requirementsFilePath := fmt.Sprintf("%s/roles/requirements.yml", t.getRepoPath())
	requirementsHashFilePath := fmt.Sprintf("%s/requirements.md5", t.getRepoPath())

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
	cmd.Dir = t.getRepoPath()

	t.setCmdEnvironment(cmd, t.repository.SSHKey.GetSshCommand())

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

func (t *task) runPlaybook() (err error) {
	args, err := t.getPlaybookArgs()
	if err != nil {
		return
	}
	cmd := exec.Command("ansible-playbook", args...) //nolint: gas
	cmd.Dir = t.getRepoPath()
	t.setCmdEnvironment(cmd, "")

	t.logCmd(cmd)
	cmd.Stdin = strings.NewReader("")
	err = cmd.Start()
	if err != nil {
		return
	}
	t.process = cmd.Process
	err = cmd.Wait()
	return
}

func (t *task) getExtraVars() (str string, err error) {
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
		user, err = t.store.GetUser(*t.task.UserID)
		if err == nil {
			taskDetails["username"] = user.Username
		}
	}

	if t.template.Type != db.TemplateTask {
		taskDetails["type"] = t.template.Type
		incomingVersion := t.task.GetIncomingVersion(t.store)
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
func (t *task) getPlaybookArgs() (args []string, err error) {
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
		t.log(err.Error())
		t.log("Could not remove command environment, if existant it will be passed to --extra-vars. This is not fatal but be aware of side effects")
	} else if extraVars != "" {
		args = append(args, "--extra-vars", extraVars)
	}

	var templateExtraArgs []string
	if t.template.Arguments != nil {
		err = json.Unmarshal([]byte(*t.template.Arguments), &templateExtraArgs)
		if err != nil {
			t.log("Could not unmarshal arguments to []string")
			return
		}
	}

	if t.template.OverrideArguments {
		args = templateExtraArgs
	} else {
		args = append(args, templateExtraArgs...)
		args = append(args, playbookName)
	}

	return
}

func (t *task) setCmdEnvironment(cmd *exec.Cmd, gitSSHCommand string) {
	env := os.Environ()
	env = append(env, fmt.Sprintf("HOME=%s", util.Config.TmpPath))
	env = append(env, fmt.Sprintf("PWD=%s", cmd.Dir))
	env = append(env, fmt.Sprintln("PYTHONUNBUFFERED=1"))
	env = append(env, extractCommandEnvironment(t.environment.JSON)...)

	//if util.Config.VariablesPassingMethod == util.VariablesPassingBoth ||
	//	util.Config.VariablesPassingMethod == util.VariablesPassingEnv {
	//
	//	if t.task.Message != "" {
	//		env = append(env, "SEMAPHORE_TASK_MESSAGE="+t.task.Message)
	//	}
	//
	//	if t.task.UserID != nil {
	//		user, err := t.store.GetUser(*t.task.UserID)
	//		if err != nil {
	//			panic("Deploy task can't find user")
	//		}
	//		env = append(env, "SEMAPHORE_TASK_USERNAME="+user.Username)
	//	}
	//
	//	if t.template.Type != db.TemplateTask {
	//		env = append(env, "SEMAPHORE_TASK_TYPE="+string(t.template.Type))
	//		incomingVersion, err := t.task.GetIncomingVersion(t.store)
	//		if err != nil {
	//			panic("Deploy task has no build task")
	//		}
	//		env = append(env, "SEMAPHORE_INCOMING_VERSION="+incomingVersion)
	//		if t.template.Type == db.TemplateBuild {
	//			env = append(env, "SEMAPHORE_TARGET_VERSION="+*t.task.Migration)
	//		}
	//	}
	//}

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
