package audit

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/semaphoreui/semaphore/db"
	log "github.com/sirupsen/logrus"
)

// A stuck lock on the sequence row must not hang the request.
const storeTimeout = 5 * time.Second

type Recorder interface {
	Record(ctx context.Context, event Event)
}

type Nop struct{}

func (Nop) Record(context.Context, Event) {}

type Options struct {
	InstanceID string
	NodeID     string
	NewID      func() string
}

type recorder struct {
	store db.AuditEventManager
	opts  Options
}

func NewRecorder(store db.AuditEventManager, opts Options) Recorder {
	if opts.NewID == nil {
		opts.NewID = uuid.NewString
	}
	return &recorder{store: store, opts: opts}
}

func (r *recorder) Record(ctx context.Context, event Event) {
	eventID := r.opts.NewID()
	fields := log.Fields{"context": "audit", "event_code": event.Kind.Code(), "action": event.Kind.Action(), "event_id": eventID}

	if err := Validate(event); err != nil {
		log.WithError(err).WithFields(fields).Error("Audit event does not match the catalog")
	}

	row, err := r.row(ctx, eventID, event)
	if err != nil {
		log.WithError(err).WithFields(fields).Error("Failed to build audit event")
		return
	}

	storeCtx, cancel := context.WithTimeout(context.Background(), storeTimeout)
	defer cancel()
	if _, err = r.store.CreateAuditEvent(storeCtx, row); err != nil {
		log.WithError(err).WithFields(fields).Error("Failed to store audit event")
	}
}

func (r *recorder) row(ctx context.Context, eventID string, event Event) (db.AuditEvent, error) {
	metadata := "{}"
	if event.Metadata != nil {
		raw, err := json.Marshal(event.Metadata)
		if err != nil {
			return db.AuditEvent{}, err
		}
		metadata = string(raw)
		if len(raw) > MaxMetadataBytes {
			// Cut JSON would not decode.
			log.WithFields(log.Fields{
				"context": "audit", "event_id": eventID, "event_code": event.Kind.Code(),
				"action": event.Kind.Action(), "metadata_bytes": len(raw),
			}).Error("Audit event metadata is too large and was dropped")
			metadata = `{"truncated":true}`
		}
	}

	entry, _ := Lookup(event.Kind)
	actor := ActorFrom(ctx)

	row := db.AuditEvent{
		EventID:               eventID,
		SchemaVersion:         SchemaVersion,
		Category:              event.Kind.Category(),
		EventCode:             event.Kind.Code(),
		Type:                  string(entry.Type),
		Action:                event.Kind.Action(),
		Outcome:               string(event.outcome()),
		Reason:                string(event.Reason),
		ActorType:             string(actor.Type),
		ActorID:               actor.ID,
		ActorName:             TruncateName(actor.Name, MaxNameBytes),
		ActorAuth:             string(actor.Auth),
		ActorTokenFingerprint: actor.TokenFingerprint,
		InstanceID:            r.opts.InstanceID,
		NodeID:                TruncateName(r.opts.NodeID, maxNodeIDBytes),
		Metadata:              metadata,
	}

	if request, ok := RequestFrom(ctx); ok {
		row.RequestID = request.ID
		row.SourceIP = TruncateName(request.IP, maxIPBytes)
		row.UserAgent = TruncateName(request.UserAgent, maxUserAgentBytes)
	}

	if event.Target != nil {
		row.TargetType = event.Target.Type
		row.TargetID = TruncateName(event.Target.ID, MaxNameBytes)
		row.TargetName = TruncateName(event.Target.Name, MaxNameBytes)
	}

	if event.ProjectID > 0 {
		projectID := event.ProjectID
		row.ProjectID = &projectID
	}

	return row, nil
}
