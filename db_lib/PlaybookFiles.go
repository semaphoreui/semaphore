package db_lib

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

// maxPlaybookFiles caps the number of playbook paths returned by FindPlaybooks
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

// FindPlaybooks walks rootDir and returns a sorted slice of paths (relative to
// rootDir) of files with a .yml or .yaml extension (case-insensitive),
// skipping .git and conventional non-playbook Ansible directories.
// The result is capped at maxPlaybookFiles entries.
func FindPlaybooks(rootDir string) ([]string, error) {
	var result []string

	err := filepath.WalkDir(rootDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if len(result) >= maxPlaybookFiles {
			return filepath.SkipAll
		}

		if d.IsDir() {
			if p != rootDir && excludedPlaybookDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(d.Name()))
		if ext != ".yml" && ext != ".yaml" {
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
