package ssh

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"testing"

	golang_ssh "golang.org/x/crypto/ssh"
)

func TestGetPublicKeyBytes(t *testing.T) {
	// Generate a dynamic ED25519 private key in memory for clean unit testing
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	pemBlock, err := golang_ssh.MarshalPrivateKey(priv, "")
	if err != nil {
		t.Fatalf("Failed to marshal private key: %v", err)
	}

	pemBytes := pem.EncodeToMemory(pemBlock)

	pubBytes, err := GetPublicKeyBytes(string(pemBytes), "")
	if err != nil {
		t.Fatalf("GetPublicKeyBytes failed: %v", err)
	}

	if len(pubBytes) == 0 {
		t.Fatalf("Derived public key is empty")
	}

	t.Logf("SUCCESS: Derived Public Key Bytes: %s", string(pubBytes))
}
