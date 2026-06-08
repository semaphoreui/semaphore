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

// fileExists reports whether the path exists and is a regular file (not a directory).
func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}

// dirExists reports whether the path exists and is a directory.
func dirExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
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
	rolePaths, collectionPaths := t.resolveGalaxyRequirements()

	if err := t.installCollectionsRequirements(collectionPaths, args.EnvironmentVars); err != nil {
		return err
	}
	if err := t.installRolesRequirements(rolePaths, args.EnvironmentVars); err != nil {
		return err
	}
	return nil
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

// installGalaxyRequirementsFile installs a single requirements file. The file is assumed to exist:
// existence and logging of missing files is handled by resolveGalaxyRequirements. Installation is
// skipped when the file content has not changed since the last successful install.
func (t *AnsibleApp) installGalaxyRequirementsFile(requirementsType GalaxyRequirementsType, requirementsFilePath string, environmentVars []string) error {
	requirementsHashFilePath := t.requirementsHashFilePath(requirementsType, requirementsFilePath)

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

// resolveGalaxyRequirements collects the requirements files that should be installed and returns
// the existing paths split by type (roles, collections).
//
// Search rules:
//   - <dir>/roles/requirements.yml and <dir>/collections/requirements.yml are type-specific
//     subdirectory paths. If the subdirectory does not exist, the path is skipped silently.
//     If the subdirectory exists but contains no requirements.yml, a warning is logged.
//   - <dir>/requirements.yml is a shared file that may contain both roles and collections, so it
//     is offered to both types. If none of the shared files exist anywhere, a single message
//     listing the searched paths is logged.
//
// <dir> is checked both as the playbook directory and as the repository root.
func (t *AnsibleApp) resolveGalaxyRequirements() (rolePaths []string, collectionPaths []string) {
	playbookDir := t.GetPlaybookDir()
	repoPath := t.getRepoPath()

	// Base directories to search, de-duplicated (playbook dir may equal repo root).
	baseDirs := []string{playbookDir}
	if repoPath != playbookDir {
		baseDirs = append(baseDirs, repoPath)
	}

	// --- Type-specific subdirectory requirements: <dir>/roles|collections/requirements.yml ---
	type subdir struct {
		reqType GalaxyRequirementsType
		dirName string
		target  *[]string
	}
	subdirs := []subdir{
		{GalaxyRole, "roles", &rolePaths},
		{GalaxyCollection, "collections", &collectionPaths},
	}

	for _, base := range baseDirs {
		for _, sd := range subdirs {
			dir := path.Join(base, sd.dirName)
			if !dirExists(dir) {
				// No roles/ or collections/ directory: nothing to install, stay silent.
				continue
			}
			reqFile := path.Join(dir, "requirements.yml")
			if fileExists(reqFile) {
				*sd.target = append(*sd.target, reqFile)
			} else {
				// Directory exists but has no requirements.yml: worth highlighting.
				t.Log("Warning: " + dir + " exists but contains no requirements.yml.\n")
			}
		}
	}

	// --- Shared requirements: <dir>/requirements.yml (may hold roles and collections) ---
	var sharedSearched []string
	var sharedFound []string
	for _, base := range baseDirs {
		reqFile := path.Join(base, "requirements.yml")
		sharedSearched = append(sharedSearched, reqFile)
		if fileExists(reqFile) {
			sharedFound = append(sharedFound, reqFile)
		}
	}

	if len(sharedFound) == 0 {
		// None of the shared requirements files exist: log once, listing where we looked.
		msg := "No requirements.yml found. Skip galaxy install process. Searched:"
		for _, p := range sharedSearched {
			msg += "\n  " + p
		}
		t.Log(msg + "\n")
	} else {
		// A shared file may contain both roles and collections, so offer it to both types.
		for _, p := range sharedFound {
			rolePaths = append(rolePaths, p)
			collectionPaths = append(collectionPaths, p)
		}
	}

	return
}

func (t *AnsibleApp) installRolesRequirements(paths []string, environmentVars []string) error {
	for _, p := range paths {
		if err := t.installGalaxyRequirementsFile(GalaxyRole, p, environmentVars); err != nil {
			return err
		}
	}
	return nil
}

func (t *AnsibleApp) installCollectionsRequirements(paths []string, environmentVars []string) error {
	for _, p := range paths {
		if err := t.installGalaxyRequirementsFile(GalaxyCollection, p, environmentVars); err != nil {
			return err
		}
	}
	return nil
}

func (t *AnsibleApp) runGalaxy(args []string, environmentVars []string) error {
	return t.Playbook.RunGalaxy(args, environmentVars)
}
