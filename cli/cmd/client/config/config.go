// Package config resolves the settings of the `semaphore client` commands.
//
// A setting comes, in this order of precedence, from a command-line flag, an
// environment variable prefixed SEMAPHORE_CLIENT_, the active context of the
// configuration file (~/.semaphore/client.yml by default) and finally a built-in
// default. The package never touches the server configuration (util.Config)
// and keeps no global state: Resolve takes everything it needs as arguments.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Environment variables the client understands. Every one of them carries the
// SEMAPHORE_CLIENT_ prefix so that none can be confused with a server setting.
const (
	EnvConfig    = "SEMAPHORE_CLIENT_CONFIG"
	EnvContext   = "SEMAPHORE_CLIENT_CONTEXT"
	EnvURL       = "SEMAPHORE_CLIENT_URL"
	EnvToken     = "SEMAPHORE_CLIENT_TOKEN"
	EnvTokenFile = "SEMAPHORE_CLIENT_TOKEN_FILE"
	EnvCACert    = "SEMAPHORE_CLIENT_CA_CERT"
	EnvInsecure  = "SEMAPHORE_CLIENT_INSECURE"
	EnvOutput    = "SEMAPHORE_CLIENT_OUTPUT"
	EnvTimeout   = "SEMAPHORE_CLIENT_TIMEOUT"
	EnvProject   = "SEMAPHORE_CLIENT_PROJECT"
	EnvNoInput   = "SEMAPHORE_CLIENT_NO_INPUT"
	EnvDebug     = "SEMAPHORE_CLIENT_DEBUG"
	// EnvCI is honoured like EnvNoInput: CI systems set it, and a prompt there
	// would hang the job.
	EnvCI = "CI"
)

// Output formats a command can print in.
const (
	OutputTable = "table"
	OutputJSON  = "json"
	OutputYAML  = "yaml"
)

// DefaultTimeout bounds one HTTP request when nothing else says otherwise.
const DefaultTimeout = 30 * time.Second

// Context is one saved server: where it is and how to authenticate.
type Context struct {
	URL string `yaml:"url"`
	// Token is the API token in clear text. Prefer TokenFile when the config
	// file is shared through dotfiles.
	Token string `yaml:"token,omitempty"`
	// TokenFile is a path to a file holding the token; "~" expands to $HOME.
	TokenFile string `yaml:"token_file,omitempty"`
	// Project is the default project (id or name) for project-scoped commands.
	Project string `yaml:"project,omitempty"`
	// CACert is a PEM bundle that verifies the server certificate.
	CACert string `yaml:"ca_cert,omitempty"`
	// Insecure disables TLS verification for this server.
	Insecure bool `yaml:"insecure_skip_tls_verify,omitempty"`
}

// Defaults are settings that apply to every context.
type Defaults struct {
	Output  string `yaml:"output,omitempty"`
	Timeout string `yaml:"timeout,omitempty"`
}

// File is the content of ~/.semaphore/client.yml.
type File struct {
	CurrentContext string             `yaml:"current_context,omitempty"`
	Contexts       map[string]Context `yaml:"contexts,omitempty"`
	Defaults       Defaults           `yaml:"defaults,omitempty"`
}

// ContextNames returns the context names sorted, for messages and listings.
func (f *File) ContextNames() []string {
	names := make([]string, 0, len(f.Contexts))
	for name := range f.Contexts {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// SetContext adds or replaces a context.
func (f *File) SetContext(name string, ctx Context) {
	if f.Contexts == nil {
		f.Contexts = map[string]Context{}
	}
	f.Contexts[name] = ctx
}

// Path returns the configuration file path: EnvConfig when set, otherwise
// ~/.semaphore/client.yml.
func Path(getenv func(string) string) (string, error) {
	if p := strings.TrimSpace(getenv(EnvConfig)); p != "" {
		return expandHome(p, getenv), nil
	}
	home := getenv("HOME")
	if home == "" {
		var err error
		if home, err = os.UserHomeDir(); err != nil {
			return "", fmt.Errorf("cannot determine the home directory: %w", err)
		}
	}
	return filepath.Join(home, ".semaphore", "client.yml"), nil
}

// expandHome replaces a leading "~" with the home directory.
func expandHome(p string, getenv func(string) string) string {
	if p == "~" || strings.HasPrefix(p, "~/") {
		home := getenv("HOME")
		if home == "" {
			home, _ = os.UserHomeDir()
		}
		return filepath.Join(home, strings.TrimPrefix(p, "~"))
	}
	return p
}

// ErrInsecurePermissions is returned by Load when the file can be written by
// the group or by others: such a file could feed the client a foreign server
// URL and send it the token.
var ErrInsecurePermissions = errors.New("configuration file is writable by other users")

// Load reads the configuration file. A missing file is not an error: an empty
// File is returned so that the environment alone can configure the client.
// warnings receives human-readable notes (file readable by others).
func Load(path string) (*File, []string, error) {
	var warnings []string

	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return &File{}, nil, nil
	}
	if err != nil {
		return nil, nil, fmt.Errorf("read %s: %w", path, err)
	}
	if info.IsDir() {
		return nil, nil, fmt.Errorf("%s is a directory, expected a file", path)
	}

	mode := info.Mode().Perm()
	if mode&0o022 != 0 {
		return nil, nil, fmt.Errorf("%w: %s has mode %04o; fix with: chmod 600 %s", ErrInsecurePermissions, path, mode, path)
	}
	if mode&0o044 != 0 {
		warnings = append(warnings, fmt.Sprintf("%s has mode %04o and can be read by other users; consider: chmod 600 %s", path, mode, path))
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, warnings, fmt.Errorf("read %s: %w", path, err)
	}

	var f File
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, warnings, fmt.Errorf("parse %s: %w", path, err)
	}
	if f.Contexts == nil {
		f.Contexts = map[string]Context{}
	}
	return &f, warnings, nil
}

// Save writes the file atomically with mode 0600 inside a 0700 directory: the
// content is written to a temporary file next to the target and renamed over
// it, so a crash never leaves a half-written or world-readable token behind.
func Save(path string, f *File) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create %s: %w", dir, err)
	}

	data, err := yaml.Marshal(f)
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".client.yml.*")
	if err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	tmpName := tmp.Name()
	cleanup := func() { _ = tmp.Close(); _ = os.Remove(tmpName) }

	if err := tmp.Chmod(0o600); err != nil {
		cleanup()
		return fmt.Errorf("write %s: %w", path, err)
	}
	if _, err := tmp.Write(data); err != nil {
		cleanup()
		return fmt.Errorf("write %s: %w", path, err)
	}
	if err := tmp.Sync(); err != nil {
		cleanup()
		return fmt.Errorf("write %s: %w", path, err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("write %s: %w", path, err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
