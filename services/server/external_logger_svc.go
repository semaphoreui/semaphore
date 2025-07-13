package server

import "github.com/semaphoreui/semaphore/util"

type ExternalLoggerService interface {
	WriteEventLog(event util.EventLogRecord) error
	WriteTaskLog(event util.TaskLogType) error
}

type ExternalLoggerServiceImpl struct {
}
