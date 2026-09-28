package db

import (
	"encoding/json"
	"fmt"
	"net/netip"
	"time"

	"github.com/google/uuid"
	"github.com/semaphoreui/semaphore/pkg/tz"
)

const AuditSchemaVersion = "1"

type AuditActorType string

const (
	AuditActorTypeUser        AuditActorType = "user"
	AuditActorTypeAnonymous   AuditActorType = "anonymous"
	AuditActorTypeAPIToken    AuditActorType = "api_token"
	AuditActorTypeRunner      AuditActorType = "runner"
	AuditActorTypeSystem      AuditActorType = "system"
	AuditActorTypeIntegration AuditActorType = "integration"

	AuditEventCodeProject       = "resource.project"
	AuditEventCodeInventory     = "resource.inventory"
	AuditEventCodeCredential    = "resource.credential"
	AuditEventCodeRepository    = "resource.repository"
	AuditEventCodeView          = "resource.view"
	AuditEventCodeTemplate      = "resource.template"
	AuditEventCodeSchedule      = "resource.schedule"
	AuditEventCodeEnvironment   = "resource.environment"
	AuditEventCodeSecretStorage = "resource.secret_storage"
	AuditEventCodeMembership    = "iam.membership"
	AuditEventCodeProjectRole   = "iam.project_role"
	AuditCategoryResource       = "resource"
	AuditCategoryIAM            = "iam"
	AuditTypeCreation           = "creation"
	AuditTypeChange             = "change"
	AuditTypeDeletion           = "deletion"
	AuditActionAdd              = "add"
	AuditActionRemove           = "remove"
	AuditActionChange           = "change"
	AuditActionCreate           = "create"
	AuditActionUpdate           = "update"
	AuditActionDelete           = "delete"
	AuditOutcomeSuccess         = "success"
)

type AuditEvent struct {
	Seq           int64           `json:"-"`
	ID            string          `json:"id"`
	Timestamp     time.Time       `json:"timestamp"`
	SchemaVersion string          `json:"schema_version"`
	EventCode     string          `json:"event_code"`
	Category      string          `json:"category"`
	Type          string          `json:"type"`
	Action        string          `json:"action"`
	Outcome       string          `json:"outcome"`
	Actor         AuditActor      `json:"actor"`
	Source        *AuditSource    `json:"source,omitempty"`
	Target        *AuditTarget    `json:"target,omitempty"`
	Scope         *AuditScope     `json:"scope,omitempty"`
	RequestID     string          `json:"request_id,omitempty"`
	InstanceID    string          `json:"instance_id"`
	NodeID        string          `json:"node_id,omitempty"`
	Metadata      json.RawMessage `json:"metadata,omitempty"`
}

type AuditActor struct {
	Type AuditActorType `json:"type"`
	ID   string         `json:"id,omitempty"`
	Name string         `json:"name,omitempty"`
}

type AuditSource struct {
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent,omitempty"`
}

type AuditTarget struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

type AuditScope struct {
	ProjectID string `json:"project_id"`
}

func NewAuditEvent(actor AuditActor) AuditEvent {
	return AuditEvent{
		ID:            uuid.NewString(),
		Timestamp:     tz.Now(),
		SchemaVersion: AuditSchemaVersion,
		Actor:         actor,
	}
}

func (event AuditEvent) Validate() error {
	if !isUUIDv4(event.ID) {
		return fmt.Errorf("event ID must be a UUIDv4")
	}
	if event.Timestamp.IsZero() || event.Timestamp.Location() != time.UTC {
		return fmt.Errorf("timestamp must be UTC")
	}
	if event.SchemaVersion != AuditSchemaVersion {
		return fmt.Errorf("schema version must be %q", AuditSchemaVersion)
	}
	for _, field := range []struct {
		name  string
		value string
	}{
		{"event code", event.EventCode},
		{"category", event.Category},
		{"type", event.Type},
		{"action", event.Action},
		{"outcome", event.Outcome},
		{"actor type", string(event.Actor.Type)},
		{"instance ID", event.InstanceID},
	} {
		if field.value == "" {
			return fmt.Errorf("%s is required", field.name)
		}
	}
	if event.Source != nil {
		if _, err := netip.ParseAddr(event.Source.IP); err != nil {
			return fmt.Errorf("source IP is invalid")
		}
	}
	if event.Target != nil {
		if event.Target.Type == "" {
			return fmt.Errorf("target type is required")
		}
		if event.Target.ID == "" {
			return fmt.Errorf("target ID is required")
		}
	}
	if event.Scope != nil && event.Scope.ProjectID == "" {
		return fmt.Errorf("project ID is required")
	}
	if event.RequestID != "" && !isUUIDv4(event.RequestID) {
		return fmt.Errorf("request ID must be a UUIDv4")
	}
	if len(event.Metadata) > 0 && !json.Valid(event.Metadata) {
		return fmt.Errorf("metadata is invalid JSON")
	}
	return nil
}

func isUUIDv4(value string) bool {
	id, err := uuid.Parse(value)
	return err == nil && id.Version() == 4
}
