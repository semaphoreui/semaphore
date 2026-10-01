package sql

import (
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

// TestMigration_2_20_6_RollbackIsNoop: the rollback file is intentionally
// empty. A comment-only file would be sent as a statement, and MySQL rejects
// a comment-only query with "Query was empty".
func TestMigration_2_20_6_RollbackIsNoop(t *testing.T) {
	for _, dialect := range []string{"sqlite", "mysql", "postgres"} {
		t.Run(dialect, func(t *testing.T) {
			for _, q := range getVersionSQL(dialect, "v2.20.6.err.sql", false) {
				assert.Empty(t, q)
			}
		})
	}
}
