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
