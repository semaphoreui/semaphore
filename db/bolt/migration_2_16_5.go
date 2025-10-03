package bolt

import (
	"github.com/Digital-Data-Co/forge/db"
	"go.etcd.io/bbolt"
)

type migration_2_16_5 struct {
	migration
}

func (m migration_2_16_5) Apply() error {
	return m.db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("project_alert_config"))
		return err
	})
}
