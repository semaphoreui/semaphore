package db

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAuditEnvelopeSupportsFutureOptionalFields(t *testing.T) {
	event := NewAuditEvent()
	event.EventCode, event.Category, event.Type, event.Action, event.Outcome, event.InstanceID = "auth.login", "authentication", "authentication", "login", "failure", "test"
	event.Actor = &AuditActor{Type: "anonymous"}
	assert.NoError(t, event.Validate())
	payload, err := json.Marshal(event)
	assert.NoError(t, err)
	assert.NotContains(t, string(payload), `"source"`)
	assert.NotContains(t, string(payload), `"target"`)
	assert.NotContains(t, string(payload), `"scope"`)
}

func TestAuditResourceValidation(t *testing.T) {
	event := validResourceEvent()
	assert.NoError(t, event.Validate())
	assert.NoError(t, event.ValidateResourceEvent())
	event.Outcome = "failure"
	assert.NoError(t, event.Validate())
	assert.Error(t, event.ValidateResourceEvent())
}

func TestAuditMetadataValidation(t *testing.T) {
	for _, metadata := range []json.RawMessage{json.RawMessage(`{"unknown":"value"}`), json.RawMessage(`{"nested":{"value":true}}`), json.RawMessage(`{"` + strings.Repeat("a", AuditMetadataMaxBytes) + `":"x"}`)} {
		event := validResourceEvent()
		event.Metadata = metadata
		assert.Error(t, event.Validate())
	}
}

func TestAuditTargetTypeValidation(t *testing.T) {
	event := validResourceEvent()
	event.Target.Type = strings.Repeat("a", 64)
	assert.NoError(t, event.Validate())
	event.Target.Type += "a"
	assert.Error(t, event.Validate())
}

func TestAuditUUIDValidationRequiresCanonicalForm(t *testing.T) {
	for _, tt := range []struct {
		name  string
		field string
		value func(string) string
	}{
		{name: "event uppercase", field: "event", value: strings.ToUpper},
		{name: "event without hyphens", field: "event", value: func(value string) string { return strings.ReplaceAll(value, "-", "") }},
		{name: "request uppercase", field: "request", value: strings.ToUpper},
		{name: "request without hyphens", field: "request", value: func(value string) string { return strings.ReplaceAll(value, "-", "") }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			event := validResourceEvent()
			if tt.field == "event" {
				event.ID = tt.value(event.ID)
			} else {
				event.RequestID = tt.value(event.RequestID)
			}
			assert.Error(t, event.Validate())
		})
	}
}

func TestNewAuditEvent(t *testing.T) {
	event := NewAuditEvent()
	id, err := uuid.Parse(event.ID)
	assert.NoError(t, err)
	assert.Equal(t, uuid.Version(4), id.Version())
	assert.Equal(t, time.UTC, event.Timestamp.Location())
}

func validResourceEvent() AuditEvent {
	event := NewAuditEvent()
	event.EventCode, event.Category, event.Type, event.Action, event.Outcome, event.InstanceID = AuditEventCodeInventory, AuditCategoryResource, AuditTypeCreation, AuditActionCreate, AuditOutcomeSuccess, "test"
	event.Actor = &AuditActor{Type: "user", ID: "42"}
	event.Source = &AuditSource{IP: "203.0.113.10"}
	event.Target = &AuditTarget{Type: "inventory", ID: "1"}
	event.Scope = &AuditScope{ProjectID: "1"}
	event.RequestID = uuid.NewString()
	return event
}
