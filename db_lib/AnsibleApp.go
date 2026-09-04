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
	if t.skipGalaxyInstall(args) {
		t.Log("Galaxy install step is skipped.\n")
		return nil
	}

	if err := t.installCollectionsRequirements(args.EnvironmentVars, galaxyExtraArgs(args, GalaxyCollection)); err != nil {
		return err
	}
	if err := t.installRolesRequirements(args.EnvironmentVars, galaxyExtraArgs(args, GalaxyRole)); err != nil {
		return err
	}
	return nil
}

// skipGalaxyInstall reports whether the Galaxy install step must be skipped.
// The template-level flag provides the default; when the template allows
// overriding it, the task-level flag takes precedence.
func (t *AnsibleApp) skipGalaxyInstall(args LocalAppInstallingArgs) bool {
	tplParams, ok := args.TplParams.(*db.AnsibleTemplateParams)
	if !ok || tplParams == nil {
		return false
	}

	skip := tplParams.SkipGalaxyInstall

	if tplParams.AllowOverrideSkipGalaxyInstall {
		if params, ok := args.Params.(*db.AnsibleTaskParams); ok && params != nil {
			skip = params.SkipGalaxyInstall
		}
	}

	return skip
}

func (t *AnsibleApp) getRepoPath() string {
	return t.Repository.GetFullPath(t.Template.ID)
}

// requirementsHashFilePath is the path to the cached hash of a requirements file. Hashes are kept
// under GetInternalPath (repository_<id>_template_<id>_internal) so they are not written next to
// repository files (especially local paths shared by multiple templates). Legacy *.md5 files beside
// requirements.yml are not read.
func (t *AnsibleApp) requirementsHashFilePath(requirementsType GalaxyRequirementsType, requirementsFilePath string) string {
	sum := md5.Sum([]byte(requirementsFilePath))
	internalDir := t.Repository.GetInternalPath(t.Template.ID)
	return path.Join(internalDir, fmt.Sprintf("requirements_%x_%s.md5", sum, requirementsType))
}

func (t *AnsibleApp) installGalaxyRequirementsFile(requirementsType GalaxyRequirementsType, requirementsFilePath string, environmentVars, extraArgs []string) error {
	requirementsHashFilePath := t.requirementsHashFilePath(requirementsType, requirementsFilePath)

	if _, err := os.Stat(requirementsFilePath); err != nil {
		t.Log("No " + requirementsFilePath + " file found. Skip galaxy install process.\n")
		return nil
	}

	if hasRequirementsChanges(requirementsFilePath, requirementsHashFilePath) {
		galaxyArgs := append([]string{
			string(requirementsType),
			"install",
			"-r",
			requirementsFilePath,
			"--force",
		}, extraArgs...)

		if err := t.runGalaxy(galaxyArgs, environmentVars); err != nil {
			return err
		}
		if err := os.MkdirAll(t.Repository.GetInternalPath(t.Template.ID), 0o755); err != nil {
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

func (t *AnsibleApp) installRolesRequirements(environmentVars, extraArgs []string) (err error) {
	// default roles path
	err = t.installGalaxyRequirementsFile(GalaxyRole, path.Join(t.GetPlaybookDir(), "roles", "requirements.yml"), environmentVars, extraArgs)
	if err != nil {
		return
	}
	err = t.installGalaxyRequirementsFile(GalaxyRole, path.Join(t.GetPlaybookDir(), "requirements.yml"), environmentVars, extraArgs)
	if err != nil {
		return
	}

	// alternative roles path
	err = t.installGalaxyRequirementsFile(GalaxyRole, path.Join(t.getRepoPath(), "roles", "requirements.yml"), environmentVars, extraArgs)
	if err != nil {
		return
	}
	err = t.installGalaxyRequirementsFile(GalaxyRole, path.Join(t.getRepoPath(), "requirements.yml"), environmentVars, extraArgs)
	return
}

func (t *AnsibleApp) installCollectionsRequirements(environmentVars, extraArgs []string) (err error) {
	// default collections path
	err = t.installGalaxyRequirementsFile(GalaxyCollection, path.Join(t.GetPlaybookDir(), "collections", "requirements.yml"), environmentVars, extraArgs)
	if err != nil {
		return
	}
	err = t.installGalaxyRequirementsFile(GalaxyCollection, path.Join(t.GetPlaybookDir(), "requirements.yml"), environmentVars, extraArgs)
	if err != nil {
		return
	}

	// alternative collections path
	err = t.installGalaxyRequirementsFile(GalaxyCollection, path.Join(t.getRepoPath(), "collections", "requirements.yml"), environmentVars, extraArgs)
	if err != nil {
		return
	}
	err = t.installGalaxyRequirementsFile(GalaxyCollection, path.Join(t.getRepoPath(), "requirements.yml"), environmentVars, extraArgs)
	return
}

func (t *AnsibleApp) runGalaxy(args []string, environmentVars []string) error {
	return t.Playbook.RunGalaxy(args, environmentVars)
}

// galaxyExtraArgs returns the template-configured arguments for one galaxy
// subcommand. Roles and collections are configured separately because the two
// accept different flags.
func galaxyExtraArgs(args LocalAppInstallingArgs, requirementsType GalaxyRequirementsType) []string {
	tplParams, ok := args.TplParams.(*db.AnsibleTemplateParams)
	if !ok || tplParams == nil {
		return nil
	}

	switch requirementsType {
	case GalaxyRole:
		return tplParams.GalaxyRoleArgs
	case GalaxyCollection:
		return tplParams.GalaxyCollectionArgs
	}
	return nil
}
