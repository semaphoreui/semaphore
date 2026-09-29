package sql

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func selectAuditEvents(t *testing.T, store *SqlDb) []db.AuditEvent {
	t.Helper()
	var rows []db.AuditEvent
	_, err := store.Sql().Select(&rows, "select * from audit_event order by seq")
	require.NoError(t, err)
	return rows
}

func fullAuditEvent() db.AuditEvent {
	projectID := 12
	return db.AuditEvent{
		EventID:               "7b0c0000-0000-4000-8000-000000000001",
		SchemaVersion:         "1",
		Category:              "iam",
		EventCode:             "iam.api_token",
		Type:                  "creation",
		Action:                "create",
		Outcome:               "success",
		Reason:                "",
		ActorType:             "user",
		ActorID:               "7",
		ActorName:             "alice",
		ActorAuth:             "api_token",
		ActorTokenFingerprint: "3f9a0c1d2e4b5a69",
		SourceIP:              "10.0.0.5",
		UserAgent:             "curl/8.5",
		TargetType:            "api_token",
		TargetID:              "3f9a0c1d2e4b5a69",
		TargetName:            "ci-token",
		ProjectID:             &projectID,
		RequestID:             "req-1",
		InstanceID:            "prod-eu",
		NodeID:                "node-a",
		Metadata:              `{"fields":["name"]}`,
	}
}

func TestAuditTablesAfterMigration(t *testing.T) {
	store := InitConfigCreateTestStore()

	lastSeq, err := store.Sql().SelectInt("select last_seq from audit_event_sequence where id = 1")
	require.NoError(t, err)
	assert.Zero(t, lastSeq)

	count, err := store.Sql().SelectInt("select count(*) from audit_export_state")
	require.NoError(t, err)
	assert.Zero(t, count)
}

func TestCreateAuditEvent_RoundTripsEveryColumn(t *testing.T) {
	store := InitConfigCreateTestStore()
	event := fullAuditEvent()

	stored, err := store.CreateAuditEvent(context.Background(), event)
	require.NoError(t, err)
	assert.Equal(t, int64(1), stored.Seq)

	rows := selectAuditEvents(t, store)
	require.Len(t, rows, 1)
	assert.WithinDuration(t, time.Now(), stored.Created, time.Minute)
	assert.True(t, stored.Created.Equal(rows[0].Created))
	event.Seq = 1
	event.Created = stored.Created
	rows[0].Created = stored.Created
	assert.Equal(t, event, rows[0])
}

func TestCreateAuditEvent_NullProject(t *testing.T) {
	store := InitConfigCreateTestStore()
	event := fullAuditEvent()
	event.ProjectID = nil

	_, err := store.CreateAuditEvent(context.Background(), event)
	require.NoError(t, err)
	assert.Nil(t, selectAuditEvents(t, store)[0].ProjectID)
}

func TestCreateAuditEvent_GapFreeUnderConcurrency(t *testing.T) {
	store := InitConfigCreateTestStore()
	const n = 20

	var wg sync.WaitGroup
	errs := make(chan error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			event := fullAuditEvent()
			event.EventID = fmt.Sprintf("id-%d", i)
			_, err := store.CreateAuditEvent(context.Background(), event)
			errs <- err
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}

	rows := selectAuditEvents(t, store)
	require.Len(t, rows, n)
	for i, row := range rows {
		assert.Equal(t, int64(i+1), row.Seq)
	}
}

func TestCreateAuditEvent_CanceledContextStoresNothing(t *testing.T) {
	store := InitConfigCreateTestStore()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := store.CreateAuditEvent(ctx, fullAuditEvent())

	assert.ErrorIs(t, err, context.Canceled)
	assert.Empty(t, selectAuditEvents(t, store))
	lastSeq, err := store.Sql().SelectInt("select last_seq from audit_event_sequence where id = 1")
	require.NoError(t, err)
	assert.Zero(t, lastSeq)
}
