package factory

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/bolt"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/semaphoreui/semaphore/util"
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
