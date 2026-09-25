package db

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAuditEvent(t *testing.T) {
	event := NewAuditEvent()
	id, err := uuid.Parse(event.ID)
	require.NoError(t, err)
	assert.Equal(t, uuid.Version(4), id.Version())
	assert.Equal(t, time.UTC, event.Timestamp.Location())
	assert.Equal(t, AuditSchemaVersion, event.SchemaVersion)
}

func TestAuditEventJSON(t *testing.T) {
	event := AuditEvent{
		Seq:           42,
		ID:            "85f5652e-34e9-4e24-8f88-6f5e06fc2fc4",
		Timestamp:     time.Date(2026, 9, 17, 8, 15, 30, 0, time.UTC),
		SchemaVersion: AuditSchemaVersion,
		EventCode:     AuditEventCodeInventory,
		Category:      AuditCategoryResource,
		Type:          AuditTypeCreation,
		Action:        AuditActionCreate,
		Outcome:       AuditOutcomeSuccess,
		Actor:         &AuditActor{Type: "user", ID: "42", Name: "alice"},
		Target:        &AuditTarget{Type: "inventory", ID: "9", Name: "production"},
		InstanceID:    "semaphore-prod",
		Metadata:      json.RawMessage(`{}`),
	}

	payload, err := json.Marshal(event)
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"id":"85f5652e-34e9-4e24-8f88-6f5e06fc2fc4",
		"timestamp":"2026-09-17T08:15:30Z",
		"schema_version":"1",
		"event_code":"resource.inventory",
		"category":"resource",
		"type":"creation",
		"action":"create",
		"outcome":"success",
		"actor":{"type":"user","id":"42","name":"alice"},
		"target":{"type":"inventory","id":"9","name":"production"},
		"instance_id":"semaphore-prod",
		"metadata":{}
	}`, string(payload))
}
