package factory

import (
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/db/bolt"
	"github.com/ansible-semaphore/semaphore/db/sql"
	"github.com/ansible-semaphore/semaphore/util"
)

func CreateStore() db.Store {
	config, err := util.Config.GetDBConfig()
	if err != nil {
		panic("Can not read configuration")
	}
	switch config.Dialect {
	case util.DbDriverMySQL:
		return &sql.SqlDb{}
	case util.DbDriverBolt:
		return &bolt.BoltDb{}
	case util.DbDriverPostgres:
		return &sql.SqlDb{}
	default:
		panic("Unsupported database dialect: " + config.Dialect)
	}
}
