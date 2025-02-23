package sql

import "github.com/go-gorp/gorp/v3"

type migration_2_12_6 struct {
	db *SqlDb
}

func (m migration_2_12_6) PreApply(tx *gorp.Transaction) (err error) {
	// Add the "path" field to the repository table
	return nil
}

func (m migration_2_12_6) PostApply(tx *gorp.Transaction) error {
	return nil
}
