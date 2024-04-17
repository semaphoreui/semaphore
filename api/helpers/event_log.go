package helpers

import (
	"net/http"

	"github.com/ansible-semaphore/semaphore/db"
	log "github.com/sirupsen/logrus"
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

func EventLog(r *http.Request, action EventLogType, event EventLogItem) {
	record := db.Event{
		ObjectType:  &event.ObjectType,
		ObjectID:    &event.ObjectID,
		Description: &event.Description,
	}

	if event.IntegrationID > 0 {
		record.IntegrationID = &event.IntegrationID
	}

	if event.UserID > 0 {
		record.UserID = &event.UserID
	}

	if event.ProjectID > 0 {
		record.ProjectID = &event.ProjectID
	}

	if _, err := Store(r).CreateEvent(record); err != nil {
		log.WithFields(log.Fields{
			"integration": event.IntegrationID,
			"user":        event.UserID,
			"project":     event.ProjectID,
			"type":        string(event.ObjectType),
			"object":      event.ObjectID,
			"action":      string(action),
		}).Error("Failed to store event")
	}
}
