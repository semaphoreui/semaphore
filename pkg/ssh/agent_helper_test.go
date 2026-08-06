package ssh

import (
	"os"
	"testing"
)

func TestGetPublicKeyBytes(t *testing.T) {
	keyAStr, err := os.ReadFile("/root/.ssh/id_ed25519_repo_a")
	if err != nil {
		t.Fatalf("Failed reading key A: %v", err)
	}

	pubBytes, err := GetPublicKeyBytes(string(keyAStr), "")
	if err != nil {
		t.Fatalf("GetPublicKeyBytes failed: %v", err)
	}

	if len(pubBytes) == 0 {
		t.Fatalf("Derived public key is empty")
	}

	t.Logf("SUCCESS: Derived Public Key Bytes: %s", string(pubBytes))
}
