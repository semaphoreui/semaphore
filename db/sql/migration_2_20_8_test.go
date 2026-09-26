package sql

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigration_2_20_8_AuditSchemaFreshSQLite(t *testing.T) {
	store := InitConfigCreateTestStore()
	t.Cleanup(store.Close)
	assertAuditMigrationSchema(t, store)
}

func TestMigration_2_20_8_AuditSchemaUpgradeFrom2_20_7(t *testing.T) {
	version := "2.20.7"
	store := InitConfigCreateTestStoreAt(&version)
	t.Cleanup(store.Close)
	require.NoError(t, db.Migrate(store, nil))
	assertAuditMigrationSchema(t, store)
}

func TestMigration_2_20_8_AuditDialectShapes(t *testing.T) {
	postgres := getVersionSQL(util.DbDriverPostgres, "v2.20.8.sql", false)
	mysql := getVersionSQL(util.DbDriverMySQL, "v2.20.8.sql", false)
	require.NotEmpty(t, postgres)
	require.NotEmpty(t, mysql)
	postgresSQL := strings.Join(postgres, "\n")
	mysqlSQL := strings.Join(mysql, "\n")
	assert.Contains(t, postgresSQL, "seq               bigint primary key")
	assert.Contains(t, postgresSQL, "timestamp with time zone")
	assert.NotContains(t, postgresSQL, "bigserial")
	assert.Contains(t, mysqlSQL, "seq               bigint primary key")
	assert.Contains(t, mysqlSQL, "datetime(6)")
	assert.NotContains(t, mysqlSQL, "auto_increment")
}

func assertAuditMigrationSchema(t *testing.T, store *SqlDb) {
	t.Helper()
	now := time.Now().UTC()
	eventID := uuid.NewString()
	insertEvent := func(seq int64, id string) error {
		_, err := store.exec(`
			insert into audit_event (
				seq, event_id, occurred_at, schema_version, event_code, category, type, action,
				outcome, actor_type, actor_id, actor_name, source_ip, source_user_agent,
				target_type, target_id, target_name, project_id, request_id, instance_id,
				node_id, metadata
			) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			seq, id, now, "1", "resource.inventory", "resource", "creation", "create",
			"success", "user", "42", "Audit User", "203.0.113.10", "migration test",
			"inventory", "9", "Inventory", "7", uuid.NewString(), "test-instance",
			"test-node", `{}`,
		)
		return err
	}

	var sequenceCount int
	var lastSeq int64
	err := store.Sql().QueryRow(
		"select count(*), max(last_seq) from audit_event_sequence where id=1",
	).Scan(&sequenceCount, &lastSeq)
	require.NoError(t, err)
	assert.Equal(t, 1, sequenceCount)
	assert.Zero(t, lastSeq)

	require.NoError(t, insertEvent(42, eventID))
	var seq int64
	err = store.Sql().QueryRow(
		store.PrepareQuery("select seq from audit_event where event_id=?"), eventID,
	).Scan(&seq)
	require.NoError(t, err)
	assert.Equal(t, int64(42), seq)
	assert.Error(t, insertEvent(43, eventID))

	destinationID := "primary-siem"
	_, err = store.exec(`
		insert into audit_export_state (
			destination_id, cursor_seq, owner_id, lease_until, lease_generation,
			last_attempt_at, last_success_at, last_error
		) values (?, ?, ?, ?, ?, ?, ?, ?)`,
		destinationID, 12, uuid.NewString(), now, 3, now, now, "test error",
	)
	require.NoError(t, err)

	var stateCount int
	err = store.Sql().QueryRow(store.PrepareQuery(`
		select count(*) from audit_export_state
		where destination_id=? and cursor_seq=12 and lease_generation=3
			and last_attempt_at is not null and last_success_at is not null
			and last_error='test error'`), destinationID).Scan(&stateCount)
	require.NoError(t, err)
	assert.Equal(t, 1, stateCount)

	_, err = store.exec("insert into audit_export_state (destination_id) values (?)", destinationID)
	assert.Error(t, err)
}
