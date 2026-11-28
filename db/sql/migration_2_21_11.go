package sql

import (
	"github.com/go-gorp/gorp/v3"
	"github.com/semaphoreui/semaphore/util"
)

// migration_2_21_11 moves webhook aliases and authorization from integrations
// to endpoints and integration groups. The data moves in v2.21.11.sql; this
// hook only clears what the SQL cannot express for one dialect.
type migration_2_21_11 struct {
	db *SqlDb
}

// PreApply drops the unnamed MySQL foreign key on project__integration.auth_secret_id,
// which v2.9.61 created without a name, so the SQL can drop the column.
// Postgres drops a column's constraints with it, and SQLite rebuilds the table.
func (m migration_2_21_11) PreApply(tx *gorp.Transaction) error {
	if m.db.GetDialect() != util.DbDriverMySQL {
		return nil
	}

	return dropMysqlForeignKey(tx, "project__integration", "auth_secret_id")
}
