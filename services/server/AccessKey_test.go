package server

import (
	"encoding/base64"
	"errors"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
)

func TestSetSecret(t *testing.T) {
	accessKey := db.AccessKey{
		Type: db.AccessKeySSH,
		Name: "test",
		SshKey: db.SshKey{
			PrivateKey: "qerphqeruqoweurqwerqqeuiqwpavqr",
		},
	}

	encryptionService := NewAccessKeyEncryptionService(nil, nil, nil, nil)

	util.Config = &util.ConfigType{}
	err := encryptionService.SerializeSecret(&accessKey)

	if err != nil {
		t.Fatal(err)
	}

	secret, err := base64.StdEncoding.DecodeString(*accessKey.Secret)

	if err != nil {
		t.Error(err)
	}

	if string(secret) != "{\"login\":\"\",\"passphrase\":\"\",\"private_key\":\"qerphqeruqoweurqwerqqeuiqwpavqr\"}" {
		t.Error("invalid secret")
	}
}

func TestGetSecret(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte(`{
	"passphrase": "123456",
	"private_key": "qerphqeruqoweurqwerqqeuiqwpavqr"
}`))
	util.Config = &util.ConfigType{}

	encryptionService := NewAccessKeyEncryptionService(nil, nil, nil, nil)

	accessKey := db.AccessKey{
		Secret: &secret,
		Type:   db.AccessKeySSH,
	}

	err := encryptionService.DeserializeSecret(&accessKey)

	if err != nil {
		t.Error(err)
	}

	if accessKey.SshKey.Passphrase != "123456" {
		t.Errorf("")
	}

	if accessKey.SshKey.PrivateKey != "qerphqeruqoweurqwerqqeuiqwpavqr" {
		t.Errorf("")
	}
}

func TestSetGetSecretWithEncryption(t *testing.T) {

	encryptionService := NewAccessKeyEncryptionService(nil, nil, nil, nil)

	accessKey := db.AccessKey{
		Name: "test",
		Type: db.AccessKeySSH,
		SshKey: db.SshKey{
			PrivateKey: "qerphqeruqoweurqwerqqeuiqwpavqr",
		},
	}

	util.Config = &util.ConfigType{
		AccessKeyEncryption: "hHYgPrhQTZYm7UFTvcdNfKJMB3wtAXtJENUButH+DmM=",
	}

	err := encryptionService.SerializeSecret(&accessKey)

	if err != nil {
		t.Error(err)
	}

	//accessKey.ClearSecret()

	err = encryptionService.DeserializeSecret(&accessKey)

	if err != nil {
		t.Error(err)
	}

	if accessKey.SshKey.PrivateKey != "qerphqeruqoweurqwerqqeuiqwpavqr" {
		t.Error("invalid secret")
	}
}

func TestSerializeSecretReadOnlyReturnsUnwrappableSentinel(t *testing.T) {
	storageType := db.AccessKeySourceStorageEnv
	accessKey := db.AccessKey{
		Type:              db.AccessKeyString,
		Name:              "test",
		String:            "value",
		SourceStorageType: &storageType,
	}

	util.Config = &util.ConfigType{}
	encryptionService := NewAccessKeyEncryptionService(nil, nil, nil, nil)

	err := encryptionService.SerializeSecret(&accessKey)
	if err == nil {
		t.Fatal("expected error for read-only storage, got nil")
	}
	if !errors.Is(err, ErrReadOnlyStorage) {
		t.Fatalf("expected error to wrap ErrReadOnlyStorage, got: %v", err)
	}
}

func TestCreateSkipsSerializationForReadOnlyStorage(t *testing.T) {
	storageType := db.AccessKeySourceStorageEnv
	key := db.AccessKey{
		Type:              db.AccessKeyString,
		Name:              "test",
		String:            "value",
		SourceStorageType: &storageType,
	}

	util.Config = &util.ConfigType{}

	repo := &mockAccessKeyRepo{}
	encryptionService := NewAccessKeyEncryptionService(nil, nil, nil, nil)
	svc := NewAccessKeyService(repo, encryptionService, nil)

	created, err := svc.Create(key)
	if err != nil {
		t.Fatalf("Create should succeed for read-only storage, got: %v", err)
	}
	if created.Name != "test" {
		t.Fatalf("expected key name 'test', got '%s'", created.Name)
	}
}

func TestRekeyAccessKeysSkipsExternalStorageKeys(t *testing.T) {
	vaultType := db.AccessKeySourceStorageVault
	envType := db.AccessKeySourceStorageEnv
	fileType := db.AccessKeySourceStorageFile
	projectID := 1

	localSecret := base64.StdEncoding.EncodeToString([]byte("local-secret-value"))
	vaultSecret := "vault-ciphertext-should-not-be-touched"
	envSecret := "env-ciphertext-should-not-be-touched"
	fileSecret := "file-ciphertext-should-not-be-touched"

	allKeys := []db.AccessKey{
		{
			ID:        1,
			Name:      "local-key",
			Type:      db.AccessKeyString,
			ProjectID: &projectID,
			Secret:    &localSecret,
		},
		{
			ID:                2,
			Name:              "vault-key",
			Type:              db.AccessKeyString,
			ProjectID:         &projectID,
			Secret:            &vaultSecret,
			SourceStorageType: &vaultType,
			SourceStorageID:   intPtr(10),
		},
		{
			ID:                3,
			Name:              "env-key",
			Type:              db.AccessKeyString,
			ProjectID:         &projectID,
			Secret:            &envSecret,
			SourceStorageType: &envType,
			SourceStorageKey:  strPtr("MY_ENV_VAR"),
		},
		{
			ID:                4,
			Name:              "file-key",
			Type:              db.AccessKeyString,
			ProjectID:         &projectID,
			Secret:            &fileSecret,
			SourceStorageType: &fileType,
			SourceStorageKey:  strPtr("/etc/secret"),
		},
	}

	var updatedIDs []int
	keyMgr := &mockAccessKeyManager{
		GetAccessKeysFn: func(_ int, _ db.GetAccessKeyOptions, params db.RetrieveQueryParams) ([]db.AccessKey, error) {
			if params.Offset > 0 {
				return nil, nil
			}
			return allKeys, nil
		},
		UpdateAccessKeyFn: func(key db.AccessKey) error {
			updatedIDs = append(updatedIDs, key.ID)
			return nil
		},
	}

	projectStore := &mockProjectStore{
		GetAllProjectsFn: func() ([]db.Project, error) {
			return []db.Project{{ID: projectID}}, nil
		},
	}

	util.Config = &util.ConfigType{}
	svc := NewAccessKeyEncryptionService(keyMgr, nil, nil, projectStore)

	err := svc.RekeyAccessKeys("")
	if err != nil {
		t.Fatalf("RekeyAccessKeys returned error: %v", err)
	}

	if len(updatedIDs) != 1 || updatedIDs[0] != 1 {
		t.Fatalf("expected only local key (ID=1) to be updated, got updates for IDs: %v", updatedIDs)
	}
}

func intPtr(i int) *int       { return &i }
func strPtr(s string) *string { return &s }

type mockAccessKeyRepo struct {
	keys []db.AccessKey
}

func (m *mockAccessKeyRepo) GetAccessKey(_ int, keyID int) (db.AccessKey, error) {
	for _, k := range m.keys {
		if k.ID == keyID {
			return k, nil
		}
	}
	return db.AccessKey{}, db.ErrNotFound
}
func (m *mockAccessKeyRepo) GetAccessKeyRefs(int, int) (db.ObjectReferrers, error) {
	return db.ObjectReferrers{}, nil
}
func (m *mockAccessKeyRepo) GetAccessKeys(int, db.GetAccessKeyOptions, db.RetrieveQueryParams) ([]db.AccessKey, error) {
	return nil, nil
}
func (m *mockAccessKeyRepo) UpdateAccessKey(db.AccessKey) error { return nil }
func (m *mockAccessKeyRepo) CreateAccessKey(k db.AccessKey) (db.AccessKey, error) {
	return k, nil
}
func (m *mockAccessKeyRepo) DeleteAccessKey(int, int) error { return nil }

// --- AccessKeyServiceImpl.Delete (secret storage + synchronized key behavior) ---

type mockSecretStorageRepoDelete struct {
	storage db.SecretStorage
}

func (m *mockSecretStorageRepoDelete) GetSecretStorages(int) ([]db.SecretStorage, error) {
	return nil, nil
}
func (m *mockSecretStorageRepoDelete) CreateSecretStorage(db.SecretStorage) (db.SecretStorage, error) {
	return db.SecretStorage{}, nil
}
func (m *mockSecretStorageRepoDelete) GetSecretStorage(_ int, _ int) (db.SecretStorage, error) {
	return m.storage, nil
}
func (m *mockSecretStorageRepoDelete) UpdateSecretStorage(db.SecretStorage) error { return nil }
func (m *mockSecretStorageRepoDelete) GetSecretStorageRefs(int, int) (db.ObjectReferrers, error) {
	return db.ObjectReferrers{}, nil
}
func (m *mockSecretStorageRepoDelete) DeleteSecretStorage(int, int) error { return nil }

type mockAccessKeyRepoDelete struct {
	key            db.AccessKey
	deleteKeyCalls int
}

func (m *mockAccessKeyRepoDelete) GetAccessKey(_ int, keyID int) (db.AccessKey, error) {
	if keyID == m.key.ID {
		return m.key, nil
	}
	return db.AccessKey{}, db.ErrNotFound
}
func (m *mockAccessKeyRepoDelete) GetAccessKeyRefs(int, int) (db.ObjectReferrers, error) {
	return db.ObjectReferrers{}, nil
}
func (m *mockAccessKeyRepoDelete) GetAccessKeys(int, db.GetAccessKeyOptions, db.RetrieveQueryParams) ([]db.AccessKey, error) {
	return nil, nil
}
func (m *mockAccessKeyRepoDelete) UpdateAccessKey(db.AccessKey) error { return nil }
func (m *mockAccessKeyRepoDelete) CreateAccessKey(k db.AccessKey) (db.AccessKey, error) {
	return k, nil
}
func (m *mockAccessKeyRepoDelete) DeleteAccessKey(_ int, _ int) error {
	m.deleteKeyCalls++
	return nil
}

type mockEncryptionDeleteSecret struct {
	deleteSecretCalls int
}

func (m *mockEncryptionDeleteSecret) SerializeSecret(*db.AccessKey) error   { return nil }
func (m *mockEncryptionDeleteSecret) DeserializeSecret(*db.AccessKey) error { return nil }
func (m *mockEncryptionDeleteSecret) FillEnvironmentSecrets(*db.Environment, bool) error {
	return nil
}
func (m *mockEncryptionDeleteSecret) DeleteSecret(*db.AccessKey) error {
	m.deleteSecretCalls++
	return nil
}
func (m *mockEncryptionDeleteSecret) RekeyAccessKeys(string) error { return nil }

func TestAccessKeyServiceDelete_WritableSynchronizedStillCallsDeleteSecret(t *testing.T) {
	projectID := 1
	storageID := 99
	st := db.AccessKeySourceStorageVault
	key := db.AccessKey{
		ID:                5,
		ProjectID:         &projectID,
		SourceStorageType: &st,
		SourceStorageID:   &storageID,
		Synchronized:      true,
	}

	repo := &mockAccessKeyRepoDelete{key: key}
	enc := &mockEncryptionDeleteSecret{}
	stRepo := &mockSecretStorageRepoDelete{
		storage: db.SecretStorage{ID: storageID, ReadOnly: false},
	}

	svc := NewAccessKeyService(repo, enc, stRepo)
	err := svc.Delete(projectID, key.ID)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if enc.deleteSecretCalls != 1 {
		t.Fatalf("expected DeleteSecret once for writable+synchronized, got %d", enc.deleteSecretCalls)
	}
	if repo.deleteKeyCalls != 1 {
		t.Fatalf("expected DeleteAccessKey once, got %d", repo.deleteKeyCalls)
	}
}

func TestAccessKeyServiceDelete_ReadOnlyNonSynchronizedSkipsDeleteSecret(t *testing.T) {
	projectID := 1
	storageID := 99
	st := db.AccessKeySourceStorageVault
	key := db.AccessKey{
		ID:                5,
		ProjectID:         &projectID,
		SourceStorageType: &st,
		SourceStorageID:   &storageID,
		Synchronized:      false,
	}

	repo := &mockAccessKeyRepoDelete{key: key}
	enc := &mockEncryptionDeleteSecret{}
	stRepo := &mockSecretStorageRepoDelete{
		storage: db.SecretStorage{ID: storageID, ReadOnly: true},
	}

	svc := NewAccessKeyService(repo, enc, stRepo)
	err := svc.Delete(projectID, key.ID)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if enc.deleteSecretCalls != 0 {
		t.Fatalf("expected no DeleteSecret for read-only non-sync, got %d", enc.deleteSecretCalls)
	}
	if repo.deleteKeyCalls != 1 {
		t.Fatalf("expected DeleteAccessKey once, got %d", repo.deleteKeyCalls)
	}
}

func TestAccessKeyServiceDelete_ReadOnlySynchronizedReturnsError(t *testing.T) {
	projectID := 1
	storageID := 99
	st := db.AccessKeySourceStorageVault
	key := db.AccessKey{
		ID:                5,
		ProjectID:         &projectID,
		SourceStorageType: &st,
		SourceStorageID:   &storageID,
		Synchronized:      true,
	}

	repo := &mockAccessKeyRepoDelete{key: key}
	enc := &mockEncryptionDeleteSecret{}
	stRepo := &mockSecretStorageRepoDelete{
		storage: db.SecretStorage{ID: storageID, ReadOnly: true},
	}

	svc := NewAccessKeyService(repo, enc, stRepo)
	err := svc.Delete(projectID, key.ID)
	if err == nil {
		t.Fatal("expected error for read-only synchronized key")
	}
	if enc.deleteSecretCalls != 0 {
		t.Fatalf("expected no DeleteSecret, got %d", enc.deleteSecretCalls)
	}
	if repo.deleteKeyCalls != 0 {
		t.Fatalf("expected DeleteAccessKey not called on error, got %d", repo.deleteKeyCalls)
	}
}
