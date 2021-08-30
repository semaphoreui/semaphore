package db

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"strconv"

	"github.com/ansible-semaphore/semaphore/util"
)

const (
	AccessKeySSH = "ssh"
	AccessKeyNone = "none"
)


// AccessKey represents a key used to access a machine with ansible from semaphore
type AccessKey struct {
	ID   int    `db:"id" json:"id"`
	Name string `db:"name" json:"name" binding:"required"`
	// 'ssh/none'
	Type string `db:"type" json:"type" binding:"required"`

	ProjectID *int    `db:"project_id" json:"project_id"`
	Key       *string `db:"key" json:"key"`
	Secret    *string `db:"secret" json:"secret"`

	Removed bool `db:"removed" json:"removed"`
}

// GetPath returns the location of the access key once written to disk
func (key AccessKey) GetPath() string {
	return util.Config.TmpPath + "/access_key_" + strconv.Itoa(key.ID)
}

func (key *AccessKey) EncryptSecret() error {
	if key.Secret == nil || *key.Secret == "" {
		return nil
	}

	if util.Config.AccessKeyEncryption == "" { // do not encrypt if secret key not specified
		return nil
	}

	plaintext := []byte(*key.Secret)

	encryption, err := base64.StdEncoding.DecodeString(util.Config.CookieEncryption)

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

	secret := string(gcm.Seal(nonce, nonce, plaintext, nil))

	key.Secret = &secret

	return nil
}

func (key AccessKey) DecryptSecret() (string, error) {
	if key.Secret == nil || *key.Secret == "" {
		return "", nil
	}

	ciphertext := []byte(*key.Secret)

	if ciphertext[len(ciphertext) - 1] == '\n' { // not encrypted string
		return *key.Secret, nil
	}

	if util.Config.AccessKeyEncryption == "" { // do not decrypt if secret key not specified
		return *key.Secret, nil
	}

	encryption, err := base64.StdEncoding.DecodeString(util.Config.CookieEncryption)
	if err != nil {
		return "", err
	}

	c, err := aes.NewCipher(encryption)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	encrypted, err := gcm.Open(nil, nonce, ciphertext, nil)

	if err != nil {
		return "", err
	}

	return string(encrypted), nil
}
