package server

import (
	"encoding/base64"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
	"testing"
)

func TestSetSecret(t *testing.T) {
	accessKey := db.AccessKey{
		Type: db.AccessKeySSH,
		SshKey: db.SshKey{
			PrivateKey: "qerphqeruqoweurqwerqqeuiqwpavqr",
		},
	}

	encryptionService := NewAccessKeyEncryptionService(nil, nil, nil)

	util.Config = &util.ConfigType{}
	err := encryptionService.SerializeSecret(&accessKey)

	if err != nil {
		t.Error(err)
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

	encryptionService := NewAccessKeyEncryptionService(nil, nil, nil)

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

	encryptionService := NewAccessKeyEncryptionService(nil, nil, nil)

	accessKey := db.AccessKey{
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
