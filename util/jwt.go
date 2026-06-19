package util

import (
	"fmt"
	"os"
	"time"

	"github.com/semaphoreui/semaphore/pkg/jwt"
)

// OptionStore is the minimal interface required to load and persist the JWT
// signing key. db.Store satisfies this interface; using an interface here
// avoids an import cycle between util and db.
type OptionStore interface {
	GetOption(key string) (string, error)
	SetOption(key string, value string) error
}

// jwtSigningKeyOption is the database option key under which the AES-GCM
// encrypted ECDSA P-256 private key PEM is stored.
const jwtSigningKeyOption = "jwt_signing_key"

// InitJWTSignerFromStore initialises the global JWT signer.
// It must be called once after the db.Store has been opened and after ConfigInit has run.
func InitJWTSignerFromStore(store OptionStore) (singer jwt.Signer, err error) {
	if !Config.JWT.Enabled {
		return
	}

	opts := jwtSignerOptions()

	pemBytes, err := loadOrCreateJWTKey(store)
	if err != nil {
		return nil, fmt.Errorf("jwt: could not load or create signing key: %w", err)
	}

	signer, err := jwt.NewECDSASignerFromPEM(pemBytes, opts)
	if err != nil {
		return nil, fmt.Errorf("jwt: failed to initialise signer: %w", err)
	}

	return signer, nil
}

// jwtSignerOptions builds SignerOptions from the current Config.
func jwtSignerOptions() jwt.SignerOptions {
	ttl := time.Hour
	if Config.JWT.DefaultTTL != "" {
		if parsed, err := time.ParseDuration(Config.JWT.DefaultTTL); err == nil {
			ttl = parsed
		} else {
			fmt.Fprintf(os.Stderr, "jwt: invalid jwt_default_ttl %q, falling back to 1h: %v\n", Config.JWT.DefaultTTL, err)
		}
	}

	maxTTL := 24 * time.Hour
	if Config.JWT.MaxTTL != "" {
		if parsed, err := time.ParseDuration(Config.JWT.MaxTTL); err == nil && parsed > 0 {
			maxTTL = parsed
		} else {
			fmt.Fprintf(os.Stderr, "jwt: invalid jwt_max_ttl %q, falling back to 24h: %v\n", Config.JWT.MaxTTL, err)
		}
	}

	return jwt.SignerOptions{
		Issuer:     Config.JWT.Issuer,
		DefaultTTL: ttl,
		MaxTTL:     maxTTL,
	}
}

// loadOrCreateJWTKey returns the raw PEM bytes of the JWT signing key. It
// reads the encrypted value from the database, decrypts it, and returns the
// plaintext PEM. If no key exists yet it generates one, persists it, and
// returns the plaintext PEM.
func loadOrCreateJWTKey(store OptionStore) ([]byte, error) {
	stored, err := store.GetOption(jwtSigningKeyOption)
	if err != nil {
		return nil, fmt.Errorf("read option: %w", err)
	}

	if stored != "" {
		return decryptJWTKey(stored)
	}

	// No key in DB yet
	pemBytes, err := jwt.GenerateKeyPEM()
	if err != nil {
		return nil, err
	}

	encrypted, err := encryptJWTKey(pemBytes)
	if err != nil {
		return nil, err
	}

	if err := store.SetOption(jwtSigningKeyOption, encrypted); err != nil {
		return nil, fmt.Errorf("persist signing key: %w", err)
	}

	return pemBytes, nil
}

// CheckJWTSigningKey reports which keyring slot decrypts the stored JWT signing
// key, for `vault check`. Returns ("", nil) when no key is stored.
func CheckJWTSigningKey(store OptionStore) (slot string, err error) {
	stored, err := store.GetOption(jwtSigningKeyOption)
	if err != nil {
		return "", fmt.Errorf("read jwt signing key option: %w", err)
	}
	if stored == "" {
		return "", nil
	}
	return Config.OptionSlot(stored), nil
}

// RekeyJWTSigningKey re-encrypts the stored JWT signing key under the current
// option keyring primary. It decrypts using the option keyring, the access
// keyring fallback, and — when supplied — oldKey (the legacy
// `vault rekey --old-key` flow). It is a no-op when no key is stored or when
// the ciphertext already matches (e.g. encryption disabled).
func RekeyJWTSigningKey(store OptionStore, oldKey string) error {
	stored, err := store.GetOption(jwtSigningKeyOption)
	if err != nil {
		return fmt.Errorf("read jwt signing key option: %w", err)
	}
	if stored == "" {
		return nil
	}

	keys := Config.OptionDecryptKeys()
	if oldKey != "" {
		keys = append(keys, oldKey)
	}

	pemBytes, err := decryptWithKeys(stored, keys)
	if err != nil {
		return fmt.Errorf("jwt: decrypt signing key for rekey: %w", err)
	}

	reEncrypted, err := Config.EncryptOption(pemBytes)
	if err != nil {
		return fmt.Errorf("jwt: re-encrypt signing key: %w", err)
	}

	if reEncrypted == stored {
		return nil
	}

	if err := store.SetOption(jwtSigningKeyOption, reEncrypted); err != nil {
		return fmt.Errorf("jwt: persist re-encrypted signing key: %w", err)
	}
	return nil
}

// encryptJWTKey encrypts pemBytes using the option keyring primary key (which
// falls back to the access key when no separate option key is configured).
func encryptJWTKey(pemBytes []byte) (string, error) {
	return Config.EncryptOption(pemBytes)
}

// decryptJWTKey reverses encryptJWTKey. It tries the option keyring and then
// the access keyring as a migration fallback, so a key written before the
// option/access split (encrypted with the access key) still loads.
func decryptJWTKey(stored string) ([]byte, error) {
	plaintext, err := Config.DecryptOption(stored)
	if err != nil {
		return nil, fmt.Errorf("jwt: decrypt signing key: %w", err)
	}
	return plaintext, nil
}
