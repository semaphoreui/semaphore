//go:build !windows

package util

// NormalizeLocalFilesystemPath returns the path unchanged: MSYS-style drive
// paths only need rewriting on Windows.
func NormalizeLocalFilesystemPath(p string) string {
	return p
}
