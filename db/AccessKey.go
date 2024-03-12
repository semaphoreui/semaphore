package db

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/ansible-semaphore/semaphore/lib"
	"io"
	"math/big"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/ansible-semaphore/semaphore/util"
)

type AccessKeyType string

const (
	AccessKeySSH           AccessKeyType = "ssh"
	AccessKeyNone          AccessKeyType = "none"
	AccessKeyLoginPassword AccessKeyType = "login_password"
)

// AccessKey represents a key used to access a machine with ansible from semaphore
type AccessKey struct {
	ID   int    `db:"id" json:"id"`
	Name string `db:"name" json:"name" binding:"required"`
	// 'ssh/login_password/none'
	Type AccessKeyType `db:"type" json:"type" binding:"required"`

	ProjectID *int `db:"project_id" json:"project_id"`

	// Secret used internally, do not assign this field.
	// You should use methods SerializeSecret to fill this field.
	Secret *string `db:"secret" json:"-"`

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

type AccessKeyRole int

const (
	AccessKeyRoleAnsibleUser = iota
	AccessKeyRoleAnsibleBecomeUser
	AccessKeyRoleAnsiblePasswordVault
	AccessKeyRoleGit
)

type AccessKeyInstallation struct {
	InstallationKey int64
	SshAgent        *lib.SshAgent
}

func (key AccessKeyInstallation) Destroy() error {
	if key.SshAgent != nil {
		return key.SshAgent.Close()
	}

	installPath := key.GetPath()
	_, err := os.Stat(installPath)
	if os.IsNotExist(err) {
		return nil
	}
	return os.Remove(installPath)
}

// GetPath returns the location of the access key once written to disk
func (key AccessKeyInstallation) GetPath() string {
	return util.Config.TmpPath + "/access_key_" + strconv.FormatInt(key.InstallationKey, 10)
}

func (key *AccessKey) startSshAgent(logger lib.Logger) (lib.SshAgent, error) {

	sshAgent := lib.SshAgent{
		Logger: logger,
		Keys: []lib.SshAgentKey{
			{
				Key:        []byte(key.SshKey.PrivateKey),
				Passphrase: []byte(key.SshKey.Passphrase),
			},
		},
		SocketFile: path.Join(util.Config.TmpPath, fmt.Sprintf("ssh-agent-%d-%d.sock", key.ID, time.Now().Unix())),
	}

	return sshAgent, sshAgent.Listen()
}

func (key *AccessKey) Install(usage AccessKeyRole, logger lib.Logger) (installation AccessKeyInstallation, err error) {
	rnd, err := rand.Int(rand.Reader, big.NewInt(1000000000))
	if err != nil {
		return
	}

	installation.InstallationKey = rnd.Int64()

	if key.Type == AccessKeyNone {
		return
	}

	installationPath := installation.GetPath()

	err = key.DeserializeSecret()

	if err != nil {
		return
	}

	switch usage {
	case AccessKeyRoleGit:
		switch key.Type {
		case AccessKeySSH:
			var agent lib.SshAgent
			agent, err = key.startSshAgent(logger)
			installation.SshAgent = &agent

			//err = os.WriteFile(installationPath, []byte(key.SshKey.PrivateKey+"\n"), 0600)
		}
	case AccessKeyRoleAnsiblePasswordVault:
		switch key.Type {
		case AccessKeyLoginPassword:
			err = os.WriteFile(installationPath, []byte(key.LoginPassword.Password), 0600)
		}
	case AccessKeyRoleAnsibleBecomeUser:
		switch key.Type {
		case AccessKeyLoginPassword:
			content := make(map[string]string)
			if len(key.LoginPassword.Login) > 0 {
				content["ansible_become_user"] = key.LoginPassword.Login
			}
			content["ansible_become_password"] = key.LoginPassword.Password
			var bytes []byte
			bytes, err = json.Marshal(content)
			if err != nil {
				return
			}
			err = os.WriteFile(installationPath, bytes, 0600)
		default:
			err = fmt.Errorf("access key type not supported for ansible user")
		}
	case AccessKeyRoleAnsibleUser:
		switch key.Type {
		case AccessKeySSH:
			var agent lib.SshAgent
			agent, err = key.startSshAgent(logger)
			installation.SshAgent = &agent
			//err = os.WriteFile(installationPath, []byte(key.SshKey.PrivateKey+"\n"), 0600)
		case AccessKeyLoginPassword:
			content := make(map[string]string)
			content["ansible_user"] = key.LoginPassword.Login
			content["ansible_password"] = key.LoginPassword.Password
			var bytes []byte
			bytes, err = json.Marshal(content)
			if err != nil {
				return
			}
			err = os.WriteFile(installationPath, bytes, 0600)

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
