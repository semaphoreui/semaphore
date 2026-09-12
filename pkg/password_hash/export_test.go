package password_hash

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/argon2"
)

// encodeArgon2idForTest builds an Argon2id PHC string with arbitrary
// parameters. Exposed to the test package so we can construct hashes
// with sub-default parameters to exercise the needsRehash branch.
func encodeArgon2idForTest(password string, memory, time uint32, threads uint8, keyLen uint32) (string, error) {
	salt := make([]byte, argon2idSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(password), salt, time, memory, threads, keyLen)
	b64 := base64.RawStdEncoding
	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, memory, time, threads,
		b64.EncodeToString(salt), b64.EncodeToString(key),
	), nil
}
