package audit

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	log "github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeStore struct {
	rows []db.AuditEvent
	err  error
	// State of the store context at the time of the call.
	ctxErr   error
	deadline time.Time
}

func (s *fakeStore) CreateAuditEvent(ctx context.Context, row db.AuditEvent) (db.AuditEvent, error) {
	s.ctxErr = ctx.Err()
	s.deadline, _ = ctx.Deadline()
	if s.err != nil {
		return row, s.err
	}
	row.Seq = int64(len(s.rows) + 1)
	row.Created = time.Date(2026, 9, 29, 10, 0, 0, 123456000, time.UTC)
	s.rows = append(s.rows, row)
	return row, nil
}

func newTestRecorder(store *fakeStore) Recorder {
	return NewRecorder(store, Options{
		InstanceID: "prod-eu",
		NodeID:     "node-a",
		NewID:      func() string { return "7b0c0000-0000-4000-8000-000000000001" },
	})
}

func onlyEnvelope(t *testing.T, store *fakeStore) Envelope {
	t.Helper()
	require.Len(t, store.rows, 1)
	envelope, err := EnvelopeFromRow(store.rows[0])
	require.NoError(t, err)
	return envelope
}

func TestRecord_BuildsTheEvent(t *testing.T) {
	store := &fakeStore{}
	ctx := WithActor(context.Background(), UserActor(7, "alice", AuthAPIToken, "3f9a0c1d2e4b5a69"))
	ctx = WithRequest(ctx, RequestInfo{ID: "req-1", IP: "10.0.0.5", UserAgent: "curl/8.5"})

	newTestRecorder(store).Record(ctx, Event{
		Kind:      IAMAPITokenCreate,
		Target:    &Target{Type: TargetAPIToken, ID: "3f9a0c1d2e4b5a69", Name: "ci-token"},
		ProjectID: 12,
	})

	assert.Equal(t, Envelope{
		EventID:       "7b0c0000-0000-4000-8000-000000000001",
		Seq:           1,
		Timestamp:     "2026-09-29T10:00:00.123456Z",
		SchemaVersion: "1",
		Category:      "iam",
		EventCode:     "iam.api_token",
		Type:          TypeCreation,
		Action:        "create",
		Outcome:       OutcomeSuccess,
		Actor:         Actor{Type: ActorUser, ID: "7", Name: "alice", Auth: AuthAPIToken, TokenFingerprint: "3f9a0c1d2e4b5a69"},
		Source:        &Source{IP: "10.0.0.5", UserAgent: "curl/8.5"},
		Target:        &Target{Type: TargetAPIToken, ID: "3f9a0c1d2e4b5a69", Name: "ci-token"},
		Scope:         &Scope{ProjectID: "12"},
		RequestID:     "req-1",
		InstanceID:    "prod-eu",
		NodeID:        "node-a",
		Metadata:      []byte(`{}`),
	}, onlyEnvelope(t, store))
}

func TestRecord_ActorKinds(t *testing.T) {
	tests := []struct {
		name  string
		ctx   context.Context
		actor Actor
	}{
		{"session user", WithActor(context.Background(), UserActor(1, "a", AuthSession, "")), Actor{Type: ActorUser, ID: "1", Name: "a", Auth: AuthSession}},
		{"api token user", WithActor(context.Background(), UserActor(1, "a", AuthAPIToken, "ff")), Actor{Type: ActorUser, ID: "1", Name: "a", Auth: AuthAPIToken, TokenFingerprint: "ff"}},
		{"system", WithActor(context.Background(), SystemActor(ComponentScheduler)), Actor{Type: ActorSystem, Name: "scheduler"}},
		{"runner", WithActor(context.Background(), RunnerActor(3, "r1")), Actor{Type: ActorRunner, ID: "3", Name: "r1"}},
		{"integration", WithActor(context.Background(), IntegrationActor(4, "gh")), Actor{Type: ActorIntegration, ID: "4", Name: "gh"}},
		{"anonymous", context.Background(), Actor{Type: ActorAnonymous}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeStore{}
			newTestRecorder(store).Record(tt.ctx, Event{Kind: IAMUserDelete})
			assert.Equal(t, tt.actor, onlyEnvelope(t, store).Actor)
		})
	}
}

func TestRecord_OutsideHTTP(t *testing.T) {
	store := &fakeStore{}
	newTestRecorder(store).Record(context.Background(), Event{Kind: AuthLogout})
	envelope := onlyEnvelope(t, store)
	assert.Nil(t, envelope.Source)
	assert.Nil(t, envelope.Scope)
	assert.Nil(t, envelope.Target)
	assert.Nil(t, store.rows[0].ProjectID)
}

func TestRecord_FailureAndMetadata(t *testing.T) {
	store := &fakeStore{}
	newTestRecorder(store).Record(context.Background(), Event{
		Kind:     AuthLogin,
		Outcome:  OutcomeFailure,
		Reason:   ReasonInvalidCredentials,
		Metadata: AuthMethodMetadata{Method: LoginMethodPassword},
	})
	envelope := onlyEnvelope(t, store)
	assert.Equal(t, OutcomeFailure, envelope.Outcome)
	assert.Equal(t, ReasonInvalidCredentials, envelope.Reason)
	assert.Equal(t, TypeStart, envelope.Type)
	assert.JSONEq(t, `{"method":"password"}`, string(envelope.Metadata))
}

func TestRecord_BoundsStrings(t *testing.T) {
	store := &fakeStore{}
	long := strings.Repeat("x", 300)
	ctx := WithActor(context.Background(), UserActor(1, long, AuthSession, ""))
	ctx = WithRequest(ctx, RequestInfo{ID: "r", IP: long, UserAgent: strings.Repeat("u", 2000) + "\xff"})
	recorder := NewRecorder(store, Options{InstanceID: "i", NodeID: long})
	recorder.Record(ctx, Event{Kind: IAMUserDelete, Target: &Target{Type: TargetUser, ID: long, Name: "a\xffb"}})

	row := store.rows[0]
	assert.Len(t, row.ActorName, MaxNameBytes)
	assert.Len(t, row.TargetID, MaxNameBytes)
	assert.Equal(t, "a�b", row.TargetName)
	assert.Len(t, row.NodeID, 255)
	assert.Len(t, row.UserAgent, 1024)
	assert.Len(t, row.SourceIP, 64)
}

func TestRecord_StoresMismatchingEvents(t *testing.T) {
	store := &fakeStore{}
	recorder := newTestRecorder(store)
	recorder.Record(context.Background(), Event{Kind: IAMUserDelete, Outcome: OutcomeFailure, Reason: ReasonForbidden})
	recorder.Record(context.Background(), Event{Kind: AuthLogin, Metadata: DenyMetadata{Method: "POST"}})

	require.Len(t, store.rows, 2)
	assert.Equal(t, "failure", store.rows[0].Outcome)
	assert.Equal(t, "forbidden", store.rows[0].Reason)
	assert.Equal(t, "start", store.rows[1].Type)
	assert.JSONEq(t, `{"method":"POST"}`, store.rows[1].Metadata)
}

func TestRecord_StoreErrorIsLoggedWithTheEvent(t *testing.T) {
	store := &fakeStore{err: errors.New("database is down")}
	hook := logtest.NewGlobal()
	defer hook.Reset()

	assert.NotPanics(t, func() {
		newTestRecorder(store).Record(context.Background(), Event{Kind: IAMUserDelete})
	})

	entry := hook.LastEntry()
	require.NotNil(t, entry)
	assert.Equal(t, log.ErrorLevel, entry.Level)
	assert.Equal(t, "iam.user", entry.Data["event_code"])
	assert.Equal(t, "7b0c0000-0000-4000-8000-000000000001", entry.Data["event_id"])
}

func TestNewRecorder_Defaults(t *testing.T) {
	store := &fakeStore{}
	NewRecorder(store, Options{InstanceID: "i"}).Record(context.Background(), Event{Kind: IAMUserDelete})
	assert.Len(t, store.rows[0].EventID, 36)
}

func TestRecord_OversizedMetadataIsReplaced(t *testing.T) {
	store := &fakeStore{}
	hook := logtest.NewGlobal()
	defer hook.Reset()

	newTestRecorder(store).Record(context.Background(), Event{
		Kind:     IAMProjectRoleChange,
		Metadata: ProjectRoleMetadata{OldRole: strings.Repeat("x", MaxMetadataBytes+1)},
	})

	require.Len(t, store.rows, 1)
	assert.JSONEq(t, `{"truncated":true}`, store.rows[0].Metadata)
	entry := hook.LastEntry()
	require.NotNil(t, entry)
	assert.Equal(t, log.ErrorLevel, entry.Level)
	assert.Equal(t, "7b0c0000-0000-4000-8000-000000000001", entry.Data["event_id"])
	assert.Equal(t, "iam.project_role", entry.Data["event_code"])
	assert.Greater(t, entry.Data["metadata_bytes"], MaxMetadataBytes)
}

// A client that disconnects must not cancel the record of its own request.
func TestRecord_StoresWithOwnTimeout(t *testing.T) {
	store := &fakeStore{}
	requestCtx, cancel := context.WithCancel(context.Background())
	cancel()

	newTestRecorder(store).Record(requestCtx, Event{Kind: IAMUserDelete})

	require.Len(t, store.rows, 1)
	assert.NoError(t, store.ctxErr)
	assert.WithinDuration(t, time.Now().Add(storeTimeout), store.deadline, time.Second)
}
