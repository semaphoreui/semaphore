package db_lib

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/ssh"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
)

func getMD5Hash(filepath string) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close() //nolint:errcheck

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func hasRequirementsChanges(requirementsFilePath string, requirementsHashFilePath string) bool {
	oldFileMD5HashBytes, err := os.ReadFile(requirementsHashFilePath)
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

	return os.WriteFile(requirementsHashFile, []byte(newFileMD5Hash), 0o644)
}

type AnsibleApp struct {
	Logger     task_logger.Logger
	Playbook   *AnsiblePlaybook
	Template   db.Template
	Repository db.Repository
	Inventory  db.Inventory

	sshKeyInstallation    ssh.AccessKeyInstallation
	becomeKeyInstallation ssh.AccessKeyInstallation

	vaultFileInstallations map[string]ssh.AccessKeyInstallation
}

func (t *AnsibleApp) SetLogger(logger task_logger.Logger) task_logger.Logger {
	t.Logger = logger
	t.Playbook.Logger = logger
	return logger
}

func (t *AnsibleApp) Run(args LocalAppRunningArgs) error {
	return t.Playbook.RunPlaybook(args.CliArgs, args.EnvironmentVars, args.Inputs, args.Callback)
}

func (t *AnsibleApp) Log(msg string) {
	t.Logger.Log(msg)
}

func (t *AnsibleApp) Clear() {
}

func (t *AnsibleApp) InstallRequirements(args LocalAppInstallingArgs) error {
	if err := t.installCollectionsRequirements(); err != nil {
		return err
	}
	if err := t.installRolesRequirements(); err != nil {
		return err
	}
	return nil
}

func (t *AnsibleApp) installVaultKeyFiles(keyInstaller AccessKeyInstaller) (err error) {
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
			install, err = keyInstaller.Install(*vault.Vault, db.AccessKeyRoleAnsiblePasswordVault, t.Logger)
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

func (t *AnsibleApp) getRepoPath() string {
	return t.Repository.GetFullPath(t.Template.ID)
}

func (t *AnsibleApp) installGalaxyRequirementsFile(requirementsType GalaxyRequirementsType, requirementsFilePath string) error {
	requirementsHashFilePath := fmt.Sprintf("%s_%s.md5", requirementsFilePath, requirementsType)

	if _, err := os.Stat(requirementsFilePath); err != nil {
		t.Log("No " + requirementsFilePath + " file found. Skip galaxy install process.\n")
		return nil
	}

	if hasRequirementsChanges(requirementsFilePath, requirementsHashFilePath) {
		if err := t.runGalaxy([]string{
			string(requirementsType),
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
		t.Log(requirementsFilePath + " has no changes. Skip galaxy install process.\n")
	}

	return nil
}

func (t *AnsibleApp) GetPlaybookDir() string {
	playbookPath := path.Join(t.getRepoPath(), t.Template.Playbook)

	return path.Dir(playbookPath)
}

type GalaxyRequirementsType string

const (
	GalaxyRole       GalaxyRequirementsType = "role"
	GalaxyCollection GalaxyRequirementsType = "collection"
)

func (t *AnsibleApp) installRolesRequirements() (err error) {
	// default roles path
	err = t.installGalaxyRequirementsFile(GalaxyRole, path.Join(t.GetPlaybookDir(), "roles", "requirements.yml"))
	if err != nil {
		return
	}
	err = t.installGalaxyRequirementsFile(GalaxyRole, path.Join(t.GetPlaybookDir(), "requirements.yml"))
	if err != nil {
		return
	}

	// alternative roles path
	err = t.installGalaxyRequirementsFile(GalaxyRole, path.Join(t.getRepoPath(), "roles", "requirements.yml"))
	if err != nil {
		return
	}
	err = t.installGalaxyRequirementsFile(GalaxyRole, path.Join(t.getRepoPath(), "requirements.yml"))
	return
}

func (t *AnsibleApp) installCollectionsRequirements() (err error) {
	// default collections path
	err = t.installGalaxyRequirementsFile(GalaxyCollection, path.Join(t.GetPlaybookDir(), "collections", "requirements.yml"))
	if err != nil {
		return
	}
	err = t.installGalaxyRequirementsFile(GalaxyCollection, path.Join(t.GetPlaybookDir(), "requirements.yml"))
	if err != nil {
		return
	}

	// alternative collections path
	err = t.installGalaxyRequirementsFile(GalaxyCollection, path.Join(t.getRepoPath(), "collections", "requirements.yml"))
	if err != nil {
		return
	}
	err = t.installGalaxyRequirementsFile(GalaxyCollection, path.Join(t.getRepoPath(), "requirements.yml"))
	return
}

func (t *AnsibleApp) runGalaxy(args []string) error {
	return t.Playbook.RunGalaxy(args)
}

// nolint: gocyclo
func (t *AnsibleApp) getPlaybookArgs(username string, incomingVersion *string) (args []string, inputs map[string]string, err error) {

	inputMap := make(map[db.AccessKeyRole]string)
	inputs = make(map[string]string)

	playbookName := t.Template.Playbook
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

	for _, secret := range t.Environment.Secrets {
		if secret.Type != db.EnvironmentSecretVar {
			continue
		}
		args = append(args, "--extra-vars", fmt.Sprintf("%s=%s", secret.Name, secret.Secret))
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

func (t *AnsibleApp) installInventory() (err error) {
	if t.Inventory.SSHKeyID != nil {
		t.sshKeyInstallation, err = t.KeyInstaller.Install(t.Inventory.SSHKey, db.AccessKeyRoleAnsibleUser, t.Logger)
		if err != nil {
			return
		}
	}

	if t.Inventory.BecomeKeyID != nil {
		t.becomeKeyInstallation, err = t.KeyInstaller.Install(t.Inventory.BecomeKey, db.AccessKeyRoleAnsibleBecomeUser, t.Logger)
		if err != nil {
			return
		}
	}

	switch t.Inventory.Type {
	case db.InventoryFile:
		err = t.cloneInventoryRepo(t.KeyInstaller)
	case db.InventoryStatic, db.InventoryStaticYaml:
		err = t.installStaticInventory()
	}

	return
}
