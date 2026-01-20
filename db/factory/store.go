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
	return CreateStoreWithConfig(config)
}

func CreateStoreWithConfig(config util.DbConfig) db.Store {

	switch config.Dialect {
	case util.DbDriverBolt:
		return bolt.CreateBoltDBWithConfig(config)

	case util.DbDriverMySQL:
	case util.DbDriverPostgres:
	case util.DbDriverSQLite:
		return sql.CreateDbWithConfig(config)

	default:
		panic("Unsupported database dialect: " + config.Dialect)
	}
	// This line should never be reached due to panic above, but satisfies linter
	return nil
}
