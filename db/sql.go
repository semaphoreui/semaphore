package db

import (
	"database/sql"
	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/go-gorp/gorp/v3"
	_ "github.com/go-sql-driver/mysql" // imports mysql driver
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"regexp"
	"time"
)

// Sql is the gorp database map
// db.Connect must be called to set this up correctly
var Sql *gorp.DbMap

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

	cfg, err := util.Config.GetDBConfig()
	if err != nil {
		return err
	}

	var dialect gorp.Dialect

	switch cfg.Dialect {
	case util.DbDriverMySQL:
		dialect = gorp.MySQLDialect{Engine: "InnoDB", Encoding: "UTF8"}
	case util.DbDriverSQLite:
		dialect = gorp.SqliteDialect{}
	}

	Sql = &gorp.DbMap{Db: db, Dialect: dialect}
	return nil
}

// Close closes the mysql connection and reports any errors
// called from main with a defer
func Close() {
	err := Sql.Db.Close()
	if err != nil {
		log.Warn("Error closing database:" + err.Error())
	}
}

func createDb() error {
	cfg, err := util.Config.GetDBConfig()
	if err != nil {
		return err
	}

	if !cfg.HasSupportMultipleDatabases() {
		return nil
	}

	connectionString, err := cfg.GetConnectionString(false)
	if err != nil {
		return err
	}

	db, err := sql.Open(cfg.Dialect.String(), connectionString)
	if err != nil {
		return err
	}

	_, err = db.Exec("create database " + cfg.DbName)

	if err != nil {
		log.Warn(err.Error())
	}

	return nil
}

func connect() (*sql.DB, error) {
	cfg, err := util.Config.GetDBConfig()
	if err != nil {
		return nil, err
	}

	connectionString, err := cfg.GetConnectionString(true)
	if err != nil {
		return nil, err
	}


	return sql.Open(cfg.Dialect.String(), connectionString)
}


var (
	autoIncrementRE = regexp.MustCompile(`(?i)\bautoincrement\b`)
)

func PrepareMigration(query string) string {
	switch Sql.Dialect.(type) {
	case gorp.MySQLDialect:
		query = autoIncrementRE.ReplaceAllString(query, "auto_increment")
	}
	return query
}
