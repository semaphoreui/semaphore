package server_services

import (
	"encoding/json"
	"github.com/semaphoreui/semaphore/db"
	"strings"
)

const RekeyBatchSize = 100

type AccessKeyEncryptionService interface {
	DeserializeSecret(key *db.AccessKey) error
	FillEnvironmentSecrets(env *db.Environment, deserializeSecret bool) error
}

func NewAccessKeyEncryptionService(
	accessKeyRepo db.AccessKeyManager,
	environmentRepo db.EnvironmentManager,
	secretStorageRepo db.SecretStorageRepository,
) AccessKeyEncryptionService {
	return &accessKeyEncryptionServiceImpl{
		accessKeyRepo:     accessKeyRepo,
		environmentRepo:   environmentRepo,
		secretStorageRepo: secretStorageRepo,
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

func createAccessKeyDeserializer(
	keyRepo db.AccessKeyManager,
	storageRepo db.SecretStorageRepository,
	key *db.AccessKey,
	encryptionService AccessKeyEncryptionService,
) AccessKeyKeyDeserializer {
	if key.SourceStorageID == nil {
		return &LocalAccessKeyDeserializer{}
	}

	return NewVaultAccessKeyDeserializer(keyRepo, storageRepo, encryptionService)
}

type accessKeyEncryptionServiceImpl struct {
	accessKeyRepo     db.AccessKeyManager
	environmentRepo   db.EnvironmentManager
	secretStorageRepo db.SecretStorageRepository
}

func (s *accessKeyEncryptionServiceImpl) DeserializeSecret(key *db.AccessKey) error {
	deserializer := createAccessKeyDeserializer(s.accessKeyRepo, s.secretStorageRepo, key, s)

	ciphertext, err := deserializer.DeserializeSecret(key)
	if err != nil {
		return err
	}

	return unmarshalAppropriateField(key, []byte(ciphertext))
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

	//var globalProps = db.AccessKeyProps
	//globalProps.IsGlobal = true
	//
	//for i := 0; ; i++ {
	//
	//	var keys []db.AccessKey
	//	err = d.getObjects(-1, globalProps, db.RetrieveQueryParams{Count: RekeyBatchSize, Offset: i * RekeyBatchSize}, nil, &keys)
	//
	//	if err != nil {
	//		return
	//	}
	//
	//	if len(keys) == 0 {
	//		break
	//	}
	//
	//	for _, key := range keys {
	//
	//		err = s.DeserializeSecret(oldKey)
	//		err = key.DeserializeSecret2(oldKey)
	//
	//		if err != nil {
	//			return err
	//		}
	//
	//		key.OverrideSecret = true
	//		err = s.accessKeyRepo.UpdateAccessKey(key)
	//
	//		if err != nil && !errors.Is(err, db.ErrNotFound) {
	//			return err
	//		}
	//	}
	//}

	return
}
