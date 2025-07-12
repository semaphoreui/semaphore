package server

import (
	"fmt"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/random"
	"github.com/semaphoreui/semaphore/pkg/ssh"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
	"path"
)

type AccessKeyInstallationService interface {
	Install(key db.AccessKey, usage db.AccessKeyRole, logger task_logger.Logger) (installation db.AccessKeyInstallation, err error)
}

func NewAccessKeyInstallationService(encryptionService AccessKeyEncryptionService) AccessKeyInstallationService {
	return &AccessKeyInstallationServiceImpl{
		encryptionService: encryptionService,
	}
}

type AccessKeyInstallationServiceImpl struct {
	encryptionService AccessKeyEncryptionService
}

func startSSHAgent(key db.AccessKey, logger task_logger.Logger) (ssh.Agent, error) {

	socketFilename := fmt.Sprintf("ssh-agent-%d-%s.sock", key.ID, random.String(10))

	var socketFile string

	if key.ProjectID == nil {
		socketFile = path.Join(util.Config.TmpPath, socketFilename)
	} else {
		socketFile = path.Join(util.Config.GetProjectTmpDir(*key.ProjectID), socketFilename)
	}

	sshAgent := ssh.Agent{
		Logger: logger,
		Keys: []ssh.AgentKey{
			{
				Key:        []byte(key.SshKey.PrivateKey),
				Passphrase: []byte(key.SshKey.Passphrase),
			},
		},
		SocketFile: socketFile,
	}

	return sshAgent, sshAgent.Listen()
}

func (s *AccessKeyInstallationServiceImpl) Install(key db.AccessKey, usage db.AccessKeyRole, logger task_logger.Logger) (installation db.AccessKeyInstallation, err error) {

	if key.Type == db.AccessKeyNone {
		return
	}

	err = s.encryptionService.DeserializeSecret(&key)

	if err != nil {
		return
	}

	switch usage {
	case db.AccessKeyRoleGit:
		switch key.Type {
		case db.AccessKeySSH:
			var agent ssh.Agent
			agent, err = startSSHAgent(key, logger)
			installation.SSHAgent = &agent
			installation.Login = key.SshKey.Login
		}
	case db.AccessKeyRoleAnsiblePasswordVault:
		switch key.Type {
		case db.AccessKeyLoginPassword:
			installation.Password = key.LoginPassword.Password
		default:
			err = fmt.Errorf("access key type not supported for ansible password vault")
		}
	case db.AccessKeyRoleAnsibleBecomeUser:
		if key.Type != db.AccessKeyLoginPassword {
			err = fmt.Errorf("access key type not supported for ansible become user")
		}
		installation.Login = key.LoginPassword.Login
		installation.Password = key.LoginPassword.Password
	case db.AccessKeyRoleAnsibleUser:
		switch key.Type {
		case db.AccessKeySSH:
			var agent ssh.Agent
			agent, err = startSSHAgent(key, logger)
			installation.SSHAgent = &agent
			installation.Login = key.SshKey.Login
		case db.AccessKeyLoginPassword:
			installation.Login = key.LoginPassword.Login
			installation.Password = key.LoginPassword.Password
		default:
			err = fmt.Errorf("access key type not supported for ansible user")
		}
	}

	return
}
