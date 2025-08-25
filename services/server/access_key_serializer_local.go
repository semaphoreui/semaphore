package server

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
	"io"
)

type LocalAccessKeyDeserializer struct {
}

func NewLocalAccessKeyDeserializer() *LocalAccessKeyDeserializer {
	return &LocalAccessKeyDeserializer{}
}

func (d *LocalAccessKeyDeserializer) DeleteSecret(key *db.AccessKey) error {
	// No-op for local deserializer
	return nil
}

func (d *LocalAccessKeyDeserializer) SerializeSecret(key *db.AccessKey) error {
	var plaintext []byte
	var err error

	switch key.Type {
	case db.AccessKeyString:
		if key.String == "" {
			key.Secret = nil
			return nil
		}
		plaintext = []byte(key.String)
	case db.AccessKeySSH:
		if key.SshKey.PrivateKey == "" {
			if key.SshKey.Login != "" || key.SshKey.Passphrase != "" {
				return fmt.Errorf("invalid ssh key")
			}
			key.Secret = nil
			return nil
		}

		plaintext, err = json.Marshal(key.SshKey)
		if err != nil {
			return err
		}
	case db.AccessKeyLoginPassword:
		if key.LoginPassword.Password == "" {
			if key.LoginPassword.Login != "" {
				return fmt.Errorf("invalid password key")
			}
			key.Secret = nil
			return nil
		}

		plaintext, err = json.Marshal(key.LoginPassword)
		if err != nil {
			return err
		}
	case db.AccessKeyNone:
		key.Secret = nil
		return nil
	default:
		return fmt.Errorf("invalid access token type")
	}

	encryptionString := util.Config.AccessKeyEncryption

	if encryptionString == "" {
		secret := base64.StdEncoding.EncodeToString(plaintext)
		key.Secret = &secret
		return nil
	}

	encryption, err := base64.StdEncoding.DecodeString(encryptionString)

	if err != nil {
		return err
	}

	c, err := aes.NewCipher(encryption)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	secret := base64.StdEncoding.EncodeToString(gcm.Seal(nonce, nonce, plaintext, nil))
	key.Secret = &secret

	return nil
}

func (d *LocalAccessKeyDeserializer) DeserializeSecret(key *db.AccessKey) (res string, err error) {
	return d.DeserializeSecret2(key, util.Config.AccessKeyEncryption)
}

func (d *LocalAccessKeyDeserializer) DeserializeSecret2(key *db.AccessKey, encryptionString string) (res string, err error) {
	if key.Secret == nil || *key.Secret == "" {
		return
	}

	ciphertext := []byte(*key.Secret)

	if ciphertext[len(*key.Secret)-1] == '\n' { // not encrypted private key, used for back compatibility
		if key.Type != db.AccessKeySSH {
			err = fmt.Errorf("invalid access key type")
			return
		}

		sshKey := db.SshKey{
			PrivateKey: *key.Secret,
		}

		var marshaled []byte
		marshaled, err = json.Marshal(sshKey)
		if err != nil {
			return
		}

		res = string(marshaled)

		return
	}

	ciphertext, err = base64.StdEncoding.DecodeString(*key.Secret)
	if err != nil {
		return
	}

	if encryptionString == "" {
		res = string(ciphertext)
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
