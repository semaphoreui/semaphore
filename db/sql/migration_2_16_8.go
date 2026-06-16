package sql

import "github.com/go-gorp/gorp/v3"

type migration_2_16_8 struct {
	db *SqlDb
}

// PreApply drops the foreign keys on task__stage before v2.16.8.sql removes the
// start_output_id / end_output_id columns. On MySQL a column cannot be dropped
// while it is part of a foreign key, and the constraint name is auto-generated,
// so it is resolved dynamically. Postgres drops the dependent FK together with
// the column, and SQLite handles it via index drops in the .sql file.
func (m migration_2_16_8) PreApply(tx *gorp.Transaction) error {
	switch m.db.Sql().Dialect.(type) {
	case gorp.MySQLDialect:
		if err := dropMysqlForeignKey(tx, "task__stage", "start_output_id"); err != nil {
			return err
		}
		return dropMysqlForeignKey(tx, "task__stage", "end_output_id")
	}
	return nil
}

// PreRollback drops the foreign key on task__output before v2.16.8.err.sql
// removes the stage_id column.
func (m migration_2_16_8) PreRollback(tx *gorp.Transaction) error {
	switch m.db.Sql().Dialect.(type) {
	case gorp.MySQLDialect:
		return dropMysqlForeignKey(tx, "task__output", "stage_id")
	}
	return nil
}
