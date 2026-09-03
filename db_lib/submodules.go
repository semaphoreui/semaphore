package db_lib

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/semaphoreui/semaphore/util"
)

func copySubmoduleStore(source, destPath string) error {
	src := filepath.Join(source, ".git", "modules")

	if _, err := os.Stat(src); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	dst := filepath.Join(destPath, ".git", "modules")

	if err := os.CopyFS(dst, os.DirFS(src)); err != nil {
		return err
	}

	if err := filepath.WalkDir(dst, removeStaleModuleIndex); err != nil {
		return err
	}

	// The copy is made by the Semaphore process, but git runs as the configured
	// process user and has to write into the module stores.
	return util.ChownTree(dst)
}

// A copied module store carries the index of the source worktree, listing files
// the fresh copy does not have yet. git rewrites such an index on checkout,
// go-git refuses with "worktree contains unstaged changes".
func removeStaleModuleIndex(path string, entry fs.DirEntry, err error) error {
	if err != nil {
		return err
	}

	if entry.IsDir() || entry.Name() != "index" {
		return nil
	}

	if _, err := os.Stat(filepath.Join(filepath.Dir(path), "HEAD")); err != nil {
		return nil
	}

	return os.Remove(path)
}
