package sql

import "github.com/go-gorp/gorp/v3"

type migration_2_20_10 struct {
	db *SqlDb
}

// PostApply adds the foreign key on access_key.user_id for MySQL. Migration
// v2.10.15 declared it inline (`add user_id int null references user(id)`),
// which MySQL parses but ignores, while MariaDB and PostgreSQL create the
// constraint. The lookup keeps MariaDB from getting a duplicate one. The
// orphaned access_key rows are removed by v2.20.10.sql before this runs.
func (m migration_2_20_10) PostApply(tx *gorp.Transaction) error {
	switch m.db.Sql().Dialect.(type) {
	case gorp.MySQLDialect:
		fkName, err := findMysqlForeignKey(tx, "access_key", "user_id")
		if err != nil || fkName != "" {
			return err
		}
		_, err = tx.Exec("alter table `access_key` add constraint `access_key__user_fk` " +
			"foreign key (`user_id`) references `user`(`id`) on delete cascade")
		return err
	}
	return nil
}
