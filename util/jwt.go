package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
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
// encrypted RSA private key PEM is stored.
const jwtSigningKeyOption = "system.jwt_signing_key"

// InitJWTSignerFromStore initialises the process-wide JWT signer using the
// RSA private key stored (encrypted) in the database. It must be called once
// after the db.Store has been opened and after ConfigInit has run.
func InitJWTSignerFromStore(store OptionStore) error {
	if !Config.JWTEnabled {
		jwt.SetDefault(nil)
		return nil
	}

	opts := jwtSignerOptions()

	pemBytes, err := loadOrCreateJWTKey(store)
	if err != nil {
		return fmt.Errorf("jwt: could not load or create signing key: %w", err)
	}

	signer, err := jwt.NewRSASignerFromPEM(pemBytes, opts)
	if err != nil {
		return fmt.Errorf("jwt: failed to initialise signer: %w", err)
	}

	jwt.SetDefault(signer)
	return nil
}

// jwtSignerOptions builds SignerOptions from the current Config.
func jwtSignerOptions() jwt.SignerOptions {
	ttl := time.Hour
	if Config.JWTTTL != "" {
		if parsed, err := time.ParseDuration(Config.JWTTTL); err == nil {
			ttl = parsed
		} else {
			fmt.Fprintf(os.Stderr, "jwt: invalid jwt_ttl %q, falling back to 1h: %v\n", Config.JWTTTL, err)
		}
	}

	issuer := Config.JWTIssuer
	if issuer == "" {
		if Config.WebHost != "" {
			issuer = Config.WebHost
		} else {
			issuer = "semaphore"
		}
	}

	audience := Config.JWTAudience
	if audience == "" {
		audience = "semaphore"
	}

	return jwt.SignerOptions{
		Issuer:   issuer,
		Audience: audience,
		TTL:      ttl,
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

	// No key in DB yet – generate, encrypt and persist.
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

// encryptJWTKey encrypts pemBytes using AES-256-GCM with the configured
// AccessKeyEncryption key and returns a base64-encoded ciphertext identical in
// format to the one used for access keys. When AccessKeyEncryption is not set
// the plaintext PEM is stored as plain base64 (same fallback as access keys).
func encryptJWTKey(pemBytes []byte) (string, error) {
	encryptionKey := Config.AccessKeyEncryption

	if encryptionKey == "" {
		return base64.StdEncoding.EncodeToString(pemBytes), nil
	}

	keyBytes, err := base64.StdEncoding.DecodeString(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("decode encryption key: %w", err)
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, pemBytes, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptJWTKey reverses encryptJWTKey.
func decryptJWTKey(stored string) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(stored)
	if err != nil {
		return nil, fmt.Errorf("base64 decode stored key: %w", err)
	}

	encryptionKey := Config.AccessKeyEncryption

	if encryptionKey == "" {
		// Stored as plain base64.
		return ciphertext, nil
	}

	keyBytes, err := base64.StdEncoding.DecodeString(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("decode encryption key: %w", err)
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("jwt: stored key is too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("jwt: decrypt signing key: %w", err)
	}

	return plaintext, nil
}
