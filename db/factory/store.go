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
		return sql.CreateDb(config.Dialect)
	case util.DbDriverBolt:
		return bolt.CreateBoltDB()
	case util.DbDriverPostgres:
		return sql.CreateDb(config.Dialect)
	case util.DbDriverSQLite:
		return sql.CreateDb(config.Dialect)
	default:
		panic("Unsupported database dialect: " + config.Dialect)
	}
	// This line should never be reached due to panic above, but satisfies linter
	return nil
}
