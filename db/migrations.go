package db

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/go-sql-driver/mysql"
	"github.com/gobuffalo/packr"
)

var dbAssets = packr.NewBox("./migrations")

// CheckExists queries the database to see if a migration table with this version id exists already
func (version *Version) CheckExists() (bool, error) {
	exists, err := Mysql.SelectInt("select count(1) as ex from migrations where version=?", version.VersionString())

	if err != nil {
		//nolint: gosimple
		switch err.(type) {
		case *mysql.MySQLError:
			// 1146 is mysql table does not exist
			if err.(*mysql.MySQLError).Number != 1146 {
				return false, err
			}

			fmt.Println("Creating migrations table")
			if _, err = Mysql.Exec(initialSQL); err != nil {
				panic(err)
			}

			return version.CheckExists()
		default:
			return false, err
		}
	}

	return exists > 0, nil
}

// Run executes a database migration
func (version *Version) Run() error {
	fmt.Printf("Executing migration %s (at %v)...\n", version.HumanoidVersion(), time.Now())

	tx, err := Mysql.Begin()
	if err != nil {
		return err
	}

	sql := version.GetSQL(version.GetPath())
	for i, query := range sql {
		fmt.Printf("\r [%d/%d]", i+1, len(sql))

		if len(query) == 0 {
			continue
		}

		if _, err := tx.Exec(query); err != nil {
			handleRollbackError(tx.Rollback())
			log.Warnf("\n ERR! Query: %v\n\n", query)
			return err
		}
	}

	if _, err := tx.Exec("insert into migrations set version=?, upgraded_date=?", version.VersionString(), time.Now()); err != nil {
		handleRollbackError(tx.Rollback())
		return err
	}

	fmt.Println()

	return tx.Commit()
}

func handleRollbackError(err error) {
	if err != nil {
		log.Warn(err.Error())
	}
}

// TryRollback attempts to rollback the database to an earlier version if a rollback exists
func (version *Version) TryRollback() {
	fmt.Printf("Rolling back %s (time: %v)...\n", version.HumanoidVersion(), time.Now())

	data := dbAssets.Bytes(version.GetErrPath())
	if len(data) == 0 {
		fmt.Println("Rollback SQL does not exist.")
		fmt.Println()
		return
	}

	sql := version.GetSQL(version.GetErrPath())
	for _, query := range sql {
		fmt.Printf(" [ROLLBACK] > %v\n", query)

		if _, err := Mysql.Exec(query); err != nil {
			fmt.Println(" [ROLLBACK] - Stopping")
			return
		}
	}
}

// MigrateAll checks for db migrations and executes them
func MigrateAll() error {
	fmt.Println("Checking DB migrations")
	didRun := false

	// go from beginning to the end
	for _, version := range Versions {
		if exists, err := version.CheckExists(); err != nil || exists {
			if exists {
				continue
			}

			return err
		}

		didRun = true
		if err := version.Run(); err != nil {
			version.TryRollback()

			return err
		}
	}

	if didRun {
		fmt.Println("Migrations Finished")
	}

	return nil
}
