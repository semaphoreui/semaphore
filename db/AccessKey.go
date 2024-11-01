package db

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/semaphoreui/semaphore/pkg/random"
	"github.com/semaphoreui/semaphore/pkg/ssh"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
	"io"
	"path"
)

type AccessKeyType string

const (
	AccessKeySSH           AccessKeyType = "ssh"
	AccessKeyNone          AccessKeyType = "none"
	AccessKeyLoginPassword AccessKeyType = "login_password"
	AccessKeyString        AccessKeyType = "string"
)

// AccessKey represents a key used to access a machine with ansible from semaphore
type AccessKey struct {
	ID   int    `db:"id" json:"id" backup:"-"`
	Name string `db:"name" json:"name" binding:"required"`
	// 'ssh/login_password/none'
	Type AccessKeyType `db:"type" json:"type" binding:"required"`

	ProjectID *int `db:"project_id" json:"project_id" backup:"-"`

	// Secret used internally, do not assign this field.
	// You should use methods SerializeSecret to fill this field.
	Secret *string `db:"secret" json:"-" backup:"-"`

	String         string        `db:"-" json:"string"`
	LoginPassword  LoginPassword `db:"-" json:"login_password"`
	SshKey         SshKey        `db:"-" json:"ssh"`
	OverrideSecret bool          `db:"-" json:"override_secret"`

	// EnvironmentID is an ID of environment which owns the access key.
	EnvironmentID *int `db:"environment_id" json:"-" backup:"-"`

	// UserID is an ID of user which owns the access key.
	UserID *int `db:"user_id" json:"-" backup:"-"`

	Empty bool `db:"-" json:"empty,omitempty"`
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

type AccessKeyRole int

const (
	AccessKeyRoleAnsibleUser = iota
	AccessKeyRoleAnsibleBecomeUser
	AccessKeyRoleAnsiblePasswordVault
	AccessKeyRoleGit
)

type AccessKeyInstallation struct {
	SSHAgent *ssh.Agent
	Login    string
	Password string
	Script   string
}

func (key AccessKeyInstallation) Destroy() error {
	if key.SSHAgent != nil {
		return key.SSHAgent.Close()
	}
	return nil
}

func (key *AccessKey) startSSHAgent(logger task_logger.Logger) (ssh.Agent, error) {
	sshAgent := ssh.Agent{
		Logger: logger,
		Keys: []ssh.AgentKey{
			{
				Key:        []byte(key.SshKey.PrivateKey),
				Passphrase: []byte(key.SshKey.Passphrase),
			},
		},
		SocketFile: path.Join(util.Config.TmpPath, fmt.Sprintf("ssh-agent-%d-%s.sock", key.ID, random.String(10))),
	}

	return sshAgent, sshAgent.Listen()
}

func (key *AccessKey) Install(usage AccessKeyRole, logger task_logger.Logger) (installation AccessKeyInstallation, err error) {

	if key.Type == AccessKeyNone {
		return
	}

	err = key.DeserializeSecret()

	if err != nil {
		return
	}

	switch usage {
	case AccessKeyRoleGit:
		switch key.Type {
		case AccessKeySSH:
			var agent ssh.Agent
			agent, err = key.startSSHAgent(logger)
			installation.SSHAgent = &agent
			installation.Login = key.SshKey.Login
		}
	case AccessKeyRoleAnsiblePasswordVault:
		switch key.Type {
		case AccessKeyLoginPassword:
			installation.Password = key.LoginPassword.Password
		default:
			err = fmt.Errorf("access key type not supported for ansible password vault")
		}
	case AccessKeyRoleAnsibleBecomeUser:
		if key.Type != AccessKeyLoginPassword {
			err = fmt.Errorf("access key type not supported for ansible become user")
		}
		installation.Login = key.LoginPassword.Login
		installation.Password = key.LoginPassword.Password
	case AccessKeyRoleAnsibleUser:
		switch key.Type {
		case AccessKeySSH:
			var agent ssh.Agent
			agent, err = key.startSSHAgent(logger)
			installation.SSHAgent = &agent
			installation.Login = key.SshKey.Login
		case AccessKeyLoginPassword:
			installation.Login = key.LoginPassword.Login
			installation.Password = key.LoginPassword.Password
		default:
			err = fmt.Errorf("access key type not supported for ansible user")
		}
	}

	return
}

func (key *AccessKey) Validate(validateSecretFields bool) error {
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
	case AccessKeyString:
		if key.String == "" {
			key.Secret = nil
			return nil
		}
		plaintext = []byte(key.String)
	case AccessKeySSH:
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
	case AccessKeyLoginPassword:
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
	case AccessKeyNone:
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

func (key *AccessKey) unmarshalAppropriateField(secret []byte) (err error) {
	switch key.Type {
	case AccessKeyString:
		key.String = string(secret)
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

func (key *AccessKey) DeserializeSecret() error {
	return key.DeserializeSecret2(util.Config.AccessKeyEncryption)
}

func (key *AccessKey) DeserializeSecret2(encryptionString string) error {
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

	if encryptionString == "" {
		err = key.unmarshalAppropriateField(ciphertext)
		if _, ok := err.(*json.SyntaxError); ok {
			err = fmt.Errorf("secret must be valid json in key '%s'", key.Name)
		}
		return err
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

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	ciphertext, err = gcm.Open(nil, nonce, ciphertext, nil)

	if err != nil {
		if err.Error() == "cipher: message authentication failed" {
			err = fmt.Errorf("cannot decrypt access key, perhaps encryption key was changed")
		}
		return err
	}

	return key.unmarshalAppropriateField(ciphertext)
}
