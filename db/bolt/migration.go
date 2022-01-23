package bolt

import (
	"encoding/json"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/db/bolt/migrations"
	"go.etcd.io/bbolt"
)

func (d *BoltDb) IsMigrationApplied(migration db.Migration) (bool, error) {
	err := d.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("migrations"))
		if b == nil {
			return db.ErrNotFound
		}

		d := b.Get([]byte(migration.Version))

		if d == nil {
			return db.ErrNotFound
		}

		return nil
	})

	if err == nil {
		return true, nil
	}

	if err == db.ErrNotFound {
		return false, nil
	}

	return false, err
}

func (d *BoltDb) ApplyMigration(migration db.Migration) (err error) {
	switch migration.Version {
	case "2.8.26":
		err = migrations.Migration_2_8_28{DB: d.db}.Apply()
	}

	if err != nil {
		return
	}

	return d.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("migrations"))

		if err != nil {
			return err
		}

		j, err := json.Marshal(migration)

		if err != nil {
			return err
		}

		return b.Put([]byte(migration.Version), j)
	})
}

func (d *BoltDb) TryRollbackMigration(migration db.Migration) {
	switch migration.Version {
	case "2.8.26":
	}
}
