package sql

import (
	"github.com/go-gorp/gorp/v3"
)

type migration_2_18_4 struct {
	db *SqlDb
}

func (m migration_2_18_4) PreApply(tx *gorp.Transaction) error {
	switch m.db.Sql().Dialect.(type) {
	case gorp.MySQLDialect:
		return dropMysqlForeignKey(tx, "project__template", "environment_id")
	case gorp.PostgresDialect:
		_, _ = tx.Exec(
			m.db.PrepareQuery("alter table `project__template` drop constraint if exists `project__template_environment_id_fkey`"))
	case gorp.SqliteDialect:
		_, _ = tx.Exec(
			m.db.PrepareQuery("drop index if exists `project__template__environment_id`"))
	}
	return nil
}
