package projects

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	"github.com/semaphoreui/semaphore/services/server"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type auditLogWriter struct{}

func (auditLogWriter) WriteEventLog(pro_interfaces.EventLogRecord) error { return nil }
func (auditLogWriter) WriteTaskLog(pro_interfaces.TaskLogRecord) error   { return nil }
func (auditLogWriter) WriteResult(any) error                             { return nil }

type auditAccessKeyService struct {
	store db.AccessKeyManager
}

func (s auditAccessKeyService) Create(key db.AccessKey) (db.AccessKey, error) {
	return s.store.CreateAccessKey(key)
}

func (auditAccessKeyService) Update(db.AccessKey) error { return nil }
func (auditAccessKeyService) GetAll(int, db.GetAccessKeyOptions, db.RetrieveQueryParams) ([]db.AccessKey, error) {
	return nil, nil
}
func (auditAccessKeyService) Delete(int, int) error { return nil }

var _ server.AccessKeyService = auditAccessKeyService{}

type auditEventStoreSpy struct {
	createCalls int
}

func (*auditEventStoreSpy) GetAuditEventsAfter(int64, int) ([]db.AuditEvent, error) {
	return nil, nil
}

func (s *auditEventStoreSpy) CreateAuditEvent(event db.AuditEvent) (db.AuditEvent, error) {
	s.createCalls++
	return event, nil
}

func auditTestStore(t *testing.T, enabled bool) (*sql.SqlDb, db.User) {
	t.Helper()
	previousConfig := util.Config
	t.Cleanup(func() { util.Config = previousConfig })

	store := sql.InitConfigCreateTestStore()
	setAuditConfig(t, enabled)
	user, err := store.CreateUserWithoutPassword(db.User{
		Username: "audit-user",
		Name:     "Audit User",
		Email:    "audit-user@example.test",
		Admin:    true,
	})
	require.NoError(t, err)
	return store, user
}

func setAuditConfig(t *testing.T, enabled bool) {
	t.Helper()
	previousConfig := util.Config
	t.Cleanup(func() { util.Config = previousConfig })
	util.Config = &util.ConfigType{
		Audit: &util.AuditConfig{Enabled: enabled, InstanceID: "test-instance"},
	}
}

func auditRequest(
	t *testing.T,
	method string,
	path string,
	body any,
	store db.Store,
	user db.User,
) *http.Request {
	t.Helper()

	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		data, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(data)
	}

	req := httptest.NewRequest(method, path, reader)
	req = helpers.SetContextValue(req, "store", store)
	req = helpers.SetContextValue(req, "user", &user)
	req = helpers.SetContextValue(req, "log_writer", auditLogWriter{})
	req = helpers.SetAuditRequestContext(req, helpers.AuditRequestContext{
		RequestID: uuid.NewString(),
		SourceIP:  "203.0.113.10",
		UserAgent: "audit test agent",
	})
	return req
}

func auditEvents(t *testing.T, store db.AuditEventStore) []db.AuditEvent {
	t.Helper()
	events, err := store.GetAuditEventsAfter(0, 100)
	require.NoError(t, err)
	return events
}

type auditEventExpectation struct {
	eventCode  string
	eventType  string
	action     string
	targetType string
	targetID   string
	targetName string
	projectID  string
	requestID  string
}

func assertAuditEvent(t *testing.T, event db.AuditEvent, expected auditEventExpectation, user db.User) {
	t.Helper()
	assert.Equal(t, db.AuditSchemaVersion, event.SchemaVersion)
	assert.Equal(t, expected.eventCode, event.EventCode)
	assert.Equal(t, db.AuditCategoryResource, event.Category)
	assert.Equal(t, expected.eventType, event.Type)
	assert.Equal(t, expected.action, event.Action)
	assert.Equal(t, db.AuditOutcomeSuccess, event.Outcome)
	assert.Equal(t, "user", event.Actor.Type)
	assert.Equal(t, strconv.Itoa(user.ID), event.Actor.ID)
	assert.Equal(t, user.Username, event.Actor.Name)
	assert.Equal(t, "203.0.113.10", event.Source.IP)
	assert.Equal(t, "audit test agent", event.Source.UserAgent)
	assert.Equal(t, expected.requestID, event.RequestID)
	requestID, err := uuid.Parse(expected.requestID)
	require.NoError(t, err)
	assert.Equal(t, uuid.RFC4122, requestID.Variant())
	assert.Equal(t, uuid.Version(4), requestID.Version())
	assert.Equal(t, expected.targetType, event.Target.Type)
	assert.Equal(t, expected.targetID, event.Target.ID)
	assert.Equal(t, expected.targetName, event.Target.Name)
	assert.Equal(t, expected.projectID, event.Scope.ProjectID)
	assert.Equal(t, "test-instance", event.InstanceID)
	assert.Empty(t, event.NodeID)
	id, err := uuid.Parse(event.ID)
	require.NoError(t, err)
	assert.Equal(t, uuid.Version(4), id.Version())
	assert.False(t, event.Timestamp.IsZero())
	assert.Equal(t, time.UTC, event.Timestamp.Location())
	assert.Empty(t, event.Metadata)
}

func requestID(t *testing.T, req *http.Request) string {
	t.Helper()
	context, ok := helpers.AuditRequestContextFrom(req)
	require.True(t, ok)
	return context.RequestID
}

func TestProjectCreateRecordsAuditEvent(t *testing.T) {
	store, user := auditTestStore(t, true)
	controller := NewProjectsController(auditAccessKeyService{store: store})
	req := auditRequest(t, http.MethodPost, "/api/projects", db.Project{Name: "audit project"}, store, user)
	reqID := requestID(t, req)
	w := httptest.NewRecorder()

	controller.AddProject(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	var project db.Project
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &project))
	projectID := strconv.Itoa(project.ID)
	events := auditEvents(t, store)
	require.Len(t, events, 1)
	event := events[0]
	assertAuditEvent(t, event, auditEventExpectation{
		eventCode:  db.AuditEventCodeProject,
		eventType:  db.AuditTypeCreation,
		action:     db.AuditActionCreate,
		targetType: "project",
		targetID:   projectID,
		targetName: "audit project",
		projectID:  projectID,
		requestID:  reqID,
	}, user)
	activities, err := store.GetEvents(project.ID, db.RetrieveQueryParams{})
	require.NoError(t, err)
	require.Len(t, activities, 1)
	assert.Equal(t, "Project created", *activities[0].Description)
}

func TestInventoryMutationsRecordAuditEvents(t *testing.T) {
	store, user := auditTestStore(t, true)
	project, err := store.CreateProject(db.Project{Name: "inventory project"})
	require.NoError(t, err)

	secret := "inventory secret must not be audited"
	create := db.Inventory{
		Name:      "initial inventory",
		ProjectID: project.ID,
		Inventory: secret,
		Type:      db.InventoryStatic,
	}
	createRequest := auditRequest(t, http.MethodPost, "/api/projects/inventory", create, store, user)
	createRequestID := requestID(t, createRequest)
	createRequest = helpers.SetContextValue(createRequest, "project", project)
	createResponse := httptest.NewRecorder()
	AddInventory(createResponse, createRequest)
	require.Equal(t, http.StatusCreated, createResponse.Code)

	events := auditEvents(t, store)
	require.Len(t, events, 1)
	created := events[0]
	assert.NotEmpty(t, created.Target.ID)
	assertAuditEvent(t, created, auditEventExpectation{
		eventCode:  db.AuditEventCodeInventory,
		eventType:  db.AuditTypeCreation,
		action:     db.AuditActionCreate,
		targetType: "inventory",
		targetID:   created.Target.ID,
		targetName: "initial inventory",
		projectID:  strconv.Itoa(project.ID),
		requestID:  createRequestID,
	}, user)
	serialized, err := json.Marshal(created)
	require.NoError(t, err)
	assert.NotContains(t, string(serialized), secret)
	assert.NotContains(t, string(serialized), "Inventory initial inventory created")

	inventory, err := store.GetInventory(project.ID, mustID(t, created.Target.ID))
	require.NoError(t, err)
	inventory.Name = "final inventory"
	updateRequest := auditRequest(t, http.MethodPut, "/api/projects/inventory", inventory, store, user)
	updateRequestID := requestID(t, updateRequest)
	updateRequest = helpers.SetContextValue(updateRequest, "inventory", inventory)
	updateResponse := httptest.NewRecorder()
	UpdateInventory(updateResponse, updateRequest)
	require.Equal(t, http.StatusNoContent, updateResponse.Code)

	events = auditEvents(t, store)
	require.Len(t, events, 2)
	updated := events[1]
	assertAuditEvent(t, updated, auditEventExpectation{
		eventCode:  db.AuditEventCodeInventory,
		eventType:  db.AuditTypeChange,
		action:     db.AuditActionUpdate,
		targetType: "inventory",
		targetID:   strconv.Itoa(inventory.ID),
		targetName: "final inventory",
		projectID:  strconv.Itoa(project.ID),
		requestID:  updateRequestID,
	}, user)

	deleteRequest := auditRequest(t, http.MethodDelete, "/api/projects/inventory", nil, store, user)
	deleteRequestID := requestID(t, deleteRequest)
	deleteRequest = helpers.SetContextValue(deleteRequest, "inventory", inventory)
	deleteResponse := httptest.NewRecorder()
	RemoveInventory(deleteResponse, deleteRequest)
	require.Equal(t, http.StatusNoContent, deleteResponse.Code)

	events = auditEvents(t, store)
	require.Len(t, events, 3)
	deleted := events[2]
	assertAuditEvent(t, deleted, auditEventExpectation{
		eventCode:  db.AuditEventCodeInventory,
		eventType:  db.AuditTypeDeletion,
		action:     db.AuditActionDelete,
		targetType: "inventory",
		targetID:   strconv.Itoa(inventory.ID),
		targetName: "final inventory",
		projectID:  strconv.Itoa(project.ID),
		requestID:  deleteRequestID,
	}, user)
}

func TestInventoryAuditDisabledAndFailurePaths(t *testing.T) {
	t.Run("project create does not create an audit event", func(t *testing.T) {
		store, user := auditTestStore(t, false)
		controller := NewProjectsController(auditAccessKeyService{store: store})
		req := auditRequest(t, http.MethodPost, "/api/projects", db.Project{Name: "disabled project"}, store, user)
		w := httptest.NewRecorder()

		controller.AddProject(w, req)

		require.Equal(t, http.StatusCreated, w.Code)
		assert.Empty(t, auditEvents(t, store))
	})

	t.Run("disabled does not create an audit event", func(t *testing.T) {
		store, user := auditTestStore(t, false)
		project, err := store.CreateProject(db.Project{Name: "disabled audit project"})
		require.NoError(t, err)
		inventory := db.Inventory{Name: "inventory", ProjectID: project.ID, Type: db.InventoryStatic}
		req := auditRequest(t, http.MethodPost, "/api/projects/inventory", inventory, store, user)
		req = helpers.SetContextValue(req, "project", project)
		w := httptest.NewRecorder()

		AddInventory(w, req)

		require.Equal(t, http.StatusCreated, w.Code)
		assert.Empty(t, auditEvents(t, store))
	})

	t.Run("mutation failure does not create a success event", func(t *testing.T) {
		store, user := auditTestStore(t, true)
		project, err := store.CreateProject(db.Project{Name: "failed audit project"})
		require.NoError(t, err)
		inventory := db.Inventory{Name: "inventory", ProjectID: project.ID + 1, Type: db.InventoryStatic}
		req := auditRequest(t, http.MethodPost, "/api/projects/inventory", inventory, store, user)
		req = helpers.SetContextValue(req, "project", project)
		w := httptest.NewRecorder()

		AddInventory(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
		assert.Empty(t, auditEvents(t, store))
	})

	t.Run("audit insert failure preserves successful response", func(t *testing.T) {
		store, user := auditTestStore(t, true)
		project, err := store.CreateProject(db.Project{Name: "audit failure project"})
		require.NoError(t, err)
		_, err = store.Sql().Exec("drop table audit_event")
		require.NoError(t, err)
		inventory := db.Inventory{Name: "inventory", ProjectID: project.ID, Type: db.InventoryStatic}
		req := auditRequest(t, http.MethodPost, "/api/projects/inventory", inventory, store, user)
		req = helpers.SetContextValue(req, "project", project)
		w := httptest.NewRecorder()

		AddInventory(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})
}

func mustID(t *testing.T, value string) int {
	t.Helper()
	id, err := strconv.Atoi(value)
	require.NoError(t, err)
	return id
}

func TestAuditResourceEventMissingContextDoesNotPersist(t *testing.T) {
	store, user := auditTestStore(t, true)
	req := httptest.NewRequest(http.MethodPost, "/api/projects", nil)
	req = helpers.SetContextValue(req, "user", &user)

	assert.NotPanics(t, func() {
		helpers.AuditResourceEvent(req, store, helpers.AuditResourceEventItem{
			Resource:   helpers.AuditResourceProject,
			Action:     helpers.EventLogCreate,
			TargetID:   "1",
			TargetName: "project",
			ProjectID:  1,
		})
	})
	assert.Empty(t, auditEvents(t, store))
}

func TestAuditResourceEventDisabledDoesNotCallStore(t *testing.T) {
	setAuditConfig(t, false)
	store := &auditEventStoreSpy{}
	req := httptest.NewRequest(http.MethodPost, "/api/projects", nil)

	helpers.AuditResourceEvent(req, store, helpers.AuditResourceEventItem{
		Resource:   helpers.AuditResourceProject,
		Action:     helpers.EventLogCreate,
		TargetID:   "1",
		TargetName: "project",
		ProjectID:  1,
	})

	assert.Zero(t, store.createCalls)
}

func TestAuditResourceEventUnknownResourceDoesNotPersist(t *testing.T) {
	setAuditConfig(t, true)
	user := db.User{ID: 1, Username: "audit-user"}
	store := &auditEventStoreSpy{}
	req := httptest.NewRequest(http.MethodPost, "/api/projects", nil)
	req = helpers.SetContextValue(req, "user", &user)
	req = helpers.SetAuditRequestContext(req, helpers.AuditRequestContext{
		RequestID: uuid.NewString(),
		SourceIP:  "203.0.113.10",
	})

	helpers.AuditResourceEvent(req, store, helpers.AuditResourceEventItem{
		Resource:   helpers.AuditResourceKind("other"),
		Action:     helpers.EventLogCreate,
		TargetID:   "1",
		TargetName: "project",
		ProjectID:  1,
	})

	assert.Zero(t, store.createCalls)
}
