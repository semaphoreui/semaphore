package db

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"

	"github.com/ansible-semaphore/semaphore/util"
)

const (
	AccessKeySSH           = "ssh"
	AccessKeyNone          = "none"
	AccessKeyLoginPassword = "login_password"
)

// AccessKey represents a key used to access a machine with ansible from semaphore
type AccessKey struct {
	ID   int    `db:"id" json:"id"`
	Name string `db:"name" json:"name" binding:"required"`
	// 'ssh/login_password/none'
	Type string `db:"type" json:"type" binding:"required"`

	ProjectID *int `db:"project_id" json:"project_id"`

	// Secret used internally, do not assign this field.
	// You should use methods SerializeSecret to fill this field.
	Secret *string `db:"secret" json:"-"`

	Removed bool `db:"removed" json:"removed"`

	LoginPassword  LoginPassword `db:"-" json:"login_password"`
	SshKey         SshKey        `db:"-" json:"ssh"`
	OverrideSecret bool          `db:"-" json:"override_secret"`
}

type LoginPassword struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type SshKey struct {
	Login      string `json:"login"`
	Passphrase string `json:"passphrase"`
	PrivateKey string `json:"private_key"`
}

type AccessKeyUsage int

const (
	AccessKeyUsageAnsibleUser = iota
	AccessKeyUsageAnsibleBecomeUser
	AccessKeyUsagePrivateKey
	AccessKeyUsageVault
)

func (key AccessKey) Install(usage AccessKeyUsage) error {
	if key.Type == AccessKeyNone {
		return nil
	}

	path := key.GetPath()

	err := key.DeserializeSecret()

	if err != nil {
		return err
	}

	switch usage {
	case AccessKeyUsagePrivateKey:
		if key.SshKey.Passphrase != "" {
			return fmt.Errorf("ssh key with passphrase not supported")
		}
		return ioutil.WriteFile(path, []byte(key.SshKey.PrivateKey + "\n"), 0600)
	case AccessKeyUsageVault:
		switch key.Type {
		case AccessKeyLoginPassword:
			return ioutil.WriteFile(path, []byte(key.LoginPassword.Password), 0600)
		}
	case AccessKeyUsageAnsibleBecomeUser:
		switch key.Type {
		case AccessKeyLoginPassword:
			content := make(map[string]string)
			content["ansible_become_user"] = key.LoginPassword.Login
			content["ansible_become_password"] = key.LoginPassword.Password
			var bytes []byte
			bytes, err = json.Marshal(content)
			if err != nil {
				return err
			}
			return ioutil.WriteFile(path, bytes, 0600)
		default:
			return fmt.Errorf("access key type not supported for ansible user")
		}
	case AccessKeyUsageAnsibleUser:
		switch key.Type {
		case AccessKeySSH:
			if key.SshKey.Passphrase != "" {
				return fmt.Errorf("ssh key with passphrase not supported")
			}
			return ioutil.WriteFile(path, []byte(key.SshKey.PrivateKey + "\n"), 0600)
		case AccessKeyLoginPassword:
			content := make(map[string]string)
			content["ansible_user"] = key.LoginPassword.Login
			content["ansible_password"] = key.LoginPassword.Password
			var bytes []byte
			bytes, err = json.Marshal(content)
			if err != nil {
				return err
			}
			return ioutil.WriteFile(path, bytes, 0600)

		default:
			return fmt.Errorf("access key type not supported for ansible user")
		}
	}

	return nil
}

// GetPath returns the location of the access key once written to disk
func (key AccessKey) GetPath() string {
	return util.Config.TmpPath + "/access_key_" + strconv.Itoa(key.ID)
}

func (key AccessKey) GetSshCommand() string {
	if key.Type != AccessKeySSH {
		panic("type must be ssh")
	}

	args := "ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i " + key.GetPath()
	if util.Config.SshConfigPath != "" {
		args += " -F " + util.Config.SshConfigPath
	}
	return args
}

func (key AccessKey) Validate(validateSecretFields bool) error {
	if key.Name == "" {
		return fmt.Errorf("name can not be empty")
	}

	if !validateSecretFields {
		return nil
	}

	switch key.Type {
	case AccessKeySSH:
		if key.SshKey.PrivateKey == "" {
			return fmt.Errorf("private key can not be empty")
		}
	case AccessKeyLoginPassword:
		if key.LoginPassword.Password == "" {
			return fmt.Errorf("password can not be empty")
		}
	}

	return nil
}

func (key *AccessKey) SerializeSecret() error {
	var plaintext []byte
	var err error

	switch key.Type {
	case AccessKeySSH:
		plaintext, err = json.Marshal(key.SshKey)
		if err != nil {
			return err
		}
	case AccessKeyLoginPassword:
		plaintext, err = json.Marshal(key.LoginPassword)
		if err != nil {
			return err
		}
	default:
		key.Secret = nil
		return nil
	}

	if util.Config.AccessKeyEncryption == "" {
		secret := base64.StdEncoding.EncodeToString(plaintext)
		key.Secret = &secret
		return nil
	}

	encryption, err := base64.StdEncoding.DecodeString(util.Config.AccessKeyEncryption)

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

func (key *AccessKey) unmarshalAppropriateField(secret []byte) (err error) {
	switch key.Type {
	case AccessKeySSH:
		sshKey := SshKey{}
		err = json.Unmarshal(secret, &sshKey)
		if err == nil {
			key.SshKey = sshKey
		}
	case AccessKeyLoginPassword:
		loginPass := LoginPassword{}
		err = json.Unmarshal(secret, &loginPass)
		if err == nil {
			key.LoginPassword = loginPass
		}
	}
	return
}

func (key *AccessKey) ResetSecret() {
	//key.Secret = nil
	key.LoginPassword = LoginPassword{}
	key.SshKey = SshKey{}
}

func (key *AccessKey) DeserializeSecret() error {
	if key.Secret == nil || *key.Secret == "" {
		return nil
	}

	ciphertext := []byte(*key.Secret)

	if ciphertext[len(*key.Secret)-1] == '\n' { // not encrypted private key, used for back compatibility
		if key.Type != AccessKeySSH {
			return fmt.Errorf("invalid access key type")
		}
		key.SshKey = SshKey{
			PrivateKey: *key.Secret,
		}
		return nil
	}

	ciphertext, err := base64.StdEncoding.DecodeString(*key.Secret)
	if err != nil {
		return err
	}

	if util.Config.AccessKeyEncryption == "" {
		err = key.unmarshalAppropriateField(ciphertext)
		if _, ok := err.(*json.SyntaxError); ok {
			err = fmt.Errorf("[ERR_INVALID_ENCRYPTION_KEY] Cannot decrypt access key, perhaps encryption key was changed")
		}
		return err
	}

	encryption, err := base64.StdEncoding.DecodeString(util.Config.AccessKeyEncryption)
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

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	ciphertext, err = gcm.Open(nil, nonce, ciphertext, nil)

	if err != nil {
		if err.Error() == "cipher: message authentication failed" {
			err = fmt.Errorf("[ERR_INVALID_ENCRYPTION_KEY] Cannot decrypt access key, perhaps encryption key was changed")
		}
		return err
	}

	return key.unmarshalAppropriateField(ciphertext)
}
