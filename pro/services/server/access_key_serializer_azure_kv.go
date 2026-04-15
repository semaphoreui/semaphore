package server

import (
	"github.com/semaphoreui/semaphore/db"
)

type AzureKvStorageTokenDeserializer interface {
	DeserializeSecret(key *db.AccessKey) error
}

type AzureKvAccessKeyDeserializer struct {
}

func NewAzureKvAccessKeyDeserializer(
	_ db.AccessKeyManager,
	_ db.SecretStorageRepository,
	_ AzureKvStorageTokenDeserializer,
) *AzureKvAccessKeyDeserializer {
	return &AzureKvAccessKeyDeserializer{}
}

func (d *AzureKvAccessKeyDeserializer) DeleteSecret(_ *db.AccessKey) error {
	return nil
}

func (d *AzureKvAccessKeyDeserializer) SerializeSecret(_ *db.AccessKey) error {
	return nil
}

func (d *AzureKvAccessKeyDeserializer) DeserializeSecret(_ *db.AccessKey) (res string, err error) {
	return
}
