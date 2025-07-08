package server_services

import (
	"context"
	"encoding/json"
	"github.com/semaphoreui/semaphore/db"
	"time"

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

func (d *VaultAccessKeyDeserializer) DeserializeSecret(key *db.AccessKey) (res string, err error) {

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

	ctx := context.TODO()

	client, err := vault.New(
		vault.WithAddress(vaultParams.URL),
		vault.WithRequestTimeout(30*time.Second),
	)

	if err != nil {
		return
	}

	if err = client.SetToken(tokenKey.String); err != nil {
		return
	}

	s, err := client.Secrets.KvV2Read(ctx, *key.SourceStorageKey, vault.WithMountPath("secret"))
	if err != nil {
		return
	}

	bytes, err := json.Marshal(s.Data.Data)
	if err != nil {
		return
	}

	res = string(bytes)
	return
}
