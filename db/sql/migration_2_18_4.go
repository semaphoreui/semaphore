package sql

import (
	"fmt"

	"github.com/go-gorp/gorp/v3"
)

type migration_2_18_4 struct {
	db *SqlDb
}

func (m migration_2_18_4) PreApply(tx *gorp.Transaction) error {
	switch m.db.Sql().Dialect.(type) {
	case gorp.MySQLDialect:
		fkName, err := tx.SelectStr(
			`SELECT CONSTRAINT_NAME FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE
             WHERE TABLE_SCHEMA = DATABASE()
               AND TABLE_NAME = 'project__template'
               AND COLUMN_NAME = 'environment_id'
               AND REFERENCED_TABLE_NAME IS NOT NULL
             LIMIT 1`)
		if err == nil && fkName != "" {
			_, _ = tx.Exec(fmt.Sprintf("alter table `project__template` drop foreign key `%s`", fkName))
		}
	case gorp.PostgresDialect:
		_, _ = tx.Exec(
			m.db.PrepareQuery("alter table `project__template` drop constraint if exists `project__template_environment_id_fkey`"))
	case gorp.SqliteDialect:
		_, _ = tx.Exec(
			m.db.PrepareQuery("drop index if exists `project__template__environment_id`"))
	}
	return nil
}
