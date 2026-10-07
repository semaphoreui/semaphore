package audit

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type retentionStore struct {
	mu      sync.Mutex
	calls   []int
	deleted int64
	lastSeq int64
	err     error
}

func (s *retentionStore) CreateAuditEvent(_ context.Context, row db.AuditEvent) (db.AuditEvent, error) {
	return row, nil
}

func (s *retentionStore) DeleteAuditEventsOlderThan(_ context.Context, days int, _ int) (int64, int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, days)
	return s.deleted, s.lastSeq, s.err
}

func (s *retentionStore) callCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.calls)
}

type recordedEvent struct {
	event Event
	actor Actor
}

type memRecorder struct {
	events []recordedEvent
}

func (r *memRecorder) Record(ctx context.Context, event Event) {
	r.events = append(r.events, recordedEvent{event: event, actor: ActorFrom(ctx)})
}

func TestRetention_RecordsWhatItDeleted(t *testing.T) {
	store, recorder := &retentionStore{deleted: 3, lastSeq: 41}, &memRecorder{}

	pruneAuditEvents(context.Background(), store, recorder, 30)

	require.Equal(t, []int{30}, store.calls)
	require.Len(t, recorder.events, 1)
	got := recorder.events[0]
	require.NoError(t, Validate(got.event))
	assert.Equal(t, AuditRetentionDelete, got.event.Kind)
	assert.Equal(t, RetentionMetadata{Deleted: 3, LastSeq: 41, RetentionDays: 30}, got.event.Metadata)
	assert.Equal(t, SystemActor(ComponentRetention), got.actor)
}

func TestRetention_RecordsOnlyWhatItDeleted(t *testing.T) {
	store, recorder := &retentionStore{}, &memRecorder{}

	pruneAuditEvents(context.Background(), store, recorder, 30)

	assert.Empty(t, recorder.events)
}

func TestRetention_PartialDeleteIsRecorded(t *testing.T) {
	store, recorder := &retentionStore{deleted: 2, lastSeq: 9, err: errors.New("connection reset")}, &memRecorder{}

	pruneAuditEvents(context.Background(), store, recorder, 7)

	require.Len(t, recorder.events, 1)
	assert.Equal(t, RetentionMetadata{Deleted: 2, LastSeq: 9, RetentionDays: 7}, recorder.events[0].event.Metadata)
}

func TestStartService_RunsRetentionAndStopsIt(t *testing.T) {
	store := &retentionStore{}
	s, err := StartService(store, &util.AuditConfig{Enabled: true, InstanceID: "i", RetentionDays: 30}, "", &fakeExporter{})
	require.NoError(t, err)

	require.Eventually(t, func() bool { return store.callCount() == 1 }, 2*time.Second, 10*time.Millisecond, "a pass at start")
	s.Stop()
	assert.Equal(t, 1, store.callCount())
}

func TestStartService_NoRetentionByDefault(t *testing.T) {
	store := &retentionStore{}
	s, err := StartService(store, &util.AuditConfig{Enabled: true, InstanceID: "i"}, "", &fakeExporter{})
	require.NoError(t, err)
	time.Sleep(50 * time.Millisecond)
	s.Stop()

	assert.Zero(t, store.callCount())
}

func TestStartService_DisabledAuditRunsNoRetention(t *testing.T) {
	store := &retentionStore{}
	s, err := StartService(store, &util.AuditConfig{RetentionDays: 30}, "", &fakeExporter{})
	require.NoError(t, err)
	time.Sleep(50 * time.Millisecond)
	s.Stop()

	assert.Zero(t, store.callCount())
}
