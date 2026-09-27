package db_lib

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/ssh"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
)

// SecretDeserializer decrypts the secret of an access key. Declared here rather
// than taken from services/server so db_lib does not depend on it.
type SecretDeserializer interface {
	DeserializeSecret(key *db.AccessKey) error
}

// InstallProjectHostConfigs loads the credential mappings of a project and
// generates the ssh config and git rewrites they describe.
//
// Every git operation of a project has to go through the same mappings, not only
// the ones a task runs, so the callers outside the task pipeline — listing the
// branches of a repository, polling a schedule for new commits — load them here.
//
// The returned installation is nil when the project has no mappings, and
// Destroy is safe to call on it either way.
func InstallProjectHostConfigs(
	store db.Store,
	deserializer SecretDeserializer,
	projectID int,
	logger task_logger.Logger,
) (*ssh.HostConfigInstallation, error) {

	hostConfigs, err := store.GetHostConfigs(projectID, db.RetrieveQueryParams{})
	if err != nil {
		return nil, err
	}

	for i := range hostConfigs {
		hostConfigs[i].SSHKey, err = store.GetAccessKey(projectID, hostConfigs[i].SSHKeyID)
		if err != nil {
			return nil, err
		}

		if err = deserializer.DeserializeSecret(&hostConfigs[i].SSHKey); err != nil {
			return nil, err
		}
	}

	return ssh.InstallHostConfigs(projectID, hostConfigs, logger)
}
