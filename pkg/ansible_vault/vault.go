// Package ansible_vault reads and writes files in the Ansible Vault format
// (version 1.1, cipher AES256) with the standard library only, so a secret
// can be handed to ansible-playbook as a vault-encrypted file without the
// ansible-vault binary.
//
// The format, as implemented by ansible-core (lib/ansible/parsing/vault,
// VaultAES256): a random 32-byte salt; PBKDF2-HMAC-SHA256 over the password
// with 10000 iterations yields 80 bytes — a 32-byte AES key, a 32-byte HMAC
// key and a 16-byte counter IV; the plaintext is padded with PKCS7 to the AES
// block size (CTR needs no padding, Ansible pads anyway and unpads on read);
// AES-256-CTR; HMAC-SHA256 of the ciphertext. The body is
// hex(hex(salt) "\n" hex(hmac) "\n" hex(ciphertext)) — hex twice, kept for
// compatibility — wrapped at 80 columns under the header line.
package ansible_vault

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

const (
	header     = "$ANSIBLE_VAULT"
	version    = "1.1"
	cipherName = "AES256"

	saltSize   = 32
	keySize    = 32
	iterations = 10000
	lineWidth  = 80
)

var (
	// ErrFormat is returned when the text is not a vault file this package reads.
	ErrFormat = errors.New("ansible vault: not a $ANSIBLE_VAULT;1.1/1.2;AES256 file")
	// ErrAuth is returned when the HMAC does not match: wrong password or
	// tampered ciphertext — the two cannot be told apart.
	ErrAuth = errors.New("ansible vault: HMAC verification failed")
)

// Encrypt returns the content of a vault file holding plaintext, encrypted
// with password. The header carries no vault id, so Ansible tries every
// --vault-id it was given on it.
func Encrypt(plaintext []byte, password string) ([]byte, error) {
	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("ansible vault: salt: %w", err)
	}
	return encryptWithSalt(plaintext, password, salt)
}

func encryptWithSalt(plaintext []byte, password string, salt []byte) ([]byte, error) {
	key1, key2, iv, err := deriveKeys(password, salt)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key1)
	if err != nil {
		return nil, fmt.Errorf("ansible vault: %w", err)
	}
	padded := pkcs7Pad(plaintext, block.BlockSize())
	ciphertext := make([]byte, len(padded))
	cipher.NewCTR(block, iv).XORKeyStream(ciphertext, padded)

	mac := hmac.New(sha256.New, key2)
	mac.Write(ciphertext)

	inner := bytes.Join([][]byte{
		hexEncode(salt),
		hexEncode(mac.Sum(nil)),
		hexEncode(ciphertext),
	}, []byte("\n"))
	body := hexEncode(inner)

	var out bytes.Buffer
	out.Grow(len(header) + len(body) + len(body)/lineWidth + 16)
	out.WriteString(header + ";" + version + ";" + cipherName + "\n")
	for len(body) > lineWidth {
		out.Write(body[:lineWidth])
		out.WriteByte('\n')
		body = body[lineWidth:]
	}
	out.Write(body)
	out.WriteByte('\n')
	return out.Bytes(), nil
}

// Decrypt returns the plaintext of a vault file. It accepts the 1.1 and 1.2
// envelopes with the AES256 cipher, verifies the HMAC before decrypting and
// returns ErrAuth when it does not match.
func Decrypt(vaulttext []byte, password string) ([]byte, error) {
	lines := bytes.Split(bytes.TrimSpace(vaulttext), []byte("\n"))
	if len(lines) < 2 {
		return nil, ErrFormat
	}
	parts := bytes.Split(bytes.TrimSpace(lines[0]), []byte(";"))
	if len(parts) < 3 || len(parts) > 4 ||
		string(parts[0]) != header ||
		(string(parts[1]) != "1.1" && string(parts[1]) != "1.2") ||
		string(parts[2]) != cipherName {
		return nil, ErrFormat
	}

	var body []byte
	for _, line := range lines[1:] {
		body = append(body, bytes.TrimSpace(line)...)
	}
	inner, err := hexDecode(body)
	if err != nil {
		return nil, err
	}
	fields := bytes.SplitN(inner, []byte("\n"), 3)
	if len(fields) != 3 {
		return nil, ErrFormat
	}
	salt, err := hexDecode(fields[0])
	if err != nil {
		return nil, err
	}
	wantMAC, err := hexDecode(fields[1])
	if err != nil {
		return nil, err
	}
	ciphertext, err := hexDecode(fields[2])
	if err != nil {
		return nil, err
	}

	key1, key2, iv, err := deriveKeys(password, salt)
	if err != nil {
		return nil, err
	}
	mac := hmac.New(sha256.New, key2)
	mac.Write(ciphertext)
	if !hmac.Equal(mac.Sum(nil), wantMAC) {
		return nil, ErrAuth
	}

	block, err := aes.NewCipher(key1)
	if err != nil {
		return nil, fmt.Errorf("ansible vault: %w", err)
	}
	padded := make([]byte, len(ciphertext))
	cipher.NewCTR(block, iv).XORKeyStream(padded, ciphertext)
	return pkcs7Unpad(padded, block.BlockSize())
}

func deriveKeys(password string, salt []byte) (key1, key2, iv []byte, err error) {
	if len(salt) == 0 {
		return nil, nil, nil, ErrFormat
	}
	derived, err := pbkdf2.Key(sha256.New, password, salt, iterations, 2*keySize+aes.BlockSize)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("ansible vault: key derivation: %w", err)
	}
	return derived[:keySize], derived[keySize : 2*keySize], derived[2*keySize:], nil
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	n := blockSize - len(data)%blockSize
	return append(append([]byte{}, data...), bytes.Repeat([]byte{byte(n)}, n)...)
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 || len(data)%blockSize != 0 {
		return nil, ErrFormat
	}
	n := int(data[len(data)-1])
	if n == 0 || n > blockSize || n > len(data) {
		return nil, ErrFormat
	}
	for _, b := range data[len(data)-n:] {
		if int(b) != n {
			return nil, ErrFormat
		}
	}
	return data[:len(data)-n], nil
}

func hexEncode(b []byte) []byte {
	out := make([]byte, hex.EncodedLen(len(b)))
	hex.Encode(out, b)
	return out
}

func hexDecode(b []byte) ([]byte, error) {
	out := make([]byte, hex.DecodedLen(len(b)))
	n, err := hex.Decode(out, b)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFormat, err)
	}
	return out[:n], nil
}
