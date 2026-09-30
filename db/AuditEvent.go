package db

import (
	"context"
	"time"
)

type AuditEvent struct {
	Seq                   int64     `db:"seq"`
	EventID               string    `db:"event_id"`
	Created               time.Time `db:"created"`
	SchemaVersion         string    `db:"schema_version"`
	Category              string    `db:"category"`
	EventCode             string    `db:"event_code"`
	Type                  string    `db:"type"`
	Action                string    `db:"action"`
	Outcome               string    `db:"outcome"`
	Reason                string    `db:"reason"`
	ActorType             string    `db:"actor_type"`
	ActorID               string    `db:"actor_id"`
	ActorName             string    `db:"actor_name"`
	ActorAuth             string    `db:"actor_auth"`
	ActorTokenFingerprint string    `db:"actor_token_fingerprint"`
	SourceIP              string    `db:"source_ip"`
	UserAgent             string    `db:"user_agent"`
	TargetType            string    `db:"target_type"`
	TargetID              string    `db:"target_id"`
	TargetName            string    `db:"target_name"`
	ProjectID             *int      `db:"project_id"`
	RequestID             string    `db:"request_id"`
	InstanceID            string    `db:"instance_id"`
	NodeID                string    `db:"node_id"`
	Metadata              string    `db:"metadata"`
}

type AuditEventManager interface {
	CreateAuditEvent(ctx context.Context, event AuditEvent) (AuditEvent, error)
}
