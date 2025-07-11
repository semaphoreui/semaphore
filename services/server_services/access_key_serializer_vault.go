package server_services

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/hashicorp/vault-client-go/schema"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/conv"

	"github.com/hashicorp/vault-client-go"
)

type VaultAccessKeyDeserializer struct {
	accessKeyRepo     db.AccessKeyManager
	secretStorageRepo db.SecretStorageRepository
	encryptionService AccessKeyEncryptionService
}

func NewVaultAccessKeyDeserializer(
	accessKeyRepo db.AccessKeyManager,
	secretStorageRepo db.SecretStorageRepository,
	encryptionService AccessKeyEncryptionService,
) *VaultAccessKeyDeserializer {
	return &VaultAccessKeyDeserializer{
		accessKeyRepo:     accessKeyRepo,
		secretStorageRepo: secretStorageRepo,
		encryptionService: encryptionService,
	}
}

func (d *VaultAccessKeyDeserializer) DeleteSecret(key *db.AccessKey) error {

	client, err := d.getClient(key)
	if err != nil {
		return err
	}

	ctx := context.TODO()

	_, err = client.Secrets.KvV2Delete(ctx, *key.SourceStorageKey, vault.WithMountPath("secret"))
	if err != nil {
		return err
	}

	return nil
}

func (d *VaultAccessKeyDeserializer) SerializeSecret(key *db.AccessKey) (err error) {

	client, err := d.getClient(key)
	if err != nil {
		return
	}

	ctx := context.TODO()

	var data map[string]any

	switch key.Type {
	case db.AccessKeyString:
		data = map[string]any{
			"string": key.String,
		}
	case db.AccessKeySSH:
		data = conv.StructToFlatMap(key.SshKey)
	case db.AccessKeyLoginPassword:
		data = conv.StructToFlatMap(key.LoginPassword)
	default:
		err = errors.New("unknown access key type: " + string(key.Type))
		return
	}

	_, err = client.Secrets.KvV2Write(
		ctx,
		*key.SourceStorageKey,
		schema.KvV2WriteRequest{
			Data: data,
		},
		vault.WithMountPath("secret"),
	)
	if err != nil {
		return
	}
	return nil
}

func (d *VaultAccessKeyDeserializer) getClient(key *db.AccessKey) (client *vault.Client, err error) {

	if key.SourceStorageID == nil || key.SourceStorageKey == nil {
		err = db.ErrNotFound
		return
	}

	storage, err := d.secretStorageRepo.GetSecretStorage(*key.ProjectID, *key.SourceStorageID)

	if err != nil {
		return
	}

	if storage.Type != db.SecretStorageTypeVault {
		err = db.ErrNotFound
	}

	keys, err := d.accessKeyRepo.GetAccessKeys(*key.ProjectID, db.GetAccessKeyOptions{
		Owner:     db.AccessKeyVault,
		StorageID: key.SourceStorageID,
	}, db.RetrieveQueryParams{})

	if err != nil {
		return
	}

	if len(keys) == 0 {
		err = db.ErrNotFound
		return
	}

	tokenKey := keys[0]

	err = d.encryptionService.DeserializeSecret(&tokenKey)
	if err != nil {
		return
	}

	var vaultParams db.VaultSecretStorageParams
	if err = storage.ExtractParams(&vaultParams); err != nil {
		return
	}

	client, err = vault.New(
		vault.WithAddress(vaultParams.URL),
		vault.WithRequestTimeout(30*time.Second),
	)

	if err != nil {
		return
	}

	if err = client.SetToken(tokenKey.String); err != nil {
		return
	}

	return
}

func (d *VaultAccessKeyDeserializer) DeserializeSecret(key *db.AccessKey) (res string, err error) {

	client, err := d.getClient(key)
	if err != nil {
		return
	}

	ctx := context.TODO()

	s, err := client.Secrets.KvV2Read(ctx, *key.SourceStorageKey, vault.WithMountPath("secret"))
	if err != nil {
		return
	}

	if key.Type == db.AccessKeyString {
		res = s.Data.Data["string"].(string)
		return
	}

	bytes, err := json.Marshal(s.Data.Data)
	if err != nil {
		return
	}

	res = string(bytes)
	return
}
