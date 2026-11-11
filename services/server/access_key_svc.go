package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
)

type AccessKeyService interface {
	Update(key db.AccessKey) error
	Create(key db.AccessKey) (newKey db.AccessKey, err error)
	GetAll(projectID int, options db.GetAccessKeyOptions, params db.RetrieveQueryParams) ([]db.AccessKey, error)
	Delete(projectID int, keyID int) (err error)

	// DeleteWithSecret deletes the key a variable group owns together with the secret that is its only
	// link (secret.KeyID). The link is removed last, in one transaction with the key row: when the secret
	// can not be removed from its storage both stay, and the deletion can be repeated.
	DeleteWithSecret(secret db.EnvironmentSecret) (err error)
}

type AccessKeyServiceImpl struct {
	accessKeyRepo     db.AccessKeyManager
	encryptionService AccessKeyEncryptionService
	secretStorageRepo db.SecretStorageRepository
	hostConfigRepo    db.HostConfigManager
	deletionRepo      db.AccessKeyDeletionManager
}

func NewAccessKeyService(
	accessKeyRepo db.AccessKeyManager,
	encryptionService AccessKeyEncryptionService,
	secretStorageRepo db.SecretStorageRepository,
	hostConfigRepo db.HostConfigManager,
	deletionRepo db.AccessKeyDeletionManager,
) AccessKeyService {
	return &AccessKeyServiceImpl{
		accessKeyRepo:     accessKeyRepo,
		encryptionService: encryptionService,
		secretStorageRepo: secretStorageRepo,
		hostConfigRepo:    hostConfigRepo,
		deletionRepo:      deletionRepo,
	}
}

// hostConfigsUsing returns the credential mappings of the project which point at
// the key. The repository rejects a change breaking one of them too, but only
// once the secret storage has already been written or emptied, which is too
// late to undo.
func (s *AccessKeyServiceImpl) hostConfigsUsing(projectID *int, keyID int) (used []db.HostConfig, err error) {
	if s.hostConfigRepo == nil || projectID == nil {
		return
	}

	hostConfigs, err := s.hostConfigRepo.GetHostConfigs(*projectID, db.RetrieveQueryParams{})
	if err != nil {
		return
	}

	for _, hostConfig := range hostConfigs {
		if hostConfig.SSHKeyID == keyID {
			used = append(used, hostConfig)
		}
	}

	return
}

func (s *AccessKeyServiceImpl) Delete(projectID int, keyID int) (err error) {
	return s.deleteKey(projectID, keyID, nil)
}

func (s *AccessKeyServiceImpl) DeleteWithSecret(secret db.EnvironmentSecret) (err error) {
	return s.deleteKey(secret.ProjectID, secret.KeyID, &secret)
}

// deleteKey removes the secret of the key from its storage and then the key row. onlyLink, when set, is
// the one variable group secret allowed to point at the key; it is deleted together with the row.
func (s *AccessKeyServiceImpl) deleteKey(projectID int, keyID int, onlyLink *db.EnvironmentSecret) (err error) {
	key, err := s.accessKeyRepo.GetAccessKey(projectID, keyID)
	if err != nil {
		return
	}

	// Checked here rather than left to the repository: the secret is removed
	// from its storage below, and a deletion refused after that has destroyed
	// the credential while leaving the mapping pointing at it.
	used, err := s.hostConfigsUsing(&projectID, keyID)
	if err != nil {
		return
	}

	if len(used) > 0 {
		err = common_errors.NewValidationError(
			"the credential is used by the mapping for " + used[0].Name)
		return
	}

	// The same for variable groups, with the name of the group for the answer. This read alone would
	// leave a window: a link accepted after it, while the secret is being removed below. The claim
	// closes it.
	refs, err := s.accessKeyRepo.GetAccessKeyRefs(projectID, keyID)
	if err != nil {
		return
	}

	for _, env := range refs.Environments {
		if onlyLink != nil && env.ID == onlyLink.EnvironmentID {
			continue
		}
		err = fmt.Errorf("%w: the key is used by the variable group %s", db.ErrInvalidOperation, env.Name)
		return
	}

	// From here on no variable group can link the key, on any node, and no link but onlyLink exists.
	exceptSecretID := 0
	if onlyLink != nil {
		exceptSecretID = onlyLink.ID
	}

	claim, err := s.deletionRepo.ClaimAccessKeyDeletion(projectID, keyID, exceptSecretID)
	if err != nil {
		return
	}

	// Set once the secret has left its storage: from then on the key is not given back.
	secretRemoved := false

	defer func() {
		if err == nil {
			return
		}

		fields := log.Fields{"project_id": projectID, "key_id": keyID}

		if secretRemoved {
			// The claim stays until it expires. Releasing it here would let a link that is being
			// written right now pass its check and stay on a key that has no secret any more.
			log.WithError(err).WithFields(fields).Error(
				"the secret of a key was removed from its storage, but the key could not be deleted")
			return
		}

		// The key is as it was: give it back, so it can be linked and its deletion repeated.
		if releaseErr := s.deletionRepo.ReleaseAccessKeyDeletion(projectID, keyID, claim); releaseErr != nil {
			log.WithError(releaseErr).WithFields(fields).Error(
				"failed to release the deletion claim of a key; it expires on its own")
		}
	}()

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
			secretRemoved = err == nil
		}

		if err != nil {
			return
		}
	}

	err = s.deleteKeyRow(projectID, keyID, onlyLink)

	// A link written while the key was being claimed is still in the table for a moment: its writer
	// reads the claim next and takes the link back. Until then the foreign key refuses the row. With the
	// secret already gone this is not the time to give up, so the row is tried again.
	for _, pause := range []time.Duration{25 * time.Millisecond, 50 * time.Millisecond, 100 * time.Millisecond, 200 * time.Millisecond} {
		if !secretRemoved || !errors.Is(err, db.ErrInvalidOperation) {
			break
		}
		time.Sleep(pause)
		err = s.deleteKeyRow(projectID, keyID, onlyLink)
	}

	return
}

func (s *AccessKeyServiceImpl) deleteKeyRow(projectID int, keyID int, onlyLink *db.EnvironmentSecret) error {
	if onlyLink != nil {
		return s.deletionRepo.DeleteAccessKeyWithSecret(projectID, keyID, onlyLink.EnvironmentID, onlyLink.ID)
	}
	return s.accessKeyRepo.DeleteAccessKey(projectID, keyID)
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

	// Before the secret is written to its storage: a type the mappings can not
	// use is rejected by the repository, but by then the remote secret has
	// already been overwritten with a payload the old row can not deserialize.
	var used []db.HostConfig
	used, err = s.hostConfigsUsing(key.ProjectID, key.ID)
	if err != nil {
		return
	}

	for _, hostConfig := range used {
		if err = hostConfig.ValidateCredential(key.Type); err != nil {
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
