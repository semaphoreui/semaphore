//go:build pro

package util

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"sort"
	"strings"
	"time"
)

func structToMap(obj interface{}) map[string]interface{} {
	b, err := json.Marshal(obj)
	if err != nil {
		return nil
	}

	// 2) unmarshal those bytes into a map
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil
	}
	return m
}

func (e *EventLogType) Write(event EventLogRecord) error {
	if !e.Enabled {
		return nil
	}

	w := getLogFileWriter(e.Format)
	if w == nil {
		return errors.New("log file writer is nil")
	}

	return w.Write(structToMap(event), e.Logger)
}

func (e *TaskLogType) Write(task TaskLogRecord) error {
	if !e.Enabled {
		return nil
	}

	w := getLogFileWriter(e.Format)
	if w == nil {
		return errors.New("log file writer is nil")
	}

	return w.Write(structToMap(task), e.Logger)
}

func (e *TaskLogType) WriteResult(task any) error {
	if !e.Enabled {
		return nil
	}

	w := getLogFileWriter(FileLogJSON)
	if w == nil {
		return errors.New("log file writer is nil")
	}

	res := structToMap(task)

	return w.Write(res, e.ResultLogger)
}

type logFileWriter interface {
	Write(row log.Fields, logger *lumberjack.Logger) error
}

type rawLogWriter struct {
}

func (w rawLogWriter) Write(row log.Fields, logger *lumberjack.Logger) error {

	if logger == nil {
		return errors.New("logger is not initialized")
	}

	keys := make([]string, 0, len(row))
	for key := range row {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Build the log line from the event's key=value pairs.
	logLine := time.Now().UTC().Format(time.RFC3339) + " "
	for _, key := range keys {
		val := fmt.Sprintf("%v", row[key])
		if strings.Index(val, " ") != -1 {
			val = "\"" + val + "\""
		}
		logLine += fmt.Sprintf("%s=%s ", key, val)
	}

	// Remove the trailing space and add a newline.
	if len(logLine) > 0 {
		logLine = logLine[:len(logLine)-1]
	}
	logLine += "\n"

	_, err := logger.Write([]byte(logLine))
	return err
}

type jsonLogWriter struct {
}

func (w jsonLogWriter) Write(row log.Fields, logger *lumberjack.Logger) error {

	if logger == nil {
		return errors.New("logger is not initialized")
	}

	row["time"] = time.Now().UTC().Format(time.RFC3339)

	s, err := json.Marshal(row)

	if err != nil {
		return err
	}

	_, err = logger.Write([]byte(string(s) + "\n"))

	return err
}

func getLogFileWriter(format FileLogFormat) logFileWriter {
	switch format {
	case FileLogRaw:
		return rawLogWriter{}
	case FileLogJSON:
		return jsonLogWriter{}
	}
	return nil
}
