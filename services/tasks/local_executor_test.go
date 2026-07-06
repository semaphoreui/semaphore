package db

import (
	"path"
	"strings"
)

// ValidatePlaybookPath checks that a playbook (or script/subdirectory for
// non-Ansible apps) path is relative and stays inside the repository bound
// to the template. Absolute paths and paths escaping the repository via ".."
// are rejected to prevent execution of arbitrary files on the host.
func ValidatePlaybookPath(playbook string, objectName string) error {
	if playbook == "" {
		return nil
	}

	// Treat backslashes as path separators so Windows-style paths
	// (C:\..., ..\..\evil.ps1) can not bypass the checks below.
	p := strings.ReplaceAll(playbook, "\\", "/")

	if path.IsAbs(p) {
		return NewValidationError(objectName + " playbook must be a relative path inside the repository")
	}

	// Windows absolute paths like "C:/..." are not caught by path.IsAbs.
	if len(p) >= 2 && p[1] == ':' {
		return NewValidationError(objectName + " playbook must be a relative path inside the repository")
	}

	cleaned := path.Clean(p)
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return NewValidationError(objectName + " playbook must not point outside the repository")
	}

	return nil
}