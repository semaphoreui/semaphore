package sql

import (
	"context"
	"slices"

	"github.com/Masterminds/squirrel"
	"github.com/semaphoreui/semaphore/db"
)

// ListAuditEvents pages the whole audit log newest first, reading limit+1 rows to know if more exist.
func (d *SqlDb) ListAuditEvents(ctx context.Context, before, after int64, limit int) (db.AuditEventPage, error) {
	var page db.AuditEventPage
	q := squirrel.Select("*").From("audit_event").Limit(uint64(limit + 1))
	switch {
	case after > 0:
		q = q.Where(squirrel.Gt{"seq": after}).OrderBy("seq asc")
	case before > 0:
		q = q.Where(squirrel.Lt{"seq": before}).OrderBy("seq desc")
	default:
		q = q.OrderBy("seq desc")
	}
	query, args, err := q.ToSql()
	if err != nil {
		return page, err
	}
	var rows []db.AuditEvent
	if _, err = d.Sql().WithContext(ctx).Select(&rows, d.PrepareQuery(query), args...); err != nil {
		return page, err
	}

	more := len(rows) > limit
	rows = rows[:min(len(rows), limit)]
	if after > 0 {
		slices.Reverse(rows)
		page.Older = after + 1
		if more {
			page.Newer = rows[0].Seq
		}
		if len(rows) > 0 {
			page.Older = rows[len(rows)-1].Seq
		}
	} else {
		if more {
			page.Older = rows[len(rows)-1].Seq
		}
		if before > 0 {
			page.Newer = before - 1
		}
	}
	page.Events = rows
	return page, nil
}
