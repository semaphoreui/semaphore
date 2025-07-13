//go:build pro

package util

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_WriteEventLog_DisabledEventLog_ReturnsNil(t *testing.T) {
	eventLog := &EventLogType{Enabled: false}
	err := eventLog.Write(EventLogRecord{})
	assert.NoError(t, err)
}

func Test_WriteEventLog_NilLogFileWriter_ReturnsError(t *testing.T) {
	eventLog := &EventLogType{Enabled: true, Format: FileLogFormat("invalid")}
	err := eventLog.Write(EventLogRecord{})
	assert.Error(t, err)
	assert.Equal(t, "log file writer is nil", err.Error())
}

func Test_WriteTaskLog_DisabledTaskLog_ReturnsNil(t *testing.T) {
	taskLog := &TaskLogType{Enabled: false}
	err := taskLog.Write(TaskLogRecord{})
	assert.NoError(t, err)
}

func Test_WriteTaskLog_NilLogFileWriter_ReturnsError(t *testing.T) {
	taskLog := &TaskLogType{Enabled: true, Format: FileLogFormat("invalid")}
	err := taskLog.Write(TaskLogRecord{})
	assert.Error(t, err)
	assert.Equal(t, "log file writer is nil", err.Error())
}

func Test_RawLogWriter_NilLogger_ReturnsError(t *testing.T) {
	writer := rawLogWriter{}
	err := writer.Write(log.Fields{"key": "value"}, nil)
	assert.Error(t, err)
	assert.Equal(t, "logger is not initialized", err.Error())
}

func Test_JsonLogWriter_NilLogger_ReturnsError(t *testing.T) {
	writer := jsonLogWriter{}
	err := writer.Write(log.Fields{"key": "value"}, nil)
	assert.Error(t, err)
	assert.Equal(t, "logger is not initialized", err.Error())
}

func Test_GetLogFileWriter_InvalidFormat_ReturnsNil(t *testing.T) {
	writer := getLogFileWriter(FileLogFormat("invalid"))
	assert.Nil(t, writer)
}

func Test_GetLogFileWriter_ValidRawFormat_ReturnsRawLogWriter(t *testing.T) {
	writer := getLogFileWriter(FileLogRaw)
	_, ok := writer.(rawLogWriter)
	assert.True(t, ok)
}

func Test_GetLogFileWriter_ValidJSONFormat_ReturnsJsonLogWriter(t *testing.T) {
	writer := getLogFileWriter(FileLogJSON)
	_, ok := writer.(jsonLogWriter)
	assert.True(t, ok)
}
