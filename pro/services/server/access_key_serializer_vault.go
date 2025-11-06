package server

import (
	"github.com/semaphoreui/semaphore/db"
)

type VaultStorageTokenDeserializer interface {
	DeserializeSecret(key *db.AccessKey) error
}

type VaultAccessKeyDeserializer struct {
}

func NewVaultAccessKeyDeserializer(
	_ db.AccessKeyManager,
	_ db.SecretStorageRepository,
	_ VaultStorageTokenDeserializer,
) *VaultAccessKeyDeserializer {
	return &VaultAccessKeyDeserializer{}
}

func (d *VaultAccessKeyDeserializer) DeleteSecret(key *db.AccessKey) error {
	return nil
}

func (d *VaultAccessKeyDeserializer) SerializeSecret(key *db.AccessKey) (err error) {
	return
}

func (d *VaultAccessKeyDeserializer) DeserializeSecret(key *db.AccessKey) (res string, err error) {
	return
}
