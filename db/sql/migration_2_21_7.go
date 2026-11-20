package sql

import (
	"github.com/go-gorp/gorp/v3"
	"github.com/semaphoreui/semaphore/db"
)

type migration_2_21_7 struct {
	db *SqlDb
}

// PostApply counts the variables written before the column existed.
func (m migration_2_21_7) PostApply(tx *gorp.Transaction) error {
	tables := []struct{ rows, count string }{
		{"select id, variables from project__inventory_host where variables <> ''",
			"update project__inventory_host set variable_count=? where id=?"},
		{"select id, variables from project__inventory where variables <> ''",
			"update project__inventory set variable_count=? where id=?"},
	}
	for _, table := range tables {
		var rows []struct {
			ID        int    `db:"id"`
			Variables string `db:"variables"`
		}
		_, err := tx.Select(&rows, m.db.PrepareQuery(table.rows))
		if err != nil {
			return err
		}
		for _, row := range rows {
			count := db.CountInventoryVariables(row.Variables)
			if count == 0 {
				continue
			}
			_, err = tx.Exec(m.db.PrepareQuery(table.count), count, row.ID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
