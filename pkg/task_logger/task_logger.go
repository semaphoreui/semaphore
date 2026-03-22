package task_logger

import (
	"os/exec"
	"time"
)

type TaskStatus string

const (
	TaskWaitingStatus       TaskStatus = "waiting"
	TaskStartingStatus      TaskStatus = "starting"
	TaskWaitingConfirmation TaskStatus = "waiting_confirmation"
	TaskConfirmed           TaskStatus = "confirmed"
	TaskRejected            TaskStatus = "rejected"
	TaskRunningStatus       TaskStatus = "running"
	TaskStoppingStatus      TaskStatus = "stopping"
	TaskStoppedStatus       TaskStatus = "stopped"
	TaskSuccessStatus       TaskStatus = "success"
	TaskFailStatus          TaskStatus = "error"
)

func UnfinishedTaskStatuses() []TaskStatus {
	return []TaskStatus{
		TaskWaitingStatus,
		TaskStartingStatus,
		TaskWaitingConfirmation,
		TaskConfirmed,
		TaskRejected,
		TaskRunningStatus,
		TaskStoppingStatus,
	}
}

func (s TaskStatus) IsValid() bool {
	switch s {
	case TaskWaitingStatus,
		TaskStartingStatus,
		TaskWaitingConfirmation,
		TaskConfirmed,
		TaskRejected,
		TaskRunningStatus,
		TaskStoppingStatus,
		TaskStoppedStatus,
		TaskSuccessStatus,
		TaskFailStatus:
		return true
	}
	return false
}

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
	default:
		res += "❓"
	}

	switch s {
	case TaskWaitingStatus:
		res += " WAITING"
	case TaskStartingStatus:
		res += " STARTING"
	case TaskWaitingConfirmation:
		res += " WAITING_CONFIRMATION"
	case TaskConfirmed:
		res += " CONFIRMED"
	case TaskRejected:
		res += " REJECTED"
	case TaskRunningStatus:
		res += " RUNNING"
	case TaskStoppingStatus:
		res += " STOPPING"
	case TaskStoppedStatus:
		res += " STOPPED"
	case TaskSuccessStatus:
		res += " SUCCESS"
	case TaskFailStatus:
		res += " ERROR"
	default:
		res += " UNKNOWN"
	}

	return
}

func (s TaskStatus) IsFinished() bool {
	return s == TaskStoppedStatus || s == TaskSuccessStatus || s == TaskFailStatus
}

// TaskStatusProgressRank orders statuses for comparing how far execution has progressed
// (higher = further). Used in HA / remote-runner paths to avoid regressing status when
// merging DB state with stale runtime-store snapshots.
func TaskStatusProgressRank(s TaskStatus) int {
	switch s {
	case TaskWaitingStatus:
		return 10
	case TaskStartingStatus:
		return 20
	case TaskWaitingConfirmation:
		return 25
	case TaskConfirmed, TaskRejected:
		return 26
	case TaskRunningStatus:
		return 40
	case TaskStoppingStatus:
		return 50
	case TaskStoppedStatus, TaskSuccessStatus, TaskFailStatus:
		return 100
	default:
		return 0
	}
}

type StatusListener func(status TaskStatus)
type LogListener func(new time.Time, msg string)

type Logger interface {
	Log(msg string)
	Logf(format string, a ...any)
	LogWithTime(now time.Time, msg string)
	LogfWithTime(now time.Time, format string, a ...any)
	LogCmd(cmd *exec.Cmd)
	SetStatus(status TaskStatus)
	AddStatusListener(l StatusListener)
	AddLogListener(l LogListener)

	SetCommit(hash, message string)

	WaitLog()
}
