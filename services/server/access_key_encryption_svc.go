package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
	pro "github.com/semaphoreui/semaphore/pro/services/server"
)

const RekeyBatchSize = 100

var ErrReadOnlyStorage = errors.New("cannot modify secret in read-only storage")

type AccessKeyEncryptionService interface {
	SerializeSecret(key *db.AccessKey) error
	DeserializeSecret(key *db.AccessKey) error
	FillEnvironmentSecrets(env *db.Environment, deserializeSecret bool) error
	DeleteSecret(key *db.AccessKey) error
	RekeyAccessKeys(oldKey string) (err error)
}

func NewAccessKeyEncryptionService(
	accessKeyRepo db.AccessKeyManager,
	environmentRepo db.EnvironmentManager,
	secretStorageRepo db.SecretStorageRepository,
	projectRepo db.ProjectStore,
) AccessKeyEncryptionService {
	return &accessKeyEncryptionServiceImpl{
		accessKeyRepo:     accessKeyRepo,
		environmentRepo:   environmentRepo,
		secretStorageRepo: secretStorageRepo,
		projectRepo:       projectRepo,
	}
}

func unmarshalAppropriateField(key *db.AccessKey, secret []byte) (err error) {
	switch key.Type {
	case db.AccessKeyString:
		key.String = string(secret)
	case db.AccessKeySSH:
		sshKey := db.SshKey{}
		err = json.Unmarshal(secret, &sshKey)
		if err == nil {
			key.SshKey = sshKey
		}
	case db.AccessKeyLoginPassword:
		loginPass := db.LoginPassword{}
		err = json.Unmarshal(secret, &loginPass)
		if err == nil {
			key.LoginPassword = loginPass
		}
	}
	return
}

type accessKeyEncryptionServiceImpl struct {
	accessKeyRepo     db.AccessKeyManager
	environmentRepo   db.EnvironmentManager
	secretStorageRepo db.SecretStorageRepository
	projectRepo       db.ProjectStore
}

func (s *accessKeyEncryptionServiceImpl) getDeserializer(key *db.AccessKey) (AccessKeyDeserializer, bool, error) {

	if key.SourceStorageType == nil {
		return &LocalAccessKeyDeserializer{}, false, nil
	}

	switch *key.SourceStorageType {
	case db.AccessKeySourceStorageEnv, db.AccessKeySourceStorageFile:
		return &LocalAccessKeyDeserializer{}, true, nil
	case db.AccessKeySourceStorageVault:
		if key.SourceStorageID == nil {
			return &LocalAccessKeyDeserializer{}, false, errors.New("vault storage id is required")
		}
	default:
		return nil, false, fmt.Errorf("unsupported secret storage type '%s'", *key.SourceStorageType)
	}

	storage, err := s.secretStorageRepo.GetSecretStorage(*key.ProjectID, *key.SourceStorageID)
	if err != nil {
		return nil, false, err
	}

	switch storage.Type {
	case db.SecretStorageTypeVault:
		return pro.NewVaultAccessKeyDeserializer(s.accessKeyRepo, s.secretStorageRepo, s), storage.ReadOnly, nil
	case db.SecretStorageTypeDvls:
		return pro.NewDvlsAccessKeyDeserializer(s.accessKeyRepo, s.secretStorageRepo, s), storage.ReadOnly, nil
	case db.SecretStorageTypeAwsSm:
		return pro.NewAwsSmAccessKeyDeserializer(s.accessKeyRepo, s.secretStorageRepo, s), storage.ReadOnly, nil
	}

	return nil, false, fmt.Errorf("unsupported secret storage type '%s'", storage.Type)
}

func (s *accessKeyEncryptionServiceImpl) DeleteSecret(key *db.AccessKey) error {
	d, _, err := s.getDeserializer(key)
	if err != nil {
		return err
	}
	return d.DeleteSecret(key)
}

func (s *accessKeyEncryptionServiceImpl) SerializeSecret(key *db.AccessKey) error {
	d, readonly, err := s.getDeserializer(key)
	if err != nil {
		return err
	}
	if readonly {
		return common_errors.NewUserErrorS("cannot modify secret in read-only storage")
	}

	err = key.Validate(true)
	if err != nil {
		return err
	}

	return d.SerializeSecret(key)
}

func (s *accessKeyEncryptionServiceImpl) DeserializeSecret(key *db.AccessKey) error {
	d, _, err := s.getDeserializer(key)
	if err != nil {
		return err
	}
	ciphertext, err := d.DeserializeSecret(key)
	if err != nil {
		return err
	}

	err = unmarshalAppropriateField(key, []byte(ciphertext))

	var syntaxError *json.SyntaxError
	if errors.As(err, &syntaxError) {
		err = fmt.Errorf("secret must be valid json in key '%s'", key.Name)
	}

	return err
}

func (s *accessKeyEncryptionServiceImpl) FillEnvironmentSecrets(env *db.Environment, deserializeSecret bool) error {
	keys, err := s.environmentRepo.GetEnvironmentSecrets(env.ProjectID, env.ID)

	if err != nil {
		return err
	}

	for _, k := range keys {
		var secretName string
		var secretType db.EnvironmentSecretType

		if k.Owner == db.AccessKeyVariable {
			secretType = db.EnvironmentSecretVar
			secretName = strings.TrimPrefix(k.Name, string(db.EnvironmentSecretVar)+".")
		} else if k.Owner == db.AccessKeyEnvironment {
			secretType = db.EnvironmentSecretEnv
			secretName = strings.TrimPrefix(k.Name, string(db.EnvironmentSecretEnv)+".")
		} else {
			secretType = db.EnvironmentSecretVar
			secretName = k.Name
		}

		if deserializeSecret {
			err = s.DeserializeSecret(&k)
			if err != nil {
				return err
			}
		}

		env.Secrets = append(env.Secrets, db.EnvironmentSecret{
			ID:     k.ID,
			Name:   secretName,
			Type:   secretType,
			Secret: k.String,
		})
	}

	return nil
}

func (s *accessKeyEncryptionServiceImpl) RekeyAccessKeys(oldKey string) (err error) {

	deserializer := NewLocalAccessKeyDeserializer()

	projects, err := s.projectRepo.GetAllProjects()
	if err != nil {
		return
	}

	for _, project := range projects {

		for i := 0; ; i++ {

			var keys []db.AccessKey
			keys, err = s.accessKeyRepo.GetAccessKeys(project.ID, db.GetAccessKeyOptions{IgnoreOwner: true}, db.RetrieveQueryParams{Count: RekeyBatchSize, Offset: i * RekeyBatchSize})

			if err != nil {
				return
			}

			if len(keys) == 0 {
				break
			}

			for _, key := range keys {

				var secret string
				secret, err = deserializer.DeserializeSecret2(&key, oldKey)

				if err != nil {
					return err
				}

				err = unmarshalAppropriateField(&key, []byte(secret))

				if err != nil {
					return err
				}

				key.OverrideSecret = true
				err = deserializer.SerializeSecret(&key)

				if err != nil {
					return err
				}

				err = s.accessKeyRepo.UpdateAccessKey(key)

				if err != nil && !errors.Is(err, db.ErrNotFound) {
					return err
				}
			}
		}
	}

	return
}
