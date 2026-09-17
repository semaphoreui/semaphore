package sql

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-gorp/gorp/v3"
	"github.com/google/uuid"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuditExportStateFreshSQLiteMigration(t *testing.T) {
	store := InitConfigCreateTestStore()
	_, err := store.GetOrCreateAuditExportState("siem")
	assert.NoError(t, err)
}

func TestAuditExportStateMigrationFrom2_20_5(t *testing.T) {
	version := "2.20.5"
	store := InitConfigCreateTestStoreAt(&version)
	require.NoError(t, db.Migrate(store, nil))
	_, err := store.GetOrCreateAuditExportState("siem")
	assert.NoError(t, err)
}

func TestAuditStateGetOrCreateAndFencing(t *testing.T) {
	store := InitConfigCreateTestStore()
	state, err := store.GetOrCreateAuditExportState("siem")
	require.NoError(t, err)
	_, err = store.GetOrCreateAuditExportState("siem")
	assert.NoError(t, err)
	owner := uuid.NewString()
	gen, ok, err := store.TryAcquireAuditExportLease("siem", owner, time.Minute)
	require.NoError(t, err)
	require.True(t, ok)
	ok, err = store.AdvanceAuditExportCursor("siem", owner, gen, 0, 1)
	require.NoError(t, err)
	assert.True(t, ok)
	_, err = store.AdvanceAuditExportCursor("siem", owner, gen, 1, 1)
	assert.Error(t, err)
	_, err = store.AdvanceAuditExportCursor("siem", owner, gen, 1, 0)
	assert.Error(t, err)
	ok, err = store.AdvanceAuditExportCursor("siem", uuid.NewString(), gen, 1, 2)
	assert.False(t, ok)
	assert.ErrorIs(t, err, db.ErrAuditExportStateStale)
	ok, err = store.AdvanceAuditExportCursor("siem", owner, gen+1, 1, 2)
	assert.False(t, ok)
	assert.ErrorIs(t, err, db.ErrAuditExportStateStale)
	ok, err = store.AdvanceAuditExportCursor("siem", owner, gen, 0, 2)
	assert.False(t, ok)
	assert.ErrorIs(t, err, db.ErrAuditExportCursorConflict)
	ok, err = store.RecordAuditExportError("siem", uuid.NewString(), gen, "error")
	assert.False(t, ok)
	assert.ErrorIs(t, err, db.ErrAuditExportStateStale)
	_, err = store.exec("update audit_export_state set lease_until=datetime('now','-1 second') where destination_id=?", "siem")
	require.NoError(t, err)
	ok, err = store.RenewAuditExportLease("siem", owner, gen, time.Minute)
	assert.False(t, ok)
	assert.ErrorIs(t, err, db.ErrAuditExportLeaseLost)
	assert.Equal(t, int64(0), state.CursorSeq)
}

func TestAuditLeaseZeroRowClassification(t *testing.T) {
	store := InitConfigCreateTestStore()
	_, err := store.GetOrCreateAuditExportState("siem")
	require.NoError(t, err)
	owner := uuid.NewString()
	generation, acquired, err := store.TryAcquireAuditExportLease("siem", owner, time.Minute)
	require.NoError(t, err)
	require.True(t, acquired)

	assert.NoError(t, store.classifyAuditExportLeaseMutation("siem", owner, generation))
	assert.ErrorIs(t, store.classifyAuditExportLeaseMutation("siem", uuid.NewString(), generation), db.ErrAuditExportStateStale)
	assert.ErrorIs(t, store.classifyAuditExportLeaseMutation("siem", owner, generation+1), db.ErrAuditExportStateStale)
	_, err = store.exec("update audit_export_state set lease_until=datetime('now','-1 second') where destination_id=?", "siem")
	require.NoError(t, err)
	assert.ErrorIs(t, store.classifyAuditExportLeaseMutation("siem", owner, generation), db.ErrAuditExportLeaseLost)
	assert.ErrorIs(t, store.classifyAuditExportLeaseMutation("missing", owner, generation), db.ErrAuditExportStateNotFound)
}

func TestAuditExportStateDestinationIDValidation(t *testing.T) {
	store := InitConfigCreateTestStore()
	owner := uuid.NewString()
	operations := []struct {
		name string
		call func(string) error
	}{
		{name: "get", call: func(destinationID string) error { _, err := store.GetAuditExportState(destinationID); return err }},
		{name: "get or create", call: func(destinationID string) error {
			_, err := store.GetOrCreateAuditExportState(destinationID)
			return err
		}},
		{name: "acquire", call: func(destinationID string) error {
			_, _, err := store.TryAcquireAuditExportLease(destinationID, owner, time.Minute)
			return err
		}},
		{name: "renew", call: func(destinationID string) error {
			_, err := store.RenewAuditExportLease(destinationID, owner, 1, time.Minute)
			return err
		}},
		{name: "advance cursor", call: func(destinationID string) error {
			_, err := store.AdvanceAuditExportCursor(destinationID, owner, 1, 0, 1)
			return err
		}},
		{name: "record attempt", call: func(destinationID string) error {
			_, err := store.RecordAuditExportAttempt(destinationID, owner, 1)
			return err
		}},
		{name: "record success", call: func(destinationID string) error {
			_, err := store.RecordAuditExportSuccess(destinationID, owner, 1)
			return err
		}},
		{name: "record error", call: func(destinationID string) error {
			_, err := store.RecordAuditExportError(destinationID, owner, 1, "error")
			return err
		}},
	}
	for _, destinationID := range []string{"", strings.Repeat("a", auditDestinationIDMaxSize+1)} {
		for _, operation := range operations {
			t.Run(operation.name+"/"+strconv.Itoa(len(destinationID)), func(t *testing.T) {
				assert.Error(t, operation.call(destinationID))
			})
		}
	}
}

func TestAuditStateLeaseContentionRenewalAndTakeover(t *testing.T) {
	store := InitConfigCreateTestStore()
	_, err := store.GetOrCreateAuditExportState("siem")
	require.NoError(t, err)
	owner := uuid.NewString()
	generation, acquired, err := store.TryAcquireAuditExportLease("siem", owner, time.Minute)
	require.NoError(t, err)
	require.True(t, acquired)

	_, acquired, err = store.TryAcquireAuditExportLease("siem", uuid.NewString(), time.Minute)
	assert.False(t, acquired)
	assert.ErrorIs(t, err, db.ErrAuditExportLeaseContended)

	renewed, err := store.RenewAuditExportLease("siem", owner, generation, time.Minute)
	require.NoError(t, err)
	assert.True(t, renewed)

	_, err = store.exec("update audit_export_state set lease_until=datetime('now','-1 second') where destination_id=?", "siem")
	require.NoError(t, err)
	newOwner := uuid.NewString()
	newGeneration, acquired, err := store.TryAcquireAuditExportLease("siem", newOwner, time.Minute)
	require.NoError(t, err)
	require.True(t, acquired)
	assert.Greater(t, newGeneration, generation)

	renewed, err = store.RenewAuditExportLease("siem", owner, generation, time.Minute)
	assert.False(t, renewed)
	assert.ErrorIs(t, err, db.ErrAuditExportStateStale)
}

func TestAuditStateRecordErrorIsIdempotentAndTruncated(t *testing.T) {
	store := InitConfigCreateTestStore()
	_, err := store.GetOrCreateAuditExportState("siem")
	require.NoError(t, err)
	owner := uuid.NewString()
	generation, acquired, err := store.TryAcquireAuditExportLease("siem", owner, time.Minute)
	require.NoError(t, err)
	require.True(t, acquired)

	message := strings.Repeat("x", auditExportErrorMaxLength+1)
	recorded, err := store.RecordAuditExportError("siem", owner, generation, message)
	require.NoError(t, err)
	assert.True(t, recorded)
	recorded, err = store.RecordAuditExportError("siem", owner, generation, message)
	assert.NoError(t, err)
	assert.True(t, recorded)

	state, err := store.GetAuditExportState("siem")
	require.NoError(t, err)
	require.NotNil(t, state.LastError)
	assert.Len(t, *state.LastError, auditExportErrorMaxLength)
}

func TestAuditStateConcurrentGetOrCreate(t *testing.T) {
	store := InitConfigCreateTestStore()
	errs := make(chan error, 2)
	for range 2 {
		go func() { _, err := store.GetOrCreateAuditExportState("siem"); errs <- err }()
	}
	assert.NoError(t, <-errs)
	assert.NoError(t, <-errs)
}

func TestAuditStateMissingAndInvalidLeaseInput(t *testing.T) {
	store := InitConfigCreateTestStore()
	_, ok, err := store.TryAcquireAuditExportLease("missing", uuid.NewString(), time.Minute)
	assert.False(t, ok)
	assert.ErrorIs(t, err, db.ErrAuditExportStateNotFound)
	_, _, err = store.TryAcquireAuditExportLease("missing", "bad", time.Minute)
	assert.Error(t, err)
	_, _, err = store.TryAcquireAuditExportLease("missing", uuid.NewString(), 0)
	assert.Error(t, err)
	_, _, err = store.TryAcquireAuditExportLease("missing", uuid.NewString(), -time.Second)
	assert.Error(t, err)
	_, _, err = store.TryAcquireAuditExportLease("missing", strings.ToUpper(uuid.NewString()), time.Minute)
	assert.Error(t, err)
}

func TestAuditLeaseDialectSQL(t *testing.T) {
	for _, tt := range []struct{ name, dialect, want string }{{"mysql", util.DbDriverMySQL, "date_add(current_timestamp(6), interval ? second)"}, {"postgres", util.DbDriverPostgres, "current_timestamp + (? * interval '1 second')"}} {
		t.Run(tt.name, func(t *testing.T) {
			d := &SqlDb{connection: SqlDbConnection{dialect: tt.dialect, sql: &gorp.DbMap{Dialect: dialectForAuditTest(tt.dialect)}}}
			actual, _, err := d.auditLeaseUntil(time.Second)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, actual)
			now, err := d.auditCurrentTime()
			assert.NoError(t, err)
			if tt.dialect == util.DbDriverMySQL {
				assert.Equal(t, "current_timestamp(6)", now)
			}
			prepared := d.PrepareQuery("update x set value=" + actual + " where id=?")
			if tt.dialect == util.DbDriverPostgres {
				assert.Contains(t, prepared, "$")
			} else {
				assert.Contains(t, prepared, "?")
			}
		})
	}
}

func TestAuditMigrationDialectShapes(t *testing.T) {
	postgres := getVersionSQL(util.DbDriverPostgres, "v2.20.6.sql", false)
	mysql := getVersionSQL(util.DbDriverMySQL, "v2.20.6.sql", false)
	assert.Contains(t, postgres[0], "bigserial primary key")
	assert.Contains(t, postgres[0], "timestamp with time zone")
	assert.Contains(t, mysql[0], "bigint primary key auto_increment")
	assert.Contains(t, mysql[0], "datetime(6)")
}

func dialectForAuditTest(dialect string) gorp.Dialect {
	if dialect == util.DbDriverPostgres {
		return gorp.PostgresDialect{}
	}
	return gorp.MySQLDialect{}
}

var _ db.Store = (*SqlDb)(nil)
