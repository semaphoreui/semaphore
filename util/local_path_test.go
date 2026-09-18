package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsWindowsLocalRepositoryPath(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{`D:\repo`, true},
		{`D:/repo`, true},
		{`D:`, true},
		{`D:repo`, false},
		{`a://example/repo.git`, false},
		{`c:\`, true},
		{`\\server\share`, true},
		{`\\`, false},
		{``, false},
		{`/usr/src`, false},
		{`https://x`, false},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			assert.Equal(t, tt.want, IsWindowsLocalRepositoryPath(tt.path))
		})
	}
}

func TestNormalizeMsysPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{"leading slash before drive", `/D:/ps-demo-script`, `D:/ps-demo-script`},
		{"leading slash before drive backslash", `/d:\ps-demo-script`, `d:\ps-demo-script`},
		{"msys drive path", `/d/ps-demo-script/extra`, `D:\ps-demo-script\extra`},
		{"msys drive root", `/d/`, `D:\`},
		{"msys drive lowercase upcased", `/c/Users`, `C:\Users`},
		{"native windows path unchanged", `D:\repo`, `D:\repo`},
		{"unix path unchanged", `/usr/src`, `/usr/src`},
		{"single letter dir unchanged", `/d`, `/d`},
		{"empty", ``, ``},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, normalizeMsysPath(tt.path))
		})
	}
}
