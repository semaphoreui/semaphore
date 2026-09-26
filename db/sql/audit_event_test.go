package sql

import (
	"testing"

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

	invalid := db.NewAuditEvent()
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

func createAuditEvent(t *testing.T, store *SqlDb) db.AuditEvent {
	t.Helper()
	event := db.NewAuditEvent()
	event.EventCode = db.AuditEventCodeInventory
	event.Category = db.AuditCategoryResource
	event.Type = db.AuditTypeCreation
	event.Action = db.AuditActionCreate
	event.Outcome = db.AuditOutcomeSuccess
	event.InstanceID = "test"
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
