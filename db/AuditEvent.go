package db

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/semaphoreui/semaphore/pkg/tz"
)

const AuditSchemaVersion = "1"

const (
	AuditEventCodeProject    = "resource.project"
	AuditEventCodeInventory  = "resource.inventory"
	AuditEventCodeCredential = "resource.credential"
	AuditEventCodeRepository = "resource.repository"
	AuditEventCodeView       = "resource.view"
	AuditEventCodeTemplate   = "resource.template"
	AuditCategoryResource    = "resource"
	AuditTypeCreation        = "creation"
	AuditTypeChange          = "change"
	AuditTypeDeletion        = "deletion"
	AuditActionCreate        = "create"
	AuditActionUpdate        = "update"
	AuditActionDelete        = "delete"
	AuditOutcomeSuccess      = "success"
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
	Actor         *AuditActor     `json:"actor"`
	Source        *AuditSource    `json:"source,omitempty"`
	Target        *AuditTarget    `json:"target,omitempty"`
	Scope         *AuditScope     `json:"scope,omitempty"`
	RequestID     string          `json:"request_id,omitempty"`
	InstanceID    string          `json:"instance_id"`
	NodeID        string          `json:"node_id,omitempty"`
	Metadata      json.RawMessage `json:"metadata,omitempty"`
}

type AuditActor struct {
	Type string `json:"type"`
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
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

func NewAuditEvent() AuditEvent {
	return AuditEvent{ID: uuid.NewString(), Timestamp: tz.Now(), SchemaVersion: AuditSchemaVersion}
}
