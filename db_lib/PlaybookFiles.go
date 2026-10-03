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
// the entries which could be the entry point of app. The result is capped at
// maxRepositoryFiles entries.
func FindRepositoryFiles(rootDir string, app db.TemplateApp) ([]string, error) {
	filter := app.RepositoryFileFilter()

	if filter.OnlyDirectories {
		return findTopLevelDirs(rootDir)
	}

	return findFiles(rootDir, app, filter.Extensions)
}

// findTopLevelDirs returns the directories directly under rootDir. It does not
// recurse: a terraform root is a directory of the repository, not any directory
// below it.
func findTopLevelDirs(rootDir string) ([]string, error) {
	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return nil, err
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

		result = append(result, entry.Name())
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
