package tasks

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/semaphoreui/semaphore/pkg/ssh"
	"os"
	"os/exec"
	"strings"

	"path"
	"strconv"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db_lib"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
)

type LocalJob struct {
	Task        db.Task
	Template    db.Template
	Inventory   db.Inventory
	Repository  db.Repository
	Environment db.Environment
	Secret      string             // Secret contains secrets received from Survey variables
	Logger      task_logger.Logger // Logger allows to send logs and status to the server

	App db_lib.LocalApp

	killed  bool // killed means that API request to stop the job has been received
	Process *os.Process

	sshKeyInstallation     ssh.AccessKeyInstallation
	becomeKeyInstallation  ssh.AccessKeyInstallation
	vaultFileInstallations map[string]ssh.AccessKeyInstallation

	KeyInstaller       db_lib.AccessKeyInstaller
	secretVarFile      string // temporary file for secret variables (Ansible)
	secretVaultPassword string // vault password for encrypting secret files
	secretVaultPasswordFile string // temporary file containing vault password
}

func (t *LocalJob) IsKilled() bool {
	return t.killed
}

func (t *LocalJob) Kill() {
	t.killed = true

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

func (t *LocalJob) SetCommit(hash, message string) {
	t.Logger.SetCommit(hash, message)
}

func (t *LocalJob) getEnvironmentExtraVars(username string, incomingVersion *string) (extraVars map[string]any, err error) {

	extraVars = make(map[string]any)

	if t.Environment.JSON != "" {
		err = json.Unmarshal([]byte(t.Environment.JSON), &extraVars)
		if err != nil {
			return
		}
	}

	taskDetails := make(map[string]any)

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

	vars := make(map[string]any)
	vars["task_details"] = taskDetails
	extraVars["semaphore_vars"] = vars

	return
}

// getEnvironmentExtraVarsJSON returns JSON for public extra vars only, secrets are handled separately
func (t *LocalJob) getEnvironmentExtraVarsJSON(username string, incomingVersion *string) (str string, err error) {
	extraVars := make(map[string]any)

	// Only include public variables from Environment.JSON
	if t.Environment.JSON != "" {
		err = json.Unmarshal([]byte(t.Environment.JSON), &extraVars)
		if err != nil {
			return
		}
	}

	// Do not include secrets from t.Secret - they will be handled in separate temp file
	// Clear the secret to avoid reprocessing
	t.Secret = "{}"

	taskDetails := make(map[string]any)

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

	vars := make(map[string]any)
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

// getSecretEnvironmentVars returns environment variables for secrets, avoiding command line exposure
func (t *LocalJob) getSecretEnvironmentVars(prefix string) (secretEnvVars []string, err error) {
	// Add environment secrets
	for _, secret := range t.Environment.Secrets {
		if secret.Type == db.EnvironmentSecretVar {
			secretEnvVars = append(secretEnvVars, fmt.Sprintf("%s%s=%s", prefix, secret.Name, secret.Secret))
		}
	}

	// Add survey secrets from t.Secret
	if t.Secret != "" {
		var extraSecretVars map[string]any
		err = json.Unmarshal([]byte(t.Secret), &extraSecretVars)
		if err != nil {
			return
		}
		for name, value := range extraSecretVars {
			if strValue, ok := value.(string); ok {
				secretEnvVars = append(secretEnvVars, fmt.Sprintf("%s%s=%s", prefix, name, strValue))
			}
		}
	}

	return
}

// generateVaultPassword creates a secure random vault password for encrypting secret files
func (t *LocalJob) generateVaultPassword() (string, error) {
	// Generate 32 bytes of random data and encode as hex
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random vault password: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// createVaultPasswordFile creates a temporary file containing the vault password
func (t *LocalJob) createVaultPasswordFile() (string, error) {
	if t.secretVaultPassword == "" {
		var err error
		t.secretVaultPassword, err = t.generateVaultPassword()
		if err != nil {
			return "", err
		}
	}

	tmpDir := util.Config.GetProjectTmpDir(t.Template.ProjectID)
	vaultPasswordFile := path.Join(tmpDir, fmt.Sprintf("vault_pass_%d", t.Task.ID))

	err := os.WriteFile(vaultPasswordFile, []byte(t.secretVaultPassword), 0600)
	if err != nil {
		return "", fmt.Errorf("failed to create vault password file: %w", err)
	}

	t.secretVaultPasswordFile = vaultPasswordFile
	return vaultPasswordFile, nil
}

// createSecretExtraVarsFile creates a temporary file with secret variables for Ansible, encrypted with ansible-vault
func (t *LocalJob) createSecretExtraVarsFile() (tempFile string, err error) {
	secretVars := make(map[string]any)

	// Add environment secrets
	for _, secret := range t.Environment.Secrets {
		if secret.Type == db.EnvironmentSecretVar {
			secretVars[secret.Name] = secret.Secret
		}
	}

	// Add survey secrets from t.Secret (before it's cleared)
	if t.Secret != "" {
		var extraSecretVars map[string]any
		err = json.Unmarshal([]byte(t.Secret), &extraSecretVars)
		if err != nil {
			return
		}
		for name, value := range extraSecretVars {
			secretVars[name] = value
		}
	}

	// If no secrets, don't create file
	if len(secretVars) == 0 {
		return "", nil
	}

	// Create temporary file
	tmpDir := util.Config.GetProjectTmpDir(t.Template.ProjectID)
	tempFile = path.Join(tmpDir, fmt.Sprintf("secret_vars_%d.json", t.Task.ID))

	jsonData, err := json.Marshal(secretVars)
	if err != nil {
		return
	}

	// Write unencrypted data first
	err = os.WriteFile(tempFile, jsonData, 0600)
	if err != nil {
		return
	}

	// Create vault password file for encryption
	vaultPasswordFile, err := t.createVaultPasswordFile()
	if err != nil {
		// Clean up the secret file if vault password creation fails
		os.Remove(tempFile)
		return "", fmt.Errorf("failed to create vault password file: %w", err)
	}

	// Encrypt the file using ansible-vault
	cmd := exec.Command("ansible-vault", "encrypt", "--vault-password-file", vaultPasswordFile, tempFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Clean up files if encryption fails
		os.Remove(tempFile)
		os.Remove(vaultPasswordFile)
		return "", fmt.Errorf("failed to encrypt secret vars file with ansible-vault: %w, output: %s", err, string(output))
	}

	// Store for cleanup
	t.secretVarFile = tempFile

	return
}

// cleanupSecretFile removes the temporary secret file and vault password file if they exist
func (t *LocalJob) cleanupSecretFile() {
	if t.secretVarFile != "" {
		if err := os.Remove(t.secretVarFile); err != nil {
			t.Log("Warning: Could not remove secret vars file: " + err.Error())
		}
		t.secretVarFile = ""
	}
	if t.secretVaultPasswordFile != "" {
		if err := os.Remove(t.secretVaultPasswordFile); err != nil {
			t.Log("Warning: Could not remove vault password file: " + err.Error())
		}
		t.secretVaultPasswordFile = ""
	}
}

// nolint: gocyclo
func (t *LocalJob) getShellArgs(username string, incomingVersion *string) (args []string, err error) {
	extraVars, err := t.getEnvironmentExtraVars(username, incomingVersion)

	if err != nil {
		t.Log(err.Error())
		t.Log("Error getting environment extra vars")
		return
	}

	templateArgs, taskArgs, err := t.getCLIArgs()
	if err != nil {
		t.Log(err.Error())
		return
	}

	// Script to run
	args = append(args, t.Template.Playbook)

	// Include extra args from template
	args = append(args, templateArgs...)

	// Include ExtraVars and Survey Vars
	for name, value := range extraVars {
		if name != "semaphore_vars" {
			args = append(args, fmt.Sprintf("%s=%s", name, value))
		}
	}

	// Include extra args from task
	args = append(args, taskArgs...)

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

	var params db.TerraformTaskParams
	err = t.Task.ExtractParams(&params)
	if err != nil {
		return
	}

	if params.Destroy {
		args = append(args, "-destroy")
	}

	for name, value := range extraVars {
		if name == "semaphore_vars" {
			continue
		}
		args = append(args, "-var", fmt.Sprintf("%s=%s", name, value))
	}

	templateArgs, taskArgs, err := t.getCLIArgs()
	if err != nil {
		t.Log(err.Error())
		return
	}

	args = append(args, templateArgs...)
	args = append(args, taskArgs...)

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

	var tplParams db.AnsibleTemplateParams

	err = t.Template.FillParams(&tplParams)
	if err != nil {
		return
	}

	var params db.AnsibleTaskParams

	err = t.Task.ExtractParams(&params)
	if err != nil {
		return
	}

	if tplParams.AllowDebug && params.Debug {
		if params.DebugLevel < 1 {
			params.DebugLevel = 4
		}

		if params.DebugLevel > 6 {
			params.DebugLevel = 6
		}

		args = append(args, "-"+strings.Repeat("v", params.DebugLevel))
	}

	if params.Diff {
		args = append(args, "--diff")
	}

	if params.DryRun {
		args = append(args, "--check")
	}

	for name, install := range t.vaultFileInstallations {
		if install.Password != "" {
			args = append(args, fmt.Sprintf("--vault-id=%s@prompt", name))
			inputs[fmt.Sprintf("Vault password (%s):", name)] = install.Password
		}
		if install.Script != "" {
			args = append(args, fmt.Sprintf("--vault-id=%s@%s", name, install.Script))
		}
	}

	extraVars, err := t.getEnvironmentExtraVarsJSON(username, incomingVersion)
	if err != nil {
		t.Log(err.Error())
		t.Log("Could not remove command environment, if existent it will be passed to --extra-vars. This is not fatal but be aware of side effects")
	} else if extraVars != "" {
		args = append(args, "--extra-vars", extraVars)
	}

	// Create temporary file for secrets to avoid exposing them in command line
	secretFile, err := t.createSecretExtraVarsFile()
	if err != nil {
		t.Log("Warning: Could not create secret vars file: " + err.Error())
	} else if secretFile != "" {
		// Add vault-id for decrypting the secret vars file
		if t.secretVaultPasswordFile != "" {
			args = append(args, fmt.Sprintf("--vault-id=secrets@%s", t.secretVaultPasswordFile))
		}
		args = append(args, "--extra-vars", "@"+secretFile)
	}

	templateArgs, taskArgs, err := t.getCLIArgs()
	if err != nil {
		t.Log(err.Error())
		return
	}

	var limit string
	var tags string
	var skipTags string

	// Fill fields from template
	if len(tplParams.Limit) > 0 {
		limit = strings.Join(tplParams.Limit, ",")
	}

	if len(tplParams.Tags) > 0 {
		tags = strings.Join(tplParams.Tags, ",")
	}

	if len(tplParams.SkipTags) > 0 {
		skipTags = strings.Join(tplParams.SkipTags, ",")
	}

	// Fill fields from task

	if tplParams.AllowOverrideLimit && params.Limit != nil {
		limit = strings.Join(params.Limit, ",")
	}

	if tplParams.AllowOverrideTags && params.Tags != nil {
		tags = strings.Join(params.Tags, ",")
	}

	if tplParams.AllowOverrideSkipTags && params.SkipTags != nil {
		skipTags = strings.Join(params.SkipTags, ",")
	}

	// Add final args

	if limit != "" {
		templateArgs = append(templateArgs, "--limit="+limit)
	}

	if tags != "" {
		templateArgs = append(templateArgs, "--tags="+tags)
	}

	if skipTags != "" {
		templateArgs = append(templateArgs, "--skip-tags="+skipTags)
	}

	args = append(args, templateArgs...)
	args = append(args, taskArgs...)
	args = append(args, playbookName)

	if line, ok := inputMap[db.AccessKeyRoleAnsibleUser]; ok {
		inputs["SSH password:"] = line
	}

	if line, ok := inputMap[db.AccessKeyRoleAnsibleBecomeUser]; ok {
		inputs["BECOME password"] = line
	}

	if line, ok := inputMap[db.AccessKeyRoleAnsibleBecomeUser]; ok {
		inputs["SUDO password"] = line
	}

	return
}

func (t *LocalJob) getCLIArgs() (templateArgs []string, taskArgs []string, err error) {

	if t.Template.Arguments != nil {
		err = json.Unmarshal([]byte(*t.Template.Arguments), &templateArgs)
		if err != nil {
			err = fmt.Errorf("invalid format of the template extra arguments, must be valid JSON")
			return
		}
	}

	if t.Template.AllowOverrideArgsInTask && t.Task.Arguments != nil {
		err = json.Unmarshal([]byte(*t.Task.Arguments), &taskArgs)
		if err != nil {
			err = fmt.Errorf("invalid format of the TaskRunner extra arguments, must be valid JSON")
			return
		}
	}

	return
}

func (t *LocalJob) getTemplateParams() (any, error) {
	var params any
	switch t.Template.App {
	case db.AppAnsible:
		params = &db.AnsibleTemplateParams{}
	case db.AppTerraform, db.AppTofu, db.AppTerragrunt:
		params = &db.TerraformTemplateParams{}
	default:
		return nil, nil
	}

	err := t.Template.FillParams(params)
	return params, err
}

func (t *LocalJob) getParams() (params any, err error) {
	switch t.Template.App {
	case db.AppAnsible:
		params = &db.AnsibleTaskParams{}
	case db.AppTerraform, db.AppTofu, db.AppTerragrunt:
		params = &db.TerraformTaskParams{}
	default:
		params = &db.DefaultTaskParams{}
	}

	err = t.Task.ExtractParams(params)

	if err != nil {
		return
	}

	return
}

func (t *LocalJob) Run(username string, incomingVersion *string, alias string) (err error) {

	defer func() {
		t.destroyKeys()
		t.destroyInventoryFile()
		t.cleanupSecretFile()
		t.App.Clear()
	}()

	t.SetStatus(task_logger.TaskRunningStatus) // It is required for local mode. Don't delete

	environmentVariables, err := t.getEnvironmentENV()
	if err != nil {
		return
	}

	tplParams, err := t.getTemplateParams()
	if err != nil {
		return
	}

	params, err := t.getParams()
	if err != nil {
		return
	}

	if t.Template.App.IsTerraform() && alias != "" {
		environmentVariables = append(environmentVariables, "TF_HTTP_ADDRESS="+util.GetPublicAliasURL("terraform", alias))
	}

	err = t.prepareRun(db_lib.LocalAppInstallingArgs{
		EnvironmentVars: environmentVariables,
		TplParams:       tplParams,
		Params:          params,
		Installer:       t.KeyInstaller,
	})

	if err != nil {
		return err
	}

	var args []string
	var inputs map[string]string

	switch t.Template.App {
	case db.AppAnsible:
		args, inputs, err = t.getPlaybookArgs(username, incomingVersion)
	case db.AppTerraform, db.AppTofu, db.AppTerragrunt:
		args, err = t.getTerraformArgs(username, incomingVersion)
	default:
		args, err = t.getShellArgs(username, incomingVersion)
	}

	if err != nil {
		return
	}

	// Add secret environment variables to avoid exposing them in command line arguments
	switch t.Template.App {
	case db.AppTerraform, db.AppTofu, db.AppTerragrunt:
		// For Terraform, use TF_VAR_ prefix for variables
		secretEnvVars, secretErr := t.getSecretEnvironmentVars("TF_VAR_")
		if secretErr != nil {
			t.Log("Warning: Failed to process secret environment variables: " + secretErr.Error())
		} else {
			environmentVariables = append(environmentVariables, secretEnvVars...)
		}
	default:
		// For Ansible and Shell, add secrets as regular environment variables  
		secretEnvVars, secretErr := t.getSecretEnvironmentVars("")
		if secretErr != nil {
			t.Log("Warning: Failed to process secret environment variables: " + secretErr.Error())
		} else {
			environmentVariables = append(environmentVariables, secretEnvVars...)
		}
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

	if t.killed {
		t.SetStatus(task_logger.TaskStoppedStatus)
		return nil
	}

	return t.App.Run(db_lib.LocalAppRunningArgs{
		CliArgs:         args,
		EnvironmentVars: environmentVariables,
		Inputs:          inputs,
		TaskParams:      params,
		TemplateParams:  tplParams,
		Callback: func(p *os.Process) {
			t.Process = p
		},
	})

}

func (t *LocalJob) prepareRun(installingArgs db_lib.LocalAppInstallingArgs) error {

	t.Log("Preparing: " + strconv.Itoa(t.Task.ID))

	if err := checkTmpDir(util.Config.GetProjectTmpDir(t.Template.ProjectID)); err != nil {
		t.Log("Creating tmp dir failed: " + err.Error())
		return err
	}

	// Override git branch from template if set
	if t.Template.GitBranch != nil && *t.Template.GitBranch != "" {
		t.Repository.GitBranch = *t.Template.GitBranch
	}

	// Override git branch from task if set
	if t.Task.GitBranch != nil && *t.Task.GitBranch != "" {
		t.Repository.GitBranch = *t.Task.GitBranch
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

	if err := t.App.InstallRequirements(installingArgs); err != nil {
		t.Log("Failed to install requirements: " + err.Error())
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
		Client:     db_lib.CreateDefaultGitClient(t.KeyInstaller),
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
		Client:     db_lib.CreateDefaultGitClient(t.KeyInstaller),
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

	commitHash, err := repo.GetLastCommitHash()

	if err != nil {
		return err
	}

	commitMessage, err := repo.GetLastCommitMessage()

	if err != nil {
		t.Log(err.Error())
	}

	t.SetCommit(commitHash, commitMessage)

	return nil
}

func (t *LocalJob) installVaultKeyFiles() (err error) {
	t.vaultFileInstallations = make(map[string]ssh.AccessKeyInstallation)

	if len(t.Template.Vaults) == 0 {
		return nil
	}

	for _, vault := range t.Template.Vaults {
		var name string
		if vault.Name != nil {
			name = *vault.Name
		} else {
			name = "default"
		}

		var install ssh.AccessKeyInstallation
		if vault.Type == db.TemplateVaultPassword {
			install, err = t.KeyInstaller.Install(*vault.Vault, db.AccessKeyRoleAnsiblePasswordVault, t.Logger)
			if err != nil {
				return
			}
		}
		if vault.Type == db.TemplateVaultScript && vault.Script != nil {
			install.Script = *vault.Script
		}

		t.vaultFileInstallations[name] = install
	}

	return
}
