package server

import (
	"github.com/semaphoreui/semaphore/db"
)

type CyberArkStorageTokenDeserializer interface {
	DeserializeSecret(key *db.AccessKey) error
}

type CyberArkAccessKeyDeserializer struct {
}

func NewCyberArkAccessKeyDeserializer(
	_ db.AccessKeyManager,
	_ db.SecretStorageRepository,
	_ CyberArkStorageTokenDeserializer,
) *CyberArkAccessKeyDeserializer {
	return &CyberArkAccessKeyDeserializer{}
}

func (d *CyberArkAccessKeyDeserializer) DeleteSecret(_ *db.AccessKey) error {
	return nil
}

func (d *CyberArkAccessKeyDeserializer) SerializeSecret(_ *db.AccessKey) error {
	return nil
}

func (d *CyberArkAccessKeyDeserializer) DeserializeSecret(_ *db.AccessKey) (res string, err error) {
	return
}
