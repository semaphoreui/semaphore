package bolt

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.etcd.io/bbolt"
)

func TestMigration_2_17_0_Apply(t *testing.T) {
	store := CreateTestStore()

	err := store.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("project"))
		if err != nil {
			return err
		}

		return b.Put([]byte("0000000001"), []byte("{}"))
	})

	assert.NoError(t, err)

	err = migration_2_17_0{migration{store.db}}.Apply()
	if err != nil {
		t.Fatal(err)
	}

	var s1 []byte
	err = store.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("project__view_0000000001"))
		s1 = b.Get([]byte("0000000001"))
		return nil
	})

	assert.NoError(t, err)
	assert.NotNil(t, s1)

	var res map[string]any
	err = json.Unmarshal(s1, &res)

	assert.NoError(t, err)
	assert.Equal(t, 1.0, res["id"])
	assert.Equal(t, "all", res["type"])
}
