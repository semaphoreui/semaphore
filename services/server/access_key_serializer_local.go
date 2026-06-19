package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/semaphoreui/semaphore/util"
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

	secret, err := util.Config.EncryptAccessSecret(plaintext)
	if err != nil {
		return err
	}
	key.Secret = &secret

	return nil
}

func (d *LocalAccessKeyDeserializer) DeserializeSecret(key *db.AccessKey) (res string, err error) {
	return d.deserializeSecretWithKeys(key, util.Config.AccessSecretDecryptKeys())
}

// DeserializeSecret2 decrypts using a single explicit key. It is kept for the
// rekey path (which supplies an old key) and for tests.
func (d *LocalAccessKeyDeserializer) DeserializeSecret2(key *db.AccessKey, encryptionString string) (res string, err error) {
	return d.deserializeSecretWithKeys(key, []string{encryptionString})
}

// deserializeSecretWithKeys decrypts the secret, trying each key in
// encryptionKeys in order (primary first, then retired secondaries) and
// returning the first success.
func (d *LocalAccessKeyDeserializer) deserializeSecretWithKeys(key *db.AccessKey, encryptionKeys []string) (res string, err error) {

	if key.SourceStorageType != nil {
		if key.SourceStorageKey == nil {
			return "", fmt.Errorf("source storage key is required")
		}

		switch *key.SourceStorageType {
		case db.AccessKeySourceStorageEnv:
			res = os.Getenv(*key.SourceStorageKey)
			return
		case db.AccessKeySourceStorageFile:

			filePath := filepath.Clean(*key.SourceStorageKey)
			if !filepath.IsAbs(filePath) {
				err = common_errors.NewUserErrorS("file path must be absolute")
				return
			}

			for _, segment := range strings.Split(filepath.ToSlash(*key.SourceStorageKey), "/") {
				if segment == ".." {
					err = common_errors.NewUserErrorS("file path must not contain traversal segments")
					return
				}
			}

			secretsBasePath := filepath.Clean(util.Config.Dirs.Secrets)
			if !filepath.IsAbs(secretsBasePath) {
				err = common_errors.NewUserErrorS("secrets path must be absolute")
				return
			}

			var relPath string
			relPath, err = filepath.Rel(secretsBasePath, filePath)
			if err != nil {
				return
			}

			if relPath == ".." || strings.HasPrefix(relPath, ".."+string(os.PathSeparator)) {
				err = common_errors.NewUserErrorS("file path must be inside secrets path")
				return
			}

			var data []byte
			data, err = os.ReadFile(filePath)
			if err != nil {
				return
			}
			res = strings.TrimSuffix(string(data), "\n")
			return
		}
	}

	if key.Secret == nil || *key.Secret == "" {
		return
	}

	secret := *key.Secret

	if secret[len(secret)-1] == '\n' { // not encrypted private key, used for back compatibility
		if key.Type != db.AccessKeySSH {
			err = fmt.Errorf("invalid access key type")
			return
		}

		sshKey := db.SshKey{
			PrivateKey: secret,
		}

		var marshaled []byte
		marshaled, err = json.Marshal(sshKey)
		if err != nil {
			return
		}

		res = string(marshaled)

		return
	}

	// abort early if the secret is not valid base64
	if _, err = base64.StdEncoding.DecodeString(secret); err != nil {
		return
	}

	if len(encryptionKeys) == 0 {
		encryptionKeys = []string{""}
	}

	var decErr error
	for _, encryptionString := range encryptionKeys {
		var plaintext []byte
		plaintext, decErr = util.DecryptAESGCM(secret, encryptionString)
		if decErr == nil {
			res = string(plaintext)
			return
		}
	}

	if decErr.Error() == "cipher: message authentication failed" {
		err = fmt.Errorf("cannot decrypt access key, perhaps encryption key was changed")
	} else {
		err = decErr
	}
	return
}
