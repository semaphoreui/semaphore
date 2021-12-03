package factory

import (
	"github.com/neo1908/semaphore/db"
	"github.com/neo1908/semaphore/db/bolt"
	"github.com/neo1908/semaphore/db/sql"
	"github.com/neo1908/semaphore/util"
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
