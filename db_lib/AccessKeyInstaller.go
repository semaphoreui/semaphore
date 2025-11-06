package db_lib

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/ssh"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
)

type AccessKeyInstaller interface {
	Install(key db.AccessKey, usage db.AccessKeyRole, logger task_logger.Logger) (installation ssh.AccessKeyInstallation, err error)
}
