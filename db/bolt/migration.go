package bolt

import (
	"encoding/json"
	"github.com/ansible-semaphore/semaphore/db"
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

func (d *BoltDb) ApplyMigration(m db.Migration) (err error) {
	switch m.Version {
	case "2.8.26":
		err = migration_2_8_28{migration{d.db}}.Apply()
	case "2.8.40":
		err = migration_2_8_40{migration{d.db}}.Apply()
	case "2.8.91":
		err = migration_2_8_91{migration{d.db}}.Apply()
	case "2.10.12":
		err = migration_2_10_12{migration{d.db}}.Apply()
	case "2.10.16":
		err = migration_2_10_16{migration{d.db}}.Apply()
	case "2.10.24":
		err = migration_2_10_24{migration{d.db}}.Apply()
	}

	if err != nil {
		return
	}

	return d.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("migrations"))

		if err != nil {
			return err
		}

		j, err := json.Marshal(m)

		if err != nil {
			return err
		}

		return b.Put([]byte(m.Version), j)
	})
}

func (d *BoltDb) TryRollbackMigration(m db.Migration) {
	switch m.Version {
	case "2.8.26":
	}
}

type migration struct {
	db *bbolt.DB
}

func (d migration) getProjectIDs() (projectIDs []string, err error) {
	err = d.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("project"))
		if b == nil {
			return nil
		}
		return b.ForEach(func(id, _ []byte) error {
			projectIDs = append(projectIDs, string(id))
			return nil
		})
	})
	return
}

// getObjects returns map of following format: map[OBJECT_ID]map[FIELD_NAME]interface{}
func (d migration) getObjects(projectID string, objectPrefix string) (map[string]map[string]interface{}, error) {
	repos := make(map[string]map[string]interface{}) // ???

	err := d.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("project__" + objectPrefix + "_" + projectID))
		if b == nil {
			return nil
		}
		return b.ForEach(func(id, body []byte) error {
			r := make(map[string]interface{})
			repos[string(id)] = r
			return json.Unmarshal(body, &r)
		})
	})

	return repos, err
}

func (d migration) setObject(projectID string, objectPrefix string, objectID string, object map[string]interface{}) error {
	return d.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("project__" + objectPrefix + "_" + projectID))
		if err != nil {
			return err
		}
		j, err := json.Marshal(object)
		if err != nil {
			return err
		}
		return b.Put([]byte(objectID), j)
	})
}
