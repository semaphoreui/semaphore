package db_lib

import (
	"io/fs"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/semaphoreui/semaphore/db"
)

// maxRepositoryFiles caps the number of paths returned by FindRepositoryFiles
// as a safety net against huge repositories.
const maxRepositoryFiles = 1000

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

// FindRepositoryFiles walks rootDir and returns a sorted slice of paths,
// relative to rootDir, of the files which could be the entry point of app.
// The result is capped at maxRepositoryFiles entries.
func FindRepositoryFiles(rootDir string, app db.TemplateApp) ([]string, error) {
	var result []string

	extensions := app.RepositoryFileExtensions()
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
