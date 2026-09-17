package sql

import (
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuditMigrationFrom2_20_5(t *testing.T) {
	version := "2.20.5"
	store := InitConfigCreateTestStoreAt(&version)
	require.NoError(t, db.Migrate(store, nil))
	assert.Positive(t, createAuditEvent(t, store).Seq)
}
func TestAuditFreshSQLiteMigration(t *testing.T) {
	store := InitConfigCreateTestStore()
	assert.Positive(t, createAuditEvent(t, store).Seq)
}
func TestAuditEventStoreInt64AndMetadata(t *testing.T) {
	store := InitConfigCreateTestStore()
	first := createAuditEvent(t, store)
	second := createAuditEvent(t, store)
	events, err := store.GetAuditEventsAfter(0, 2)
	require.NoError(t, err)
	assert.Equal(t, first.Seq, events[0].Seq)
	assert.Equal(t, second.Seq, events[1].Seq)
	assert.True(t, first.Timestamp.Equal(events[0].Timestamp))
	assert.JSONEq(t, `{}`, string(events[0].Metadata))
	_, err = store.GetAuditEventsAfter(0, 0)
	assert.Error(t, err)
	_, err = store.GetAuditEventsAfter(0, 1001)
	assert.Error(t, err)
}
func TestAuditEventStoreRejectsDuplicateEventID(t *testing.T) {
	store := InitConfigCreateTestStore()
	event := createAuditEvent(t, store)
	_, err := store.CreateAuditEvent(event)
	assert.Error(t, err)
}
func TestAuditEventStoreReadsAfterCursorWithBatchLimit(t *testing.T) {
	store := InitConfigCreateTestStore()
	first := createAuditEvent(t, store)
	second := createAuditEvent(t, store)
	third := createAuditEvent(t, store)

	events, err := store.GetAuditEventsAfter(first.Seq, 2)
	require.NoError(t, err)
	assert.Equal(t, []int64{second.Seq, third.Seq}, []int64{events[0].Seq, events[1].Seq})

	events, err = store.GetAuditEventsAfter(first.Seq, 1)
	require.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, second.Seq, events[0].Seq)
}
func TestAuditEventToRowUsesTypedNulls(t *testing.T) {
	event := db.NewAuditEvent()
	event.EventCode, event.Category, event.Type, event.Action, event.Outcome, event.InstanceID = "auth.login", "authentication", "authentication", "login", "failure", "test"
	event.Actor = &db.AuditActor{Type: "anonymous"}
	event.Metadata = []byte(`{}`)

	row := auditEventToRow(event)
	assert.Nil(t, row.SourceIP)
	assert.Nil(t, row.SourceUserAgent)
	assert.Nil(t, row.TargetType)
	assert.Nil(t, row.TargetID)
	assert.Nil(t, row.TargetName)
	assert.Nil(t, row.ProjectID)
	assert.Nil(t, row.ActorID)
	assert.Nil(t, row.ActorName)
	assert.Nil(t, row.RequestID)
	assert.Nil(t, row.NodeID)
	require.NotNil(t, row.Metadata)
	assert.Equal(t, "{}", *row.Metadata)
}
func TestAuditHistorySurvivesProductDeletion(t *testing.T) {
	store := InitConfigCreateTestStore()
	user, err := store.CreateUserWithoutPassword(db.User{Username: "audit", Name: "Audit", Email: "audit@example.com"})
	require.NoError(t, err)
	project, err := store.CreateProject(db.Project{Name: "audit"})
	require.NoError(t, err)
	event := createAuditEvent(t, store)
	event.ID = uuid.NewString()
	event.Actor.ID, event.Scope.ProjectID = strconv.Itoa(user.ID), strconv.Itoa(project.ID)
	_, err = store.CreateAuditEvent(event)
	require.NoError(t, err)
	_, err = store.exec("delete from project where id=?", project.ID)
	require.NoError(t, err)
	_, err = store.exec("delete from `user` where id=?", user.ID)
	require.NoError(t, err)
	events, err := store.GetAuditEventsAfter(0, 10)
	require.NoError(t, err)
	assert.Len(t, events, 2)
}
func createAuditEvent(t *testing.T, store *SqlDb) db.AuditEvent {
	t.Helper()
	event := db.NewAuditEvent()
	event.EventCode, event.Category, event.Type, event.Action, event.Outcome, event.InstanceID = db.AuditEventCodeInventory, db.AuditCategoryResource, db.AuditTypeCreation, db.AuditActionCreate, db.AuditOutcomeSuccess, "test"
	event.Actor = &db.AuditActor{Type: "user", ID: "1"}
	event.Source = &db.AuditSource{IP: "203.0.113.10"}
	event.Target = &db.AuditTarget{Type: "inventory", ID: "1"}
	event.Scope = &db.AuditScope{ProjectID: "1"}
	event.RequestID = uuid.NewString()
	event.Metadata = []byte(`{}`)
	saved, err := store.CreateAuditEvent(event)
	require.NoError(t, err)
	return saved
}
