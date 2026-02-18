package server

import (
	"github.com/semaphoreui/semaphore/db"
)

type DvlsStorageTokenDeserializer interface {
	DeserializeSecret(key *db.AccessKey) error
}

type DvlsAccessKeyDeserializer struct {
}

func NewDvlsAccessKeyDeserializer(
	_ db.AccessKeyManager,
	_ db.SecretStorageRepository,
	_ VaultStorageTokenDeserializer,
) *DvlsAccessKeyDeserializer {
	return &DvlsAccessKeyDeserializer{}
}

func (d *DvlsAccessKeyDeserializer) DeleteSecret(key *db.AccessKey) error {
	return nil
}

func (d *DvlsAccessKeyDeserializer) SerializeSecret(key *db.AccessKey) (err error) {
	return
}

func (d *DvlsAccessKeyDeserializer) DeserializeSecret(key *db.AccessKey) (res string, err error) {
	return
}
