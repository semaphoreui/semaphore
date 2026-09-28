package sql

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuditEventSequenceAllocation(t *testing.T) {
	store := InitConfigCreateTestStore()
	t.Cleanup(store.Close)

	first := createAuditEvent(t, store)
	second := createAuditEvent(t, store)

	assert.Positive(t, first.Seq)
	assert.Equal(t, first.Seq+1, second.Seq)
}

func TestAuditEventDuplicateIDRollsBackSequence(t *testing.T) {
	store := InitConfigCreateTestStore()
	t.Cleanup(store.Close)

	first := createAuditEvent(t, store)
	_, err := store.CreateAuditEvent(first)
	assert.Error(t, err)

	next := createAuditEvent(t, store)
	assert.Equal(t, first.Seq+1, next.Seq)

	var allocated int64
	require.NoError(t, store.selectOne(&allocated, "select last_seq from audit_event_sequence where id=?", 1))
	assert.Equal(t, next.Seq, allocated)
}

func TestAuditEventRejectsInvalidMetadata(t *testing.T) {
	store := InitConfigCreateTestStore()
	t.Cleanup(store.Close)

	invalid := validAuditEvent()
	invalid.Metadata = []byte(`{`)
	_, err := store.CreateAuditEvent(invalid)
	assert.ErrorContains(t, err, "metadata is invalid JSON")

	var count int
	require.NoError(t, store.selectOne(
		&count,
		"select count(*) from audit_event where event_id=?",
		invalid.ID,
	))
	assert.Zero(t, count)
}

func TestAuditEventRejectsEmptyActorType(t *testing.T) {
	store := InitConfigCreateTestStore()
	t.Cleanup(store.Close)

	invalid := validAuditEvent()
	invalid.Actor.Type = ""
	_, err := store.CreateAuditEvent(invalid)
	assert.ErrorContains(t, err, "actor type is required")
}

func TestAuditEventRejectsMissingRequiredEnvelopeFields(t *testing.T) {
	store := InitConfigCreateTestStore()
	t.Cleanup(store.Close)

	tests := []struct {
		name   string
		field  string
		mutate func(*db.AuditEvent)
	}{
		{"event ID", "event ID", func(event *db.AuditEvent) { event.ID = "" }},
		{"timestamp", "timestamp", func(event *db.AuditEvent) { event.Timestamp = time.Time{} }},
		{"schema version", "schema version", func(event *db.AuditEvent) { event.SchemaVersion = "" }},
		{"event code", "event code", func(event *db.AuditEvent) { event.EventCode = "" }},
		{"category", "category", func(event *db.AuditEvent) { event.Category = "" }},
		{"type", "type", func(event *db.AuditEvent) { event.Type = "" }},
		{"action", "action", func(event *db.AuditEvent) { event.Action = "" }},
		{"outcome", "outcome", func(event *db.AuditEvent) { event.Outcome = "" }},
		{"actor type", "actor type", func(event *db.AuditEvent) { event.Actor.Type = "" }},
		{"instance ID", "instance ID", func(event *db.AuditEvent) { event.InstanceID = "" }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			event := validAuditEvent()
			test.mutate(&event)

			_, err := store.CreateAuditEvent(event)

			assert.ErrorContains(t, err, test.field)
		})
	}

	var allocated int64
	require.NoError(t, store.selectOne(&allocated, "select last_seq from audit_event_sequence where id=?", 1))
	assert.Zero(t, allocated)
}

func TestAuditEventAllowsMissingOptionalContext(t *testing.T) {
	store := InitConfigCreateTestStore()
	t.Cleanup(store.Close)
	event := validAuditEvent()
	event.Source = nil
	event.Target = nil
	event.Scope = nil
	event.RequestID = ""

	_, err := store.CreateAuditEvent(event)

	assert.NoError(t, err)
}

func TestAuditEventRejectsInvalidEnvelopeValues(t *testing.T) {
	store := InitConfigCreateTestStore()
	t.Cleanup(store.Close)

	tests := []struct {
		name   string
		field  string
		mutate func(*db.AuditEvent)
	}{
		{"event ID is not UUIDv4", "event ID", func(event *db.AuditEvent) { event.ID = "not-a-uuid" }},
		{"timestamp is not UTC", "timestamp", func(event *db.AuditEvent) { event.Timestamp = event.Timestamp.In(time.FixedZone("test", 3600)) }},
		{"unsupported schema version", "schema version", func(event *db.AuditEvent) { event.SchemaVersion = "2" }},
		{"source has invalid IP", "source IP", func(event *db.AuditEvent) { event.Source.IP = "not-an-ip" }},
		{"target has no type", "target type", func(event *db.AuditEvent) { event.Target.Type = "" }},
		{"target has no ID", "target ID", func(event *db.AuditEvent) { event.Target.ID = "" }},
		{"scope has no project ID", "project ID", func(event *db.AuditEvent) { event.Scope.ProjectID = "" }},
		{"request ID is not UUIDv4", "request ID", func(event *db.AuditEvent) { event.RequestID = "not-a-uuid" }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			event := validAuditEvent()
			test.mutate(&event)

			_, err := store.CreateAuditEvent(event)

			assert.ErrorContains(t, err, test.field)
		})
	}
}

func createAuditEvent(t *testing.T, store *SqlDb) db.AuditEvent {
	t.Helper()
	event := validAuditEvent()
	saved, err := store.CreateAuditEvent(event)
	require.NoError(t, err)
	return saved
}

func validAuditEvent() db.AuditEvent {
	event := db.NewAuditEvent(db.AuditActor{Type: db.AuditActorTypeUser, ID: "1"})
	event.EventCode = db.AuditEventCodeInventory
	event.Category = db.AuditCategoryResource
	event.Type = db.AuditTypeCreation
	event.Action = db.AuditActionCreate
	event.Outcome = db.AuditOutcomeSuccess
	event.InstanceID = "test"
	event.Source = &db.AuditSource{IP: "203.0.113.10"}
	event.Target = &db.AuditTarget{Type: "inventory", ID: "1"}
	event.Scope = &db.AuditScope{ProjectID: "1"}
	event.RequestID = uuid.NewString()
	event.Metadata = []byte(`{}`)
	return event
}
