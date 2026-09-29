package audit

import (
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvelopeFromRow(t *testing.T) {
	projectID := 12
	envelope, err := EnvelopeFromRow(db.AuditEvent{
		Seq: 42, EventID: "e1", Created: time.Date(2026, 9, 29, 10, 0, 0, 123456000, time.UTC),
		SchemaVersion: "1", Category: "iam", EventCode: "iam.api_token", Type: "creation", Action: "create",
		Outcome: "success", ActorType: "user", ActorID: "7", ActorName: "alice", ActorAuth: "api_token",
		ActorTokenFingerprint: "ff", SourceIP: "10.0.0.5", UserAgent: "curl/8.5", TargetType: "api_token",
		TargetID: "ff", TargetName: "ci", ProjectID: &projectID, RequestID: "r1", InstanceID: "prod-eu",
		NodeID: "node-a", Metadata: `{"a":1}`,
	})
	require.NoError(t, err)

	assert.Equal(t, Envelope{
		EventID: "e1", Seq: 42, Timestamp: "2026-09-29T10:00:00.123456Z", SchemaVersion: "1",
		Category: "iam", EventCode: "iam.api_token", Type: TypeCreation, Action: "create", Outcome: OutcomeSuccess,
		Actor:     Actor{Type: ActorUser, ID: "7", Name: "alice", Auth: AuthAPIToken, TokenFingerprint: "ff"},
		Source:    &Source{IP: "10.0.0.5", UserAgent: "curl/8.5"},
		Target:    &Target{Type: "api_token", ID: "ff", Name: "ci"},
		Scope:     &Scope{ProjectID: "12"},
		RequestID: "r1", InstanceID: "prod-eu", NodeID: "node-a",
		Metadata: []byte(`{"a":1}`),
	}, envelope)
}

func TestEnvelopeFromRow_OptionalParts(t *testing.T) {
	envelope, err := EnvelopeFromRow(db.AuditEvent{ActorType: "system", ActorName: "server", Metadata: "{}"})
	require.NoError(t, err)
	assert.Nil(t, envelope.Source, "no request ID means not an HTTP event")
	assert.Nil(t, envelope.Target)
	assert.Nil(t, envelope.Scope)
}

func TestEnvelopeFromRow_InvalidMetadata(t *testing.T) {
	_, err := EnvelopeFromRow(db.AuditEvent{Seq: 3, Metadata: "not json"})
	assert.Error(t, err)
}
