package tasks

import (
	"encoding/json"
	"fmt"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/lib"
	"github.com/ansible-semaphore/semaphore/util"
	"os"
	"path"
	"strconv"
)

type LocalJob struct {
	// Received constant fields
	task        db.Task
	template    db.Template
	inventory   db.Inventory
	repository  db.Repository
	environment db.Environment
	playbook    *lib.AnsiblePlaybook
	logger      lib.Logger

	// Internal field
	process *os.Process
}

func (t *LocalJob) Log(msg string) {
	t.logger.Log(msg)
}

func (t *LocalJob) getEnvironmentExtraVars() (str string, err error) {
	extraVars := make(map[string]interface{})

	if t.environment.JSON != "" {
		err = json.Unmarshal([]byte(t.environment.JSON), &extraVars)
		if err != nil {
			return
		}
	}

	taskDetails := make(map[string]interface{})

	taskDetails["id"] = t.task.ID

	if t.task.Message != "" {
		taskDetails["message"] = t.task.Message
	}

	//if t.task.UserID != nil {
	//	var user db.User
	//	user, err = t.pool.store.GetUser(*t.task.UserID)
	//	if err == nil {
	//		taskDetails["username"] = user.Username
	//	}
	//}

	//if t.template.Type != db.TemplateTask {
	//	taskDetails["type"] = t.template.Type
	//	incomingVersion := t.task.GetIncomingVersion(t.pool.store)
	//	if incomingVersion != nil {
	//		taskDetails["incoming_version"] = incomingVersion
	//	}
	//	if t.template.Type == db.TemplateBuild {
	//		taskDetails["target_version"] = t.task.Version
	//	}
	//}

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

func (t *LocalJob) getEnvironmentENV() (arr []string, err error) {
	environmentVars := make(map[string]string)

	if t.environment.ENV != nil {
		err = json.Unmarshal([]byte(*t.environment.ENV), &environmentVars)
		if err != nil {
			return
		}
	}

	for key, val := range environmentVars {
		arr = append(arr, fmt.Sprintf("%s=%s", key, val))
	}

	return
}

// nolint: gocyclo
func (t *LocalJob) getPlaybookArgs() (args []string, err error) {
	playbookName := t.task.Playbook
	if playbookName == "" {
		playbookName = t.template.Playbook
	}

	var inventory string
	switch t.inventory.Type {
	case db.InventoryFile:
		inventory = t.inventory.Inventory
	case db.InventoryStatic, db.InventoryStaticYaml:
		inventory = util.Config.TmpPath + "/inventory_" + strconv.Itoa(t.task.ID)
		if t.inventory.Type == db.InventoryStaticYaml {
			inventory += ".yml"
		}
	default:
		err = fmt.Errorf("invalid invetory type")
		return
	}

	args = []string{
		"-i", inventory,
	}

	if t.inventory.SSHKeyID != nil {
		switch t.inventory.SSHKey.Type {
		case db.AccessKeySSH:
			args = append(args, "--private-key="+t.inventory.SSHKey.GetPath())
			//args = append(args, "--extra-vars={\"ansible_ssh_private_key_file\": \""+t.inventory.SSHKey.GetPath()+"\"}")
			if t.inventory.SSHKey.SshKey.Login != "" {
				args = append(args, "--extra-vars={\"ansible_user\": \""+t.inventory.SSHKey.SshKey.Login+"\"}")
			}
		case db.AccessKeyLoginPassword:
			args = append(args, "--extra-vars=@"+t.inventory.SSHKey.GetPath())
		case db.AccessKeyNone:
		default:
			err = fmt.Errorf("access key does not suite for inventory's user credentials")
			return
		}
	}

	if t.inventory.BecomeKeyID != nil {
		switch t.inventory.BecomeKey.Type {
		case db.AccessKeyLoginPassword:
			args = append(args, "--extra-vars=@"+t.inventory.BecomeKey.GetPath())
		case db.AccessKeyNone:
		default:
			err = fmt.Errorf("access key does not suite for inventory's sudo user credentials")
			return
		}
	}

	if t.task.Debug {
		args = append(args, "-vvvv")
	}

	if t.task.Diff {
		args = append(args, "--diff")
	}

	if t.task.DryRun {
		args = append(args, "--check")
	}

	if t.template.VaultKeyID != nil {
		args = append(args, "--vault-password-file", t.template.VaultKey.GetPath())
	}

	extraVars, err := t.getEnvironmentExtraVars()
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

	if t.task.Limit != "" {
		t.Log("--limit=" + t.task.Limit)
		taskExtraArgs = append(taskExtraArgs, "--limit="+t.task.Limit)
	}

	args = append(args, templateExtraArgs...)
	args = append(args, taskExtraArgs...)
	args = append(args, playbookName)

	return
}

func (t *LocalJob) destroyKeys() {
	err := t.inventory.SSHKey.Destroy()
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

func (t *LocalJob) Run() (err error) {
	err = t.prepareRun()
	if err != nil {
		return err
	}

	defer func() {
		t.destroyKeys()
	}()

	args, err := t.getPlaybookArgs()
	if err != nil {
		return
	}

	environmentVariables, err := t.getEnvironmentENV()
	if err != nil {
		return
	}

	return t.playbook.RunPlaybook(args, &environmentVariables, func(p *os.Process) {
		t.process = p
	})

}

func (t *LocalJob) prepareRun() error {
	defer func() {
		//t.pool.resourceLocker <- &resourceLock{lock: false, holder: t}

		//log.Info("Stopped preparing TaskRunner " + strconv.Itoa(t.task.ID))
		//log.Info("Release resource locker with TaskRunner " + strconv.Itoa(t.task.ID))
		//
		//t.createTaskEvent()

		err := t.repository.SSHKey.Destroy()
		if err != nil {
			t.Log("Can't destroy repository access key, error: " + err.Error())
		}
	}()

	t.Log("Preparing: " + strconv.Itoa(t.task.ID))

	if err := checkTmpDir(util.Config.TmpPath); err != nil {
		t.Log("Creating tmp dir failed: " + err.Error())
		return err
	}

	if t.repository.GetType() == db.RepositoryLocal {
		if _, err := os.Stat(t.repository.GitURL); err != nil {
			t.Log("Failed in finding static repository at " + t.repository.GitURL + ": " + err.Error())
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

	if err := t.installRequirements(); err != nil {
		t.Log("Running galaxy failed: " + err.Error())
		return err
	}

	if err := t.installVaultKeyFile(); err != nil {
		t.Log("Failed to install vault password file: " + err.Error())
		return err
	}

	return nil
}

func (t *LocalJob) updateRepository() error {
	repo := lib.GitRepository{
		Logger:     t.logger,
		TemplateID: t.template.ID,
		Repository: t.repository,
		Client:     lib.CreateDefaultGitClient(),
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

	repo := lib.GitRepository{
		Logger:     t.logger,
		TemplateID: t.template.ID,
		Repository: t.repository,
		Client:     lib.CreateDefaultGitClient(),
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

func (t *LocalJob) installRequirements() error {
	if err := t.installCollectionsRequirements(); err != nil {
		return err
	}
	if err := t.installRolesRequirements(); err != nil {
		return err
	}
	return nil
}

func (t *LocalJob) getRepoPath() string {
	repo := lib.GitRepository{
		Logger:     t.logger,
		TemplateID: t.template.ID,
		Repository: t.repository,
		Client:     lib.CreateDefaultGitClient(),
	}

	return repo.GetFullPath()
}

func (t *LocalJob) installRolesRequirements() error {
	requirementsFilePath := fmt.Sprintf("%s/roles/requirements.yml", t.getRepoPath())
	requirementsHashFilePath := fmt.Sprintf("%s.md5", requirementsFilePath)

	if _, err := os.Stat(requirementsFilePath); err != nil {
		t.Log("No roles/requirements.yml file found. Skip galaxy install process.\n")
		return nil
	}

	if hasRequirementsChanges(requirementsFilePath, requirementsHashFilePath) {
		if err := t.runGalaxy([]string{
			"role",
			"install",
			"-r",
			requirementsFilePath,
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

func (t *LocalJob) getPlaybookDir() string {
	playbookPath := path.Join(t.getRepoPath(), t.template.Playbook)

	return path.Dir(playbookPath)
}

func (t *LocalJob) installCollectionsRequirements() error {
	requirementsFilePath := path.Join(t.getPlaybookDir(), "collections", "requirements.yml")
	requirementsHashFilePath := fmt.Sprintf("%s.md5", requirementsFilePath)

	if _, err := os.Stat(requirementsFilePath); err != nil {
		t.Log("No collections/requirements.yml file found. Skip galaxy install process.\n")
		return nil
	}

	if hasRequirementsChanges(requirementsFilePath, requirementsHashFilePath) {
		if err := t.runGalaxy([]string{
			"collection",
			"install",
			"-r",
			requirementsFilePath,
			"--force",
		}); err != nil {
			return err
		}
		if err := writeMD5Hash(requirementsFilePath, requirementsHashFilePath); err != nil {
			return err
		}
	} else {
		t.Log("collections/requirements.yml has no changes. Skip galaxy install process.\n")
	}

	return nil
}

func (t *LocalJob) runGalaxy(args []string) error {
	return nil
	//return t.job.RunGalaxy(args)
}

func (t *LocalJob) installVaultKeyFile() error {
	if t.template.VaultKeyID == nil {
		return nil
	}

	return t.template.VaultKey.Install(db.AccessKeyRoleAnsiblePasswordVault)
}
