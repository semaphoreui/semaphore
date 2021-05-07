package bolt

import (
	"fmt"
	"github.com/ansible-semaphore/semaphore/util"
	"go.etcd.io/bbolt"
)

type BoltDb struct {
	db *bbolt.DB
}

func makeObjectId(tableName string, ids ...int) ([]byte, error) {
	n := len(ids)

	id := tableName
	for i := 0; i < n; i++ {
		id += fmt.Sprintf("_%010d", ids[i])
	}

	return []byte(id), nil
}

func (d *BoltDb) Migrate() error {
	return nil
}

func (d *BoltDb) Connect() error {
	config, err := util.Config.GetDBConfig()
	if err != nil {
		return err
	}
	db, err := bbolt.Open(config.Hostname, 0666, nil)
	if err != nil {
		return err
	}
	d.db = db
	return nil
}

func (d *BoltDb) Close() error {
	return d.db.Close()
}
