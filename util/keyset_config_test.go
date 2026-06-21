package util

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mkKeyFolder creates a temp dir with one file per (name -> content) entry.
func mkKeyFolder(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600))
	}
	return dir
}

// mkKeyFile writes a single key file (with a trailing newline, to exercise trimming).
func mkKeyFile(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "k")
	require.NoError(t, os.WriteFile(p, []byte(content+"\n"), 0o600))
	return p
}

// TestEncryptionKeyConfig_Variants exercises every valid way to configure the
// encryption keys, asserting the resolved active ids and a full encrypt/decrypt
// round-trip (access + option) for each.
func TestEncryptionKeyConfig_Variants(t *testing.T) {
	keyA, keyB, keyC := genKey(0x0A), genKey(0x0B), genKey(0x0C)

	cases := []struct {
		name         string
		setup        func(t *testing.T) (enc *EncryptionKeysConfig, flatAccess, flatOption string)
		accessActive string // expected active access material ("" = encryption disabled)
		optionActive string // expected active option material ("" = falls back to access)
	}{
		{
			name:  "no keys at all (disabled)",
			setup: func(t *testing.T) (*EncryptionKeysConfig, string, string) { return nil, "", "" },
		},
		{
			name:         "flat access only",
			setup:        func(t *testing.T) (*EncryptionKeysConfig, string, string) { return nil, keyA, "" },
			accessActive: keyA,
		},
		{
			name:         "flat access + flat option",
			setup:        func(t *testing.T) (*EncryptionKeysConfig, string, string) { return nil, keyA, keyB },
			accessActive: keyA,
			optionActive: keyB,
		},
		{
			name: "inline map, value, active access label",
			setup: func(t *testing.T) (*EncryptionKeysConfig, string, string) {
				return &EncryptionKeysConfig{Keys: map[string]KeySource{"a": {Value: keyA}}, Active: ActivePointers{SecretsKey: "a"}}, "", ""
			},
			accessActive: keyA,
		},
		{
			name: "inline map, separate access + option labels",
			setup: func(t *testing.T) (*EncryptionKeysConfig, string, string) {
				return &EncryptionKeysConfig{
					Keys:   map[string]KeySource{"a": {Value: keyA}, "b": {Value: keyB}},
					Active: ActivePointers{SecretsKey: "a", OptionsKey: "b"},
				}, "", ""
			},
			accessActive: keyA,
			optionActive: keyB,
		},
		{
			name: "inline map, KeySource from file",
			setup: func(t *testing.T) (*EncryptionKeysConfig, string, string) {
				return &EncryptionKeysConfig{Keys: map[string]KeySource{"a": {File: mkKeyFile(t, keyA)}}, Active: ActivePointers{SecretsKey: "a"}}, "", ""
			},
			accessActive: keyA,
		},
		{
			name: "keys_folder, active by *_file",
			setup: func(t *testing.T) (*EncryptionKeysConfig, string, string) {
				dir := mkKeyFolder(t, map[string]string{"acc.txt": keyA, "opt.txt": keyB})
				return &EncryptionKeysConfig{KeysFolder: dir, Active: ActivePointers{SecretsKeyFile: "acc.txt", OptionsKeyFile: "opt.txt"}}, "", ""
			},
			accessActive: keyA,
			optionActive: keyB,
		},
		{
			name: "keys_folder, active by label (filename)",
			setup: func(t *testing.T) (*EncryptionKeysConfig, string, string) {
				dir := mkKeyFolder(t, map[string]string{"acc.txt": keyA})
				return &EncryptionKeysConfig{KeysFolder: dir, Active: ActivePointers{SecretsKey: "acc.txt"}}, "", ""
			},
			accessActive: keyA,
		},
		{
			name: "inline map + keys_folder combined",
			setup: func(t *testing.T) (*EncryptionKeysConfig, string, string) {
				dir := mkKeyFolder(t, map[string]string{"opt.txt": keyB})
				return &EncryptionKeysConfig{
					Keys:       map[string]KeySource{"a": {Value: keyA}},
					KeysFolder: dir,
					Active:     ActivePointers{SecretsKey: "a", OptionsKeyFile: "opt.txt"},
				}, "", ""
			},
			accessActive: keyA,
			optionActive: keyB,
		},
		{
			name: "active label wins over flat field",
			setup: func(t *testing.T) (*EncryptionKeysConfig, string, string) {
				return &EncryptionKeysConfig{Keys: map[string]KeySource{"a": {Value: keyA}}, Active: ActivePointers{SecretsKey: "a"}}, keyC, ""
			},
			accessActive: keyA, // the active label wins over the flat keyC
		},
		{
			name: "active *_file absolute path, no keys_folder",
			setup: func(t *testing.T) (*EncryptionKeysConfig, string, string) {
				return &EncryptionKeysConfig{Active: ActivePointers{SecretsKeyFile: mkKeyFile(t, keyA)}}, "", ""
			},
			accessActive: keyA,
		},
		{
			name: "16/24-byte keys (AES-128/192) accepted",
			setup: func(t *testing.T) (*EncryptionKeysConfig, string, string) {
				k16 := base64.StdEncoding.EncodeToString(make([]byte, 16))
				return nil, k16, ""
			},
			accessActive: base64.StdEncoding.EncodeToString(make([]byte, 16)),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			enc, flatA, flatO := c.setup(t)
			ks, err := resolveEncryptionKeysFrom(enc, flatA, flatO)
			require.NoError(t, err)
			Config = &ConfigType{keys: &keyringStore{}, AccessKeyEncryption: flatA, OptionEncryption: flatO}
			Config.keys.current.Store(ks)

			assert.Equal(t, keyID(c.accessActive), Config.ActiveAccessKeyID(), "active access id")

			effOption := c.optionActive
			if effOption == "" {
				effOption = c.accessActive // option falls back to access
			}
			assert.Equal(t, keyID(effOption), Config.ActiveOptionKeyID(), "active option id")

			// Access round-trip; id prefix present only when a key is active.
			ct, err := Config.EncryptAccessSecret([]byte("secret-data"))
			require.NoError(t, err)
			if c.accessActive != "" {
				assert.True(t, strings.HasPrefix(ct, keyID(c.accessActive)+":"), "access ciphertext stamped with active id")
			} else {
				assert.NotContains(t, ct, ":", "disabled encryption: no id prefix")
			}
			pt, err := Config.DecryptAccessSecret(ct)
			require.NoError(t, err)
			assert.Equal(t, "secret-data", string(pt))

			// Option round-trip.
			oct, err := Config.EncryptOption([]byte("opt-data"))
			require.NoError(t, err)
			opt, err := Config.DecryptOption(oct)
			require.NoError(t, err)
			assert.Equal(t, "opt-data", string(opt))
		})
	}
}

// TestEncryptionKeyConfig_Errors covers every invalid configuration.
func TestEncryptionKeyConfig_Errors(t *testing.T) {
	keyA := genKey(0x0A)

	cases := []struct {
		name  string
		setup func(t *testing.T) (enc *EncryptionKeysConfig, flatAccess, flatOption string)
	}{
		{"value and file both set", func(t *testing.T) (*EncryptionKeysConfig, string, string) {
			return &EncryptionKeysConfig{Keys: map[string]KeySource{"a": {Value: keyA, File: "/tmp/x"}}}, "", ""
		}},
		{"active label not in keys", func(t *testing.T) (*EncryptionKeysConfig, string, string) {
			return &EncryptionKeysConfig{Keys: map[string]KeySource{"a": {Value: keyA}}, Active: ActivePointers{SecretsKey: "nope"}}, "", ""
		}},
		{"active option label not in keys", func(t *testing.T) (*EncryptionKeysConfig, string, string) {
			return &EncryptionKeysConfig{Keys: map[string]KeySource{"a": {Value: keyA}}, Active: ActivePointers{OptionsKey: "nope"}}, "", ""
		}},
		{"active file missing", func(t *testing.T) (*EncryptionKeysConfig, string, string) {
			return &EncryptionKeysConfig{Active: ActivePointers{SecretsKeyFile: "/no/such/file"}}, "", ""
		}},
		{"keys_folder missing", func(t *testing.T) (*EncryptionKeysConfig, string, string) {
			return &EncryptionKeysConfig{KeysFolder: "/no/such/folder"}, "", ""
		}},
		{"invalid base64 key in map", func(t *testing.T) (*EncryptionKeysConfig, string, string) {
			return &EncryptionKeysConfig{Keys: map[string]KeySource{"a": {Value: "not-base64!!!"}}}, "", ""
		}},
		{"wrong-length key (20 bytes)", func(t *testing.T) (*EncryptionKeysConfig, string, string) {
			bad := base64.StdEncoding.EncodeToString(make([]byte, 20))
			return &EncryptionKeysConfig{Keys: map[string]KeySource{"a": {Value: bad}}}, "", ""
		}},
		{"invalid key in keys_folder", func(t *testing.T) (*EncryptionKeysConfig, string, string) {
			dir := mkKeyFolder(t, map[string]string{"bad.txt": "garbage-not-base64!!!"})
			return &EncryptionKeysConfig{KeysFolder: dir}, "", ""
		}},
		{"invalid flat access key", func(t *testing.T) (*EncryptionKeysConfig, string, string) {
			return nil, "not-base64!!!", ""
		}},
		{"invalid flat option key", func(t *testing.T) (*EncryptionKeysConfig, string, string) {
			return nil, keyA, "not-base64!!!"
		}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			enc, a, o := c.setup(t)
			_, err := resolveEncryptionKeysFrom(enc, a, o)
			require.Error(t, err)
		})
	}
}

// TestEncryptionKeysFile_AllFormats exercises the full encryption.keys_file path
// (file read + parse + resolve) for every supported file shape and format.
func TestEncryptionKeysFile_AllFormats(t *testing.T) {
	keyA, keyB := genKey(0x0A), genKey(0x0B)

	cases := []struct {
		name  string
		build func(t *testing.T) string // writes the keys file (+ any supporting files), returns its path
	}{
		{"inline map, JSON", func(t *testing.T) string {
			p := filepath.Join(t.TempDir(), "keys.json")
			require.NoError(t, os.WriteFile(p, []byte(
				`{"keys":{"a":{"value":"`+keyA+`"},"b":{"value":"`+keyB+`"}},"active":{"secrets_key":"a","options_key":"b"}}`), 0o600))
			return p
		}},
		{"inline map, YAML", func(t *testing.T) string {
			p := filepath.Join(t.TempDir(), "keys.yaml")
			require.NoError(t, os.WriteFile(p, []byte(
				"keys:\n  a: {value: \""+keyA+"\"}\n  b: {value: \""+keyB+"\"}\nactive:\n  secrets_key: a\n  options_key: b\n"), 0o600))
			return p
		}},
		{"inline map with file refs", func(t *testing.T) string {
			dir := t.TempDir()
			ka := filepath.Join(dir, "a.key")
			require.NoError(t, os.WriteFile(ka, []byte(keyA), 0o600))
			p := filepath.Join(dir, "keys.yaml")
			require.NoError(t, os.WriteFile(p, []byte("keys:\n  a: {file: "+ka+"}\nactive:\n  secrets_key: a\n"), 0o600))
			return p
		}},
		{"keys_folder, no extension (k8s-style)", func(t *testing.T) string {
			dir := t.TempDir()
			folder := filepath.Join(dir, "kf")
			require.NoError(t, os.MkdirAll(folder, 0o700))
			require.NoError(t, os.WriteFile(filepath.Join(folder, "acc.txt"), []byte(keyA), 0o600))
			require.NoError(t, os.WriteFile(filepath.Join(folder, "opt.txt"), []byte(keyB), 0o600))
			p := filepath.Join(dir, "encryption_keys") // no .yaml/.yml extension
			require.NoError(t, os.WriteFile(p, []byte(
				"keys_folder: "+folder+"\nactive:\n  secrets_key_file: acc.txt\n  options_key_file: opt.txt\n"), 0o600))
			return p
		}},
		{"map + folder combined", func(t *testing.T) string {
			dir := t.TempDir()
			folder := filepath.Join(dir, "kf")
			require.NoError(t, os.MkdirAll(folder, 0o700))
			require.NoError(t, os.WriteFile(filepath.Join(folder, "opt.txt"), []byte(keyB), 0o600))
			p := filepath.Join(dir, "keys.yaml")
			require.NoError(t, os.WriteFile(p, []byte(
				"keys:\n  a: {value: \""+keyA+"\"}\nkeys_folder: "+folder+"\nactive:\n  secrets_key: a\n  options_key_file: opt.txt\n"), 0o600))
			return p
		}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			path := c.build(t)
			Config = &ConfigType{Encryption: &EncryptionConfig{KeysFile: path}}
			resolveEncryptionKeys()

			require.Equal(t, keyID(keyA), Config.ActiveAccessKeyID())

			ct, err := Config.EncryptAccessSecret([]byte("x"))
			require.NoError(t, err)
			assert.True(t, strings.HasPrefix(ct, keyID(keyA)+":"))
			pt, err := Config.DecryptAccessSecret(ct)
			require.NoError(t, err)
			assert.Equal(t, "x", string(pt))
		})
	}
}
