package sql

import (
	"strings"
	"testing"

	"github.com/go-gorp/gorp/v3"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
)

func nonEmptyQueries(queries []string) (res []string) {
	for _, q := range queries {
		if q != "" {
			res = append(res, q)
		}
	}
	return
}

// TestMigration_2_20_10_RendersPerDialect: SQLite already has the user_id
// foreign keys on task/session (v2.19.14), so both files must render to
// nothing there — a comment-only statement would be sent to the database.
// MySQL and Postgres get the orphan cleanup followed by the two constraints.
func TestMigration_2_20_10_RendersPerDialect(t *testing.T) {
	tests := []struct {
		dialect     string
		forward     int
		constraints int
		rollback    int
	}{
		{util.DbDriverSQLite, 0, 0, 0},
		{"mysql", 5, 2, 2},
		{"postgres", 4, 2, 2},
	}
	for _, tt := range tests {
		t.Run(tt.dialect, func(t *testing.T) {
			forward := nonEmptyQueries(getVersionSQL(tt.dialect, "v2.20.10.sql", false))
			assert.Len(t, forward, tt.forward)
			joined := strings.Join(forward, "\n")
			assert.Equal(t, tt.constraints, strings.Count(joined, "add constraint"))
			if tt.constraints > 0 {
				assert.Contains(t, joined, "`session__user_fk` foreign key (`user_id`) references `user`(`id`) on delete cascade")
				assert.Contains(t, joined, "`task__user_fk` foreign key (`user_id`) references `user`(`id`) on delete set null")
			}

			rollback := nonEmptyQueries(getVersionSQL(tt.dialect, "v2.20.10.err.sql", false))
			assert.Len(t, rollback, tt.rollback)
			assert.Equal(t, tt.rollback, strings.Count(strings.Join(rollback, "\n"), "drop foreign key"))
		})
	}
}

// Postgres has no `drop foreign key`; prepareMigration must turn the rollback
// into `drop constraint` with double-quoted identifiers.
func TestMigration_2_20_10_PostgresRollback(t *testing.T) {
	d := &SqlDb{connection: SqlDbConnection{sql: &gorp.DbMap{Dialect: gorp.PostgresDialect{}}}}

	queries := nonEmptyQueries(getVersionSQL("postgres", "v2.20.10.err.sql", false))
	assert.Len(t, queries, 2)
	assert.Equal(t, `alter table "task" drop constraint "task__user_fk"`, d.prepareMigration(queries[0]))
	assert.Equal(t, `alter table "session" drop constraint "session__user_fk"`, d.prepareMigration(queries[1]))
}
