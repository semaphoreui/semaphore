package audit

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pro_interfaces"
)

type Resource string

const (
	ResourceProject       Resource = "project"
	ResourceInventory     Resource = "inventory"
	ResourceCredential    Resource = "credential"
	ResourceRepository    Resource = "repository"
	ResourceView          Resource = "view"
	ResourceTemplate      Resource = "template"
	ResourceSchedule      Resource = "schedule"
	ResourceEnvironment   Resource = "environment"
	ResourceSecretStorage Resource = "secret_storage"
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

type ProjectMemberAction string

const (
	ProjectMemberAdd        ProjectMemberAction = "add"
	ProjectMemberRemove     ProjectMemberAction = "remove"
	ProjectMemberChangeRole ProjectMemberAction = "change_role"
)

type ProjectMemberEvent struct {
	Action      ProjectMemberAction
	Actor       Actor
	Request     Request
	ProjectID   int
	UserID      int
	UserName    string
	Role        string
	OldRole     string
	NewRole     string
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
	case ResourceEnvironment:
		return resourceMapping{
			activityType: db.EventEnvironment,
			eventCode:    db.AuditEventCodeEnvironment,
			targetType:   "environment",
		}, nil
	case ResourceSecretStorage:
		return resourceMapping{
			activityType: db.EventSecretStorage,
			eventCode:    db.AuditEventCodeSecretStorage,
			targetType:   "secret_storage",
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

type memberActionMapping struct {
	eventLogAction string
	eventCode      string
	auditAction    string
}

func mapProjectMemberAction(action ProjectMemberAction) (memberActionMapping, error) {
	switch action {
	case ProjectMemberAdd:
		return memberActionMapping{"create", db.AuditEventCodeMembership, db.AuditActionAdd}, nil
	case ProjectMemberRemove:
		return memberActionMapping{"delete", db.AuditEventCodeMembership, db.AuditActionRemove}, nil
	case ProjectMemberChangeRole:
		return memberActionMapping{"update", db.AuditEventCodeProjectRole, db.AuditActionChange}, nil
	default:
		return memberActionMapping{}, fmt.Errorf("unsupported project member action %q", action)
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

func newActivityEvent(
	actor Actor, projectID, targetID int,
	targetType db.EventObjectType, description string,
) db.Event {
	event := db.Event{
		ObjectID:    &targetID,
		ObjectType:  &targetType,
		Description: &description,
	}
	if actor.ID > 0 {
		event.UserID = &actor.ID
	}
	if projectID > 0 {
		event.ProjectID = &projectID
	}
	return event
}

func (s *Service) newAuditEvent(
	actor Actor, request Request, projectID int,
	target db.AuditTarget,
) db.AuditEvent {
	event := db.NewAuditEvent()
	event.Outcome = db.AuditOutcomeSuccess
	event.Actor = &db.AuditActor{
		Type: "user",
		ID:   strconv.Itoa(actor.ID),
		Name: actor.Name,
	}
	event.Source = &db.AuditSource{
		IP:        request.SourceIP,
		UserAgent: request.UserAgent,
	}
	event.Target = &target
	event.Scope = &db.AuditScope{ProjectID: strconv.Itoa(projectID)}
	event.RequestID = request.ID
	event.InstanceID = s.settings.InstanceID
	event.NodeID = s.settings.NodeID
	return event
}

func (s *Service) recordProjections(
	activityEvent db.Event, eventLogAction string, auditEvent *db.AuditEvent,
) error {
	_, activityErr := s.store.CreateEvent(activityEvent)
	eventLogErr := s.logWriter.WriteEventLog(pro_interfaces.EventLogRecord{
		Action:        eventLogAction,
		ProjectID:     activityEvent.ProjectID,
		UserID:        activityEvent.UserID,
		IntegrationID: activityEvent.IntegrationID,
		Description:   activityEvent.Description,
	})

	var auditErr error
	if auditEvent != nil {
		_, auditErr = s.store.CreateAuditEvent(*auditEvent)
	}
	return errors.Join(activityErr, eventLogErr, auditErr)
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
	activityEvent := newActivityEvent(event.Actor, event.ProjectID, event.TargetID,
		resourceFields.activityType, event.Description)

	var auditEvent *db.AuditEvent
	if s.settings.Enabled {
		target := db.AuditTarget{
			Type: resourceFields.targetType,
			ID:   strconv.Itoa(event.TargetID),
			Name: event.TargetName,
		}
		value := s.newAuditEvent(event.Actor, event.Request, event.ProjectID, target)
		value.EventCode = resourceFields.eventCode
		value.Category = db.AuditCategoryResource
		value.Type = actionFields.eventType
		value.Action = actionFields.action
		auditEvent = &value
	}

	return s.recordProjections(activityEvent, string(event.Action), auditEvent)
}

func (s *Service) RecordProjectMember(event ProjectMemberEvent) error {
	actionFields, err := mapProjectMemberAction(event.Action)
	if err != nil {
		return err
	}

	activityEvent := newActivityEvent(event.Actor, event.ProjectID, event.UserID,
		db.EventUser, event.Description)

	var auditEvent *db.AuditEvent
	if s.settings.Enabled {
		var metadata any
		switch event.Action {
		case ProjectMemberAdd:
			metadata = struct {
				Role string `json:"role,omitempty"`
			}{event.Role}
		case ProjectMemberRemove:
			selfRemoval := event.Actor.ID == event.UserID
			metadata = struct {
				Role        string `json:"role,omitempty"`
				SelfRemoval bool   `json:"self_removal"`
			}{event.Role, selfRemoval}
		case ProjectMemberChangeRole:
			metadata = struct {
				OldRole string `json:"old_role,omitempty"`
				NewRole string `json:"new_role,omitempty"`
			}{event.OldRole, event.NewRole}
		}
		auditMetadata, err := json.Marshal(metadata)
		if err != nil {
			return err
		}

		target := db.AuditTarget{
			Type: "user",
			ID:   strconv.Itoa(event.UserID),
			Name: event.UserName,
		}
		value := s.newAuditEvent(event.Actor, event.Request, event.ProjectID, target)
		value.EventCode = actionFields.eventCode
		value.Category = db.AuditCategoryIAM
		value.Type = db.AuditTypeChange
		value.Action = actionFields.auditAction
		value.Metadata = auditMetadata
		auditEvent = &value
	}

	return s.recordProjections(activityEvent, actionFields.eventLogAction, auditEvent)
}
