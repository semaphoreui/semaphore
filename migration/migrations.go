package migration

import (
	"fmt"
	"github.com/castawaylabs/semaphore/database"
	"github.com/castawaylabs/semaphore/util"
	"github.com/go-sql-driver/mysql"
	"time"
)

func (version *DBVersion) CheckExists() (bool, error) {
	exists, err := database.Mysql.SelectInt("select count(1) as ex from migrations where version=?", version.VersionString())

	if err != nil {
		switch err.(type) {
		case *mysql.MySQLError:
			// 1146 is mysql table does not exist
			if err.(*mysql.MySQLError).Number != 1146 {
				return false, err
			}

			fmt.Println("Creating migrations table")
			if _, err := database.Mysql.Exec(initialSQL); err != nil {
				panic(err)
			}

			return version.CheckExists()
		default:
			return false, err
		}
	}

	return exists > 0, nil
}

func (version *DBVersion) Run() error {
	fmt.Printf("Executing migration %s (at %v)...\n", version.HumanoidVersion(), time.Now())

	tx, err := database.Mysql.Begin()
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
			tx.Rollback()
			fmt.Printf("\n ERR! Query: %v\n\n", query)
			return err
		}
	}

	if _, err := tx.Exec("insert into migrations set version=?, upgraded_date=?", version.VersionString(), time.Now()); err != nil {
		tx.Rollback()
		return err
	}

	fmt.Println()

	return tx.Commit()
}

func (version *DBVersion) TryRollback() {
	fmt.Printf("Rolling back %s (time: %v)...\n", version.HumanoidVersion(), time.Now())

	if _, err := util.Asset(version.GetErrPath()); err != nil {
		fmt.Println("Rollback SQL doesn't exist.")
		fmt.Println()
		return
	}

	sql := version.GetSQL(version.GetErrPath())
	for _, query := range sql {
		fmt.Printf(" [ROLLBACK] > %v\n", query)

		if _, err := database.Mysql.Exec(query); err != nil {
			fmt.Println(" [ROLLBACK] - Stopping")
			return
		}
	}
}

func MigrateAll() error {
	// go from beginning to the end
	for _, version := range Versions {
		if exists, err := version.CheckExists(); err != nil || exists == true {
			if exists == true {
				fmt.Printf("Skipping %s\n", version.HumanoidVersion())
				continue
			}

			return err
		}

		if err := version.Run(); err != nil {
			version.TryRollback()

			return err
		}
	}

	return nil
}
