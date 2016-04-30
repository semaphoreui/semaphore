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
	url := util.Config.MySQL.Username + ":" + util.Config.MySQL.Password + "@tcp(" + util.Config.MySQL.Hostname + ")/?parseTime=true&interpolateParams=true"

	db, err := sql.Open("mysql", url)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	if _, err := db.Exec("create database if not exists " + util.Config.MySQL.DbName); err != nil {
		panic(err)
	}

	if _, err := db.Exec("use " + util.Config.MySQL.DbName); err != nil {
		panic(err)
	}

	Mysql = &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{Engine: "InnoDB", Encoding: "UTF8"}}
	return nil
}
