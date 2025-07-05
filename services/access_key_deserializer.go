package services

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
)

type AccessKeyKeyDeserializer interface {
	DeserializeSecret(key *db.AccessKey) (string, error)
}

type VaultAccessKeyKeyDeserializer struct {
	accessKeyRepo db.AccessKeyManager
}

func (d *VaultAccessKeyKeyDeserializer) DeserializeSecret(key *db.AccessKey) (res string, err error) {
	return
}

type DatabaseAccessKeyKeyDeserializer struct {
}

func (d *DatabaseAccessKeyKeyDeserializer) DeserializeSecret(key *db.AccessKey) (res string, err error) {
	return d.DeserializeSecret2(key, util.Config.AccessKeyEncryption)
}

func (d *DatabaseAccessKeyKeyDeserializer) DeserializeSecret2(key *db.AccessKey, encryptionString string) (res string, err error) {
	if key.Secret == nil || *key.Secret == "" {
		return
	}

	ciphertext := []byte(*key.Secret)

	if ciphertext[len(*key.Secret)-1] == '\n' { // not encrypted private key, used for back compatibility
		if key.Type != db.AccessKeySSH {
			err = fmt.Errorf("invalid access key type")
			return
		}
		key.SshKey = db.SshKey{
			PrivateKey: *key.Secret,
		}
		return
	}

	ciphertext, err = base64.StdEncoding.DecodeString(*key.Secret)
	if err != nil {
		return
	}

	if encryptionString == "" {
		err = unmarshalAppropriateField(key, ciphertext)
		var syntaxError *json.SyntaxError
		if errors.As(err, &syntaxError) {
			err = fmt.Errorf("secret must be valid json in key '%s'", key.Name)
		}

		return
	}

	encryption, err := base64.StdEncoding.DecodeString(encryptionString)
	if err != nil {
		return
	}

	c, err := aes.NewCipher(encryption)
	if err != nil {
		return
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		err = fmt.Errorf("ciphertext too short")
		return
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	ciphertext, err = gcm.Open(nil, nonce, ciphertext, nil)

	if err != nil {
		if err.Error() == "cipher: message authentication failed" {
			err = fmt.Errorf("cannot decrypt access key, perhaps encryption key was changed")
		}
		return
	}

	res = string(ciphertext)
	return
}
