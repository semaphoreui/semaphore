package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func twoContexts() *File {
	return &File{
		Contexts: map[string]Context{
			DefaultContextName: {URL: "https://prod.example.com", Token: "prod-token\n", Project: "3"},
			"staging":          {URL: "https://staging.example.com", Token: "staging-token", CACert: "~/ca.pem", Insecure: true},
		},
		Defaults: Defaults{Output: "json", Timeout: "45"},
	}
}

func TestPath(t *testing.T) {
	p, err := Path(env(map[string]string{"HOME": "/home/u"}))
	require.NoError(t, err)
	assert.Equal(t, filepath.Join("/home/u", ".semaphore", "client.yml"), p)

	p, err = Path(env(map[string]string{"HOME": "/home/u", EnvConfig: "~/cfg/sem.yml"}))
	require.NoError(t, err)
	assert.Equal(t, filepath.Join("/home/u", "cfg", "sem.yml"), p)

	p, err = Path(env(map[string]string{EnvConfig: "/etc/semaphore/client.yml"}))
	require.NoError(t, err)
	assert.Equal(t, "/etc/semaphore/client.yml", p)
}

func TestResolve_Precedence(t *testing.T) {
	file := twoContexts()
	tests := []struct {
		name  string
		flags Flags
		env   map[string]string
		check func(t *testing.T, e Effective)
	}{
		{
			name: "file only: the context named default wins",
			check: func(t *testing.T, e Effective) {
				assert.Equal(t, DefaultContextName, e.ContextName)
				assert.Equal(t, SourceDefault, e.Sources.Context)
				assert.Equal(t, "https://prod.example.com", e.URL)
				assert.Equal(t, SourceContext, e.Sources.URL)
				assert.Equal(t, "prod-token", e.Token, "token is trimmed")
				assert.Equal(t, SourceContext, e.Sources.Token)
				assert.Equal(t, "3", e.Project)
				assert.Equal(t, OutputJSON, e.Output)
				assert.Equal(t, 45*time.Second, e.Timeout)
				assert.False(t, e.Insecure)
				assert.False(t, e.NoInput)
			},
		},
		{
			name: "env context overrides the default one",
			env:  map[string]string{EnvContext: "staging", "HOME": "/home/u"},
			check: func(t *testing.T, e Effective) {
				assert.Equal(t, "staging", e.ContextName)
				assert.Equal(t, SourceEnv, e.Sources.Context)
				assert.Equal(t, "https://staging.example.com", e.URL)
				assert.Equal(t, "staging-token", e.Token)
				assert.True(t, e.Insecure)
				assert.Equal(t, "/home/u/ca.pem", e.CACert, "~ expands")
				assert.Empty(t, e.Project)
			},
		},
		{
			name:  "flag context overrides env context",
			flags: Flags{Context: DefaultContextName},
			env:   map[string]string{EnvContext: "staging"},
			check: func(t *testing.T, e Effective) {
				assert.Equal(t, DefaultContextName, e.ContextName)
				assert.Equal(t, SourceFlag, e.Sources.Context)
			},
		},
		{
			name: "env URL overrides context URL, token still from context",
			env:  map[string]string{EnvURL: "https://env.example.com"},
			check: func(t *testing.T, e Effective) {
				assert.Equal(t, "https://env.example.com", e.URL)
				assert.Equal(t, SourceEnv, e.Sources.URL)
				assert.Equal(t, "prod-token", e.Token)
				assert.Equal(t, SourceContext, e.Sources.Token)
			},
		},
		{
			name:  "flag URL overrides env URL",
			flags: Flags{URL: "https://flag.example.com"},
			env:   map[string]string{EnvURL: "https://env.example.com"},
			check: func(t *testing.T, e Effective) {
				assert.Equal(t, "https://flag.example.com", e.URL)
				assert.Equal(t, SourceFlag, e.Sources.URL)
			},
		},
		{
			name: "env token with env URL",
			env:  map[string]string{EnvURL: "https://ci.example.com", EnvToken: " ci-token "},
			check: func(t *testing.T, e Effective) {
				assert.Equal(t, "ci-token", e.Token)
				assert.Equal(t, SourceEnv, e.Sources.Token)
			},
		},
		{
			name:  "env token with flag URL",
			flags: Flags{URL: "https://flag.example.com"},
			env:   map[string]string{EnvToken: "ci-token"},
			check: func(t *testing.T, e Effective) {
				assert.Equal(t, "ci-token", e.Token)
				assert.Equal(t, SourceEnv, e.Sources.Token)
			},
		},
		{
			name:  "project: flag > env > context",
			flags: Flags{Project: "flag-proj"},
			env:   map[string]string{EnvProject: "env-proj"},
			check: func(t *testing.T, e Effective) {
				assert.Equal(t, "flag-proj", e.Project)
				assert.Equal(t, SourceFlag, e.Sources.Project)
			},
		},
		{
			name: "project from env",
			env:  map[string]string{EnvProject: "env-proj"},
			check: func(t *testing.T, e Effective) {
				assert.Equal(t, "env-proj", e.Project)
				assert.Equal(t, SourceEnv, e.Sources.Project)
			},
		},
		{
			name:  "output: flag > env > defaults; case-insensitive",
			flags: Flags{Output: "YAML"},
			env:   map[string]string{EnvOutput: "table"},
			check: func(t *testing.T, e Effective) {
				assert.Equal(t, OutputYAML, e.Output)
				assert.Equal(t, SourceFlag, e.Sources.Output)
			},
		},
		{
			name: "output from env",
			env:  map[string]string{EnvOutput: "table"},
			check: func(t *testing.T, e Effective) {
				assert.Equal(t, OutputTable, e.Output)
				assert.Equal(t, SourceEnv, e.Sources.Output)
			},
		},
		{
			name:  "timeout: flag > env > defaults",
			flags: Flags{Timeout: 7 * time.Second},
			env:   map[string]string{EnvTimeout: "2m"},
			check: func(t *testing.T, e Effective) { assert.Equal(t, 7*time.Second, e.Timeout) },
		},
		{
			name:  "timeout from env as duration",
			env:   map[string]string{EnvTimeout: "2m"},
			check: func(t *testing.T, e Effective) { assert.Equal(t, 2*time.Minute, e.Timeout) },
		},
		{
			name: "insecure, no-input, debug from env",
			env:  map[string]string{EnvInsecure: "yes", EnvNoInput: "1", EnvDebug: "true"},
			check: func(t *testing.T, e Effective) {
				assert.True(t, e.Insecure)
				assert.True(t, e.NoInput)
				assert.True(t, e.Debug)
			},
		},
		{
			name:  "CI implies no-input",
			env:   map[string]string{EnvCI: "true"},
			check: func(t *testing.T, e Effective) { assert.True(t, e.NoInput) },
		},
		{
			name:  "flags switch bools on",
			flags: Flags{Insecure: true, NoInput: true, Debug: true},
			check: func(t *testing.T, e Effective) {
				assert.True(t, e.Insecure)
				assert.True(t, e.NoInput)
				assert.True(t, e.Debug)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.env == nil {
				tt.env = map[string]string{}
			}
			e, err := Resolve(tt.flags, env(tt.env), file)
			require.NoError(t, err)
			tt.check(t, e)
		})
	}
}

func TestResolve_NoFile(t *testing.T) {
	e, err := Resolve(Flags{}, env(map[string]string{EnvURL: "https://x", EnvToken: "t"}), nil)
	require.NoError(t, err)
	assert.Equal(t, DefaultContextName, e.ContextName)
	assert.Equal(t, SourceNone, e.Sources.Context, "no file, so nothing was taken from a context")
	assert.Equal(t, "https://x", e.URL)
	assert.Equal(t, "t", e.Token)
	assert.Equal(t, OutputTable, e.Output)
	assert.Equal(t, SourceDefault, e.Sources.Output)
	assert.Equal(t, DefaultTimeout, e.Timeout)

	e, err = Resolve(Flags{}, env(map[string]string{}), &File{})
	require.NoError(t, err)
	assert.Empty(t, e.URL)
	assert.ErrorIs(t, e.RequireURL(), ErrNoURL)
}

func TestResolve_EnvTokenRequiresEnvOrFlagURL(t *testing.T) {
	file := twoContexts()

	_, err := Resolve(Flags{}, env(map[string]string{EnvToken: "ci-token"}), file)
	assert.ErrorIs(t, err, ErrTokenWithoutURL)
	assert.ErrorContains(t, err, `context "default"`)

	_, err = Resolve(Flags{Context: "staging"}, env(map[string]string{EnvToken: "ci-token"}), file)
	assert.ErrorIs(t, err, ErrTokenWithoutURL, "a token from CI must not follow --context to another server")

	tokenFile := filepath.Join(t.TempDir(), "tok")
	require.NoError(t, os.WriteFile(tokenFile, []byte("file-token\n"), 0o600))
	_, err = Resolve(Flags{}, env(map[string]string{EnvTokenFile: tokenFile}), file)
	assert.ErrorIs(t, err, ErrTokenWithoutURL)

	e, err := Resolve(Flags{}, env(map[string]string{EnvURL: "https://ci", EnvTokenFile: tokenFile}), file)
	require.NoError(t, err)
	assert.Equal(t, "file-token", e.Token)
	assert.Equal(t, tokenFile, e.TokenFile)
	assert.Equal(t, SourceEnv, e.Sources.Token)
}

func TestResolve_TokenAndTokenFileConflict(t *testing.T) {
	_, err := Resolve(Flags{}, env(map[string]string{EnvURL: "https://ci", EnvToken: "a", EnvTokenFile: "/tmp/b"}), nil)
	assert.ErrorIs(t, err, ErrTokenAndTokenFile)
}

func TestResolve_ContextTokenFile(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "tok"), []byte("  ctx-file-token\n\n"), 0o600))
	file := &File{Contexts: map[string]Context{DefaultContextName: {URL: "https://a", TokenFile: "~/tok"}}}

	e, err := Resolve(Flags{}, env(map[string]string{"HOME": dir}), file)
	require.NoError(t, err)
	assert.Equal(t, "ctx-file-token", e.Token)
	assert.Equal(t, filepath.Join(dir, "tok"), e.TokenFile)
	assert.Equal(t, SourceContext, e.Sources.Token)

	// A broken token file is reported through TokenError, not as a failure of
	// Resolve: commands that need no token must still run.
	file.Contexts[DefaultContextName] = Context{URL: "https://a", TokenFile: filepath.Join(dir, "missing")}
	e, err = Resolve(Flags{}, env(map[string]string{}), file)
	require.NoError(t, err)
	assert.Empty(t, e.Token)
	assert.ErrorContains(t, e.TokenError, "read token file")

	require.NoError(t, os.WriteFile(filepath.Join(dir, "empty"), []byte("\n"), 0o600))
	file.Contexts[DefaultContextName] = Context{URL: "https://a", TokenFile: filepath.Join(dir, "empty")}
	e, err = Resolve(Flags{}, env(map[string]string{}), file)
	require.NoError(t, err)
	assert.ErrorContains(t, e.TokenError, "is empty")
}

func TestResolve_UnknownContext(t *testing.T) {
	_, err := Resolve(Flags{Context: "nope"}, env(map[string]string{}), twoContexts())
	assert.ErrorIs(t, err, ErrUnknownContext)
	assert.ErrorContains(t, err, "default, staging")

	_, err = Resolve(Flags{}, env(map[string]string{EnvContext: "gone"}), &File{})
	assert.ErrorIs(t, err, ErrUnknownContext)
	assert.ErrorContains(t, err, "(none)")

	// The implicit default is allowed to be absent: that is how a setup made
	// of environment variables alone looks.
	e, err := Resolve(Flags{}, env(map[string]string{EnvURL: "https://env", EnvToken: "t"}), &File{})
	require.NoError(t, err)
	assert.Equal(t, DefaultContextName, e.ContextName)
	assert.Equal(t, SourceNone, e.Sources.Context)
	assert.Equal(t, "https://env", e.URL)
}

func TestResolve_BadValues(t *testing.T) {
	_, err := Resolve(Flags{Output: "xml"}, env(map[string]string{}), nil)
	assert.ErrorIs(t, err, ErrBadOutput)

	_, err = Resolve(Flags{}, env(map[string]string{EnvTimeout: "soon"}), nil)
	assert.ErrorContains(t, err, EnvTimeout)

	_, err = Resolve(Flags{}, env(map[string]string{EnvTimeout: "0"}), nil)
	assert.ErrorContains(t, err, "positive")

	_, err = Resolve(Flags{}, env(map[string]string{}), &File{Defaults: Defaults{Timeout: "-5s"}})
	assert.ErrorContains(t, err, "defaults.timeout")
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		in   string
		want time.Duration
		ok   bool
	}{
		{"30", 30 * time.Second, true},
		{" 45 ", 45 * time.Second, true},
		{"2m", 2 * time.Minute, true},
		{"1h30m", 90 * time.Minute, true},
		{"0", 0, false},
		{"-1s", 0, false},
		{"abc", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			d, err := parseDuration(tt.in)
			if !tt.ok {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, d)
		})
	}
}

func TestTruthy(t *testing.T) {
	for _, s := range []string{"1", "true", "TRUE", "yes", "on", "Y", " true "} {
		assert.True(t, truthy(s), s)
	}
	for _, s := range []string{"", "0", "false", "no", "off", "maybe"} {
		assert.False(t, truthy(s), s)
	}
}

func TestLoad_MissingFileIsEmpty(t *testing.T) {
	f, warnings, err := Load(filepath.Join(t.TempDir(), "client.yml"))
	require.NoError(t, err)
	assert.Empty(t, warnings)
	assert.Equal(t, &File{}, f)
}

func TestLoad_ParsesYAML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "client.yml")
	require.NoError(t, os.WriteFile(path, []byte(`
contexts:
  prod:
    url: https://prod.example.com
    token: abc
    project: 3
  staging:
    url: https://staging.example.com
    token_file: ~/.semaphore/staging.token
    ca_cert: ~/.semaphore/staging-ca.pem
    insecure_skip_tls_verify: true
defaults:
  output: table
  timeout: 60
`), 0o600))

	f, warnings, err := Load(path)
	require.NoError(t, err)
	assert.Empty(t, warnings)
	assert.Equal(t, []string{"prod", "staging"}, f.ContextNames())
	assert.Equal(t, Context{URL: "https://prod.example.com", Token: "abc", Project: "3"}, f.Contexts["prod"])
	assert.Equal(t, Context{URL: "https://staging.example.com", TokenFile: "~/.semaphore/staging.token", CACert: "~/.semaphore/staging-ca.pem", Insecure: true}, f.Contexts["staging"])
	assert.Equal(t, Defaults{Output: "table", Timeout: "60"}, f.Defaults)
}

func TestLoad_FilePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permissions")
	}
	tests := []struct {
		name    string
		mode    os.FileMode
		wantErr error
		warns   bool
	}{
		{"0600 is fine", 0o600, nil, false},
		{"0400 is fine", 0o400, nil, false},
		{"0644 warns: readable by others", 0o644, nil, true},
		{"0640 warns: readable by group", 0o640, nil, true},
		{"0666 refused: writable by others", 0o666, ErrInsecurePermissions, false},
		{"0620 refused: writable by group", 0o620, ErrInsecurePermissions, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "client.yml")
			require.NoError(t, os.WriteFile(path, []byte("contexts: {}\n"), 0o600))
			require.NoError(t, os.Chmod(path, tt.mode))

			_, warnings, err := Load(path)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.ErrorContains(t, err, "chmod 600")
				return
			}
			require.NoError(t, err)
			if tt.warns {
				require.Len(t, warnings, 1)
				assert.Contains(t, warnings[0], "chmod 600")
			} else {
				assert.Empty(t, warnings)
			}
		})
	}
}

func TestLoad_Errors(t *testing.T) {
	dir := t.TempDir()
	_, _, err := Load(dir)
	assert.ErrorContains(t, err, "is a directory")

	path := filepath.Join(dir, "bad.yml")
	require.NoError(t, os.WriteFile(path, []byte("contexts: [not a map"), 0o600))
	_, _, err = Load(path)
	assert.ErrorContains(t, err, "parse")
}

func TestSave_CreatesDirAndFileWithTightPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permissions")
	}
	path := filepath.Join(t.TempDir(), "nested", ".semaphore", "client.yml")
	f := &File{}
	f.SetContext("prod", Context{URL: "https://prod.example.com", Token: "secret"})

	require.NoError(t, Save(path, f))

	dirInfo, err := os.Stat(filepath.Dir(path))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o700), dirInfo.Mode().Perm())

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	// No temporary file is left behind.
	entries, err := os.ReadDir(filepath.Dir(path))
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "client.yml", entries[0].Name())

	loaded, warnings, err := Load(path)
	require.NoError(t, err)
	assert.Empty(t, warnings)
	assert.Equal(t, f, loaded)

	// Overwriting keeps the mode even if someone loosened it meanwhile.
	require.NoError(t, os.Chmod(path, 0o644))
	f.SetContext("staging", Context{URL: "https://staging"})
	require.NoError(t, Save(path, f))
	info, err = os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
}

func TestSave_YAMLShape(t *testing.T) {
	path := filepath.Join(t.TempDir(), "client.yml")
	f := &File{Contexts: map[string]Context{"prod": {URL: "https://p", Token: "t"}}}
	require.NoError(t, Save(path, f))
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "contexts:\n    prod:\n        url: https://p\n        token: t\n", string(data))
}
