// Package passwordhash hashes and verifies user login passwords.
//
// New hashes are produced with Argon2id and stored in the PHC string
// format ($argon2id$v=19$m=...,t=...,p=...$<salt>$<hash>). Verify also
// accepts legacy bcrypt hashes ($2a$ / $2b$ / $2y$) so existing users
// keep working across the upgrade. When a legacy hash (or an Argon2id
// hash produced with weaker parameters than the current defaults) is
// verified successfully, Verify reports needsRehash=true so the caller
// can transparently re-hash with Hash on the next login.
package password_hash

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

const (
	argon2idMemoryKiB = 19456
	argon2idTime      = 2
	argon2idThreads   = 1
	argon2idSaltLen   = 16
	argon2idKeyLen    = 32
)

var (
	ErrUnknownHashFormat = errors.New("password_hash: unknown hash format")
	ErrMalformedHash     = errors.New("password_hash: malformed hash")
)

// Hash returns an Argon2id PHC-encoded hash of password.
func Hash(password string) (string, error) {
	salt := make([]byte, argon2idSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	key := argon2.IDKey(
		[]byte(password),
		salt,
		argon2idTime,
		argon2idMemoryKiB,
		argon2idThreads,
		argon2idKeyLen,
	)

	b64 := base64.RawStdEncoding
	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argon2idMemoryKiB,
		argon2idTime,
		argon2idThreads,
		b64.EncodeToString(salt),
		b64.EncodeToString(key),
	), nil
}

// Verify checks password against hash. ok is true on a match. needsRehash
// is true when the match succeeded against a legacy or sub-default hash
// and the caller should re-hash with Hash. A non-nil err signals a
// malformed stored hash; callers should treat err the same as a failed
// match (ok=false) at the auth boundary.
func Verify(password, hash string) (ok bool, needsRehash bool, err error) {
	switch {
	case strings.HasPrefix(hash, "$argon2id$"):
		return verifyArgon2id(password, hash)
	case strings.HasPrefix(hash, "$2a$"),
		strings.HasPrefix(hash, "$2b$"),
		strings.HasPrefix(hash, "$2y$"):
		bErr := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
		if bErr != nil {
			return false, false, nil
		}
		return true, true, nil
	default:
		return false, false, ErrUnknownHashFormat
	}
}

func verifyArgon2id(password, hash string) (bool, bool, error) {
	parts := strings.Split(hash, "$")
	// "", "argon2id", "v=19", "m=...,t=...,p=...", "<salt>", "<key>"
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" {
		return false, false, ErrMalformedHash
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, false, ErrMalformedHash
	}
	if version != argon2.Version {
		return false, false, ErrMalformedHash
	}

	var memory uint32
	var time uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads); err != nil {
		return false, false, ErrMalformedHash
	}

	b64 := base64.RawStdEncoding
	salt, err := b64.DecodeString(parts[4])
	if err != nil {
		return false, false, ErrMalformedHash
	}
	expected, err := b64.DecodeString(parts[5])
	if err != nil {
		return false, false, ErrMalformedHash
	}

	derived := argon2.IDKey(
		[]byte(password),
		salt,
		time,
		memory,
		threads,
		uint32(len(expected)),
	)

	if subtle.ConstantTimeCompare(derived, expected) != 1 {
		return false, false, nil
	}

	needsRehash := memory != argon2idMemoryKiB ||
		time != argon2idTime ||
		threads != argon2idThreads ||
		len(expected) != argon2idKeyLen
	return true, needsRehash, nil
}
