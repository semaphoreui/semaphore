package util

import (
	"path/filepath"
	"runtime"
	"strings"
)

// NormalizeLocalFilesystemPath converts paths that Git Bash / MSYS often produce
// into a form native Windows APIs accept. Other paths are returned unchanged.
func NormalizeLocalFilesystemPath(p string) string {
	if runtime.GOOS != "windows" {
		return p
	}
	// "/D:/path" -> "D:/path", "/d:\path" -> "d:\path" (strip leading "/")
	if len(p) >= 3 && p[0] == '/' && isDriveLetter(p[1]) && p[2] == ':' {
		return p[1:]
	}
	// "/d/path" (MSYS) -> "D:\path"
	if len(p) >= 3 && p[0] == '/' && isDriveLetter(p[1]) && p[2] == '/' {
		rest := p[3:]
		drive := strings.ToUpper(string(p[1]))
		if rest == "" {
			return drive + ":\\"
		}
		return drive + ":" + string(filepath.Separator) + filepath.FromSlash(rest)
	}
	return p
}

func isDriveLetter(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}

// IsWindowsLocalRepositoryPath reports whether p is a Windows drive-letter or UNC
// path used as a local (non-Git) repository URL.
func IsWindowsLocalRepositoryPath(p string) bool {
	if len(p) < 2 {
		return false
	}
	if p[0] == '\\' && p[1] == '\\' {
		return len(p) > 2
	}
	if !isDriveLetter(p[0]) || p[1] != ':' {
		return false
	}
	if len(p) == 2 {
		return true
	}
	if p[2] != '\\' && p[2] != '/' {
		return false
	}
	if len(p) > 3 && p[2] == '/' && p[3] == '/' {
		return false
	}
	return true
}
