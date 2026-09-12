package util

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
)

// EncryptAESGCM encrypts a plaintext using AES-256-GCM with the given base64-encoded key. If the key is empty, it returns the plaintext as base64.
func EncryptAESGCM(plaintext []byte, encodedKey string) (string, error) {
	if encodedKey == "" {
		return base64.StdEncoding.EncodeToString(plaintext), nil
	}

	gcm, err := newGCM(encodedKey)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(gcm.Seal(nonce, nonce, plaintext, nil)), nil
}

// DecryptAESGCM decrypts an AES-256-GCM ciphertext.
func DecryptAESGCM(encodedCiphertext, encodedKey string) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encodedCiphertext)
	if err != nil {
		return nil, fmt.Errorf("base64 decode: %w", err)
	}

	if encodedKey == "" {
		return ciphertext, nil
	}

	gcm, err := newGCM(encodedKey)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, payload := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, payload, nil)
}

func newGCM(encodedKey string) (cipher.AEAD, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(encodedKey)
	if err != nil {
		return nil, fmt.Errorf("decode encryption key: %w", err)
	}
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func GeneratePrivateKey(privateKeyFile io.Writer) (publicKey string, err error) {
	// 1. Generate RSA Private Key (2048 bits)
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return
	}

	// 2. Encode the private key to PKCS#1 ASN.1 PEM
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPem := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	// 3. Write private key to file
	if err = pem.Encode(privateKeyFile, privateKeyPem); err != nil {
		return
	}

	publicKeyBytes := x509.MarshalPKCS1PublicKey(&privateKey.PublicKey)
	publicKeyPem := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	var b bytes.Buffer
	publicKeyFile := bufio.NewWriter(&b)

	if err = pem.Encode(publicKeyFile, publicKeyPem); err != nil {
		return
	}

	if err = publicKeyFile.Flush(); err != nil {
		return
	}

	publicKey = b.String()
	return
}
