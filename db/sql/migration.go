package sql

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/go-gorp/gorp/v3"
	"regexp"
	"strings"
	"time"
)

var (
	autoIncrementRE = regexp.MustCompile(`(?i)\bautoincrement\b`)
	serialRE        = regexp.MustCompile(`(?i)\binteger primary key autoincrement\b`)
	dateTimeTypeRE  = regexp.MustCompile(`(?i)\bdatetime\b`)
	tinyintRE       = regexp.MustCompile(`(?i)\btinyint\b`)
	longtextRE      = regexp.MustCompile(`(?i)\blongtext\b`)
	ifExistsRE      = regexp.MustCompile(`(?i)\bif exists\b`)
	dropForeignKey  = regexp.MustCompile(`(?i)\bdrop foreign key\b`)
)

// getVersionPath is the humanoid version with the file format appended
func getVersionPath(version db.Migration) string {
	return version.HumanoidVersion() + ".sql"
}

// getVersionErrPath is the humanoid version with '.err' and file format appended
func getVersionErrPath(version db.Migration) string {
	return version.HumanoidVersion() + ".err.sql"
}

// getVersionSQL takes a path to an SQL file and returns it from packr as
// a slice of strings separated by newlines
func getVersionSQL(path string) (queries []string) {
	sql, err := dbAssets.MustString(path)
	if err != nil {
		panic(err)
	}
	queries = strings.Split(strings.ReplaceAll(sql, ";\r\n", ";\n"), ";\n")
	return
}

// prepareMigration converts migration SQLite-query to current dialect.
// Supported MySQL and Postgres dialects.
func (d *SqlDb) prepareMigration(query string) string {
	switch d.sql.Dialect.(type) {
	case gorp.MySQLDialect:
		query = autoIncrementRE.ReplaceAllString(query, "auto_increment")
		query = ifExistsRE.ReplaceAllString(query, "")
	case gorp.PostgresDialect:
		query = serialRE.ReplaceAllString(query, "serial primary key")
		query = identifierQuoteRE.ReplaceAllString(query, "\"")
		query = dateTimeTypeRE.ReplaceAllString(query, "timestamp")
		query = tinyintRE.ReplaceAllString(query, "smallint")
		query = longtextRE.ReplaceAllString(query, "text")
		query = dropForeignKey.ReplaceAllString(query, "drop constraint")
	}
	return query
}

// IsMigrationApplied queries the database to see if a migration table with this version id exists already
func (d *SqlDb) IsMigrationApplied(migration db.Migration) (bool, error) {
	exists, err := d.sql.SelectInt(d.prepareQuery("select count(1) as ex from migrations where migration=?"), migration.Version)

	if err != nil {
		fmt.Println("Creating migrations table")
		query := d.prepareMigration(initialSQL)
		if _, err = d.exec(query); err != nil {
			panic(err)
		}

		return d.IsMigrationApplied(migration)
	}

	return exists > 0, nil
}

// ApplyMigration runs executes a database migration
func (d *SqlDb) ApplyMigration(migration db.Migration) error {
	tx, err := d.sql.Begin()
	if err != nil {
		return err
	}

	query := getVersionSQL(getVersionPath(migration))
	for i, query := range query {
		fmt.Printf("\r [%d/%d]", i+1, len(query))

		if len(query) == 0 {
			continue
		}

		q := d.prepareMigration(query)
		_, err = tx.Exec(q)
		if err != nil {
			handleRollbackError(tx.Rollback())
			log.Warnf("\n ERR! Query: %s\n\n", q)
			log.Fatalf(err.Error())
			return err
		}
	}

	if _, err := tx.Exec(d.prepareQuery("insert into migrations(migration, upgraded_date) values (?, ?)"), migration.Version, time.Now()); err != nil {
		handleRollbackError(tx.Rollback())
		return err
	}

	switch migration.Version {
	case "2.8.26":
		rows, err2 := d.sql.Query("SELECT id, git_url FROM project__repository")
		if err2 == nil {
			defer rows.Close()
			for rows.Next() {
				var id, url string

				err3 := rows.Scan(&id, &url)
				if err3 != nil {
					continue
				}

				branch := "master"
				parts := strings.Split(url, "#")
				if len(parts) > 1 {
					url, branch = parts[0], parts[1]
				}
				_, _ = d.sql.Exec("UPDATE project__repository "+
					"SET git_url = ?, git_branch = ? "+
					"WHERE id = ?", url, branch, id)
			}
		}
	}

	fmt.Println()

	return tx.Commit()
}

// TryRollbackMigration attempts to rollback the database to an earlier version if a rollback exists
func (d *SqlDb) TryRollbackMigration(version db.Migration) {
	data := dbAssets.Bytes(getVersionErrPath(version))
	if len(data) == 0 {
		fmt.Println("Rollback SQL does not exist.")
		fmt.Println()
		return
	}

	query := getVersionSQL(getVersionErrPath(version))
	for _, query := range query {
		fmt.Printf(" [ROLLBACK] > %v\n", query)

		if _, err := d.exec(d.prepareMigration(query)); err != nil {
			fmt.Println(" [ROLLBACK] - Stopping")
			return
		}
	}
}
