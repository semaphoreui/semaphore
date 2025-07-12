//go:build !pro

package server

import (
	"github.com/semaphoreui/semaphore/db"
)

type VaultAccessKeyDeserializer struct {
}

func NewVaultAccessKeyDeserializer(
	accessKeyRepo db.AccessKeyManager,
	secretStorageRepo db.SecretStorageRepository,
	encryptionService AccessKeyEncryptionService,
) *VaultAccessKeyDeserializer {
	return &VaultAccessKeyDeserializer{}
}

func (d *VaultAccessKeyDeserializer) DeleteSecret(key *db.AccessKey) error {
	return nil
}

func (d *VaultAccessKeyDeserializer) SerializeSecret(key *db.AccessKey) (err error) {
	return nil
}

func (d *VaultAccessKeyDeserializer) DeserializeSecret(key *db.AccessKey) (res string, err error) {
	return
}
