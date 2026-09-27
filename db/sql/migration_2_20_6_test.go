package sql

import (
	"strings"
	"testing"

	"github.com/go-gorp/gorp/v3"
	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMigration_2_20_6_DropsEventBackup recreates the state of an installation
// upgraded through v2.8.20: event_backup_5784568 still references user(id)
// without an ON DELETE action, so any user with pre-2.8.20 events cannot be
// deleted. The migration must drop the table and unblock the deletion.
func TestMigration_2_20_6_DropsEventBackup(t *testing.T) {
	preDrop := "2.20.5"
	store := InitConfigCreateTestStoreAt(&preDrop)

	user, err := store.CreateUserWithoutPassword(db.User{
		Username: "jdoe", Name: "John Doe", Email: "jdoe@example.com",
	})
	require.NoError(t, err)

	// The test store starts from the 2.15.1 dump where the backup table has
	// ON DELETE SET NULL, so rebuild it the way v2.7.4 + v2.8.20 left it.
	_, err = store.exec("drop table event_backup_5784568")
	require.NoError(t, err)
	_, err = store.exec("create table event_backup_5784568 (" +
		"id integer primary key autoincrement, project_id int, description text, " +
		"created datetime not null, user_id int null references user(id))")
	require.NoError(t, err)
	_, err = store.exec(
		"insert into event_backup_5784568 (description, created, user_id) values ('old', '2021-01-01', ?)",
		user.ID)
	require.NoError(t, err)

	assert.ErrorIs(t, store.DeleteUser(user.ID), db.ErrInvalidOperation)

	require.NoError(t, db.Migrate(store, nil))

	count, err := store.Sql().SelectInt(
		"select count(*) from sqlite_master where type = 'table' and name = 'event_backup_5784568'")
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	require.NoError(t, store.DeleteUser(user.ID))
	_, err = store.GetUser(user.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)
}

// TestMigration_2_20_6_IsIdempotent covers installations created from the
// 2.15.1 SQLite dump or later where the backup table may already be absent.
func TestMigration_2_20_6_IsIdempotent(t *testing.T) {
	preDrop := "2.20.5"
	store := InitConfigCreateTestStoreAt(&preDrop)

	_, err := store.exec("drop table event_backup_5784568")
	require.NoError(t, err)

	assert.NoError(t, db.Migrate(store, nil))
}

func TestPrepareMigration_MysqlKeepsDropTableIfExists(t *testing.T) {
	d := &SqlDb{connection: SqlDbConnection{sql: &gorp.DbMap{Dialect: gorp.MySQLDialect{}}}}

	tests := []struct {
		name     string
		query    string
		expected string
	}{
		{"drop table", "drop table if exists `event_backup_5784568`", "drop table if exists `event_backup_5784568`"},
		{"drop index", "drop index if exists `task__workflow_run_id`", "drop index `task__workflow_run_id`"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, d.prepareMigration(tt.query))
		})
	}
}

// TestMigration_2_20_6_RollbackHasNoCommentOnlyStatements: a statement that
// consists of comments only is sent to the server as is, and MySQL rejects
// it with "Query was empty". Blank statements are skipped by the runner.
func TestMigration_2_20_6_RollbackHasNoCommentOnlyStatements(t *testing.T) {
	for _, dialect := range []string{"sqlite", "mysql", "postgres"} {
		t.Run(dialect, func(t *testing.T) {
			var statements int
			for _, q := range getVersionSQL(dialect, "v2.20.6.err.sql", false) {
				if q == "" {
					continue
				}
				statements++
				var code []string
				for _, line := range strings.Split(q, "\n") {
					line = strings.TrimSpace(line)
					if line != "" && !strings.HasPrefix(line, "--") {
						code = append(code, line)
					}
				}
				assert.NotEmpty(t, code, "comment-only statement: %q", q)
			}
			assert.Greater(t, statements, 0)
		})
	}
}

// TestMigration_2_20_6_Rollback undoes the alert tables and columns the
// migration adds, and the migration applies again afterwards.
func TestMigration_2_20_6_Rollback(t *testing.T) {
	store := InitConfigCreateTestStoreAt(nil)

	tables := []string{"project__alert", "project__template_alert", "project__schedule_alert", "task__alert_send"}
	columns := map[string]string{
		"project__template": "alert_mode",
		"project__schedule": "alert_mode",
		"task":              "alert_snapshot",
	}

	tableExists := func(name string) bool {
		n, err := store.Sql().SelectInt(
			"select count(*) from sqlite_master where type = 'table' and name = ?", name)
		require.NoError(t, err)
		return n == 1
	}
	columnExists := func(table, column string) bool {
		n, err := store.Sql().SelectInt(
			"select count(*) from pragma_table_info(?) where name = ?", table, column)
		require.NoError(t, err)
		return n == 1
	}

	for _, name := range tables {
		require.True(t, tableExists(name), name)
	}
	for table, column := range columns {
		require.True(t, columnExists(table, column), table+"."+column)
	}

	require.NoError(t, db.Rollback(store, "2.20.5"))

	for _, name := range tables {
		assert.False(t, tableExists(name), name)
	}
	for table, column := range columns {
		assert.False(t, columnExists(table, column), table+"."+column)
	}
	applied, err := store.IsMigrationApplied(db.Migration{Version: "2.20.6"})
	require.NoError(t, err)
	assert.False(t, applied)

	require.NoError(t, db.Migrate(store, nil))

	for _, name := range tables {
		assert.True(t, tableExists(name), name)
	}
	for table, column := range columns {
		assert.True(t, columnExists(table, column), table+"."+column)
	}
}
