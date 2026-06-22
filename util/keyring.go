package util

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// keyset is the resolved, immutable runtime key registry. byID maps a
// content-addressed key id (see keyID) to its base64 key material; accessID and
// optionID name the active (encrypting) key per purpose. legacyAccess/legacyOption
// are the flat fields, used only to decrypt un-prefixed (pre-key-id) ciphertext.
//
// A keyset is swapped atomically on reload (see keyringStore) so rotation needs no
// restart and the hot encryption/decryption paths never lock.
type keyset struct {
	byID         map[string]string // key id -> base64 material
	accessID     string            // active access key id ("" => encryption disabled)
	optionID     string            // active option key id ("" => falls back to access)
	legacyAccess string            // flat access_key_encryption (un-prefixed access secrets)
	legacyOption string            // flat option_encryption (un-prefixed option values)
}

// keyringStore holds the resolved keyset behind an atomic pointer so it can be
// hot-swapped during rotation without restarting. Reads (encryption/decryption)
// are lock-free Loads; reloadMu serializes reloads (SIGHUP + the watcher) so their
// compare-and-swap does not interleave; it never touches the read path.
type keyringStore struct {
	current  atomic.Pointer[keyset]
	reloadMu sync.Mutex
}

func keysetsEqual(a, b *keyset) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.accessID != b.accessID || a.optionID != b.optionID ||
		a.legacyAccess != b.legacyAccess || a.legacyOption != b.legacyOption ||
		len(a.byID) != len(b.byID) {
		return false
	}
	for id, mat := range a.byID {
		if b.byID[id] != mat {
			return false
		}
	}
	return true
}

// currentKeyset returns the active keyset, or a fallback built from the flat
// fields when ConfigInit has not run (e.g. a unit test that sets the flat field
// directly).
func (conf *ConfigType) currentKeyset() *keyset {
	if conf.keys != nil {
		if ks := conf.keys.current.Load(); ks != nil {
			return ks
		}
	}
	ks := &keyset{
		byID:         map[string]string{},
		legacyAccess: conf.AccessKeyEncryption,
		legacyOption: conf.OptionEncryption,
	}
	if conf.AccessKeyEncryption != "" {
		id := keyID(conf.AccessKeyEncryption)
		ks.byID[id] = conf.AccessKeyEncryption
		ks.accessID = id
	}
	if conf.OptionEncryption != "" {
		id := keyID(conf.OptionEncryption)
		ks.byID[id] = conf.OptionEncryption
		ks.optionID = id
	}
	return ks
}

// --- encryption (stamps the active key id) ---

// EncryptAccessSecret encrypts plaintext with the active access key and returns
// the "<id>:<b64ct>" envelope (no prefix when encryption is disabled).
func (conf *ConfigType) EncryptAccessSecret(plaintext []byte) (string, error) {
	ks := conf.currentKeyset()
	return ks.encrypt(plaintext, ks.accessID)
}

// EncryptOption encrypts plaintext with the active option key, falling back to the
// access key when no separate option key is configured.
func (conf *ConfigType) EncryptOption(plaintext []byte) (string, error) {
	ks := conf.currentKeyset()
	id := ks.optionID
	if id == "" {
		id = ks.accessID
	}
	return ks.encrypt(plaintext, id)
}

func (k *keyset) encrypt(plaintext []byte, id string) (string, error) {
	ct, err := EncryptAESGCM(plaintext, k.byID[id]) // byID[""] == "" => passthrough
	if err != nil {
		return "", err
	}
	return encodeEnvelope(id, ct), nil
}

// --- decryption (id lookup, or legacy no-prefix trial) ---

// DecryptAccessSecret decrypts a stored access secret: a direct key-id lookup when
// the value carries an "<id>:" prefix, else the legacy no-prefix trial path.
func (conf *ConfigType) DecryptAccessSecret(stored string) ([]byte, error) {
	ks := conf.currentKeyset()
	return ks.decrypt(stored, ks.legacyAccessCandidates())
}

// DecryptOption decrypts a stored option (JWT) value, with the legacy access
// fallback for values written before the option/access split.
func (conf *ConfigType) DecryptOption(stored string) ([]byte, error) {
	ks := conf.currentKeyset()
	return ks.decrypt(stored, ks.legacyOptionCandidates())
}

// DecryptAccessSecretWithKey decrypts a stored secret with a single explicit key,
// stripping any id prefix. Used by the rekey `--old-key` path.
func (conf *ConfigType) DecryptAccessSecretWithKey(stored, key string) ([]byte, error) {
	_, ct, _ := parseEnvelope(stored)
	return DecryptAESGCM(ct, key)
}

func (k *keyset) decrypt(stored string, legacy []string) ([]byte, error) {
	id, ct, hasID := parseEnvelope(stored)
	if hasID {
		material, ok := k.byID[id]
		if !ok {
			return nil, fmt.Errorf("encryption key id %q not found in keyset (the key encrypting this value is missing)", id)
		}
		return DecryptAESGCM(ct, material)
	}
	return decryptWithKeys(ct, legacy)
}

// legacyAccessCandidates returns the keys to trial-decrypt an un-prefixed access
// secret: the flat access key first, then every registry key. The empty
// (passthrough) key is excluded unless there are no real keys at all, so a real
// ciphertext is never "successfully" decrypted to garbage by the empty key.
func (k *keyset) legacyAccessCandidates() []string {
	return k.legacyCandidates(k.legacyAccess)
}

// legacyOptionCandidates is like legacyAccessCandidates but tries the flat option
// key, then the flat access key, then the registry.
func (k *keyset) legacyOptionCandidates() []string {
	return k.legacyCandidates(k.legacyOption, k.legacyAccess)
}

func (k *keyset) legacyCandidates(flats ...string) []string {
	fl := len(flats)
	by := len(k.byID)
	maxInt := int(^uint(0) >> 1)

	var out []string
	if by > maxInt-fl {
		out = make([]string, 0)
	} else {
		capOut := by + fl
		out = make([]string, 0, capOut)
	}

	for _, f := range flats {
		if f != "" {
			out = append(out, f)
		}
	}
	for _, m := range k.byID {
		if m != "" {
			out = append(out, m)
		}
	}
	out = dedupeKeys(out)
	if len(out) == 0 {
		return []string{""} // encryption genuinely disabled
	}
	return out
}

// --- diagnostics / accessors (used by vault check & rekey) ---

// ActiveAccessKeyID returns the id of the key that encrypts new access secrets.
func (conf *ConfigType) ActiveAccessKeyID() string { return conf.currentKeyset().accessID }

// ActiveOptionKeyID returns the id of the key that encrypts new options (falling
// back to the access key id when no option key is configured).
func (conf *ConfigType) ActiveOptionKeyID() string {
	ks := conf.currentKeyset()
	if ks.optionID != "" {
		return ks.optionID
	}
	return ks.accessID
}

// HasKeyID reports whether the keyset holds a key with the given id.
func (conf *ConfigType) HasKeyID(id string) bool {
	_, ok := conf.currentKeyset().byID[id]
	return ok
}

// KeyIDs returns the ids of every key in the active keyset (the registry).
func (conf *ConfigType) KeyIDs() []string {
	ks := conf.currentKeyset()
	out := make([]string, 0, len(ks.byID))
	for id := range ks.byID {
		out = append(out, id)
	}
	return out
}

// SecretKeyID returns the key id stamped on a stored value, or "" for a legacy
// (un-prefixed) value.
func SecretKeyID(stored string) string {
	id, _, hasID := parseEnvelope(stored)
	if !hasID {
		return ""
	}
	return id
}

// ClassifyAccessSecret / ClassifyOptionSecret return a `vault check` status label
// for a stored value: "active:<id>", "rekey pending:<id>", "legacy (no id)", or
// "MISSING KEY <id>".
func (conf *ConfigType) ClassifyAccessSecret(stored string) string {
	return conf.currentKeyset().classify(stored, conf.ActiveAccessKeyID())
}

func (conf *ConfigType) classifyOptionSecret(stored string) string {
	return conf.currentKeyset().classify(stored, conf.ActiveOptionKeyID())
}

func (k *keyset) classify(stored, activeID string) string {
	id, _, hasID := parseEnvelope(stored)
	if !hasID {
		return "legacy (no id)"
	}
	if _, ok := k.byID[id]; !ok {
		return "MISSING KEY " + id
	}
	if id == activeID {
		return "active:" + id
	}
	return "rekey pending:" + id
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
