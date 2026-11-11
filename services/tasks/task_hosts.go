package tasks

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"unicode/utf8"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
)

//go:embed callback_plugins/*.py
var hostsCallbacks embed.FS

const HostsFileEnv = "SEMAPHORE_HOSTS_FILE"

// HostsCallbackPlugin returns the callback plugin source embedded in the binary.
func HostsCallbackPlugin() []byte {
	data, err := hostsCallbacks.ReadFile("callback_plugins/semaphore_hosts.py")
	if err != nil {
		panic(err)
	}
	return data
}

const hostsFileName = "hosts.json"

const maxHostNameLength = 250

// hostsCapture is the file the Ansible callback plugin writes per-host
// statistics to while the task runs.
type hostsCapture struct {
	dir  string
	file string
}

// beginHostsCapture creates a directory for the task under dir with an empty
// hosts file in it. Both belong to the user the task process runs as: the
// directory is 0700, the file 0600.
func beginHostsCapture(dir string, taskID int) (*hostsCapture, error) {
	taskDir, err := os.MkdirTemp(dir, fmt.Sprintf("task_%d_hosts_", taskID))
	if err != nil {
		return nil, err
	}

	file := filepath.Join(taskDir, hostsFileName)
	if err = os.WriteFile(file, nil, 0o600); err == nil {
		if err = util.ChownDir(taskDir); err == nil {
			err = util.ChownDir(file)
		}
	}
	if err != nil {
		_ = os.RemoveAll(taskDir)
		return nil, err
	}

	return &hostsCapture{dir: taskDir, file: file}, nil
}

func (c *hostsCapture) Env() []string {
	return []string{HostsFileEnv + "=" + c.file}
}

// CallbackDir extracts the embedded callback plugin next to the hosts file.
// The caller adds the returned directory to ANSIBLE_CALLBACK_PLUGINS.
func (c *hostsCapture) CallbackDir() (string, error) {
	target := filepath.Join(c.dir, "callback_plugins")
	if err := os.MkdirAll(target, 0o700); err != nil {
		return "", err
	}
	if err := util.ChownDir(target); err != nil {
		return "", err
	}

	entries, err := fs.ReadDir(hostsCallbacks, "callback_plugins")
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := hostsCallbacks.ReadFile("callback_plugins/" + entry.Name())
		if err != nil {
			return "", err
		}
		out := filepath.Join(target, entry.Name())
		if err = os.WriteFile(out, data, 0o600); err != nil {
			return "", err
		}
		if err = util.ChownDir(out); err != nil {
			return "", err
		}
	}

	return target, nil
}

// Read returns the document the plugin wrote, or nil if it never ran.
func (c *hostsCapture) Read() (*string, error) {
	data, err := os.ReadFile(c.file)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	document := string(data)
	return &document, nil
}

func (c *hostsCapture) Close() {
	_ = os.RemoveAll(c.dir)
}

// parseHostsDocument turns the document written by the callback plugin into
// rows sorted by host name.
func parseHostsDocument(document string, projectID int, taskID int) ([]db.AnsibleTaskHost, error) {
	var doc struct {
		Version int
		Hosts   map[string]map[string]json.RawMessage
	}
	if err := json.Unmarshal([]byte(document), &doc); err != nil {
		return nil, fmt.Errorf("hosts document: %w", err)
	}
	if doc.Version != 1 {
		return nil, fmt.Errorf("hosts document: unsupported version %d", doc.Version)
	}
	if doc.Hosts == nil {
		return nil, fmt.Errorf("hosts document: hosts is missing")
	}

	hosts := make([]db.AnsibleTaskHost, 0, len(doc.Hosts))
	for name, stats := range doc.Hosts {
		if name == "" {
			return nil, fmt.Errorf("hosts document: empty host name")
		}
		if utf8.RuneCountInString(name) > maxHostNameLength {
			return nil, fmt.Errorf("hosts document: host name %.20q... is longer than %d characters", name, maxHostNameLength)
		}

		host := db.AnsibleTaskHost{ProjectID: projectID, TaskID: taskID, Host: name}
		counters := map[string]*int{
			"ok":          &host.Ok,
			"changed":     &host.Changed,
			"unreachable": &host.Unreachable,
			"failed":      &host.Failed,
			"skipped":     &host.Skipped,
			"rescued":     &host.Rescued,
			"ignored":     &host.Ignored,
		}
		for key, counter := range counters {
			raw, ok := stats[key]
			if !ok {
				continue
			}
			if err := json.Unmarshal(raw, counter); err != nil || *counter < 0 {
				return nil, fmt.Errorf("hosts document: host %q: %s must be a non-negative integer", name, key)
			}
		}
		hosts = append(hosts, host)
	}

	sort.Slice(hosts, func(i, j int) bool { return hosts[i].Host < hosts[j].Host })
	return hosts, nil
}
