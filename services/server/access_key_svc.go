package server

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/semaphoreui/semaphore/util"
)

type AccessKeyService interface {
	Update(key db.AccessKey) error
	Create(key db.AccessKey) (newKey db.AccessKey, err error)
	GetAll(projectID int, options db.GetAccessKeyOptions, params db.RetrieveQueryParams) ([]db.AccessKey, error)
	Delete(projectID int, keyID int) (err error)
}

type AccessKeyServiceImpl struct {
	accessKeyRepo     db.AccessKeyManager
	encryptionService AccessKeyEncryptionService
	secretStorageRepo db.SecretStorageRepository
}

func NewAccessKeyService(
	accessKeyRepo db.AccessKeyManager,
	encryptionService AccessKeyEncryptionService,
	secretStorageRepo db.SecretStorageRepository,
) AccessKeyService {
	return &AccessKeyServiceImpl{
		accessKeyRepo:     accessKeyRepo,
		encryptionService: encryptionService,
		secretStorageRepo: secretStorageRepo,
	}
}

func (s *AccessKeyServiceImpl) Delete(projectID int, keyID int) (err error) {
	key, err := s.accessKeyRepo.GetAccessKey(projectID, keyID)
	if err != nil {
		return
	}

	if key.SourceStorageID != nil {
		var storage db.SecretStorage
		storage, err = s.secretStorageRepo.GetSecretStorage(projectID, *key.SourceStorageID)
		if err != nil {
			return
		}

		if storage.ReadOnly || key.Synchronized {
			// Do nothing

			//if key.Synchronized {
			//	err = common_errors.NewUserErrorS("cannot delete synchronized secret from read-only storage")
			//}
		} else {
			err = s.encryptionService.DeleteSecret(&key)
		}

		if err != nil {
			return
		}
	}

	err = s.accessKeyRepo.DeleteAccessKey(projectID, keyID)

	return
}

func (s *AccessKeyServiceImpl) GetAll(projectID int, options db.GetAccessKeyOptions, params db.RetrieveQueryParams) ([]db.AccessKey, error) {
	return s.accessKeyRepo.GetAccessKeys(projectID, options, params)
}

// generateSSHKeyPair returns a new private key in PEM format and the matching
// public key in OpenSSH authorized_keys format.
func generateSSHKeyPair() (privateKey string, publicKey string, err error) {
	var buf bytes.Buffer

	publicKey, err = util.GeneratePrivateKey(&buf)
	if err != nil {
		return
	}

	privateKey = buf.String()
	return
}

// encodePublicKeyPlain builds the JSON document stored in the non-secret
// "plain" column so the UI can show the public key.
func encodePublicKeyPlain(publicKey string) (string, error) {
	doc := struct {
		PublicKey string `json:"public_key"`
	}{PublicKey: publicKey}

	b, err := json.Marshal(doc)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// assignGeneratedSSHKey replaces the key's secret with a freshly generated
// SSH key pair and exposes the public half through the plain field.
func assignGeneratedSSHKey(key *db.AccessKey) error {
	if key.Type != db.AccessKeySSH {
		return common_errors.NewUserErrorS("generate_ssh_key is only allowed for ssh keys")
	}

	privateKey, publicKey, err := generateSSHKeyPair()
	if err != nil {
		return err
	}

	plain, err := encodePublicKeyPlain(publicKey)
	if err != nil {
		return err
	}

	key.SshKey.PrivateKey = privateKey
	key.SshKey.Passphrase = ""
	key.Plain = &plain
	key.IgnorePlain = false

	return nil
}

func (s *AccessKeyServiceImpl) Create(key db.AccessKey) (newKey db.AccessKey, err error) {
	// Plain is derived data, never taken from the caller.
	key.Plain = nil

	if key.GenerateSSHKey {
		err = assignGeneratedSSHKey(&key)
		if err != nil {
			return
		}
	}

	// SerializeSecret encrypts/persists the secret for writable backends. For read-only
	// external storage the secret is not stored in Semaphore, so SerializeSecret fails
	// with ErrReadOnlyStorage; we still create the access key row (metadata / reference).
	// A generated key is the exception: nobody else holds the private half, so
	// a storage that cannot persist it must reject the request.
	err = s.encryptionService.SerializeSecret(&key)
	if err != nil && (key.GenerateSSHKey || !errors.Is(err, ErrReadOnlyStorage)) {
		return
	}

	newKey, err = s.accessKeyRepo.CreateAccessKey(key)
	return
}

func (s *AccessKeyServiceImpl) Update(key db.AccessKey) (err error) {
	// Plain is derived data, never taken from the caller.
	key.Plain = nil

	if !key.OverrideSecret {
		err = s.accessKeyRepo.UpdateAccessKey(key)
		return
	}

	if key.GenerateSSHKey && key.IsNativelyReadOnly() {
		// Env/file sources are never written, so a generated private key
		// would have nowhere to live. Read-only vaults are rejected later
		// by SerializeSecret.
		err = common_errors.NewUserError(ErrReadOnlyStorage)
		return
	}

	var oldKey db.AccessKey
	oldKey, err = s.accessKeyRepo.GetAccessKey(*key.ProjectID, key.ID)
	if err != nil {
		return
	}

	if oldKey.SourceStorageType != nil && !oldKey.IsNativelyReadOnly() {
		// validate if it is secure to override secret storage

		var oldSt db.SecretStorage
		oldSt, err = s.secretStorageRepo.GetSecretStorage(*key.ProjectID, *oldKey.SourceStorageID)
		if err != nil {
			return
		}

		if !oldSt.ReadOnly && (key.SourceStorageID == nil || *oldKey.SourceStorageID != *key.SourceStorageID) {
			err = common_errors.NewUserErrorS("cannot override secret storage")
			return
		}
	}

	if key.GenerateSSHKey {
		err = assignGeneratedSSHKey(&key)
		if err != nil {
			return
		}
	}

	if !key.IsNativelyReadOnly() {
		err = s.encryptionService.SerializeSecret(&key)
		if err != nil {
			return
		}
	}

	err = s.accessKeyRepo.UpdateAccessKey(key)

	return
}
