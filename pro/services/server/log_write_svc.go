package server

import (
	"github.com/semaphoreui/semaphore/pro_interfaces"
)

type LogWriteServiceImpl struct {
}

// NewLogWriteService creates a new instance of LogWriteServiceImpl.
func NewLogWriteService() pro_interfaces.LogWriteService {
	return &LogWriteServiceImpl{}
}

func (l *LogWriteServiceImpl) WriteEventLog(event pro_interfaces.EventLogRecord) error {
	return nil
}

func (l *LogWriteServiceImpl) WriteTaskLog(task pro_interfaces.TaskLogRecord) error {
	return nil
}
func (l *LogWriteServiceImpl) WriteResult(task any) error {
	return nil
}
