package audit

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

func EnvelopeFromRow(row db.AuditEvent) (Envelope, error) {
	if !json.Valid([]byte(row.Metadata)) {
		return Envelope{}, fmt.Errorf("audit event %d: metadata is not JSON", row.Seq)
	}

	envelope := Envelope{
		EventID:       row.EventID,
		Seq:           row.Seq,
		Timestamp:     row.Created.UTC().Format(TimestampLayout),
		SchemaVersion: row.SchemaVersion,
		Category:      row.Category,
		EventCode:     row.EventCode,
		Type:          Type(row.Type),
		Action:        row.Action,
		Outcome:       Outcome(row.Outcome),
		Reason:        Reason(row.Reason),
		Actor: Actor{
			Type:             ActorType(row.ActorType),
			ID:               row.ActorID,
			Name:             row.ActorName,
			Auth:             AuthMethod(row.ActorAuth),
			TokenFingerprint: row.ActorTokenFingerprint,
		},
		RequestID:  row.RequestID,
		InstanceID: row.InstanceID,
		NodeID:     row.NodeID,
		Metadata:   json.RawMessage(row.Metadata),
	}

	if row.RequestID != "" {
		envelope.Source = &Source{IP: row.SourceIP, UserAgent: row.UserAgent}
	}
	if row.TargetType != "" {
		envelope.Target = &Target{Type: row.TargetType, ID: row.TargetID, Name: row.TargetName}
	}
	if row.ProjectID != nil {
		envelope.Scope = &Scope{ProjectID: strconv.Itoa(*row.ProjectID)}
	}

	return envelope, nil
}
