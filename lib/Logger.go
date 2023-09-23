package lib

import (
	"os/exec"
	"time"
)

type TaskStatus string

const (
	TaskWaitingStatus  TaskStatus = "waiting"
	TaskStartingStatus TaskStatus = "starting"
	TaskRunningStatus  TaskStatus = "running"
	TaskStoppingStatus TaskStatus = "stopping"
	TaskStoppedStatus  TaskStatus = "stopped"
	TaskSuccessStatus  TaskStatus = "success"
	TaskFailStatus     TaskStatus = "error"
)

func (s TaskStatus) IsFinished() bool {
	return s == TaskStoppedStatus || s == TaskSuccessStatus || s == TaskFailStatus
}

type Logger interface {
	Log(msg string)
	Log2(msg string, now time.Time)
	LogCmd(cmd *exec.Cmd)
	SetStatus(status TaskStatus)
}
