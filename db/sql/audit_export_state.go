package sql

import (
	"fmt"
	"math"
	"time"
	"unicode/utf8"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
)

const (
	auditExportErrorMaxLength = 1024
	auditDestinationIDMaxSize = 255
)

type auditLeaseSnapshot struct {
	OwnerID         *string `db:"owner_id"`
	LeaseGeneration int64   `db:"lease_generation"`
	Live            int64   `db:"lease_live"`
}

func (d *SqlDb) GetAuditExportState(destinationID string) (db.AuditExportState, error) {
	if err := validateAuditDestinationID(destinationID); err != nil {
		return db.AuditExportState{}, err
	}
	var state db.AuditExportState
	query, args, err := sq.Select(
		"destination_id", "cursor_seq", "owner_id", "lease_until", "lease_generation",
		"last_attempt_at", "last_success_at", "last_error",
	).
		From("audit_export_state").
		Where(sq.Eq{"destination_id": destinationID}).
		ToSql()
	if err != nil {
		return state, err
	}
	err = d.selectOne(&state, query, args...)
	if err == db.ErrNotFound {
		err = db.ErrAuditExportStateNotFound
	}
	return state, err
}

func (d *SqlDb) GetOrCreateAuditExportState(destinationID string) (db.AuditExportState, error) {
	if err := validateAuditDestinationID(destinationID); err != nil {
		return db.AuditExportState{}, err
	}
	query, args, err := sq.Insert("audit_export_state").
		Columns("destination_id").
		Values(destinationID).
		ToSql()
	if err != nil {
		return db.AuditExportState{}, err
	}
	_, insertErr := d.exec(query, args...)
	if insertErr == nil {
		return d.GetAuditExportState(destinationID)
	}
	state, getErr := d.GetAuditExportState(destinationID)
	if getErr == nil {
		return state, nil
	}
	return db.AuditExportState{}, insertErr
}

func (d *SqlDb) TryAcquireAuditExportLease(destinationID, ownerID string, ttl time.Duration) (int64, bool, error) {
	if err := validateAuditDestinationID(destinationID); err != nil {
		return 0, false, err
	}
	if err := validateOwner(ownerID); err != nil {
		return 0, false, err
	}
	until, seconds, err := d.auditLeaseUntil(ttl)
	if err != nil {
		return 0, false, err
	}
	now, err := d.auditCurrentTime()
	if err != nil {
		return 0, false, err
	}
	result, err := d.exec("update audit_export_state set owner_id=?, lease_until="+until+", lease_generation=lease_generation+1 where destination_id=? and (lease_until is null or lease_until<="+now+")", ownerID, seconds, destinationID)
	if err != nil {
		return 0, false, err
	}
	ok, err := mutationSucceeded(result)
	if err != nil || ok {
		if err != nil {
			return 0, false, err
		}
		state, err := d.GetAuditExportState(destinationID)
		if err != nil {
			return 0, false, err
		}
		if state.OwnerID == nil || *state.OwnerID != ownerID {
			return 0, false, db.ErrAuditExportLeaseLost
		}
		return state.LeaseGeneration, true, nil
	}
	if _, err := d.GetAuditExportState(destinationID); err != nil {
		return 0, false, err
	}
	return 0, false, db.ErrAuditExportLeaseContended
}

func (d *SqlDb) RenewAuditExportLease(destinationID, ownerID string, generation int64, ttl time.Duration) (bool, error) {
	if err := validateAuditDestinationID(destinationID); err != nil {
		return false, err
	}
	if err := validateOwner(ownerID); err != nil {
		return false, err
	}
	until, seconds, err := d.auditLeaseUntil(ttl)
	if err != nil {
		return false, err
	}
	now, err := d.auditCurrentTime()
	if err != nil {
		return false, err
	}
	result, err := d.exec("update audit_export_state set lease_until="+until+" where destination_id=? and owner_id=? and lease_generation=? and lease_until>"+now, seconds, destinationID, ownerID, generation)
	if err != nil {
		return false, err
	}
	ok, err := mutationSucceeded(result)
	if err != nil || ok {
		return ok, err
	}
	if err := d.classifyAuditExportLeaseMutation(destinationID, ownerID, generation); err != nil {
		return false, err
	}
	return true, nil
}

func (d *SqlDb) AdvanceAuditExportCursor(destinationID, ownerID string, generation, expectedSeq, newSeq int64) (bool, error) {
	if err := validateAuditDestinationID(destinationID); err != nil {
		return false, err
	}
	if err := validateOwner(ownerID); err != nil {
		return false, err
	}
	if expectedSeq < 0 || newSeq < 0 || newSeq <= expectedSeq {
		return false, fmt.Errorf("audit export cursor must advance")
	}
	now, err := d.auditCurrentTime()
	if err != nil {
		return false, err
	}
	result, err := d.exec("update audit_export_state set cursor_seq=? where destination_id=? and owner_id=? and lease_generation=? and cursor_seq=? and lease_until>"+now, newSeq, destinationID, ownerID, generation, expectedSeq)
	if err != nil {
		return false, err
	}
	ok, err := mutationSucceeded(result)
	if err != nil || ok {
		return ok, err
	}
	if err := d.classifyAuditExportLeaseMutation(destinationID, ownerID, generation); err != nil {
		return false, err
	}
	return false, db.ErrAuditExportCursorConflict
}

func (d *SqlDb) RecordAuditExportAttempt(destinationID, ownerID string, generation int64) (bool, error) {
	if err := validateAuditDestinationID(destinationID); err != nil {
		return false, err
	}
	now, err := d.auditCurrentTime()
	if err != nil {
		return false, err
	}
	if err := validateOwner(ownerID); err != nil {
		return false, err
	}
	result, err := d.exec(
		"update audit_export_state set last_attempt_at="+now+
			" where destination_id=? and owner_id=? and lease_generation=? and lease_until>"+now,
		destinationID, ownerID, generation,
	)
	if err != nil {
		return false, err
	}
	ok, err := mutationSucceeded(result)
	if err != nil || ok {
		return ok, err
	}
	if err := d.classifyAuditExportLeaseMutation(destinationID, ownerID, generation); err != nil {
		return false, err
	}
	return true, nil
}

func (d *SqlDb) RecordAuditExportSuccess(destinationID, ownerID string, generation int64) (bool, error) {
	if err := validateAuditDestinationID(destinationID); err != nil {
		return false, err
	}
	now, err := d.auditCurrentTime()
	if err != nil {
		return false, err
	}
	if err := validateOwner(ownerID); err != nil {
		return false, err
	}
	result, err := d.exec(
		"update audit_export_state set last_success_at="+now+", last_error=null"+
			" where destination_id=? and owner_id=? and lease_generation=? and lease_until>"+now,
		destinationID, ownerID, generation,
	)
	if err != nil {
		return false, err
	}
	ok, err := mutationSucceeded(result)
	if err != nil || ok {
		return ok, err
	}
	if err := d.classifyAuditExportLeaseMutation(destinationID, ownerID, generation); err != nil {
		return false, err
	}
	return true, nil
}

func (d *SqlDb) RecordAuditExportError(destinationID, ownerID string, generation int64, message string) (bool, error) {
	if err := validateAuditDestinationID(destinationID); err != nil {
		return false, err
	}
	if err := validateOwner(ownerID); err != nil {
		return false, err
	}
	now, err := d.auditCurrentTime()
	if err != nil {
		return false, err
	}
	result, err := d.exec(
		"update audit_export_state set last_error=?"+
			" where destination_id=? and owner_id=? and lease_generation=? and lease_until>"+now,
		truncateAuditExportError(message), destinationID, ownerID, generation,
	)
	if err != nil {
		return false, err
	}
	ok, err := mutationSucceeded(result)
	if err != nil || ok {
		return ok, err
	}
	if err := d.classifyAuditExportLeaseMutation(destinationID, ownerID, generation); err != nil {
		return false, err
	}
	return true, nil
}

func (d *SqlDb) classifyAuditExportLeaseMutation(destinationID, ownerID string, generation int64) error {
	now, err := d.auditCurrentTime()
	if err != nil {
		return err
	}
	var snapshot auditLeaseSnapshot
	err = d.selectOne(&snapshot, "select owner_id, lease_generation, case when lease_until>"+now+" then 1 else 0 end as lease_live from audit_export_state where destination_id=?", destinationID)
	if err == db.ErrNotFound {
		return db.ErrAuditExportStateNotFound
	}
	if err != nil {
		return err
	}
	if snapshot.OwnerID == nil || *snapshot.OwnerID != ownerID || snapshot.LeaseGeneration != generation {
		return db.ErrAuditExportStateStale
	}
	if snapshot.Live == 0 {
		return db.ErrAuditExportLeaseLost
	}
	return nil
}

func (d *SqlDb) auditLeaseUntil(ttl time.Duration) (string, int64, error) {
	if ttl <= 0 {
		return "", 0, fmt.Errorf("audit export lease TTL must be positive")
	}
	seconds := int64(math.Ceil(ttl.Seconds()))
	switch d.GetDialect() {
	case util.DbDriverMySQL:
		return "date_add(current_timestamp(6), interval ? second)", seconds, nil
	case util.DbDriverPostgres:
		return "current_timestamp + (? * interval '1 second')", seconds, nil
	case util.DbDriverSQLite:
		return "datetime('now', '+' || ? || ' seconds')", seconds, nil
	}
	return "", 0, fmt.Errorf("unsupported database dialect %q", d.GetDialect())
}

func (d *SqlDb) auditCurrentTime() (string, error) {
	switch d.GetDialect() {
	case util.DbDriverMySQL:
		return "current_timestamp(6)", nil
	case util.DbDriverPostgres, util.DbDriverSQLite:
		return "current_timestamp", nil
	}
	return "", fmt.Errorf("unsupported database dialect %q", d.GetDialect())
}

func mutationSucceeded(result interface{ RowsAffected() (int64, error) }) (bool, error) {
	rows, err := result.RowsAffected()
	return rows == 1, err
}

func validateOwner(value string) error {
	id, err := uuid.Parse(value)
	if err != nil || id.Version() != 4 || id.String() != value {
		return fmt.Errorf("audit export owner ID must be a UUIDv4")
	}
	return nil
}

func validateAuditDestinationID(value string) error {
	if value == "" || len(value) > auditDestinationIDMaxSize {
		return fmt.Errorf("audit export destination ID is invalid")
	}
	return nil
}

func truncateAuditExportError(value string) string {
	if utf8.RuneCountInString(value) <= auditExportErrorMaxLength {
		return value
	}
	return string([]rune(value)[:auditExportErrorMaxLength])
}
