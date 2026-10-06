package sql

import (
	"context"
	"time"

	"github.com/go-gorp/gorp/v3"
	"github.com/semaphoreui/semaphore/db"
)

// The counter row stays locked until commit, so committed seq has no gaps.
func (d *SqlDb) CreateAuditEvent(ctx context.Context, event db.AuditEvent) (db.AuditEvent, error) {
	// The context also bounds the wait for a free connection.
	tx, err := d.Sql().WithContext(ctx).(*gorp.DbMap).Begin()
	if err != nil {
		return event, err
	}
	exec := tx.WithContext(ctx)

	_, err = exec.Exec(d.PrepareQuery("update audit_event_sequence set last_seq = last_seq + 1 where id = 1"))
	if err != nil {
		handleRollbackError(tx.Rollback())
		return event, err
	}

	event.Seq, err = exec.SelectInt(d.PrepareQuery("select last_seq from audit_event_sequence where id = 1"))
	if err != nil {
		handleRollbackError(tx.Rollback())
		return event, err
	}

	event.Created, err = d.auditNow(exec)
	if err != nil {
		handleRollbackError(tx.Rollback())
		return event, err
	}

	if err = exec.Insert(&event); err != nil {
		handleRollbackError(tx.Rollback())
		return event, err
	}

	return event, tx.Commit()
}

// Database time, so HA nodes with skewed clocks agree.
func (d *SqlDb) auditNow(tx gorp.SqlExecutor) (time.Time, error) {
	var now time.Time
	switch d.Sql().Dialect.(type) {
	case gorp.MySQLDialect:
		// Text, because the driver would read a DATETIME in the DSN loc.
		var text string
		if err := tx.QueryRow("select date_format(utc_timestamp(6), '%Y-%m-%d %H:%i:%s.%f')").Scan(&text); err != nil {
			return now, err
		}
		return time.ParseInLocation("2006-01-02 15:04:05.000000", text, time.UTC)
	case gorp.PostgresDialect:
		err := tx.QueryRow("select clock_timestamp()").Scan(&now)
		return now.UTC(), err
	default:
		var text string
		if err := tx.QueryRow("select strftime('%Y-%m-%d %H:%M:%f', 'now')").Scan(&text); err != nil {
			return now, err
		}
		return time.Parse("2006-01-02 15:04:05.000", text)
	}
}

// Walks seq ranges up to the last old event, and keeps a newer one there if the database clock went back.
func (d *SqlDb) DeleteAuditEventsBefore(ctx context.Context, cutoff time.Time, batch int) (int64, int64, error) {
	batch = max(batch, 1)
	exec := d.Sql().WithContext(ctx)
	last, err := exec.SelectNullInt(d.PrepareQuery("select max(seq) from audit_event where created < ?"), cutoff.UTC())
	if err != nil || !last.Valid {
		return 0, 0, err
	}
	first, err := exec.SelectNullInt(d.PrepareQuery("select min(seq) from audit_event"))
	if err != nil || !first.Valid || first.Int64 > last.Int64 {
		return 0, 0, err
	}
	var deleted, finished int64
	for from := first.Int64; from <= last.Int64; from = finished + 1 {
		upTo := min(from+int64(batch)-1, last.Int64)
		res, err := exec.Exec(d.PrepareQuery("delete from audit_event where seq >= ? and seq <= ? and created < ?"), from, upTo, cutoff.UTC())
		if err != nil {
			return deleted, finished, err
		}
		n, err := res.RowsAffected()
		if err != nil {
			return deleted, finished, err
		}
		deleted += n
		finished = upTo
	}
	if deleted == 0 {
		// Another node deleted these rows first.
		return 0, 0, nil
	}
	return deleted, last.Int64, nil
}
