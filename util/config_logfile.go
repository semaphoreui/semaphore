//go:build !pro

package util

func (e *EventLogType) Write(event EventLogRecord) error {
	return nil
}

func (e *TaskLogType) Write(task TaskLogRecord) error {
	return nil
}
