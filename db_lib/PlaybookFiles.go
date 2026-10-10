package db_lib

import (
	"errors"
	"io"
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
	// carries. os.Root then confines the listing to the repository, which Clean
	// alone does not: a directory symlink in the checkout would otherwise be
	// followed out of it.
	rel := strings.TrimPrefix(filepath.Clean("/"+filepath.FromSlash(dir)), string(filepath.Separator))
	if rel == "" {
		rel = "."
	}

	root, err := os.OpenRoot(rootDir)
	if err != nil {
		return nil, err
	}
	defer root.Close() //nolint:errcheck

	f, err := root.Open(rel)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close() //nolint:errcheck

	// A path naming a file is half-typed input like any other, not an error.
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, nil
	}

	// The directory being listed is itself a valid entry point - a project with
	// main.tf beside its README has no subdirectory to offer otherwise.
	prefix := ""
	if rel != "." {
		prefix = filepath.ToSlash(rel) + "/"
	}

	result := []string{"."}
	if prefix != "" {
		result = []string{strings.TrimSuffix(prefix, "/")}
	}

	// Read in batches and keep only the entries which make the cut, so a
	// directory of any size costs the same. They are kept in order because
	// File.ReadDir returns directory order, which would otherwise make the
	// result an arbitrary subset of a large directory.
	for {
		entries, err := f.ReadDir(readDirBatch)

		for _, entry := range entries {
			// Dot directories are tooling state - .git, .terraform - never a root.
			if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
				continue
			}

			result = insertCapped(result, prefix+entry.Name())
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
	}

	return result, nil
}

// readDirBatch is how many directory entries are read at a time.
const readDirBatch = 256

// insertCapped puts name into the sorted result and drops whatever falls past
// maxRepositoryFiles, so the slice never outgrows the limit.
func insertCapped(result []string, name string) []string {
	at := sort.SearchStrings(result, name)
	if at >= maxRepositoryFiles {
		return result
	}

	result = append(result, "")
	copy(result[at+1:], result[at:])
	result[at] = name

	if len(result) > maxRepositoryFiles {
		result = result[:maxRepositoryFiles]
	}

	return result
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

			// Dot directories are tooling state - .git, .venv, .tox - and can
			// hold enough matching files to fill the cap on their own.
			if strings.HasPrefix(d.Name(), ".") ||
				(skipAnsibleDirs && ansibleLayoutDirs[d.Name()]) {
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
