package db

import (
	"database/sql"

	"github.com/ansible-semaphore/semaphore/util"
	_ "github.com/go-sql-driver/mysql" // imports mysql driver
	"gopkg.in/gorp.v1"
	"time"
	log "github.com/Sirupsen/logrus"
)

// Mysql is the gorp database map
// db.Connect must be called to set this up correctly
var Mysql *gorp.DbMap

// DatabaseTimeFormat represents the format that dredd uses to validate the datetime.
// This is not the same as the raw value we pass to a new object so
// we need to use this to coerce raw values to meet the API standard
// /^\d{4}-(?:0[0-9]{1}|1[0-2]{1})-[0-9]{2}T\d{2}:\d{2}:\d{2}Z$/
const DatabaseTimeFormat = "2006-01-02T15:04:05:99Z"

// GetParsedTime returns the timestamp as it will retrieved from the database
// This allows us to create timestamp consistency on return values from create requests
func GetParsedTime(t time.Time) time.Time {
	parsedTime, err := time.Parse(DatabaseTimeFormat,t.Format(DatabaseTimeFormat))
	if err != nil {
		log.Error(err)
	}
	return parsedTime
}
// Connect ensures that the db is connected and mapped properly with gorp
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
