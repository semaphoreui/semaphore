//go:build !windows

package util

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppProcessGroup(t *testing.T) {
	config := ConfigType{Process: &ConfigProcess{}}

	attributes := config.GetAppSysProcAttr()
	if attributes == nil || !attributes.Setsid {
		t.Fatal("app process must start in a new process group")
	}
}

func TestParseLinuxCredentialUint(t *testing.T) {
	tests := []struct {
		in      string
		wantVal uint32
		wantOk  bool
	}{
		{"1", 1, true},
		{"65534", 65534, true},
		{strconv.FormatUint(1<<32-1, 10), 1<<32 - 1, true},
		{"0", 0, false},
		{"4294967296", 0, false},
		{"-1", 0, false},
		{"notint", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, ok := parseLinuxCredentialUint(tt.in)
			if ok != tt.wantOk || got != tt.wantVal {
				t.Fatalf("parseLinuxCredentialUint(%q) = (%v, %v), want (%v, %v)", tt.in, got, ok, tt.wantVal, tt.wantOk)
			}
		})
	}
}

type chownCall struct {
	path string
	uid  int
	gid  int
}

func TestChownTree_ContentFirstRootLastNoSymlinkFollowed(t *testing.T) {
	root := filepath.Join(t.TempDir(), "agent_1")
	require.NoError(t, os.MkdirAll(filepath.Join(root, "codex", "sessions"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(root, "codex", "config.toml"), []byte("x"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "mcp.json"), []byte("x"), 0o600))

	outside := filepath.Join(t.TempDir(), "outside")
	require.NoError(t, os.WriteFile(outside, []byte("x"), 0o600))
	require.NoError(t, os.Symlink(outside, filepath.Join(root, "link")))

	var calls []chownCall
	err := chownTree(root, 1001, 1002, func(name string, uid int, gid int) error {
		calls = append(calls, chownCall{name, uid, gid})
		return nil
	})
	require.NoError(t, err)

	position := make(map[string]int, len(calls))
	for i, call := range calls {
		assert.Equal(t, 1001, call.uid)
		assert.Equal(t, 1002, call.gid)
		position[call.path] = i
	}

	for _, p := range []string{
		root,
		filepath.Join(root, "codex"),
		filepath.Join(root, "codex", "sessions"),
		filepath.Join(root, "codex", "config.toml"),
		filepath.Join(root, "mcp.json"),
		filepath.Join(root, "link"),
	} {
		assert.Contains(t, position, p)
	}
	assert.Len(t, calls, 6)
	assert.NotContains(t, position, outside, "a link is given away itself, not its target")

	assert.Equal(t, len(calls)-1, position[root], "the root opens the tree, so it goes last")
	assert.Less(t, position[filepath.Join(root, "codex", "config.toml")], position[filepath.Join(root, "codex")])
	assert.Less(t, position[filepath.Join(root, "codex", "sessions")], position[filepath.Join(root, "codex")])
}

func TestChownTree_StopsOnError(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "mcp.json"), []byte("x"), 0o600))

	err := chownTree(root, 1001, 1002, func(string, int, int) error { return os.ErrPermission })
	assert.ErrorIs(t, err, os.ErrPermission)

	err = chownTree(filepath.Join(root, "missing"), 1001, 1002, func(string, int, int) error { return nil })
	assert.Error(t, err)
}

// ChownTree uses the process config user and keeps the modes. The current
// user stands for it here: giving a file to yourself needs no privileges.
func TestChownTree_ProcessUserKeepsModes(t *testing.T) {
	previous := Config
	t.Cleanup(func() { Config = previous })

	root := filepath.Join(t.TempDir(), "agent_1")
	require.NoError(t, os.MkdirAll(root, 0o700))
	file := filepath.Join(root, "mcp.json")
	require.NoError(t, os.WriteFile(file, []byte("x"), 0o600))

	t.Run("no process user", func(t *testing.T) {
		Config = &ConfigType{Process: &ConfigProcess{}}
		require.NoError(t, ChownTree(root))
	})

	t.Run("process uid and gid", func(t *testing.T) {
		uid, gid := uint32(os.Getuid()), uint32(os.Getgid())
		Config = &ConfigType{Process: &ConfigProcess{UID: &uid, GID: &gid}}
		require.NoError(t, ChownTree(root))

		dirInfo, err := os.Stat(root)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o700), dirInfo.Mode().Perm())

		fileInfo, err := os.Stat(file)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), fileInfo.Mode().Perm())
	})
}
