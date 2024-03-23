package lib

import (
	"os/exec"
	"strings"
	"time"
)

type TaskStatus string

const (
	TaskWaitingStatus       TaskStatus = "waiting"
	TaskStartingStatus      TaskStatus = "starting"
	TaskWaitingConfirmation TaskStatus = "waiting_confirmation"
	TaskConfirmed           TaskStatus = "confirmed"
	TaskRunningStatus       TaskStatus = "running"
	TaskStoppingStatus      TaskStatus = "stopping"
	TaskStoppedStatus       TaskStatus = "stopped"
	TaskSuccessStatus       TaskStatus = "success"
	TaskFailStatus          TaskStatus = "error"
)

func (s TaskStatus) IsNotifiable() bool {
	return s == TaskSuccessStatus || s == TaskFailStatus || s == TaskWaitingConfirmation
}

func (s TaskStatus) Format() (res string) {

	switch s {
	case TaskFailStatus:
		res += "❌"
	case TaskSuccessStatus:
		res += "✅"
	case TaskStoppedStatus:
		res += "⏹️"
	case TaskWaitingConfirmation:
		res += "⚠️"
	}
	res += strings.ToUpper(string(s))

	return
}

func (s TaskStatus) IsFinished() bool {
	return s == TaskStoppedStatus || s == TaskSuccessStatus || s == TaskFailStatus
}

type Logger interface {
	Log(msg string)
	Log2(msg string, now time.Time)
	LogCmd(cmd *exec.Cmd)
	SetStatus(status TaskStatus)
}
