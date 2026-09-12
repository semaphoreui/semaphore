package sql

import "github.com/go-gorp/gorp/v3"

type migration_2_7_8 struct {
	db *SqlDb
}

func (m migration_2_7_8) PreApply(tx *gorp.Transaction) error {
	switch m.db.Sql().Dialect.(type) {
	case gorp.MySQLDialect:
		if err := dropMysqlForeignKey(tx, "project__inventory", "key_id"); err != nil {
			return err
		}
		return dropMysqlForeignKey(tx, "project__template", "ssh_key_id")
	}
	// On Postgres dropping the column also drops its foreign key automatically,
	// and SQLite never reaches this migration (it starts from the 2.15.1
	// baseline schema), so no other dialect needs handling here.
	return nil
}
