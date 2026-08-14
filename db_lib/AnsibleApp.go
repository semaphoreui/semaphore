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

	// Keys are tried one at a time instead of being loaded into a single agent:
	// ssh stops at the first key the host accepts, and a key can authenticate
	// but still have no access to the requested repository (a deploy key, for
	// example), which would hide the remaining keys.
	//
	// Each attempt installs what its key can reach and leaves the rest to the
	// next one, so a requirements file whose roles live in different repositories
	// is covered by the keys of the template together.
	keys := t.galaxyInstallKeys()

	// Without a usable key the requirements are installed with the environment
	// of the task, which is what a public repository needs.
	if len(keys) == 0 {
		return t.installRequirements(args.EnvironmentVars, true)
	}

	var err error
	for i, key := range keys {
		if i > 0 {
			t.Log(fmt.Sprintf("Galaxy install failed, retrying with key %q.\n", key.Name))
		}

		// Only the first attempt forces a fresh install: ansible-galaxy --force
		// removes a role before fetching it again, so forcing on a later attempt
		// would delete the roles an earlier key installed whenever this key can
		// not reach them.
		if err = t.installRequirementsWithKey(args, key, i == 0); err == nil {
			return nil
		}
	}

	return err
}

// galaxyInstallKeys returns the repository key followed by the template's SSH keys.
// galaxyInstallKeys returns the keys to try, the repository key first. Only ssh
// keys are returned: they are the ones an agent can serve, and a repository
// without one (a public repository, or one reached over https) must still be
// able to install public requirements, which the empty result allows.
func (t *AnsibleApp) galaxyInstallKeys() []db.AccessKey {
	keys := make([]db.AccessKey, 0, 1+len(t.Template.Keys))

	for _, key := range append([]db.AccessKey{t.Repository.SSHKey}, t.Template.Keys...) {
		if key.Type == db.AccessKeySSH {
			keys = append(keys, key)
		}
	}

	return keys
}

func (t *AnsibleApp) installRequirementsWithKey(args LocalAppInstallingArgs, key db.AccessKey, force bool) error {
	keyInstallation, err := args.Installer.Install(key, db.AccessKeyRoleGit, t.Logger)
	if err != nil {
		return err
	}

	defer func() {
		// The agent holds key material, so a failure to close it is worth
		// reporting even though it does not fail the installation.
		if destroyErr := keyInstallation.Destroy(); destroyErr != nil {
			t.Log(fmt.Sprintf("Can't destroy galaxy key %q, error: %s\n", key.Name, destroyErr))
		}
	}()

	environmentVars := append(append([]string{}, args.EnvironmentVars...), keyInstallation.GetGitEnv()...)

	return t.installRequirements(environmentVars, force)
}

// installRequirements installs the collection and role requirements files with
// the given environment.
func (t *AnsibleApp) installRequirements(environmentVars []string, force bool) error {
	if err := t.installCollectionsRequirements(environmentVars, force); err != nil {
		return err
	}

	return t.installRolesRequirements(environmentVars, force)
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

// galaxyInstallArgs builds the ansible-galaxy command line. force re-fetches
// roles which are already installed; without it they are left alone, which is
// what a retry with another key needs so the roles installed by an earlier key
// survive. ansible-galaxy removes a role before re-fetching it, so forcing on a
// retry deletes what the previous key installed whenever this key cannot reach
// the same repository.
func galaxyInstallArgs(requirementsType GalaxyRequirementsType, requirementsFilePath string, force bool) []string {
	args := []string{
		string(requirementsType),
		"install",
		"-r",
		requirementsFilePath,
	}

	if force {
		args = append(args, "--force")
	}

	return args
}

func (t *AnsibleApp) installGalaxyRequirementsFile(requirementsType GalaxyRequirementsType, requirementsFilePath string, environmentVars []string, force bool) error {
	requirementsHashFilePath := t.requirementsHashFilePath(requirementsType, requirementsFilePath)

	if _, err := os.Stat(requirementsFilePath); err != nil {
		t.Log("No " + requirementsFilePath + " file found. Skip galaxy install process.\n")
		return nil
	}

	if hasRequirementsChanges(requirementsFilePath, requirementsHashFilePath) {
		if err := t.runGalaxy(galaxyInstallArgs(requirementsType, requirementsFilePath, force), environmentVars); err != nil {
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

func (t *AnsibleApp) installRolesRequirements(environmentVars []string, force bool) (err error) {
	// default roles path
	err = t.installGalaxyRequirementsFile(GalaxyRole, path.Join(t.GetPlaybookDir(), "roles", "requirements.yml"), environmentVars, force)
	if err != nil {
		return
	}
	err = t.installGalaxyRequirementsFile(GalaxyRole, path.Join(t.GetPlaybookDir(), "requirements.yml"), environmentVars, force)
	if err != nil {
		return
	}

	// alternative roles path
	err = t.installGalaxyRequirementsFile(GalaxyRole, path.Join(t.getRepoPath(), "roles", "requirements.yml"), environmentVars, force)
	if err != nil {
		return
	}
	err = t.installGalaxyRequirementsFile(GalaxyRole, path.Join(t.getRepoPath(), "requirements.yml"), environmentVars, force)
	return
}

func (t *AnsibleApp) installCollectionsRequirements(environmentVars []string, force bool) (err error) {
	// default collections path
	err = t.installGalaxyRequirementsFile(GalaxyCollection, path.Join(t.GetPlaybookDir(), "collections", "requirements.yml"), environmentVars, force)
	if err != nil {
		return
	}
	err = t.installGalaxyRequirementsFile(GalaxyCollection, path.Join(t.GetPlaybookDir(), "requirements.yml"), environmentVars, force)
	if err != nil {
		return
	}

	// alternative collections path
	err = t.installGalaxyRequirementsFile(GalaxyCollection, path.Join(t.getRepoPath(), "collections", "requirements.yml"), environmentVars, force)
	if err != nil {
		return
	}
	err = t.installGalaxyRequirementsFile(GalaxyCollection, path.Join(t.getRepoPath(), "requirements.yml"), environmentVars, force)
	return
}

func (t *AnsibleApp) runGalaxy(args []string, environmentVars []string) error {
	return t.Playbook.RunGalaxy(args, environmentVars)
}
