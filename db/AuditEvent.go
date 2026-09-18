package db

import (
	"encoding/json"
	"fmt"
	"net/netip"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/semaphoreui/semaphore/pkg/tz"
)

const AuditSchemaVersion = "1"

const (
	AuditEventCodeProject    = "resource.project"
	AuditEventCodeInventory  = "resource.inventory"
	AuditEventCodeTemplate   = "resource.template"
	AuditEventCodeSchedule   = "resource.schedule"
	AuditEventCodeRepository = "resource.repository"
	AuditEventCodeCredential = "resource.credential"
	AuditCategoryResource    = "resource"
	AuditTypeCreation        = "creation"
	AuditTypeChange          = "change"
	AuditTypeDeletion        = "deletion"
	AuditActionCreate        = "create"
	AuditActionUpdate        = "update"
	AuditActionDelete        = "delete"
	AuditOutcomeSuccess      = "success"
	AuditMetadataMaxBytes    = 16 * 1024
)

const (
	auditStringMaxLength = 255
	auditKindMaxLength   = 64
)

type AuditEvent struct {
	Seq           int64           `db:"seq" json:"-"`
	ID            string          `db:"event_id" json:"id"`
	Timestamp     time.Time       `db:"occurred_at" json:"timestamp"`
	SchemaVersion string          `db:"schema_version" json:"schema_version"`
	EventCode     string          `db:"event_code" json:"event_code"`
	Category      string          `db:"category" json:"category"`
	Type          string          `db:"type" json:"type"`
	Action        string          `db:"action" json:"action"`
	Outcome       string          `db:"outcome" json:"outcome"`
	Actor         *AuditActor     `db:"-" json:"actor"`
	Source        *AuditSource    `db:"-" json:"source,omitempty"`
	Target        *AuditTarget    `db:"-" json:"target,omitempty"`
	Scope         *AuditScope     `db:"-" json:"scope,omitempty"`
	RequestID     string          `db:"request_id" json:"request_id,omitempty"`
	InstanceID    string          `db:"instance_id" json:"instance_id"`
	NodeID        string          `db:"node_id" json:"node_id,omitempty"`
	Metadata      json.RawMessage `db:"metadata" json:"metadata,omitempty"`
}
type AuditActor struct {
	Type string `db:"actor_type" json:"type"`
	ID   string `db:"actor_id" json:"id,omitempty"`
	Name string `db:"actor_name" json:"name,omitempty"`
}
type AuditSource struct {
	IP        string `db:"source_ip" json:"ip"`
	UserAgent string `db:"source_user_agent" json:"user_agent,omitempty"`
}
type AuditTarget struct {
	Type string `db:"target_type" json:"type"`
	ID   string `db:"target_id" json:"id"`
	Name string `db:"target_name" json:"name,omitempty"`
}
type AuditScope struct {
	ProjectID string `db:"project_id" json:"project_id"`
}

func NewAuditEvent() AuditEvent {
	return AuditEvent{ID: uuid.NewString(), Timestamp: tz.Now(), SchemaVersion: AuditSchemaVersion}
}

// Validate checks envelope invariants shared by every current and future family.
func (event AuditEvent) Validate() error {
	if !isUUIDv4(event.ID) {
		return fmt.Errorf("audit event id must be a UUIDv4")
	}
	if event.Timestamp.IsZero() || event.Timestamp.Location() != time.UTC {
		return fmt.Errorf("audit event timestamp must be UTC")
	}
	if event.SchemaVersion != AuditSchemaVersion {
		return fmt.Errorf("audit event schema version must be %q", AuditSchemaVersion)
	}
	for _, value := range []string{event.EventCode, event.InstanceID} {
		if value == "" || len(value) > auditStringMaxLength {
			return fmt.Errorf("audit event envelope field is invalid")
		}
	}
	if len(event.NodeID) > auditStringMaxLength {
		return fmt.Errorf("audit event envelope field is invalid")
	}
	for _, value := range []string{event.Category, event.Type, event.Action, event.Outcome} {
		if value == "" || len(value) > auditKindMaxLength {
			return fmt.Errorf("audit event envelope field is invalid")
		}
	}
	if event.Actor == nil || event.Actor.Type == "" || len(event.Actor.Type) > auditKindMaxLength || len(event.Actor.ID) > auditStringMaxLength || utf8.RuneCountInString(event.Actor.Name) > auditStringMaxLength {
		return fmt.Errorf("audit event actor is invalid")
	}
	if event.Source != nil {
		if _, err := netip.ParseAddr(event.Source.IP); err != nil || len(event.Source.UserAgent) > 1024 {
			return fmt.Errorf("audit event source is invalid")
		}
	}
	if event.Target != nil && (event.Target.Type == "" || event.Target.ID == "" || len(event.Target.Type) > auditKindMaxLength || len(event.Target.ID) > auditStringMaxLength || utf8.RuneCountInString(event.Target.Name) > auditStringMaxLength) {
		return fmt.Errorf("audit event target is invalid")
	}
	if event.Scope != nil && (event.Scope.ProjectID == "" || len(event.Scope.ProjectID) > auditStringMaxLength) {
		return fmt.Errorf("audit event scope is invalid")
	}
	if event.RequestID != "" && !isUUIDv4(event.RequestID) {
		return fmt.Errorf("audit event request ID must be a UUIDv4")
	}
	return event.validateMetadata()
}

// ValidateResourceEvent adds constraints only for the initial HTTP resource family.
func (event AuditEvent) ValidateResourceEvent() error {
	if err := event.Validate(); err != nil {
		return err
	}
	if !isResourceCode(event.EventCode) || event.Category != AuditCategoryResource || event.Outcome != AuditOutcomeSuccess || !resourceActionMatchesType(event.Type, event.Action) {
		return fmt.Errorf("audit event is not a valid resource event")
	}
	if event.Actor.Type != "user" || event.Actor.ID == "" || event.Source == nil || event.Target == nil || event.Scope == nil || !isUUIDv4(event.RequestID) {
		return fmt.Errorf("audit resource event is missing required context")
	}
	return nil
}

func (event AuditEvent) validateMetadata() error {
	if len(event.Metadata) == 0 {
		return nil
	}
	if len(event.Metadata) > AuditMetadataMaxBytes {
		return fmt.Errorf("audit event metadata exceeds %d bytes", AuditMetadataMaxBytes)
	}
	var values map[string]any
	if json.Unmarshal(event.Metadata, &values) != nil || values == nil {
		return fmt.Errorf("audit event metadata must be a JSON object")
	}
	// No current resource event needs metadata. Future families must add their own
	// explicit scalar key allowlist before accepting metadata.
	if len(values) != 0 {
		return fmt.Errorf("audit event metadata contains unsupported keys")
	}
	return nil
}
func isUUIDv4(value string) bool {
	id, err := uuid.Parse(value)
	return err == nil && id.Version() == 4 && id.String() == value
}
func isResourceCode(value string) bool {
	switch value {
	case AuditEventCodeProject, AuditEventCodeInventory, AuditEventCodeTemplate, AuditEventCodeSchedule, AuditEventCodeRepository, AuditEventCodeCredential:
		return true
	}
	return false
}
func resourceActionMatchesType(t, a string) bool {
	return t == AuditTypeCreation && a == AuditActionCreate || t == AuditTypeChange && a == AuditActionUpdate || t == AuditTypeDeletion && a == AuditActionDelete
}
