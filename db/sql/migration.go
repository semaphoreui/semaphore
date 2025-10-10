package sql

import (
	"bytes"
	"fmt"
	"path"
	"regexp"
	"strings"

	"text/template"

	"github.com/go-gorp/gorp/v3"
	"github.com/semaphoreui/semaphore/pkg/tz"
	"github.com/semaphoreui/semaphore/util"

	"github.com/semaphoreui/semaphore/db"
	log "github.com/sirupsen/logrus"
)

var (
	autoIncrementRE  = regexp.MustCompile(`(?i)\bautoincrement\b`)
	serialRE         = regexp.MustCompile(`(?i)\binteger primary key autoincrement\b`)
	dateTimeTypeRE   = regexp.MustCompile(`(?i)\bdatetime\b`)
	tinyintRE        = regexp.MustCompile(`(?i)\btinyint\b`)
	longtextRE       = regexp.MustCompile(`(?i)\blongtext\b`)
	ifExistsRE       = regexp.MustCompile(`(?i)\bif exists\b`)
	changeRE         = regexp.MustCompile(`^alter table \x60(\w+)\x60 change \x60(\w+)\x60 \x60(\w+)\x60 ([\w\(\)]+)( autoincrement)?( not null)?$`)
	dropForeignKeyRE = regexp.MustCompile(`(?i)\bdrop foreign key\b`)
)

// getVersionPath is the humanoid version with the file format appended
func getVersionPath(version db.Migration) string {
	return version.HumanoidVersion() + ".sql"
}

// getVersionErrPath is the humanoid version with '.err' and file format appended
func getVersionErrPath(version db.Migration) string {
	return version.HumanoidVersion() + ".err.sql"
}

// getVersionSQL takes a path to an SQL file and returns it from embed.FS
// a slice of strings separated by newlines
func getVersionSQL(dialect string, name string, ignoreErrors bool) (queries []string) {
	sql, err := dbAssets.ReadFile(path.Join("migrations", name))
	if err != nil {
		if ignoreErrors {
			log.WithError(err).Warnf("migration %s not found", name)
			return nil
		} else {
			panic(err)
		}
	}

	processedSql, err := preprocessSqlDialect(dialect, string(sql))

	if err != nil {
		panic(err)
	}

	queries = strings.Split(strings.ReplaceAll(processedSql, ";\r\n", ";\n"), ";\n")
	for i := range queries {
		queries[i] = strings.Trim(queries[i], "\r\n\t ")
	}
	return
}

func getDialectConfig(dialect string) interface{} {
	type Config struct {
		Sqlite     bool
		Mysql      bool
		Postgresql bool
	}

	conf := Config{}

	switch dialect {
	case util.DbDriverSQLite:
		conf.Sqlite = true
	case "mysql":
		conf.Mysql = true
	case "postgres":
		conf.Postgresql = true
	}

	return conf
}

func preprocessSqlDialect(dialect string, sql string) (string, error) {

	tmpl, err := template.New("sql").Parse(sql)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, getDialectConfig(dialect))
	if err != nil {
		panic(err)
	}
	return buf.String(), nil
}

// prepareMigration converts migration SQLite-query to current dialect.
// Supported MySQL and Postgres dialects.
func (d *SqlDb) prepareMigration(query string) string {
	switch d.Sql().Dialect.(type) {
	case gorp.MySQLDialect:
		query = autoIncrementRE.ReplaceAllString(query, "auto_increment")
		query = ifExistsRE.ReplaceAllString(query, "")
	case gorp.PostgresDialect:
		m := changeRE.FindStringSubmatch(query)
		if m != nil {
			tableName := m[1]
			oldColumnName := m[2]
			newColumnName := m[3]
			columnType := m[4]
			//autoincrement := m[5] != ""
			columnNotNull := m[6] != ""

			var queries []string
			queries = append(queries,
				"alter table `"+tableName+"` alter column `"+oldColumnName+"` type "+columnType)

			if columnNotNull {
				queries = append(queries,
					"alter table `"+tableName+"` alter column `"+oldColumnName+"` set not null")
			} else {
				queries = append(queries,
					"alter table `"+tableName+"` alter column `"+oldColumnName+"` drop not null")
			}

			if oldColumnName != newColumnName {
				queries = append(queries,
					"alter table `"+tableName+"` rename column `"+oldColumnName+"` to `"+newColumnName+"`")
			}

			query = strings.Join(queries, "; ")
		}

		query = dateTimeTypeRE.ReplaceAllString(query, "timestamp")
		query = tinyintRE.ReplaceAllString(query, "smallint")
		query = longtextRE.ReplaceAllString(query, "text")
		query = serialRE.ReplaceAllString(query, "serial primary key")
		query = dropForeignKeyRE.ReplaceAllString(query, "drop constraint")
		query = identifierQuoteRE.ReplaceAllString(query, "\"")
	}
	return query
}

// IsMigrationApplied queries the database to see if a migration table with this version id exists already
func (d *SqlDb) IsMigrationApplied(migration db.Migration) (bool, error) {
	initialized, err := d.IsInitialized()

	if err != nil {
		return false, err
	}

	if !initialized {
		return false, nil
	}

	exists, err := d.Sql().SelectInt(
		d.PrepareQuery("select count(1) as ex from migrations where version = ?"),
		migration.Version)

	if err != nil {
		return false, err
	}

	return exists > 0, nil
}

// ApplyMigration runs executes a database migration
func (d *SqlDb) ApplyMigration(migration db.Migration) error {
	initialized, err := d.IsInitialized()

	if err != nil {
		return err
	}

	if !initialized {
		fmt.Println("Creating migrations table")
		query := d.prepareMigration(initialSQL)
		if query == "" {
			return nil
		}
		_, err = d.exec(query)
		if err != nil {
			return err
		}
	}

	tx, err := d.Sql().Begin()
	if err != nil {
		return err
	}

	switch migration.Version {
	case "2.10.24":
		err = migration_2_10_24{db: d}.PreApply(tx)
	}

	if err != nil {
		handleRollbackError(tx.Rollback())
		return err
	}

	queries := getVersionSQL(d.GetDialect(), getVersionPath(migration), false)
	for i, query := range queries {
		fmt.Printf("\r [%d/%d]", i+1, len(query))

		if len(query) == 0 {
			continue
		}

		q := d.prepareMigration(query)
		if q == "" {
			continue
		}

		_, err = tx.Exec(q)
		if err != nil {
			handleRollbackError(tx.Rollback())
			log.Warnf("\n ERR! Query: %s\n\n", q)
			log.Fatal(err.Error())
			return err
		}
	}

	switch migration.Version {
	case "2.8.26":
		err = migration_2_8_26{db: d}.PostApply(tx)
	case "2.8.42":
		err = migration_2_8_42{db: d}.PostApply(tx)
	}

	if err != nil {
		handleRollbackError(tx.Rollback())
		return err
	}

	_, err = tx.Exec(d.PrepareQuery("insert into migrations(version, upgraded_date) values (?, ?)"), migration.Version, tz.Now())
	if err != nil {
		handleRollbackError(tx.Rollback())
		return err
	}

	fmt.Println()

	return tx.Commit()
}

// TryRollbackMigration attempts to rollback the database to an earlier version if a rollback exists
func (d *SqlDb) TryRollbackMigration(version db.Migration) {
	var err error

	tx, err := d.Sql().Begin()
	if err != nil {
		panic(err)
	}

	defer func() {
		if err == nil {
			err = tx.Commit()
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"context": "migration",
					"version": version.Version,
				}).Error("failed to commit undo migration transaction")
			}
		} else {
			_ = tx.Rollback()
			log.Error(err)
		}
	}()

	queries := getVersionSQL(d.GetDialect(), getVersionErrPath(version), true)

	for _, query := range queries {
		fmt.Printf(" [ROLLBACK] > %v\n", query)
		q := d.prepareMigration(query)
		if q == "" {
			continue
		}
		if _, err = d.execTx(tx, q); err != nil {
			fmt.Println(" [ROLLBACK] - Stopping")
			return
		}
	}

	_, err = d.execTx(tx, "delete from migrations where version=?", version.Version)
}
