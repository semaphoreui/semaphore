package util

import (
	"path/filepath"
	"runtime"
	"testing"
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
		if got := IsWindowsLocalRepositoryPath(tt.path); got != tt.want {
			t.Errorf("IsWindowsLocalRepositoryPath(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestNormalizeLocalFilesystemPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Run("leading_slash_before_drive", func(t *testing.T) {
			got := NormalizeLocalFilesystemPath("/D:/ps-demo-script")
			want := "D:/ps-demo-script"
			if got != want {
				t.Fatalf("got %q want %q", got, want)
			}
		})
		t.Run("msys_drive_path", func(t *testing.T) {
			got := NormalizeLocalFilesystemPath("/d/ps-demo-script/extra")
			want := "D:" + string(filepath.Separator) + filepath.FromSlash("ps-demo-script/extra")
			if got != want {
				t.Fatalf("got %q want %q", got, want)
			}
		})
	} else {
		p := "/D:/ps-demo-script"
		if got := NormalizeLocalFilesystemPath(p); got != p {
			t.Fatalf("non-Windows should leave path unchanged: got %q", got)
		}
	}
}
