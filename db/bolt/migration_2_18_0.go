package bolt

import (
	"go.etcd.io/bbolt"
)

type migration_2_18_0 struct {
	migration
}

func (m migration_2_18_0) Apply() error {
	return m.db.Update(func(tx *bbolt.Tx) error {
		// For BoltDB, we don't need to alter table structure
		// The new fields will be automatically handled by the Go struct
		// when objects are saved/loaded
		return nil
	})
}
