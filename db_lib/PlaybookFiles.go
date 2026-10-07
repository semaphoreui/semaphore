package db_lib

import (
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/semaphoreui/semaphore/db"
)

// maxRepositoryFiles caps the number of paths returned by FindRepositoryFiles.
// The picker is a suggestion list, not a file browser: a path which is not in it
// can still be typed.
const maxRepositoryFiles = 30

// ansibleLayoutDirs are conventional Ansible directories that never contain a
// top-level playbook. They are skipped for Ansible so the result is not flooded
// with yml files that are not playbooks, and kept for every other app, where
// they are ordinary directories.
var ansibleLayoutDirs = map[string]bool{
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

// FindRepositoryFiles returns a sorted slice of paths, relative to rootDir, of
// the entries which could be the entry point of app.
//
// dir narrows a directory listing to the directories inside it, so a path can be
// built one step at a time. It is ignored by the apps which list files. The
// result is capped at maxRepositoryFiles entries.
func FindRepositoryFiles(rootDir string, app db.TemplateApp, dir string) ([]string, error) {
	filter := app.RepositoryFileFilter()

	if filter.OnlyDirectories {
		return findDirs(rootDir, dir)
	}

	return findFiles(rootDir, app, filter.Extensions)
}

// findDirs returns the directories directly inside dir, a path relative to
// rootDir, prefixed with it. It does not recurse: the next level is listed when
// the user asks for it by typing the separator.
//
// A directory which does not exist yields no entries rather than an error: dir
// is what the user has typed so far and is routinely half a name.
func findDirs(rootDir string, dir string) ([]string, error) {
	// Anchoring to "/" before cleaning resolves away any ".." the request
	// carries, so the join can not escape the repository.
	rel := strings.TrimPrefix(filepath.Clean("/"+filepath.FromSlash(dir)), string(filepath.Separator))

	entries, err := os.ReadDir(filepath.Join(rootDir, rel))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	prefix := filepath.ToSlash(rel)
	if prefix != "" && prefix != "." {
		prefix += "/"
	} else {
		prefix = ""
	}

	var result []string

	for _, entry := range entries {
		if len(result) >= maxRepositoryFiles {
			break
		}

		// Dot directories are tooling state - .git, .terraform - never a root.
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		result = append(result, prefix+entry.Name())
	}

	sort.Strings(result)

	return result, nil
}

func findFiles(rootDir string, app db.TemplateApp, extensions []string) ([]string, error) {
	var result []string

	skipAnsibleDirs := app == db.AppAnsible || app == ""

	err := filepath.WalkDir(rootDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if len(result) >= maxRepositoryFiles {
			return filepath.SkipAll
		}

		if d.IsDir() {
			if p == rootDir {
				return nil
			}

			if d.Name() == ".git" || (skipAnsibleDirs && ansibleLayoutDirs[d.Name()]) {
				return filepath.SkipDir
			}

			return nil
		}

		if len(extensions) > 0 &&
			!slices.Contains(extensions, strings.ToLower(filepath.Ext(d.Name()))) {
			return nil
		}

		rel, err := filepath.Rel(rootDir, p)
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
