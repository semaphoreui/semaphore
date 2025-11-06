package helpers

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type EventLogItem struct {
	IntegrationID int
	UserID        int
	ProjectID     int

	ObjectType  db.EventObjectType
	ObjectID    int
	Description string
}

type EventLogType string

const (
	EventLogCreate EventLogType = "create"
	EventLogUpdate EventLogType = "update"
	EventLogDelete EventLogType = "delete"
)

func EventLog(r *http.Request, action EventLogType, item EventLogItem) {
	event := db.Event{
		ObjectType:  &item.ObjectType,
		ObjectID:    &item.ObjectID,
		Description: &item.Description,
	}

	if item.IntegrationID > 0 {
		event.IntegrationID = &item.IntegrationID
	}

	if item.UserID > 0 {
		event.UserID = &item.UserID
	}

	if item.ProjectID > 0 {
		event.ProjectID = &item.ProjectID
	}

	logFields := event.ToFields()
	logFields["action"] = string(action)

	if _, err := Store(r).CreateEvent(event); err != nil {
		log.WithFields(logFields).Error("Failed to store event")
	}

	logWriter := GetFromContext(r, "log_writer").(pro_interfaces.LogWriteService)

	if err := logWriter.WriteEventLog(pro_interfaces.EventLogRecord{
		Action:        string(action),
		ProjectID:     event.ProjectID,
		UserID:        event.UserID,
		IntegrationID: event.IntegrationID,
		Description:   event.Description,
	}); err != nil {
		log.WithFields(logFields).Error("Failed to store event in log file")
	}
}
