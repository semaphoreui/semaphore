package util

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"
)

// keyIDLen is how many bytes of the SHA-256 fingerprint form the key id. 8 bytes
// (64 bits) is ample: the birthday bound is ~2^32 keys for a 50% collision, and an
// install has a handful of keys ever.
const keyIDLen = 8

// keyID returns a stable, content-addressed identifier for a base64-encoded key:
// base64url(sha256(rawKeyBytes))[:keyIDLen]. The id is intrinsic to the key
// material, so the id<->key binding cannot be repointed at a different key (a
// changed key yields a new id). Empty or invalid material yields "" — meaning "no
// id": encryption is disabled / the value is stored as passthrough with no prefix.
func keyID(material string) string {
	if material == "" {
		return ""
	}
	raw, err := base64.StdEncoding.DecodeString(material)
	if err != nil || len(raw) == 0 {
		return ""
	}
	sum := sha256.Sum256(raw)
	return base64.RawURLEncoding.EncodeToString(sum[:keyIDLen])
}

// encodeEnvelope prefixes a base64 ciphertext with its key id: "<id>:<b64ct>".
// An empty id returns the ciphertext unchanged (legacy / passthrough, no prefix).
func encodeEnvelope(id, b64ct string) string {
	if id == "" {
		return b64ct
	}
	return id + ":" + b64ct
}

// parseEnvelope splits a stored secret into its key id and base64 ciphertext.
// A value with no "<id>:" prefix is legacy (hasID == false) and the whole string
// is the ciphertext. ':' never appears in standard base64 (A-Za-z0-9+/=) nor in a
// base64url id (A-Za-z0-9-_), so the first ':' is an unambiguous separator; the
// isKeyID guard ignores any stray colon in legacy data.
func parseEnvelope(s string) (id, b64ct string, hasID bool) {
	i := strings.IndexByte(s, ':')
	if i <= 0 {
		return "", s, false
	}
	if prefix := s[:i]; isKeyID(prefix) {
		return prefix, s[i+1:], true
	}
	return "", s, false
}

// isKeyID reports whether s is shaped like a key id: exactly the length keyID
// emits and solely base64url characters. The exact-length check keeps a stray
// colon in legacy data (e.g. a PEM "Proc-Type:" line) from being mistaken for an
// id prefix.
func isKeyID(s string) bool {
	if len(s) != base64.RawURLEncoding.EncodedLen(keyIDLen) {
		return false
	}
	for _, c := range s {
		switch {
		case c >= 'A' && c <= 'Z', c >= 'a' && c <= 'z', c >= '0' && c <= '9', c == '-', c == '_':
		default:
			return false
		}
	}
	return true
}
