package server

import (
	"github.com/semaphoreui/semaphore/db"
)

type AwsSmStorageTokenDeserializer interface {
	DeserializeSecret(key *db.AccessKey) error
}

type AwsSmAccessKeyDeserializer struct {
}

func NewAwsSmAccessKeyDeserializer(
	_ db.AccessKeyManager,
	_ db.SecretStorageRepository,
	_ AwsSmStorageTokenDeserializer,
) *AwsSmAccessKeyDeserializer {
	return &AwsSmAccessKeyDeserializer{}
}

func (d *AwsSmAccessKeyDeserializer) DeleteSecret(_ *db.AccessKey) error {
	return nil
}

func (d *AwsSmAccessKeyDeserializer) SerializeSecret(_ *db.AccessKey) error {
	return nil
}

func (d *AwsSmAccessKeyDeserializer) DeserializeSecret(_ *db.AccessKey) (res string, err error) {
	return
}
