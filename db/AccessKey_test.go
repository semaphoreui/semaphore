package db

import (
	"encoding/base64"
	"github.com/ansible-semaphore/semaphore/util"
	"testing"
)

func TestSetSecret(t *testing.T) {
	accessKey := AccessKey{
		Type: AccessKeySSH,
		SshKey: SshKey{
			PrivateKey: "qerphqeruqoweurqwerqqeuiqwpavqr",
		},
	}

	util.Config = &util.ConfigType{}
	err := accessKey.SerializeSecret()

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

	accessKey := AccessKey{
		Secret: &secret,
		Type:   AccessKeySSH,
	}

	err := accessKey.DeserializeSecret()

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
	accessKey := AccessKey{
		Type: AccessKeySSH,
		SshKey: SshKey{
			PrivateKey: "qerphqeruqoweurqwerqqeuiqwpavqr",
		},
	}

	util.Config = &util.ConfigType{
		AccessKeyEncryption: "hHYgPrhQTZYm7UFTvcdNfKJMB3wtAXtJENUButH+DmM=",
	}

	err := accessKey.SerializeSecret()

	if err != nil {
		t.Error(err)
	}

	//accessKey.ClearSecret()

	err = accessKey.DeserializeSecret()

	if err != nil {
		t.Error(err)
	}

	if accessKey.SshKey.PrivateKey != "qerphqeruqoweurqwerqqeuiqwpavqr" {
		t.Error("invalid secret")
	}
}
