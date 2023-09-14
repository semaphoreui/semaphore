package lib

import (
	"github.com/ansible-semaphore/semaphore/db"
	"os/exec"
	"time"
)

type Logger interface {
	Log(msg string)
	Log2(msg string, now time.Time)
	LogCmd(cmd *exec.Cmd)
	SetStatus(status db.TaskStatus)
}
