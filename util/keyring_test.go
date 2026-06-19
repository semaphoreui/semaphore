package util

import (
	"bytes"
	"encoding/base64"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// genKey returns a valid base64-encoded 32-byte AES key whose bytes are all b.
func genKey(b byte) string {
	return base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{b}, 32))
}

// configWithKeyrings builds a Config whose runtime keyrings are set directly,
// bypassing ConfigInit. option may be nil (falls back to the access keyring).
func configWithKeyrings(access, option *runtimeKeyring) *ConfigType {
	c := &ConfigType{keys: &keyringStore{}}
	c.keys.access.Store(access)
	c.keys.option.Store(option)
	return c
}

func TestKeyring_AccessEncryptDecryptRoundTrip(t *testing.T) {
	keyA := genKey(0x01)
	Config = configWithKeyrings(&runtimeKeyring{primary: keyA}, nil)

	ciphertext, err := Config.EncryptAccessSecret([]byte("super-secret"))
	require.NoError(t, err)

	plaintext, err := decryptWithKeys(ciphertext, Config.AccessSecretDecryptKeys())
	require.NoError(t, err)
	assert.Equal(t, "super-secret", string(plaintext))
}

func TestKeyring_SecondaryDecryptsAfterRotation(t *testing.T) {
	keyOld := genKey(0x01)
	keyNew := genKey(0x02)

	// Encrypt under the old key.
	Config = configWithKeyrings(&runtimeKeyring{primary: keyOld}, nil)
	ciphertext, err := Config.EncryptAccessSecret([]byte("rotate-me"))
	require.NoError(t, err)

	// Rotate: new primary, old key retired to secondary.
	Config = configWithKeyrings(&runtimeKeyring{primary: keyNew, secondary: []string{keyOld}}, nil)

	// Old data still decrypts via the secondary...
	plaintext, err := decryptWithKeys(ciphertext, Config.AccessSecretDecryptKeys())
	require.NoError(t, err)
	assert.Equal(t, "rotate-me", string(plaintext))

	// ...and decrypting with the primary alone fails.
	_, err = DecryptAESGCM(ciphertext, Config.AccessSecretPrimaryKey())
	assert.Error(t, err)
}

func TestKeyring_PrimaryTriedBeforeSecondary(t *testing.T) {
	keyA := genKey(0x01)
	keyB := genKey(0x02)
	Config = configWithKeyrings(&runtimeKeyring{primary: keyA, secondary: []string{keyB}}, nil)

	keys := Config.AccessSecretDecryptKeys()
	require.Len(t, keys, 2)
	assert.Equal(t, keyA, keys[0])
	assert.Equal(t, keyB, keys[1])
}

func TestKeyring_EmptyKeyIsPassthrough(t *testing.T) {
	Config = configWithKeyrings(&runtimeKeyring{}, nil)

	ciphertext, err := Config.EncryptAccessSecret([]byte("plain"))
	require.NoError(t, err)
	// With no key, EncryptAESGCM returns base64 of the plaintext.
	assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("plain")), ciphertext)

	plaintext, err := decryptWithKeys(ciphertext, Config.AccessSecretDecryptKeys())
	require.NoError(t, err)
	assert.Equal(t, "plain", string(plaintext))
}

func TestKeyring_OptionFallsBackToAccessWhenUnset(t *testing.T) {
	keyA := genKey(0x01)
	Config = configWithKeyrings(&runtimeKeyring{primary: keyA}, nil)

	assert.False(t, Config.optionConfigured())
	// Option encryption uses the access key when no option key is configured.
	assert.Equal(t, keyA, Config.OptionPrimaryKey())

	ciphertext, err := Config.EncryptOption([]byte("jwt-pem"))
	require.NoError(t, err)

	plaintext, err := Config.DecryptOption(ciphertext)
	require.NoError(t, err)
	assert.Equal(t, "jwt-pem", string(plaintext))
}

func TestKeyring_OptionDecryptsAccessFallback(t *testing.T) {
	keyAccess := genKey(0x01)
	keyOption := genKey(0x02)

	// A JWT key written before the split — encrypted with the access key.
	Config = configWithKeyrings(&runtimeKeyring{primary: keyAccess}, nil)
	legacy, err := Config.EncryptOption([]byte("legacy-jwt"))
	require.NoError(t, err)

	// Now a separate option key is configured.
	Config = configWithKeyrings(&runtimeKeyring{primary: keyAccess}, &runtimeKeyring{primary: keyOption})

	// New encryption uses the option primary...
	fresh, err := Config.EncryptOption([]byte("fresh-jwt"))
	require.NoError(t, err)
	freshPlain, err := DecryptAESGCM(fresh, keyOption)
	require.NoError(t, err)
	assert.Equal(t, "fresh-jwt", string(freshPlain))

	// ...and the legacy key still decrypts via the access fallback.
	plaintext, err := Config.DecryptOption(legacy)
	require.NoError(t, err)
	assert.Equal(t, "legacy-jwt", string(plaintext))
}

func TestKeyring_OptionSlotLabels(t *testing.T) {
	keyAccess := genKey(0x01)
	keyOption := genKey(0x02)
	keyOptionOld := genKey(0x03)

	t.Run("no option key configured", func(t *testing.T) {
		Config = configWithKeyrings(&runtimeKeyring{primary: keyAccess}, nil)
		ct, err := Config.EncryptOption([]byte("x"))
		require.NoError(t, err)
		assert.Equal(t, "primary", Config.OptionSlot(ct))
	})

	t.Run("option configured, on option primary", func(t *testing.T) {
		Config = configWithKeyrings(
			&runtimeKeyring{primary: keyAccess},
			&runtimeKeyring{primary: keyOption, secondary: []string{keyOptionOld}},
		)
		ct, err := Config.EncryptOption([]byte("x"))
		require.NoError(t, err)
		assert.Equal(t, "option:primary", Config.OptionSlot(ct))
	})

	t.Run("option configured, on option secondary", func(t *testing.T) {
		// Encrypt with the old option key, then make it a secondary.
		Config = configWithKeyrings(&runtimeKeyring{primary: keyOptionOld}, nil)
		ct, err := Config.EncryptOption([]byte("x"))
		require.NoError(t, err)

		Config = configWithKeyrings(
			&runtimeKeyring{primary: keyAccess},
			&runtimeKeyring{primary: keyOption, secondary: []string{keyOptionOld}},
		)
		assert.Equal(t, "option:secondary[0]", Config.OptionSlot(ct))
	})

	t.Run("option configured, access fallback", func(t *testing.T) {
		Config = configWithKeyrings(&runtimeKeyring{primary: keyAccess}, nil)
		ct, err := Config.EncryptOption([]byte("x"))
		require.NoError(t, err)

		Config = configWithKeyrings(&runtimeKeyring{primary: keyAccess}, &runtimeKeyring{primary: keyOption})
		assert.Equal(t, "access-fallback (migrate)", Config.OptionSlot(ct))
	})
}

func TestResolveKeySource(t *testing.T) {
	t.Run("inline value", func(t *testing.T) {
		v, err := resolveKeySource(KeySource{Value: "abc"}, "k")
		require.NoError(t, err)
		assert.Equal(t, "abc", v)
	})

	t.Run("from file, trimmed", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "key")
		require.NoError(t, os.WriteFile(path, []byte("  filekey\n"), 0o600))

		v, err := resolveKeySource(KeySource{File: path}, "k")
		require.NoError(t, err)
		assert.Equal(t, "filekey", v)
	})

	t.Run("value and file are mutually exclusive", func(t *testing.T) {
		_, err := resolveKeySource(KeySource{Value: "a", File: "/tmp/x"}, "k")
		require.Error(t, err)
		assert.ErrorContains(t, err, "mutually exclusive")
	})

	t.Run("missing file errors", func(t *testing.T) {
		_, err := resolveKeySource(KeySource{File: "/no/such/file/here"}, "k")
		require.Error(t, err)
	})

	t.Run("empty source", func(t *testing.T) {
		v, err := resolveKeySource(KeySource{}, "k")
		require.NoError(t, err)
		assert.Empty(t, v)
	})
}

func TestResolveKeyring(t *testing.T) {
	keyA := genKey(0x01)
	keyB := genKey(0x02)

	t.Run("structured primary wins over flat", func(t *testing.T) {
		// Structured wins even when the flat fallback differs — this is what
		// makes hot rotation possible.
		rk, err := resolveKeyring(&Keyring{Primary: KeySource{Value: keyB}}, keyA, "k")
		require.NoError(t, err)
		assert.Equal(t, keyB, rk.primary)
	})

	t.Run("flat used when structured empty", func(t *testing.T) {
		rk, err := resolveKeyring(nil, keyA, "k")
		require.NoError(t, err)
		assert.Equal(t, keyA, rk.primary)
	})

	t.Run("secondaries resolved and empty ones dropped", func(t *testing.T) {
		rk, err := resolveKeyring(&Keyring{
			Primary:   KeySource{Value: keyA},
			Secondary: []KeySource{{Value: keyB}, {Value: ""}},
		}, "", "k")
		require.NoError(t, err)
		assert.Equal(t, keyA, rk.primary)
		assert.Equal(t, []string{keyB}, rk.secondary)
	})
}

func TestResolveEncryptionKeys(t *testing.T) {
	keyAccess := genKey(0x01)

	t.Run("no keys file, legacy flat access key, option falls back", func(t *testing.T) {
		Config = &ConfigType{AccessKeyEncryption: keyAccess}
		resolveEncryptionKeys()
		assert.Equal(t, keyAccess, Config.accessRing().primary)
		assert.False(t, Config.optionConfigured())
		assert.Equal(t, keyAccess, Config.OptionPrimaryKey())
	})

	t.Run("invalid keys file panics at boot", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "keys.json")
		require.NoError(t, os.WriteFile(path, []byte(`{"access_key":{"primary":{"value":"not-base64!!!"}}}`), 0o600))
		Config = &ConfigType{EncryptionKeysFile: path}
		assert.Panics(t, func() { resolveEncryptionKeys() })
	})
}

func TestResolveEncryptionKeysFrom(t *testing.T) {
	keyAccess := genKey(0x01)
	keyOption := genKey(0x02)

	t.Run("nil config, flat access only, option nil", func(t *testing.T) {
		access, option, err := resolveEncryptionKeysFrom(nil, keyAccess, "")
		require.NoError(t, err)
		assert.Equal(t, keyAccess, access.primary)
		assert.Nil(t, option)
	})

	t.Run("flat option key, old single-key scheme, no rotation", func(t *testing.T) {
		access, option, err := resolveEncryptionKeysFrom(nil, keyAccess, keyOption)
		require.NoError(t, err)
		assert.Equal(t, keyAccess, access.primary)
		require.NotNil(t, option)
		assert.Equal(t, keyOption, option.primary)
		assert.Empty(t, option.secondary)
	})

	t.Run("structured access from file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "access.key")
		require.NoError(t, os.WriteFile(path, []byte(keyAccess+"\n"), 0o600))

		access, _, err := resolveEncryptionKeysFrom(&EncryptionKeysConfig{
			AccessKey: &Keyring{Primary: KeySource{File: path}},
		}, "", "")
		require.NoError(t, err)
		assert.Equal(t, keyAccess, access.primary)
	})

	t.Run("structured option key wins over flat OptionEncryption", func(t *testing.T) {
		keyOption2 := genKey(0x03)
		_, option, err := resolveEncryptionKeysFrom(&EncryptionKeysConfig{
			OptionKey: &Keyring{Primary: KeySource{Value: keyOption2}},
		}, keyAccess, keyOption)
		require.NoError(t, err)
		require.NotNil(t, option)
		assert.Equal(t, keyOption2, option.primary)
	})

	t.Run("option key from file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "option.key")
		require.NoError(t, os.WriteFile(path, []byte(keyOption+"\n"), 0o600))

		_, option, err := resolveEncryptionKeysFrom(&EncryptionKeysConfig{
			OptionKey: &Keyring{Primary: KeySource{File: path}},
		}, keyAccess, "")
		require.NoError(t, err)
		require.NotNil(t, option)
		assert.Equal(t, keyOption, option.primary)
	})

	t.Run("mutually exclusive value+file errors", func(t *testing.T) {
		_, _, err := resolveEncryptionKeysFrom(&EncryptionKeysConfig{
			AccessKey: &Keyring{Primary: KeySource{Value: keyAccess, File: "/tmp/x"}},
		}, "", "")
		require.Error(t, err)
	})

	t.Run("invalid key errors", func(t *testing.T) {
		_, _, err := resolveEncryptionKeysFrom(&EncryptionKeysConfig{
			AccessKey: &Keyring{Primary: KeySource{Value: "not-base64!!!"}},
		}, "", "")
		require.Error(t, err)
	})
}

func TestOptionEncryptionFlatKey(t *testing.T) {
	keyAccess := genKey(0x01)
	keyOption := genKey(0x02)

	// Old single-key scheme for options: a distinct flat key, no keyring file.
	Config = &ConfigType{
		AccessKeyEncryption: keyAccess,
		OptionEncryption:    keyOption,
	}
	resolveEncryptionKeys()

	require.True(t, Config.optionConfigured())
	assert.Equal(t, keyOption, Config.OptionPrimaryKey())

	ct, err := Config.EncryptOption([]byte("jwt-pem"))
	require.NoError(t, err)

	// Encrypted under the option key, not the access key.
	pt, err := DecryptAESGCM(ct, keyOption)
	require.NoError(t, err)
	assert.Equal(t, "jwt-pem", string(pt))

	got, err := Config.DecryptOption(ct)
	require.NoError(t, err)
	assert.Equal(t, "jwt-pem", string(got))

	assert.Equal(t, "option:primary", Config.OptionSlot(ct))
}

func TestReloadEncryptionKeys(t *testing.T) {
	keyA := genKey(0x01)
	keyB := genKey(0x02)

	dir := t.TempDir()
	keysPath := filepath.Join(dir, "encryption_keys.json")
	write := func(body string) {
		require.NoError(t, os.WriteFile(keysPath, []byte(body), 0o600))
	}

	// Initial keys file: access primary keyA.
	write(`{"access_key":{"primary":{"value":"` + keyA + `"}}}`)
	Config = &ConfigType{EncryptionKeysFile: keysPath}
	resolveEncryptionKeys()
	require.Equal(t, keyA, Config.accessRing().primary)

	// Encrypt some data under keyA.
	oldCiphertext, err := Config.EncryptAccessSecret([]byte("rotate-me"))
	require.NoError(t, err)

	// Rotate via the keys file: new primary keyB, old keyA moved to secondary.
	write(`{"access_key":{"primary":{"value":"` + keyB + `"},"secondary":[{"value":"` + keyA + `"}]}}`)
	require.NoError(t, ReloadEncryptionKeys())

	// New primary took effect without a restart.
	assert.Equal(t, keyB, Config.accessRing().primary)
	newCiphertext, err := Config.EncryptAccessSecret([]byte("fresh"))
	require.NoError(t, err)
	fresh, err := DecryptAESGCM(newCiphertext, keyB)
	require.NoError(t, err)
	assert.Equal(t, "fresh", string(fresh))

	// Data encrypted under keyA still decrypts via the secondary.
	plaintext, err := decryptWithKeys(oldCiphertext, Config.AccessSecretDecryptKeys())
	require.NoError(t, err)
	assert.Equal(t, "rotate-me", string(plaintext))

	// An invalid reload is rejected and leaves the active keyring untouched.
	write(`{"access_key":{"primary":{"value":"not-base64!!!"}}}`)
	require.Error(t, ReloadEncryptionKeys())
	assert.Equal(t, keyB, Config.accessRing().primary)
}

func TestEncryptionKeysFile_DedicatedFileRotation(t *testing.T) {
	keyA := genKey(0x01)
	keyB := genKey(0x02)

	dir := t.TempDir()
	keysPath := filepath.Join(dir, "encryption_keys.json")
	write := func(body string) {
		require.NoError(t, os.WriteFile(keysPath, []byte(body), 0o600))
	}

	// Startup: the keyring comes from the dedicated file, not the main config.
	write(`{"access_key":{"primary":{"value":"` + keyA + `"}}}`)
	Config = &ConfigType{EncryptionKeysFile: keysPath}
	resolveEncryptionKeys()
	require.Equal(t, keyA, Config.accessRing().primary)

	old, err := Config.EncryptAccessSecret([]byte("rotate-me"))
	require.NoError(t, err)

	// No change yet — the watcher's check is a no-op.
	changed, err := ReloadEncryptionKeysIfChanged()
	require.NoError(t, err)
	assert.False(t, changed)

	// Rotate by editing the dedicated file: new primary keyB, old keyA secondary.
	write(`{"access_key":{"primary":{"value":"` + keyB + `"},"secondary":[{"value":"` + keyA + `"}]}}`)
	changed, err = ReloadEncryptionKeysIfChanged()
	require.NoError(t, err)
	assert.True(t, changed)
	assert.Equal(t, keyB, Config.accessRing().primary)

	// A subsequent check with no further edits is a no-op again.
	changed, err = ReloadEncryptionKeysIfChanged()
	require.NoError(t, err)
	assert.False(t, changed)

	// Data encrypted under keyA still decrypts via the secondary.
	pt, err := decryptWithKeys(old, Config.AccessSecretDecryptKeys())
	require.NoError(t, err)
	assert.Equal(t, "rotate-me", string(pt))
}

func TestEncryptionKeysFile_ReferencedKeyFilesYAML(t *testing.T) {
	keyAccess := genKey(0x01)
	keyOption := genKey(0x02)

	dir := t.TempDir()
	accessKeyPath := filepath.Join(dir, "access.key")
	optionKeyPath := filepath.Join(dir, "option.key")
	require.NoError(t, os.WriteFile(accessKeyPath, []byte(keyAccess), 0o600))
	require.NoError(t, os.WriteFile(optionKeyPath, []byte(keyOption), 0o600))

	keysPath := filepath.Join(dir, "keys.yaml")
	require.NoError(t, os.WriteFile(keysPath, []byte(
		"access_key:\n  primary:\n    file: "+accessKeyPath+"\n"+
			"option_key:\n  primary:\n    file: "+optionKeyPath+"\n"), 0o600))

	Config = &ConfigType{EncryptionKeysFile: keysPath}
	resolveEncryptionKeys()
	assert.Equal(t, keyAccess, Config.accessRing().primary)
	require.True(t, Config.optionConfigured())
	assert.Equal(t, keyOption, Config.OptionPrimaryKey())

	// Changing a referenced key file's *content* (structure unchanged) is also
	// detected and applied without a restart.
	keyAccess2 := genKey(0x03)
	require.NoError(t, os.WriteFile(accessKeyPath, []byte(keyAccess2), 0o600))
	changed, err := ReloadEncryptionKeysIfChanged()
	require.NoError(t, err)
	assert.True(t, changed)
	assert.Equal(t, keyAccess2, Config.accessRing().primary)
}

// TestReloadEncryptionKeys_ConcurrentReadsAreRaceFree hot-rotates the keyring
// while many goroutines encrypt and decrypt — run with -race, this is the core
// safety guarantee behind rotation without a restart. Both configs carry the
// same two keys (primary/secondary swapped), so every ciphertext stays
// decryptable through the rotation.
func TestReloadEncryptionKeys_ConcurrentReadsAreRaceFree(t *testing.T) {
	keyA := genKey(0x01)
	keyB := genKey(0x02)

	dir := t.TempDir()
	keysPath := filepath.Join(dir, "encryption_keys.json")
	cfgA := `{"access_key":{"primary":{"value":"` + keyA + `"},"secondary":[{"value":"` + keyB + `"}]}}`
	cfgB := `{"access_key":{"primary":{"value":"` + keyB + `"},"secondary":[{"value":"` + keyA + `"}]}}`

	require.NoError(t, os.WriteFile(keysPath, []byte(cfgA), 0o600))
	Config = &ConfigType{EncryptionKeysFile: keysPath}
	resolveEncryptionKeys()

	var wg sync.WaitGroup
	stop := make(chan struct{})

	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				ct, err := Config.EncryptAccessSecret([]byte("data"))
				if err != nil {
					t.Errorf("encrypt: %v", err)
					return
				}
				if _, err := decryptWithKeys(ct, Config.AccessSecretDecryptKeys()); err != nil {
					t.Errorf("decrypt: %v", err)
					return
				}
			}
		}()
	}

	for i := 0; i < 50; i++ {
		body := cfgA
		if i%2 == 1 {
			body = cfgB
		}
		require.NoError(t, os.WriteFile(keysPath, []byte(body), 0o600))
		require.NoError(t, ReloadEncryptionKeys())
	}

	close(stop)
	wg.Wait()
}
