package db

import (
	"database/sql"

	"github.com/ansible-semaphore/semaphore/util"
	_ "github.com/go-sql-driver/mysql" // imports mysql driver
	"gopkg.in/gorp.v1"
	log "github.com/Sirupsen/logrus"
)

// Mysql is the gorp database map
// db.Connect must be called to set this up correctly
var Mysql *gorp.DbMap

// Connect to MySQL and initialize the Mysql object
func Connect() error {
	db, err := connect()
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		if err = createDb(); err != nil {
			return err
		}

		db, err = connect()
		if err != nil {
			return err
		}

		if err = db.Ping(); err != nil {
			return err
		}
	}

	Mysql = &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{Engine: "InnoDB", Encoding: "UTF8"}}
	return nil
}

// Close closes the mysql connection and reports any errors
// called from main with a defer
func Close() {
	err := Mysql.Db.Close()
	if err != nil {
		log.Warn("Error closing database:" + err.Error())
	}
}

func createDb() error {
	cfg := util.Config.MySQL
	url := cfg.Username + ":" + cfg.Password + "@tcp(" + cfg.Hostname + ")/?parseTime=true&interpolateParams=true"

	db, err := sql.Open("mysql", url)
	if err != nil {
		return err
	}

	if _, err := db.Exec("create database if not exists " + cfg.DbName); err != nil {
		return err
	}

	return nil
}

func connect() (*sql.DB, error) {
	cfg := util.Config.MySQL
	url := cfg.Username + ":" + cfg.Password + "@tcp(" + cfg.Hostname + ")/" + cfg.DbName + "?parseTime=true&interpolateParams=true"

	return sql.Open("mysql", url)
}
