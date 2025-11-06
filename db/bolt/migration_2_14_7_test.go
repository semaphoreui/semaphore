package bolt

import (
	"go.etcd.io/bbolt"
	"testing"
)

func TestMigration_2_14_7_Apply(t *testing.T) {
	store := CreateTestStore()

	err := store.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("project"))
		if err != nil {
			return err
		}

		err = b.Put([]byte("0000000001"), []byte("{}"))
		if err != nil {
			return err
		}

		// Create templates

		r, err := tx.CreateBucketIfNotExists([]byte("project__template_0000000001"))
		if err != nil {
			return err
		}

		err = r.Put([]byte("0000000001"),
			[]byte("{\"id\":\"1\",\"project_id\":\"1\"}"))
		if err != nil {
			return err
		}

		// Create schedules

		r, err = tx.CreateBucketIfNotExists([]byte("project__schedule_0000000001"))
		if err != nil {
			return err
		}

		err = r.Put([]byte("0000000001"),
			[]byte("{\"id\":\"1\",\"project_id\":\"1\",\"template_id\":1}")) // correct

		err = r.Put([]byte("0000000002"),
			[]byte("{\"id\":\"1\",\"project_id\":\"1\",\"template_id\":100}")) // incorrect

		return err
	})

	if err != nil {
		t.Fatal(err)
	}

	err = migration_2_14_7{migration{store.db}}.Apply()
	if err != nil {
		t.Fatal(err)
	}

	var s1, s2 []byte
	err = store.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("project__schedule_0000000001"))
		s1 = b.Get([]byte("0000000001"))
		s2 = b.Get([]byte("0000000002"))
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}

	if s1 == nil {
		t.Fatal("Correct schedule should not be deleted")
	}

	if s2 != nil {
		t.Fatal("Incorrect schedule should be deleted")
	}
}
