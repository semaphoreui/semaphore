package db_lib

import (
	"github.com/Digital-Data-Co/semaphore/db"
	"github.com/Digital-Data-Co/semaphore/pkg/ssh"
	"github.com/Digital-Data-Co/semaphore/pkg/task_logger"
)

type AccessKeyInstaller interface {
	Install(key db.AccessKey, usage db.AccessKeyRole, logger task_logger.Logger) (installation ssh.AccessKeyInstallation, err error)
}
