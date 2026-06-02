package artifacts

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed ansible/callback_plugins/*.py
var ansibleCallbacks embed.FS

// AnsibleCallbackEnv extracts the embedded Ansible callback plugins into
// `dir`/callback_plugins/ and returns the environment variables needed to
// register them with ansible-playbook without replacing user-defined plugins
// or stdout callbacks. Returns (envVars, nil) on success; on any extraction
// failure it returns (nil, err) and the caller should fall back to running
// without artifact capture.
func AnsibleCallbackEnv(dir string) ([]string, error) {
	if dir == "" {
		return nil, fmt.Errorf("artifacts: ansible callback dir is empty")
	}
	target := filepath.Join(dir, "callback_plugins")
	if err := os.MkdirAll(target, 0o755); err != nil {
		return nil, err
	}

	entries, err := fs.ReadDir(ansibleCallbacks, "ansible/callback_plugins")
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := ansibleCallbacks.ReadFile("ansible/callback_plugins/" + entry.Name())
		if err != nil {
			return nil, err
		}
		out := filepath.Join(target, entry.Name())
		if err := os.WriteFile(out, data, 0o644); err != nil {
			return nil, err
		}
	}

	// ANSIBLE_CALLBACK_PLUGINS *prepends* search paths, so user-defined
	// callbacks shipped alongside the playbook still take precedence.
	// ANSIBLE_CALLBACKS_ENABLED activates our aggregate callback by name
	// without touching the user's stdout callback selection.
	return []string{
		"ANSIBLE_CALLBACK_PLUGINS=" + target,
		"ANSIBLE_CALLBACKS_ENABLED=semaphore_artifacts",
	}, nil
}
