package database

import (
	"database/sql"

	"github.com/ansible-semaphore/semaphore/util"
	_ "github.com/go-sql-driver/mysql" // imports mysql driver
	"gopkg.in/gorp.v1"
)

var Mysql *gorp.DbMap

// Mysql database
func Connect() error {
	url := util.Config.MySQL.Username + ":" + util.Config.MySQL.Password + "@tcp(" + util.Config.MySQL.Hostname + ")/" + util.Config.MySQL.DbName + "?parseTime=true&interpolateParams=true"

	db, err := sql.Open("mysql", url)
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	Mysql = &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{Engine: "InnoDB", Encoding: "UTF8"}}

	return nil
}
