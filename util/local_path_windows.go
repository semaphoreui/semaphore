//go:build windows

package util

// NormalizeLocalFilesystemPath converts paths that Git Bash / MSYS often produce
// into a form native Windows APIs accept.
func NormalizeLocalFilesystemPath(p string) string {
	return normalizeMsysPath(p)
}
