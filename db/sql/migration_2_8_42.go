package sql

import "github.com/go-gorp/gorp/v3"

type migration_2_8_42 struct {
	db *SqlDb
}

func (m migration_2_8_42) PostApply(tx *gorp.Transaction) error {
	switch m.db.sql.Dialect.(type) {
	case gorp.MySQLDialect:
		_, _ = tx.Exec(m.db.PrepareQuery("alter table `task` drop foreign key `task_ibfk_3`"))
	case gorp.PostgresDialect:
		_, err := tx.Exec(
			m.db.PrepareQuery("alter table `task` drop constraint if exists `task_build_task_id_fkey`"))
		return err
	}
	return nil
}
