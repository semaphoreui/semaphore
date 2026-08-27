package db

import (
	"path"
	"strings"

	"github.com/semaphoreui/semaphore/pkg/common_errors"
)

const (
	wdNotRelativeMsg = "template working directory must be a relative path inside the repository"
	wdOutsideRepoMsg = "template working directory must not point outside the repository"
)

// ValidateWorkingDirectoryLexically validates both the POSIX and Windows
// interpretations of the working directory. Validation is conservative because it runs before the
// path is stored, while the runner that eventually uses it may run on any
// operating system.
// It does not access the filesystem or resolve symbolic links.
func ValidateWorkingDirectoryLexically(wd string) error {
	if wd == "" {
		return nil
	}

	// Check the working directory using POSIX path semantics only.
	cleaned := path.Clean(wd)
	if path.IsAbs(cleaned) {
		return common_errors.NewValidationError(wdNotRelativeMsg)
	}
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return common_errors.NewValidationError(wdOutsideRepoMsg)
	}

	// Check the working directory using Windows path semantics.
	cleaned = path.Clean(strings.ReplaceAll(wd, "\\", "/"))
	// Like filepath.IsLocal on Windows, conservatively reject any colon:
	// https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/internal/filepathlite/path_windows.go;l=31
	if path.IsAbs(cleaned) || strings.Contains(cleaned, ":") {
		return common_errors.NewValidationError(wdNotRelativeMsg)
	}
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return common_errors.NewValidationError(wdOutsideRepoMsg)
	}

	return nil
}
