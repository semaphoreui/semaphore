package audit

import (
	"errors"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type storeFake struct {
	db.Store
	activityEvents []db.Event
	auditEvents    []db.AuditEvent
	activityErr    error
	auditErr       error
}

func (f *storeFake) CreateEvent(event db.Event) (db.Event, error) {
	f.activityEvents = append(f.activityEvents, event)
	return event, f.activityErr
}

func (f *storeFake) CreateAuditEvent(event db.AuditEvent) (db.AuditEvent, error) {
	f.auditEvents = append(f.auditEvents, event)
	return event, f.auditErr
}

type logWriterFake struct {
	events []pro_interfaces.EventLogRecord
	err    error
}

func (f *logWriterFake) WriteEventLog(event pro_interfaces.EventLogRecord) error {
	f.events = append(f.events, event)
	return f.err
}

func (f *logWriterFake) WriteTaskLog(pro_interfaces.TaskLogRecord) error { return nil }

func (f *logWriterFake) WriteResult(any) error { return nil }

func TestMapResource(t *testing.T) {
	tests := []struct {
		resource Resource
		mapping  resourceMapping
	}{
		{ResourceProject, resourceMapping{db.EventProject, db.AuditEventCodeProject, "project"}},
		{ResourceInventory, resourceMapping{db.EventInventory, db.AuditEventCodeInventory, "inventory"}},
		{ResourceCredential, resourceMapping{db.EventKey, db.AuditEventCodeCredential, "credential"}},
		{ResourceRepository, resourceMapping{db.EventRepository, db.AuditEventCodeRepository, "repository"}},
		{ResourceView, resourceMapping{db.EventView, db.AuditEventCodeView, "view"}},
		{ResourceTemplate, resourceMapping{db.EventTemplate, db.AuditEventCodeTemplate, "template"}},
		{ResourceSchedule, resourceMapping{db.EventSchedule, db.AuditEventCodeSchedule, "schedule"}},
		{ResourceEnvironment, resourceMapping{db.EventEnvironment, db.AuditEventCodeEnvironment, "environment"}},
	}

	for _, test := range tests {
		t.Run(string(test.resource), func(t *testing.T) {
			mapping, err := mapResource(test.resource)

			require.NoError(t, err)
			assert.Equal(t, test.mapping, mapping)
		})
	}
}

func TestMapAction(t *testing.T) {
	tests := []struct {
		action  Action
		mapping actionMapping
	}{
		{ActionCreate, actionMapping{db.AuditTypeCreation, db.AuditActionCreate}},
		{ActionUpdate, actionMapping{db.AuditTypeChange, db.AuditActionUpdate}},
		{ActionDelete, actionMapping{db.AuditTypeDeletion, db.AuditActionDelete}},
	}

	for _, test := range tests {
		t.Run(string(test.action), func(t *testing.T) {
			mapping, err := mapAction(test.action)

			require.NoError(t, err)
			assert.Equal(t, test.mapping, mapping)
		})
	}
}

func TestRecordResourceEnabledRecordsBothProjections(t *testing.T) {
	store := &storeFake{}
	logWriter := &logWriterFake{}
	service := NewService(store, logWriter, Settings{
		Enabled: true, InstanceID: "instance-1", NodeID: "node-1",
	})

	err := service.RecordResource(ResourceEvent{
		Resource: ResourceCredential,
		Action:   ActionUpdate,
		Actor: Actor{
			ID: 42, Name: "alice",
		},
		Request: Request{
			ID: "request-1", SourceIP: "203.0.113.10", UserAgent: "test agent",
		},
		ProjectID: 7, TargetID: 9, TargetName: "production",
		Description: "Access Key production updated",
	})

	require.NoError(t, err)
	require.Len(t, store.activityEvents, 1)
	activityEvent := store.activityEvents[0]
	assert.Equal(t, 42, *activityEvent.UserID)
	assert.Equal(t, 7, *activityEvent.ProjectID)
	assert.Equal(t, 9, *activityEvent.ObjectID)
	assert.Equal(t, db.EventKey, *activityEvent.ObjectType)
	assert.Equal(t, "Access Key production updated", *activityEvent.Description)
	assert.Equal(t, []pro_interfaces.EventLogRecord{{
		Action:      "update",
		ProjectID:   activityEvent.ProjectID,
		UserID:      activityEvent.UserID,
		Description: activityEvent.Description,
	}}, logWriter.events)

	require.Len(t, store.auditEvents, 1)
	auditEvent := store.auditEvents[0]
	assert.Equal(t, db.AuditEventCodeCredential, auditEvent.EventCode)
	assert.Equal(t, db.AuditCategoryResource, auditEvent.Category)
	assert.Equal(t, db.AuditTypeChange, auditEvent.Type)
	assert.Equal(t, db.AuditActionUpdate, auditEvent.Action)
	assert.Equal(t, db.AuditOutcomeSuccess, auditEvent.Outcome)
	assert.Equal(t, &db.AuditActor{Type: "user", ID: "42", Name: "alice"}, auditEvent.Actor)
	assert.Equal(t, &db.AuditSource{IP: "203.0.113.10", UserAgent: "test agent"}, auditEvent.Source)
	assert.Equal(t, &db.AuditTarget{Type: "credential", ID: "9", Name: "production"}, auditEvent.Target)
	assert.Equal(t, &db.AuditScope{ProjectID: "7"}, auditEvent.Scope)
	assert.Equal(t, "request-1", auditEvent.RequestID)
	assert.Equal(t, "instance-1", auditEvent.InstanceID)
	assert.Equal(t, "node-1", auditEvent.NodeID)
}

func TestRecordResourceDisabledRecordsOnlyActivity(t *testing.T) {
	store := &storeFake{}
	logWriter := &logWriterFake{}
	service := NewService(store, logWriter, Settings{})

	err := service.RecordResource(ResourceEvent{
		Resource: ResourceProject, Action: ActionCreate, Description: "Project created",
	})

	require.NoError(t, err)
	assert.Equal(t, "Project created", *store.activityEvents[0].Description)
	assert.Len(t, logWriter.events, 1)
	assert.Empty(t, store.auditEvents)
}

func TestRecordResourceAttemptsAllSinksAfterFailures(t *testing.T) {
	activityErr := errors.New("activity failed")
	legacyErr := errors.New("legacy failed")
	auditErr := errors.New("audit failed")
	store := &storeFake{activityErr: activityErr, auditErr: auditErr}
	logWriter := &logWriterFake{err: legacyErr}
	service := NewService(store, logWriter, Settings{Enabled: true})

	err := service.RecordResource(ResourceEvent{
		Resource: ResourceProject, Action: ActionDelete, ProjectID: 3, TargetID: 3,
	})

	assert.ErrorIs(t, err, activityErr)
	assert.ErrorIs(t, err, legacyErr)
	assert.ErrorIs(t, err, auditErr)
	assert.Len(t, store.activityEvents, 1)
	assert.Len(t, logWriter.events, 1)
	assert.Len(t, store.auditEvents, 1)
}

func TestRecordResourceRejectsUnknownResource(t *testing.T) {
	store := &storeFake{}
	logWriter := &logWriterFake{}
	service := NewService(store, logWriter, Settings{Enabled: true})

	err := service.RecordResource(ResourceEvent{Resource: "unknown"})

	assert.ErrorContains(t, err, "unsupported audit resource")
	assert.Empty(t, store.activityEvents)
	assert.Empty(t, logWriter.events)
	assert.Empty(t, store.auditEvents)
}

func TestRecordResourceRejectsUnknownAction(t *testing.T) {
	store := &storeFake{}
	logWriter := &logWriterFake{}
	service := NewService(store, logWriter, Settings{Enabled: true})

	err := service.RecordResource(ResourceEvent{Resource: ResourceProject, Action: "unknown"})

	assert.ErrorContains(t, err, "unsupported audit action")
	assert.Empty(t, store.activityEvents)
	assert.Empty(t, logWriter.events)
	assert.Empty(t, store.auditEvents)
}
