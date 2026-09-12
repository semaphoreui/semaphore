package sql

import "github.com/go-gorp/gorp/v3"

type migration_2_10_24 struct {
	db *SqlDb
}

func (m migration_2_10_24) PreApply(tx *gorp.Transaction) error {
	switch m.db.Sql().Dialect.(type) {
	case gorp.MySQLDialect:
		return dropMysqlForeignKey(tx, "project__template", "vault_key_id")
	case gorp.PostgresDialect:
		_, err := tx.Exec(
			m.db.PrepareQuery("alter table `project__template` drop constraint if exists `project__template_vault_key_id_fkey`"))
		return err
	}
	return nil
}
