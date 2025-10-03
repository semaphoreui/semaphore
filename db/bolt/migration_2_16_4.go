package bolt

import (
	"go.etcd.io/bbolt"
)

type migration_2_16_4 struct {
	migration
}

func (m migration_2_16_4) Apply() error {
	return m.db.Update(func(tx *bbolt.Tx) error {
		// Create the task_files bucket
		_, err := tx.CreateBucketIfNotExists([]byte("task_files"))
		return err
	})
}
