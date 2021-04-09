package bolt

import (
	"github.com/ansible-semaphore/semaphore/util"
	bolt "go.etcd.io/bbolt"
)

type BoltDb struct {
	db *bolt.DB
}

func (d *BoltDb) Migrate() {

}

func (d *BoltDb) Connect() error {
	config, err := util.Config.GetDBConfig()
	if err != nil {
		return err
	}
	db, err := bolt.Open(config.Hostname, 0666, nil)
	if err != nil {
		return err
	}
	d.db = db
	return nil
}

func (d *BoltDb) Close() error {
	return d.db.Close()
}
