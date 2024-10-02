package tasks

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"

	"path"
	"strconv"

	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/db_lib"
	"github.com/ansible-semaphore/semaphore/pkg/task_logger"
	"github.com/ansible-semaphore/semaphore/util"
)

type LocalJob struct {
	// Received constant fields
	Task        db.Task
	Template    db.Template
	Inventory   db.Inventory
	Repository  db.Repository
	Environment db.Environment
	Secret      string
	Logger      task_logger.Logger

	App db_lib.LocalApp

	// Internal field
	Process *os.Process

	sshKeyInstallation    db.AccessKeyInstallation
	becomeKeyInstallation db.AccessKeyInstallation

	vaultFileInstallations map[string]db.AccessKeyInstallation
}

func (t *LocalJob) Kill() {
	if t.Process == nil {
		return
	}
	err := t.Process.Kill()
	if err != nil {
		t.Log(err.Error())
	}
}

func (t *LocalJob) Log(msg string) {
	t.Logger.Log(msg)
}

func (t *LocalJob) SetStatus(status task_logger.TaskStatus) {
	t.Logger.SetStatus(status)
}

func (t *LocalJob) getEnvironmentExtraVars(username string, incomingVersion *string) (extraVars map[string]interface{}, err error) {

	extraVars = make(map[string]interface{})

	if t.Environment.JSON != "" {
		err = json.Unmarshal([]byte(t.Environment.JSON), &extraVars)
		if err != nil {
			return
		}
	}

	taskDetails := make(map[string]interface{})

	taskDetails["id"] = t.Task.ID

	if t.Task.Message != "" {
		taskDetails["message"] = t.Task.Message
	}

	taskDetails["username"] = username
	taskDetails["url"] = t.Task.GetUrl()

	if t.Template.Type != db.TemplateTask {
		taskDetails["type"] = t.Template.Type
		if incomingVersion != nil {
			taskDetails["incoming_version"] = incomingVersion
		}
		if t.Template.Type == db.TemplateBuild {
			taskDetails["target_version"] = t.Task.Version
		}
	}

	vars := make(map[string]interface{})
	vars["task_details"] = taskDetails
	extraVars["semaphore_vars"] = vars

	return
}

func (t *LocalJob) getEnvironmentExtraVarsJSON(username string, incomingVersion *string) (str string, err error) {
	extraVars := make(map[string]interface{})
	extraSecretVars := make(map[string]interface{})

	if t.Environment.JSON != "" {
		err = json.Unmarshal([]byte(t.Environment.JSON), &extraVars)
		if err != nil {
			return
		}
	}
	if t.Secret != "" {
		err = json.Unmarshal([]byte(t.Secret), &extraSecretVars)
		if err != nil {
			return
		}
	}
	t.Secret = "{}"

	maps.Copy(extraVars, extraSecretVars)

	taskDetails := make(map[string]interface{})

	taskDetails["id"] = t.Task.ID

	if t.Task.Message != "" {
		taskDetails["message"] = t.Task.Message
	}

	taskDetails["username"] = username
	taskDetails["url"] = t.Task.GetUrl()

	if t.Template.Type != db.TemplateTask {
		taskDetails["type"] = t.Template.Type
		if incomingVersion != nil {
			taskDetails["incoming_version"] = incomingVersion
		}
		if t.Template.Type == db.TemplateBuild {
			taskDetails["target_version"] = t.Task.Version
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

func (t *LocalJob) getEnvironmentENV() (res []string, err error) {
	environmentVars := make(map[string]string)

	if t.Environment.ENV != nil {
		err = json.Unmarshal([]byte(*t.Environment.ENV), &environmentVars)
		if err != nil {
			return
		}
	}

	for key, val := range environmentVars {
		res = append(res, fmt.Sprintf("%s=%s", key, val))
	}

	for _, secret := range t.Environment.Secrets {
		if secret.Type != db.EnvironmentSecretEnv {
			continue
		}
		res = append(res, fmt.Sprintf("%s=%s", secret.Name, secret.Secret))
	}

	return
}

// nolint: gocyclo
func (t *LocalJob) getShellArgs(username string, incomingVersion *string) (args []string, err error) {
	extraVars, err := t.getEnvironmentExtraVars(username, incomingVersion)

	if err != nil {
		t.Log(err.Error())
		t.Log("Error getting environment extra vars")
		return
	}

	var templateExtraArgs []string
	if t.Template.Arguments != nil {
		err = json.Unmarshal([]byte(*t.Template.Arguments), &templateExtraArgs)
		if err != nil {
			t.Log("Invalid format of the template extra arguments, must be valid JSON")
			return
		}
	}

	var taskExtraArgs []string
	if t.Template.AllowOverrideArgsInTask && t.Task.Arguments != nil {
		err = json.Unmarshal([]byte(*t.Task.Arguments), &taskExtraArgs)
		if err != nil {
			t.Log("Invalid format of the TaskRunner extra arguments, must be valid JSON")
			return
		}
	}

	// Script to run
	args = append(args, t.Template.Playbook)

	// Include Environment Secret Vars
	for _, secret := range t.Environment.Secrets {
		if secret.Type == db.EnvironmentSecretVar {
			args = append(args, fmt.Sprintf("%s=%s", secret.Name, secret.Secret))
		}
	}

	// Include extra args from template
	args = append(args, templateExtraArgs...)

	// Include ExtraVars and Survey Vars
	for name, value := range extraVars {
		if name != "semaphore_vars" {
			args = append(args, fmt.Sprintf("%s=%s", name, value))
		}
	}

	// Include extra args from task
	args = append(args, taskExtraArgs...)

	return
}

// nolint: gocyclo
func (t *LocalJob) getTerraformArgs(username string, incomingVersion *string) (args []string, err error) {

	args = []string{}

	extraVars, err := t.getEnvironmentExtraVars(username, incomingVersion)

	if err != nil {
		t.Log(err.Error())
		t.Log("Could not remove command environment, if existent it will be passed to --extra-vars. This is not fatal but be aware of side effects")
		return
	}

	for name, value := range extraVars {
		if name == "semaphore_vars" {
			continue
		}
		args = append(args, "-var", fmt.Sprintf("%s=%s", name, value))
	}

	for _, secret := range t.Environment.Secrets {
		if secret.Type != db.EnvironmentSecretVar {
			continue
		}
		args = append(args, "-var", fmt.Sprintf("%s=%s", secret.Name, secret.Secret))
	}

	return
}

// nolint: gocyclo
func (t *LocalJob) getPlaybookArgs(username string, incomingVersion *string) (args []string, inputs map[string]string, err error) {

	inputMap := make(map[db.AccessKeyRole]string)
	inputs = make(map[string]string)

	playbookName := t.Task.Playbook
	if playbookName == "" {
		playbookName = t.Template.Playbook
	}

	var inventoryFilename string
	switch t.Inventory.Type {
	case db.InventoryFile:
		if t.Inventory.RepositoryID == nil {
			inventoryFilename = t.Inventory.GetFilename()
		} else {
			inventoryFilename = path.Join(t.tmpInventoryFullPath(), t.Inventory.GetFilename())
		}
	case db.InventoryStatic, db.InventoryStaticYaml:
		inventoryFilename = t.tmpInventoryFullPath()
	default:
		err = fmt.Errorf("invalid inventory type")
		return
	}

	args = []string{
		"-i", inventoryFilename,
	}

	if t.Inventory.SSHKeyID != nil {
		switch t.Inventory.SSHKey.Type {
		case db.AccessKeySSH:
			if t.sshKeyInstallation.Login != "" {
				args = append(args, "--user", t.sshKeyInstallation.Login)
			}
		case db.AccessKeyLoginPassword:
			if t.sshKeyInstallation.Login != "" {
				args = append(args, "--user", t.sshKeyInstallation.Login)
			}
			if t.sshKeyInstallation.Password != "" {
				args = append(args, "--ask-pass")
				inputMap[db.AccessKeyRoleAnsibleUser] = t.sshKeyInstallation.Password
			}
		case db.AccessKeyNone:
		default:
			err = fmt.Errorf("access key does not suite for inventory's user credentials")
			return
		}
	}

	if t.Inventory.BecomeKeyID != nil {
		switch t.Inventory.BecomeKey.Type {
		case db.AccessKeyLoginPassword:
			if t.becomeKeyInstallation.Login != "" {
				args = append(args, "--become-user", t.becomeKeyInstallation.Login)
			}
			if t.becomeKeyInstallation.Password != "" {
				args = append(args, "--ask-become-pass")
				inputMap[db.AccessKeyRoleAnsibleBecomeUser] = t.becomeKeyInstallation.Password
			}
		case db.AccessKeyNone:
		default:
			err = fmt.Errorf("access key does not suite for inventory's sudo user credentials")
			return
		}
	}

	if t.Task.Debug {
		args = append(args, "-vvvv")
	}

	if t.Task.Diff {
		args = append(args, "--diff")
	}

	if t.Task.DryRun {
		args = append(args, "--check")
	}

	for name, install := range t.vaultFileInstallations {
		if install.Password != "" {
			args = append(args, fmt.Sprintf("--vault-id=%s@prompt", name))
			inputs[fmt.Sprintf("Vault password (%s):", name)] = install.Password
		}
	}

	extraVars, err := t.getEnvironmentExtraVarsJSON(username, incomingVersion)
	if err != nil {
		t.Log(err.Error())
		t.Log("Could not remove command environment, if existent it will be passed to --extra-vars. This is not fatal but be aware of side effects")
	} else if extraVars != "" {
		args = append(args, "--extra-vars", extraVars)
	}

	for _, secret := range t.Environment.Secrets {
		if secret.Type != db.EnvironmentSecretVar {
			continue
		}
		args = append(args, "--extra-vars", fmt.Sprintf("%s=%s", secret.Name, secret.Secret))
	}

	var templateExtraArgs []string
	if t.Template.Arguments != nil {
		err = json.Unmarshal([]byte(*t.Template.Arguments), &templateExtraArgs)
		if err != nil {
			t.Log("Invalid format of the template extra arguments, must be valid JSON")
			return
		}
	}

	var taskExtraArgs []string
	if t.Template.AllowOverrideArgsInTask && t.Task.Arguments != nil {
		err = json.Unmarshal([]byte(*t.Task.Arguments), &taskExtraArgs)
		if err != nil {
			t.Log("Invalid format of the TaskRunner extra arguments, must be valid JSON")
			return
		}
	}

	if t.Task.Limit != "" {
		t.Log("--limit=" + t.Task.Limit)
		taskExtraArgs = append(taskExtraArgs, "--limit="+t.Task.Limit)
	}

	args = append(args, templateExtraArgs...)
	args = append(args, taskExtraArgs...)
	args = append(args, playbookName)

	if line, ok := inputMap[db.AccessKeyRoleAnsibleUser]; ok {
		inputs["SSH password:"] = line
	}

	if line, ok := inputMap[db.AccessKeyRoleAnsibleBecomeUser]; ok {
		inputs["BECOME password"] = line
	}

	return
}

func (t *LocalJob) Run(username string, incomingVersion *string) (err error) {

	t.SetStatus(task_logger.TaskRunningStatus) // It is required for local mode. Don't delete

	err = t.prepareRun()
	if err != nil {
		return err
	}

	defer func() {
		t.destroyKeys()
		t.destroyInventoryFile()
	}()

	var args []string
	var inputs map[string]string

	switch t.Template.App {
	case db.AppAnsible:
		args, inputs, err = t.getPlaybookArgs(username, incomingVersion)
	case db.AppTerraform, db.AppTofu:
		args, err = t.getTerraformArgs(username, incomingVersion)
	default:
		args, err = t.getShellArgs(username, incomingVersion)
	}

	if err != nil {
		return
	}

	environmentVariables, err := t.getEnvironmentENV()
	if err != nil {
		return
	}

	if t.Inventory.SSHKey.Type == db.AccessKeySSH && t.Inventory.SSHKeyID != nil {
		environmentVariables = append(environmentVariables, fmt.Sprintf("SSH_AUTH_SOCK=%s", t.sshKeyInstallation.SSHAgent.SocketFile))
	}

	if t.Template.Type != db.TemplateTask {

		environmentVariables = append(environmentVariables, fmt.Sprintf("SEMAPHORE_TASK_TYPE=%s", t.Template.Type))

		if incomingVersion != nil {
			environmentVariables = append(
				environmentVariables,
				fmt.Sprintf("SEMAPHORE_TASK_INCOMING_VERSION=%s", *incomingVersion))
		}

		if t.Template.Type == db.TemplateBuild && t.Task.Version != nil {
			environmentVariables = append(
				environmentVariables,
				fmt.Sprintf("SEMAPHORE_TASK_TARGET_VERSION=%s", *t.Task.Version))
		}
	}

	return t.App.Run(args, &environmentVariables, inputs, func(p *os.Process) {
		t.Process = p
	})

}

func (t *LocalJob) prepareRun() error {
	t.Log("Preparing: " + strconv.Itoa(t.Task.ID))

	if err := checkTmpDir(util.Config.TmpPath); err != nil {
		t.Log("Creating tmp dir failed: " + err.Error())
		return err
	}

	if t.Repository.GetType() == db.RepositoryLocal {
		if _, err := os.Stat(t.Repository.GitURL); err != nil {
			t.Log("Failed in finding static repository at " + t.Repository.GitURL + ": " + err.Error())
			return err
		}
	} else {
		if err := t.updateRepository(); err != nil {
			t.Log("Failed updating repository: " + err.Error())
			return err
		}
		if err := t.checkoutRepository(); err != nil {
			t.Log("Failed to checkout repository to required commit: " + err.Error())
			return err
		}
	}

	if err := t.installInventory(); err != nil {
		t.Log("Failed to install inventory: " + err.Error())
		return err
	}

	if err := t.App.InstallRequirements(); err != nil {
		t.Log("Running galaxy failed: " + err.Error())
		return err
	}

	if err := t.installVaultKeyFiles(); err != nil {
		t.Log("Failed to install vault password files: " + err.Error())
		return err
	}

	return nil
}

func (t *LocalJob) updateRepository() error {
	repo := db_lib.GitRepository{
		Logger:     t.Logger,
		TemplateID: t.Template.ID,
		Repository: t.Repository,
		Client:     db_lib.CreateDefaultGitClient(),
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

func (t *LocalJob) checkoutRepository() error {

	repo := db_lib.GitRepository{
		Logger:     t.Logger,
		TemplateID: t.Template.ID,
		Repository: t.Repository,
		Client:     db_lib.CreateDefaultGitClient(),
	}

	err := repo.ValidateRepo()

	if err != nil {
		return err
	}

	if t.Task.CommitHash != nil {
		// checkout to commit if it is provided for TaskRunner
		return repo.Checkout(*t.Task.CommitHash)
	}

	// store commit to TaskRunner table

	//commitHash, err := repo.GetLastCommitHash()
	//
	//if err != nil {
	//	return err
	//}
	//
	//commitMessage, _ := repo.GetLastCommitMessage()
	//
	//t.task.CommitHash = &commitHash
	//t.task.CommitMessage = commitMessage
	//
	//return t.pool.store.UpdateTask(t.task)
	return nil
}

func (t *LocalJob) installVaultKeyFiles() (err error) {
	t.vaultFileInstallations = make(map[string]db.AccessKeyInstallation)

	if t.Template.Vaults == nil || len(t.Template.Vaults) == 0 {
		return nil
	}

	for _, vault := range t.Template.Vaults {
		var name string
		if vault.Name != nil {
			name = *vault.Name
		} else {
			name = "default"
		}

		var install db.AccessKeyInstallation
		install, err = vault.Vault.Install(db.AccessKeyRoleAnsiblePasswordVault, t.Logger)
		if err != nil {
			return
		}
		t.vaultFileInstallations[name] = install
	}

	return
}
