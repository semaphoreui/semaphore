package sql

import (
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
)

const (
	auditEventBatchMax = 1000
)

type auditEventRow struct {
	Seq             int64     `db:"seq"`
	ID              string    `db:"event_id"`
	Timestamp       time.Time `db:"occurred_at"`
	SchemaVersion   string    `db:"schema_version"`
	EventCode       string    `db:"event_code"`
	Category        string    `db:"category"`
	Type            string    `db:"type"`
	Action          string    `db:"action"`
	Outcome         string    `db:"outcome"`
	ActorType       string    `db:"actor_type"`
	ActorID         *string   `db:"actor_id"`
	ActorName       *string   `db:"actor_name"`
	SourceIP        *string   `db:"source_ip"`
	SourceUserAgent *string   `db:"source_user_agent"`
	TargetType      *string   `db:"target_type"`
	TargetID        *string   `db:"target_id"`
	TargetName      *string   `db:"target_name"`
	ProjectID       *string   `db:"project_id"`
	RequestID       *string   `db:"request_id"`
	InstanceID      string    `db:"instance_id"`
	NodeID          *string   `db:"node_id"`
	Metadata        *string   `db:"metadata"`
}

func (d *SqlDb) CreateAuditEvent(event db.AuditEvent) (db.AuditEvent, error) {
	if err := event.Validate(); err != nil {
		return db.AuditEvent{}, err
	}
	row := auditEventToRow(event)
	query, args, err := sq.Insert("audit_event").
		Columns(
			"event_id", "occurred_at", "schema_version", "event_code", "category", "type",
			"action", "outcome", "actor_type", "actor_id", "actor_name", "source_ip",
			"source_user_agent", "target_type", "target_id", "target_name", "project_id",
			"request_id", "instance_id", "node_id", "metadata",
		).
		Values(
			row.ID, row.Timestamp, row.SchemaVersion, row.EventCode, row.Category, row.Type,
			row.Action, row.Outcome, row.ActorType, row.ActorID, row.ActorName, row.SourceIP,
			row.SourceUserAgent, row.TargetType, row.TargetID, row.TargetName, row.ProjectID,
			row.RequestID, row.InstanceID, row.NodeID, row.Metadata,
		).
		ToSql()
	if err != nil {
		return db.AuditEvent{}, err
	}
	seq, err := d.insertAuditEvent(query, args...)
	if err != nil {
		return db.AuditEvent{}, err
	}
	event.Seq = seq
	return event, nil
}
func (d *SqlDb) insertAuditEvent(query string, args ...any) (int64, error) {
	if d.GetDialect() == util.DbDriverPostgres {
		var seq int64
		err := d.Sql().QueryRow(d.PrepareQuery(query+" returning seq"), args...).Scan(&seq)
		return seq, err
	}
	result, err := d.Sql().Exec(d.PrepareQuery(query), args...)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}
func (d *SqlDb) GetAuditEventsAfter(seq int64, limit int) ([]db.AuditEvent, error) {
	if seq < 0 || limit <= 0 || limit > auditEventBatchMax {
		return nil, fmt.Errorf("audit event cursor or batch limit is invalid")
	}
	var rows []auditEventRow
	query, args, err := sq.Select(
		"seq", "event_id", "occurred_at", "schema_version", "event_code", "category", "type",
		"action", "outcome", "actor_type", "actor_id", "actor_name", "source_ip",
		"source_user_agent", "target_type", "target_id", "target_name", "project_id",
		"request_id", "instance_id", "node_id", "metadata",
	).
		From("audit_event").
		Where("seq > ?", seq).
		OrderBy("seq asc").
		Suffix("limit ?", limit).
		ToSql()
	if err != nil {
		return nil, err
	}
	if _, err := d.selectAll(&rows, query, args...); err != nil {
		return nil, err
	}
	events := make([]db.AuditEvent, 0, len(rows))
	for _, row := range rows {
		events = append(events, auditEventFromRow(row))
	}
	return events, nil
}
func auditEventFromRow(r auditEventRow) db.AuditEvent {
	e := db.AuditEvent{Seq: r.Seq, ID: r.ID, Timestamp: r.Timestamp.UTC(), SchemaVersion: r.SchemaVersion, EventCode: r.EventCode, Category: r.Category, Type: r.Type, Action: r.Action, Outcome: r.Outcome, Actor: &db.AuditActor{Type: r.ActorType}, InstanceID: r.InstanceID}
	if r.ActorID != nil {
		e.Actor.ID = *r.ActorID
	}
	if r.ActorName != nil {
		e.Actor.Name = *r.ActorName
	}
	if r.SourceIP != nil {
		e.Source = &db.AuditSource{IP: *r.SourceIP}
		if r.SourceUserAgent != nil {
			e.Source.UserAgent = *r.SourceUserAgent
		}
	}
	if r.TargetType != nil {
		e.Target = &db.AuditTarget{Type: *r.TargetType}
		if r.TargetID != nil {
			e.Target.ID = *r.TargetID
		}
		if r.TargetName != nil {
			e.Target.Name = *r.TargetName
		}
	}
	if r.ProjectID != nil {
		e.Scope = &db.AuditScope{ProjectID: *r.ProjectID}
	}
	if r.RequestID != nil {
		e.RequestID = *r.RequestID
	}
	if r.NodeID != nil {
		e.NodeID = *r.NodeID
	}
	if r.Metadata != nil {
		e.Metadata = []byte(*r.Metadata)
	}
	return e
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
