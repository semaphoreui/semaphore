package helpers

import (
	"net/http"
	"strconv"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	"github.com/semaphoreui/semaphore/util"
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

type AuditResourceKind string

const (
	AuditResourceProject   AuditResourceKind = "project"
	AuditResourceInventory AuditResourceKind = "inventory"
)

type AuditResourceEventItem struct {
	Resource   AuditResourceKind
	Action     EventLogType
	TargetID   string
	TargetName string
	ProjectID  int
}

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

// AuditResourceEvent records a best-effort canonical resource audit event after a mutation succeeds.
func AuditResourceEvent(
	r *http.Request,
	store db.AuditEventStore,
	item AuditResourceEventItem,
) {
	if util.Config == nil || util.Config.Audit == nil || !util.Config.Audit.Enabled {
		return
	}

	requestContext, ok := AuditRequestContextFrom(r)
	if !ok {
		log.Error("Failed to record audit event: missing audit request context")
		return
	}
	userValue, ok := GetOkFromContext(r, "user")
	user, ok := userValue.(*db.User)
	if !ok || user == nil {
		log.Error("Failed to record audit event: missing authenticated user")
		return
	}
	if store == nil {
		log.Error("Failed to record audit event: missing audit store")
		return
	}

	eventCode, targetType, ok := auditResourceFields(item.Resource)
	if !ok {
		log.Error("Failed to record audit event: unsupported resource")
		return
	}

	eventType, auditAction, ok := auditResourceAction(item.Action)
	if !ok {
		log.Error("Failed to record audit event: unsupported action")
		return
	}

	event := db.NewAuditEvent()
	event.EventCode = eventCode
	event.Category = db.AuditCategoryResource
	event.Type = eventType
	event.Action = auditAction
	event.Outcome = db.AuditOutcomeSuccess
	event.Actor = &db.AuditActor{Type: "user", ID: strconv.Itoa(user.ID), Name: user.Username}
	event.Source = &db.AuditSource{IP: requestContext.SourceIP, UserAgent: requestContext.UserAgent}
	event.Target = &db.AuditTarget{Type: targetType, ID: item.TargetID, Name: item.TargetName}
	event.Scope = &db.AuditScope{ProjectID: strconv.Itoa(item.ProjectID)}
	event.RequestID = requestContext.RequestID
	event.InstanceID = util.Config.Audit.InstanceID
	if util.Config.HA != nil {
		event.NodeID = util.Config.HA.NodeID
	}

	if err := event.ValidateResourceEvent(); err != nil {
		log.Error("Failed to validate audit event")
		return
	}
	if _, err := store.CreateAuditEvent(event); err != nil {
		log.Error("Failed to store audit event")
	}
}

func auditResourceFields(resource AuditResourceKind) (eventCode string, targetType string, ok bool) {
	switch resource {
	case AuditResourceProject:
		return db.AuditEventCodeProject, "project", true
	case AuditResourceInventory:
		return db.AuditEventCodeInventory, "inventory", true
	default:
		return "", "", false
	}
}

func auditResourceAction(action EventLogType) (eventType string, auditAction string, ok bool) {
	switch action {
	case EventLogCreate:
		return db.AuditTypeCreation, db.AuditActionCreate, true
	case EventLogUpdate:
		return db.AuditTypeChange, db.AuditActionUpdate, true
	case EventLogDelete:
		return db.AuditTypeDeletion, db.AuditActionDelete, true
	default:
		return "", "", false
	}
}
