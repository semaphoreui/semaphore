package db_lib

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/semaphoreui/semaphore/db"
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
}

func (t *AnsibleApp) SetLogger(logger task_logger.Logger) task_logger.Logger {
	t.Logger = logger
	t.Playbook.Logger = logger
	return logger
}

func (t *AnsibleApp) Run(args LocalAppRunningArgs) error {
	// Use "default" key for backward compatibility
	cliArgs := args.CliArgs["default"]
	return t.Playbook.RunPlaybook(cliArgs, args.EnvironmentVars, args.Inputs, args.Callback)
}

func (t *AnsibleApp) Log(msg string) {
	t.Logger.Log(msg)
}

func (t *AnsibleApp) Clear() {
}

func (t *AnsibleApp) InstallRequirements(args LocalAppInstallingArgs) error {
	if err := t.installCollectionsRequirements(args.EnvironmentVars); err != nil {
		return err
	}
	if err := t.installRolesRequirements(args.EnvironmentVars); err != nil {
		return err
	}
	return nil
}

func (t *AnsibleApp) getRepoPath() string {
	return t.Repository.GetFullPath(t.Template.ID)
}

func (t *AnsibleApp) installGalaxyRequirementsFile(requirementsType GalaxyRequirementsType, requirementsFilePath string, environmentVars []string) error {
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
		}, environmentVars); err != nil {
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

func (t *AnsibleApp) installRolesRequirements(environmentVars []string) (err error) {
	// default roles path
	err = t.installGalaxyRequirementsFile(GalaxyRole, path.Join(t.GetPlaybookDir(), "roles", "requirements.yml"), environmentVars)
	if err != nil {
		return
	}
	err = t.installGalaxyRequirementsFile(GalaxyRole, path.Join(t.GetPlaybookDir(), "requirements.yml"), environmentVars)
	if err != nil {
		return
	}

	// alternative roles path
	err = t.installGalaxyRequirementsFile(GalaxyRole, path.Join(t.getRepoPath(), "roles", "requirements.yml"), environmentVars)
	if err != nil {
		return
	}
	err = t.installGalaxyRequirementsFile(GalaxyRole, path.Join(t.getRepoPath(), "requirements.yml"), environmentVars)
	return
}

func (t *AnsibleApp) installCollectionsRequirements(environmentVars []string) (err error) {
	// default collections path
	err = t.installGalaxyRequirementsFile(GalaxyCollection, path.Join(t.GetPlaybookDir(), "collections", "requirements.yml"), environmentVars)
	if err != nil {
		return
	}
	err = t.installGalaxyRequirementsFile(GalaxyCollection, path.Join(t.GetPlaybookDir(), "requirements.yml"), environmentVars)
	if err != nil {
		return
	}

	// alternative collections path
	err = t.installGalaxyRequirementsFile(GalaxyCollection, path.Join(t.getRepoPath(), "collections", "requirements.yml"), environmentVars)
	if err != nil {
		return
	}
	err = t.installGalaxyRequirementsFile(GalaxyCollection, path.Join(t.getRepoPath(), "requirements.yml"), environmentVars)
	return
}

func (t *AnsibleApp) runGalaxy(args []string, environmentVars []string) error {
	return t.Playbook.RunGalaxy(args, environmentVars)
}
