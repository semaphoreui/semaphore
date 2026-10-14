package audit

import (
	"context"
	"math/rand/v2"
	"time"

	"github.com/semaphoreui/semaphore/db"
	log "github.com/sirupsen/logrus"
)

const (
	retentionInterval = time.Hour
	retentionBatch    = 1000
	retentionTimeout  = 10 * time.Minute
)

// runRetention deletes old events at start and then hourly. Every node runs it, a second delete finds nothing.
func runRetention(ctx context.Context, store db.AuditEventManager, recorder Recorder, days int, done chan<- struct{}) {
	defer close(done)
	for {
		pruneAuditEvents(ctx, store, recorder, days)
		// Jitter keeps HA nodes from deleting at the same moment.
		delay := retentionInterval - retentionInterval/10 + rand.N(retentionInterval/5)
		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}
	}
}

func pruneAuditEvents(ctx context.Context, store db.AuditEventManager, recorder Recorder, days int) {
	parent := ctx
	ctx, cancel := context.WithTimeout(parent, retentionTimeout)
	defer cancel()
	deleted, lastSeq, err := store.DeleteAuditEventsOlderThan(ctx, days, retentionBatch)
	if err != nil {
		entry := log.WithError(err).WithField("context", "audit_retention")
		switch {
		case parent.Err() != nil:
			entry.Debug("Audit retention stopped")
		case ctx.Err() != nil:
			entry.WithField("deleted", deleted).Warn("Audit retention pass timed out, continues next hour")
		default:
			entry.Error("Failed to delete old audit events")
		}
	}
	if deleted == 0 {
		return
	}
	recorder.Record(WithActor(context.Background(), SystemActor(ComponentRetention)), Event{
		Kind:     AuditRetentionDelete,
		Metadata: RetentionMetadata{Deleted: deleted, LastSeq: lastSeq, RetentionDays: days},
	})
}
