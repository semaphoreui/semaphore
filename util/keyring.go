package util

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// runtimeKeyring holds resolved (post file-read) base64 keys. The primary key
// is used for all new encryption; secondary keys are tried, in order, only when
// decrypting. An empty primary with no secondaries means encryption is disabled
// (EncryptAESGCM/DecryptAESGCM treat an empty key as base64 passthrough).
//
// A runtimeKeyring is immutable once built; rotation swaps the whole value
// behind an atomic pointer (see keyringStore) so it can change without a
// restart and without locking the hot encryption/decryption paths.
type runtimeKeyring struct {
	primary   string
	secondary []string
}

// keyringStore holds the resolved keyrings behind atomic pointers so they can
// be hot-swapped during key rotation without restarting the server. Reads
// (encryption/decryption) are lock-free Loads; a reload Stores new values.
// reloadMu serializes reloads (SIGHUP and the file watcher) so their
// compare-and-swap does not interleave; it never touches the read path.
type keyringStore struct {
	access   atomic.Pointer[runtimeKeyring]
	option   atomic.Pointer[runtimeKeyring]
	reloadMu sync.Mutex
}

// keyringsEqual reports whether two keyrings carry the same key material.
func keyringsEqual(a, b *runtimeKeyring) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.primary != b.primary || len(a.secondary) != len(b.secondary) {
		return false
	}
	for i := range a.secondary {
		if a.secondary[i] != b.secondary[i] {
			return false
		}
	}
	return true
}

// candidates returns the decryption candidate keys, primary first.
func (k *runtimeKeyring) candidates() []string {
	out := make([]string, 0, 1+len(k.secondary))
	out = append(out, k.primary)
	out = append(out, k.secondary...)
	return out
}

// dedupeKeys returns keys with later duplicates removed, preserving order.
func dedupeKeys(keys []string) []string {
	seen := make(map[string]struct{}, len(keys))
	out := keys[:0]
	for _, k := range keys {
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, k)
	}
	return out
}

// accessRing returns the currently active access keyring. When ConfigInit has
// not run (e.g. a unit test that sets the flat AccessKeyEncryption field
// directly) it falls back to a keyring built from that flat field.
func (conf *ConfigType) accessRing() *runtimeKeyring {
	if conf.keys != nil {
		if rk := conf.keys.access.Load(); rk != nil {
			return rk
		}
	}
	return &runtimeKeyring{primary: conf.AccessKeyEncryption}
}

// optionRing returns the currently active option keyring. When no option key is
// configured it falls back to the access keyring, so the JWT signing key keeps
// using the access key exactly as before the option/access split.
func (conf *ConfigType) optionRing() *runtimeKeyring {
	if conf.keys != nil {
		if rk := conf.keys.option.Load(); rk != nil {
			return rk
		}
	}
	return conf.accessRing()
}

// EncryptAccessSecret encrypts plaintext with the access keyring primary key.
func (conf *ConfigType) EncryptAccessSecret(plaintext []byte) (string, error) {
	return EncryptAESGCM(plaintext, conf.accessRing().primary)
}

// AccessSecretPrimaryKey returns the base64 key used to encrypt access secrets.
func (conf *ConfigType) AccessSecretPrimaryKey() string {
	return conf.accessRing().primary
}

// AccessSecretDecryptKeys returns the candidate keys for decrypting an access
// secret: the primary first, then each retired secondary.
func (conf *ConfigType) AccessSecretDecryptKeys() []string {
	return conf.accessRing().candidates()
}

// EncryptOption encrypts plaintext with the option keyring primary key (which
// falls back to the access key when no option key is configured).
func (conf *ConfigType) EncryptOption(plaintext []byte) (string, error) {
	return EncryptAESGCM(plaintext, conf.optionRing().primary)
}

// OptionPrimaryKey returns the base64 key used to encrypt options.
func (conf *ConfigType) OptionPrimaryKey() string {
	return conf.optionRing().primary
}

// OptionOwnDecryptKeys returns the option keyring's own candidate keys, without
// the access keyring migration fallback.
func (conf *ConfigType) OptionOwnDecryptKeys() []string {
	return conf.optionRing().candidates()
}

// OptionDecryptKeys returns the candidate keys for decrypting an option: the
// option keyring's keys first, then the access keyring's keys as a migration
// fallback (a pre-split JWT key was encrypted with the access key). Duplicates
// are removed, so when the option keyring falls back to the access keyring this
// is just the access candidates.
func (conf *ConfigType) OptionDecryptKeys() []string {
	keys := append(conf.optionRing().candidates(), conf.accessRing().candidates()...)
	return dedupeKeys(keys)
}

// DecryptOption decrypts ciphertext, trying the option keyring then the access
// keyring fallback, returning the first success.
func (conf *ConfigType) DecryptOption(ciphertext string) ([]byte, error) {
	return decryptWithKeys(ciphertext, conf.OptionDecryptKeys())
}

// optionConfigured reports whether a separate option key is configured (as
// opposed to falling back to the access keyring).
func (conf *ConfigType) optionConfigured() bool {
	return conf.keys != nil && conf.keys.option.Load() != nil
}

// OptionSlot returns a human-readable label for which key decrypts ciphertext
// under the option keyring (with access fallback), for `vault check`:
// "option:primary", "option:secondary[i]", "access-fallback (migrate)" when a
// separate option key is configured, or "primary"/"secondary[i]" when it falls
// back to the access keyring, or "FAILED".
func (conf *ConfigType) OptionSlot(ciphertext string) string {
	if conf.optionConfigured() {
		for i, key := range conf.optionRing().candidates() {
			if _, err := DecryptAESGCM(ciphertext, key); err == nil {
				if i == 0 {
					return "option:primary"
				}
				return fmt.Sprintf("option:secondary[%d]", i-1)
			}
		}
		for _, key := range conf.accessRing().candidates() {
			if _, err := DecryptAESGCM(ciphertext, key); err == nil {
				return "access-fallback (migrate)"
			}
		}
		return "FAILED"
	}

	for i, key := range conf.accessRing().candidates() {
		if _, err := DecryptAESGCM(ciphertext, key); err == nil {
			if i == 0 {
				return "primary"
			}
			return fmt.Sprintf("secondary[%d]", i-1)
		}
	}
	return "FAILED"
}

// decryptWithKeys tries each key in order and returns the first successful
// decryption, or the last error if every candidate fails.
func decryptWithKeys(ciphertext string, keys []string) ([]byte, error) {
	if len(keys) == 0 {
		keys = []string{""}
	}
	var lastErr error
	for _, key := range keys {
		plaintext, err := DecryptAESGCM(ciphertext, key)
		if err == nil {
			return plaintext, nil
		}
		lastErr = err
	}
	return nil, lastErr
}
