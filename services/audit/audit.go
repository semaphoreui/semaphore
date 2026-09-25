package audit

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pro_interfaces"
)

type Resource string

const (
	ResourceProject    Resource = "project"
	ResourceInventory  Resource = "inventory"
	ResourceCredential Resource = "credential"
	ResourceRepository Resource = "repository"
	ResourceView       Resource = "view"
	ResourceTemplate   Resource = "template"
	ResourceSchedule   Resource = "schedule"
)

type Action string

const (
	ActionCreate Action = "create"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
)

type Actor struct {
	ID   int
	Name string
}

type Request struct {
	ID        string
	SourceIP  string
	UserAgent string
}

type ResourceEvent struct {
	Resource    Resource
	Action      Action
	Actor       Actor
	Request     Request
	ProjectID   int
	TargetID    int
	TargetName  string
	Description string
}

type Settings struct {
	Enabled    bool
	InstanceID string
	NodeID     string
}

type resourceMapping struct {
	activityType db.EventObjectType
	eventCode    string
	targetType   string
}

func mapResource(resource Resource) (resourceMapping, error) {
	switch resource {
	case ResourceProject:
		return resourceMapping{
			activityType: db.EventProject,
			eventCode:    db.AuditEventCodeProject,
			targetType:   "project",
		}, nil
	case ResourceInventory:
		return resourceMapping{
			activityType: db.EventInventory,
			eventCode:    db.AuditEventCodeInventory,
			targetType:   "inventory",
		}, nil
	case ResourceCredential:
		return resourceMapping{
			activityType: db.EventKey,
			eventCode:    db.AuditEventCodeCredential,
			targetType:   "credential",
		}, nil
	case ResourceRepository:
		return resourceMapping{
			activityType: db.EventRepository,
			eventCode:    db.AuditEventCodeRepository,
			targetType:   "repository",
		}, nil
	case ResourceView:
		return resourceMapping{
			activityType: db.EventView,
			eventCode:    db.AuditEventCodeView,
			targetType:   "view",
		}, nil
	case ResourceTemplate:
		return resourceMapping{
			activityType: db.EventTemplate,
			eventCode:    db.AuditEventCodeTemplate,
			targetType:   "template",
		}, nil
	case ResourceSchedule:
		return resourceMapping{
			activityType: db.EventSchedule,
			eventCode:    db.AuditEventCodeSchedule,
			targetType:   "schedule",
		}, nil
	default:
		return resourceMapping{}, fmt.Errorf("unsupported audit resource %q", resource)
	}
}

type actionMapping struct {
	eventType string
	action    string
}

func mapAction(action Action) (actionMapping, error) {
	switch action {
	case ActionCreate:
		return actionMapping{db.AuditTypeCreation, db.AuditActionCreate}, nil
	case ActionUpdate:
		return actionMapping{db.AuditTypeChange, db.AuditActionUpdate}, nil
	case ActionDelete:
		return actionMapping{db.AuditTypeDeletion, db.AuditActionDelete}, nil
	default:
		return actionMapping{}, fmt.Errorf("unsupported audit action %q", action)
	}
}

type Service struct {
	store     db.Store
	logWriter pro_interfaces.LogWriteService
	settings  Settings
}

func NewService(
	store db.Store,
	logWriter pro_interfaces.LogWriteService,
	settings Settings,
) *Service {
	return &Service{
		store:     store,
		logWriter: logWriter,
		settings:  settings,
	}
}

func (s *Service) RecordResource(event ResourceEvent) error {
	resourceFields, err := mapResource(event.Resource)
	if err != nil {
		return err
	}
	actionFields, err := mapAction(event.Action)
	if err != nil {
		return err
	}
	activityEvent := db.Event{
		ObjectID:    &event.TargetID,
		ObjectType:  &resourceFields.activityType,
		Description: &event.Description,
	}
	if event.Actor.ID > 0 {
		activityEvent.UserID = &event.Actor.ID
	}
	if event.ProjectID > 0 {
		activityEvent.ProjectID = &event.ProjectID
	}

	_, eventStoreErr := s.store.CreateEvent(activityEvent)
	eventLogErr := s.logWriter.WriteEventLog(pro_interfaces.EventLogRecord{
		Action:        string(event.Action),
		ProjectID:     activityEvent.ProjectID,
		UserID:        activityEvent.UserID,
		IntegrationID: activityEvent.IntegrationID,
		Description:   activityEvent.Description,
	})

	var auditStoreErr error
	if s.settings.Enabled {
		auditEvent := db.NewAuditEvent()
		auditEvent.Category = db.AuditCategoryResource
		auditEvent.Outcome = db.AuditOutcomeSuccess
		auditEvent.Actor = &db.AuditActor{
			Type: "user",
			ID:   strconv.Itoa(event.Actor.ID),
			Name: event.Actor.Name,
		}
		auditEvent.Source = &db.AuditSource{
			IP:        event.Request.SourceIP,
			UserAgent: event.Request.UserAgent,
		}
		auditEvent.Target = &db.AuditTarget{
			Type: resourceFields.targetType,
			ID:   strconv.Itoa(event.TargetID),
			Name: event.TargetName,
		}
		auditEvent.Scope = &db.AuditScope{ProjectID: strconv.Itoa(event.ProjectID)}
		auditEvent.RequestID = event.Request.ID
		auditEvent.InstanceID = s.settings.InstanceID
		auditEvent.NodeID = s.settings.NodeID

		auditEvent.EventCode = resourceFields.eventCode
		auditEvent.Type = actionFields.eventType
		auditEvent.Action = actionFields.action

		_, auditStoreErr = s.store.CreateAuditEvent(auditEvent)
	}

	return errors.Join(eventStoreErr, eventLogErr, auditStoreErr)
}
