package db_lib

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/semaphoreui/semaphore/db"
)

// maxPlaybookFiles caps the number of paths returned by FindPlaybooks
// as a safety net against huge repositories.
const maxPlaybookFiles = 1000

// excludedPlaybookDirs are conventional Ansible directories that never contain
// top-level playbooks (roles, variable files, templates, etc). They are skipped
// while walking the repository to avoid flooding the result with non-playbook
// yml/yaml files.
var excludedPlaybookDirs = map[string]bool{
	".git":           true,
	"roles":          true,
	"group_vars":     true,
	"host_vars":      true,
	"files":          true,
	"templates":      true,
	"library":        true,
	"filter_plugins": true,
	"module_utils":   true,
	"meta":           true,
	"handlers":       true,
	"defaults":       true,
	"vars":           true,
	"molecule":       true,
	"tests":          true,
}

// excludedCommonDirs are dependency and tool cache directories that never
// hold what a template points at. Hidden directories (.git, .github,
// .terraform, .venv, ...) are always skipped as well.
var excludedCommonDirs = map[string]bool{
	"node_modules": true,
	"__pycache__":  true,
	"venv":         true,
}

// repositoryFileMatcher describes which entries of a repository are offered
// as completions for the template's playbook field of one app.
type repositoryFileMatcher struct {
	// excludedDirs are directory names skipped while walking.
	excludedDirs map[string]bool
	// extensions are lower-case file extensions to match; empty means any file.
	extensions []string
	// fileNames are exact file names to match, in addition to extensions.
	fileNames []string
	// directories means the result lists the directories that contain a
	// matching file instead of the files themselves.
	directories bool
}

func (m repositoryFileMatcher) matches(name string) bool {
	if len(m.extensions) == 0 && len(m.fileNames) == 0 {
		return true
	}

	for _, n := range m.fileNames {
		if name == n {
			return true
		}
	}

	ext := strings.ToLower(filepath.Ext(name))
	for _, e := range m.extensions {
		if ext == e {
			return true
		}
	}

	return false
}

func repositoryFileMatcherFor(app db.TemplateApp) repositoryFileMatcher {
	switch app {
	case db.AppAnsible, "":
		return repositoryFileMatcher{excludedDirs: excludedPlaybookDirs, extensions: []string{".yml", ".yaml"}}
	case db.AppTerraform, db.AppTofu:
		// The field is the subdirectory holding the root module.
		return repositoryFileMatcher{excludedDirs: excludedCommonDirs, extensions: []string{".tf", ".tofu"}, directories: true}
	case db.AppTerragrunt:
		return repositoryFileMatcher{excludedDirs: excludedCommonDirs, fileNames: []string{"terragrunt.hcl"}, directories: true}
	case db.AppBash:
		return repositoryFileMatcher{excludedDirs: excludedCommonDirs, extensions: []string{".sh", ".bash"}}
	case db.AppPython:
		return repositoryFileMatcher{excludedDirs: excludedCommonDirs, extensions: []string{".py"}}
	case db.AppPowerShell:
		return repositoryFileMatcher{excludedDirs: excludedCommonDirs, extensions: []string{".ps1"}}
	default:
		return repositoryFileMatcher{excludedDirs: excludedCommonDirs}
	}
}

// FindPlaybooks walks rootDir and returns a sorted slice of paths (relative
// to rootDir, slash-separated) that suit the playbook field of a template of
// the given app: playbooks for Ansible, module directories for
// Terraform/OpenTofu/Terragrunt, scripts for Bash/Python/PowerShell, and any
// file for other apps. A nil app means Ansible. The repository root itself
// is never listed, hidden
// directories are skipped and symlinks are not followed. The result is capped
// at maxPlaybookFiles entries.
func FindPlaybooks(rootDir string, app *db.TemplateApp) ([]string, error) {
	templateApp := db.AppAnsible
	if app != nil {
		templateApp = *app
	}

	matcher := repositoryFileMatcherFor(templateApp)

	var result []string
	seenDirs := make(map[string]bool)

	err := filepath.WalkDir(rootDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if len(result) >= maxPlaybookFiles {
			return filepath.SkipAll
		}

		if d.IsDir() {
			if p != rootDir && (strings.HasPrefix(d.Name(), ".") || matcher.excludedDirs[d.Name()]) {
				return filepath.SkipDir
			}
			return nil
		}

		if !matcher.matches(d.Name()) {
			return nil
		}

		entry := p
		if matcher.directories {
			entry = filepath.Dir(p)
			if entry == rootDir || seenDirs[entry] {
				return nil
			}
			seenDirs[entry] = true
		}

		rel, err := filepath.Rel(rootDir, entry)
		if err != nil {
			return err
		}

		result = append(result, filepath.ToSlash(rel))

		return nil
	})

	if err != nil {
		return nil, err
	}

	sort.Strings(result)

	return result, nil
}
