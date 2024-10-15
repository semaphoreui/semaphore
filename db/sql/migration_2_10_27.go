package sql

import "github.com/go-gorp/gorp/v3"

type migration_2_10_27 struct {
	db *SqlDb
}

func (m migration_2_10_27) PostApply(tx *gorp.Transaction) error {
	switch m.db.sql.Dialect.(type) {
	case gorp.MySQLDialect:
		_, _ = tx.Exec(m.db.PrepareQuery("alter table `task` modify `hosts_limit` text default null;"))
	case gorp.PostgresDialect:
		_, err := tx.Exec(
			m.db.PrepareQuery("alter table `task` alter column `hosts_limit` type text, alter column `hosts_limit` set default null;"))
		return err
	}
	return nil
}
