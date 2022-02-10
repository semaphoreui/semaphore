package bolt

import (
	"encoding/json"
	"go.etcd.io/bbolt"
	"testing"
)

func TestMigration_2_8_28_Apply(t *testing.T) {
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

		r, err := tx.CreateBucketIfNotExists([]byte("project__repository_0000000001"))
		if err != nil {
			return err
		}

		err = r.Put([]byte("0000000001"),
			[]byte("{\"id\":\"1\",\"project_id\":\"1\",\"git_url\": \"git@github.com/test/test#main\"}"))

		return err
	})

	if err != nil {
		t.Fatal(err)
	}

	err = migration_2_8_28{migration{store.db}}.Apply()
	if err != nil {
		t.Fatal(err)
	}

	var repo map[string]interface{}
	err = store.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("project__repository_0000000001"))
		str := string(b.Get([]byte("0000000001")))
		return json.Unmarshal([]byte(str), &repo)
	})
	if err != nil {
		t.Fatal(err)
	}

	if repo["git_url"].(string) != "git@github.com/test/test" {
		t.Fatal("invalid url")
	}

	if repo["git_branch"].(string) != "main" {
		t.Fatal("invalid branch")
	}
}

func TestMigration_2_8_28_Apply2(t *testing.T) {
	store := CreateTestStore()

	err := store.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("project"))
		if err != nil {
			return err
		}

		err = b.Put([]byte("0000000001"), []byte("{}"))

		return err
	})

	if err != nil {
		t.Fatal(err)
	}

	err = migration_2_8_28{migration{store.db}}.Apply()
	if err != nil {
		t.Fatal(err)
	}
}
