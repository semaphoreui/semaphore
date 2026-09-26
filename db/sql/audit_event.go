package sql

import (
	"encoding/json"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/semaphoreui/semaphore/db"
)

type auditEventRow struct {
	Seq             int64
	ID              string
	Timestamp       time.Time
	SchemaVersion   string
	EventCode       string
	Category        string
	Type            string
	Action          string
	Outcome         string
	ActorType       string
	ActorID         *string
	ActorName       *string
	SourceIP        *string
	SourceUserAgent *string
	TargetType      *string
	TargetID        *string
	TargetName      *string
	ProjectID       *string
	RequestID       *string
	InstanceID      string
	NodeID          *string
	Metadata        *string
}

func (d *SqlDb) CreateAuditEvent(event db.AuditEvent) (db.AuditEvent, error) {
	if len(event.Metadata) > 0 && !json.Valid(event.Metadata) {
		return db.AuditEvent{}, errors.New("create audit event: metadata is invalid JSON")
	}
	row := auditEventToRow(event)
	tx, err := d.Sql().Begin()
	if err != nil {
		return db.AuditEvent{}, err
	}
	defer func() { _ = tx.Rollback() }()

	// Hold the singleton row through commit so visible sequences always form a committed prefix.
	result, err := tx.Exec(
		d.PrepareQuery("update audit_event_sequence set last_seq=last_seq+1 where id=?"),
		1,
	)
	if err != nil {
		return db.AuditEvent{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return db.AuditEvent{}, err
	}
	if affected != 1 {
		return db.AuditEvent{}, errors.New("create audit event: sequence row is missing")
	}
	if err := tx.SelectOne(
		&row.Seq,
		d.PrepareQuery("select last_seq from audit_event_sequence where id=?"),
		1,
	); err != nil {
		return db.AuditEvent{}, err
	}
	query, args, err := sq.Insert("audit_event").SetMap(map[string]any{
		"seq":               row.Seq,
		"event_id":          row.ID,
		"occurred_at":       row.Timestamp,
		"schema_version":    row.SchemaVersion,
		"event_code":        row.EventCode,
		"category":          row.Category,
		"type":              row.Type,
		"action":            row.Action,
		"outcome":           row.Outcome,
		"actor_type":        row.ActorType,
		"actor_id":          row.ActorID,
		"actor_name":        row.ActorName,
		"source_ip":         row.SourceIP,
		"source_user_agent": row.SourceUserAgent,
		"target_type":       row.TargetType,
		"target_id":         row.TargetID,
		"target_name":       row.TargetName,
		"project_id":        row.ProjectID,
		"request_id":        row.RequestID,
		"instance_id":       row.InstanceID,
		"node_id":           row.NodeID,
		"metadata":          row.Metadata,
	}).ToSql()
	if err != nil {
		return db.AuditEvent{}, err
	}
	if _, err := tx.Exec(d.PrepareQuery(query), args...); err != nil {
		return db.AuditEvent{}, err
	}
	if err := tx.Commit(); err != nil {
		return db.AuditEvent{}, err
	}
	event.Seq = row.Seq
	return event, nil
}

func auditEventToRow(event db.AuditEvent) auditEventRow {
	row := auditEventRow{
		ID:            event.ID,
		Timestamp:     event.Timestamp,
		SchemaVersion: event.SchemaVersion,
		EventCode:     event.EventCode,
		Category:      event.Category,
		Type:          event.Type,
		Action:        event.Action,
		Outcome:       event.Outcome,
		ActorType:     event.Actor.Type,
		ActorID:       stringOrNil(event.Actor.ID),
		ActorName:     stringOrNil(event.Actor.Name),
		RequestID:     stringOrNil(event.RequestID),
		InstanceID:    event.InstanceID,
		NodeID:        stringOrNil(event.NodeID),
		Metadata:      stringOrNil(string(event.Metadata)),
	}
	if event.Source != nil {
		row.SourceIP = stringOrNil(event.Source.IP)
		row.SourceUserAgent = stringOrNil(event.Source.UserAgent)
	}
	if event.Target != nil {
		row.TargetType = stringOrNil(event.Target.Type)
		row.TargetID = stringOrNil(event.Target.ID)
		row.TargetName = stringOrNil(event.Target.Name)
	}
	if event.Scope != nil {
		row.ProjectID = stringOrNil(event.Scope.ProjectID)
	}
	return row
}

func stringOrNil(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
