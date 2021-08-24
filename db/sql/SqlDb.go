package sql

import (
	"database/sql"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/go-gorp/gorp/v3"
	_ "github.com/go-sql-driver/mysql" // imports mysql driver
	"github.com/gobuffalo/packr"
	_ "github.com/lib/pq"
	"github.com/masterminds/squirrel"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type SqlDb struct {
	sql *gorp.DbMap
}

var initialSQL = `
create table ` + "`migrations`" + ` (
	` + "`version`" + ` varchar(255) not null primary key,
	` + "`upgraded_date`" + ` datetime null,
	` + "`notes`" + ` text null
);
`
var dbAssets = packr.NewBox("./migrations")

func containsStr(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func handleRollbackError(err error) {
	if err != nil {
		log.Warn(err.Error())
	}
}

var (
	autoIncrementRE = regexp.MustCompile(`(?i)\bautoincrement\b`)
	serialRE = regexp.MustCompile(`(?i)\binteger primary key autoincrement\b`)
	identifierQuoteRE = regexp.MustCompile("`")
	dateTimeTypeRE = regexp.MustCompile(`(?i)\bdatetime\b`)
	tinyintRE = regexp.MustCompile(`(?i)\btinyint\b`)
	longtextRE = regexp.MustCompile(`(?i)\blongtext\b`)
)

// validateMutationResult checks the success of the update query
func validateMutationResult(res sql.Result, err error) error {
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()

	if err != nil {
		return err
	}

	if affected == 0 {
		return db.ErrNotFound
	}

	return nil
}

func (d *SqlDb) prepareQueryWithDialect(query string, dialect gorp.Dialect) string {
	switch dialect.(type) {
	case gorp.PostgresDialect:
		var queryBuilder strings.Builder
		argNum := 1
		for _, r := range []rune(query) {
			switch r {
			case '?':
				queryBuilder.WriteString("$" + strconv.Itoa(argNum))
				argNum++
			case '`':
				queryBuilder.WriteRune('"')
			default:
				queryBuilder.WriteRune(r)
			}
		}
		query = queryBuilder.String()
	}
	return query
}

func (d *SqlDb) prepareQuery(query string) string {
	return d.prepareQueryWithDialect(query, d.sql.Dialect)
}


func (d *SqlDb) insert(primaryKeyColumnName string, query string, args ...interface{}) (int, error) {
	var insertId int64

	switch d.sql.Dialect.(type) {
	case gorp.PostgresDialect:
		query += " returning " + primaryKeyColumnName

		err := d.sql.QueryRow(d.prepareQuery(query), args...).Scan(&insertId)

		if err != nil {
			return 0, err
		}
	default:
		res, err := d.exec(query, args...)

		if err != nil {
			return 0, err
		}

		insertId, err = res.LastInsertId()

		if err != nil {
			return 0, err
		}
	}

	return int(insertId), nil
}

func (d *SqlDb) exec(query string, args ...interface{}) (sql.Result, error) {
	q := d.prepareQuery(query)
	return d.sql.Exec(q, args...)
}

func (d *SqlDb) selectOne(holder interface{}, query string, args ...interface{}) error {
	return d.sql.SelectOne(holder, d.prepareQuery(query), args...)
}

func (d *SqlDb) selectNullStr(query string, args ...interface{}) (sql.NullString, error) {
	return d.sql.SelectNullStr(d.prepareQuery(query), args...)
}

func (d *SqlDb) selectAll(i interface{}, query string, args ...interface{}) ([]interface{}, error) {
	q := d.prepareQuery(query)
	return d.sql.Select(i, q, args...)
}

// prepareMigration converts migration SQLite-query to current dialect.
// Supported MySQL and Postgres dialects.
func (d *SqlDb) prepareMigration(query string) string {
	switch d.sql.Dialect.(type) {
	case gorp.MySQLDialect:
		query = autoIncrementRE.ReplaceAllString(query, "auto_increment")
	case gorp.PostgresDialect:
		query = serialRE.ReplaceAllString(query, "serial primary key")
		query = identifierQuoteRE.ReplaceAllString(query, "\"")
		query = dateTimeTypeRE.ReplaceAllString(query, "timestamp")
		query = tinyintRE.ReplaceAllString(query, "smallint")
		query = longtextRE.ReplaceAllString(query, "text")
	}
	return query
}

// isMigrationApplied queries the database to see if a migration table with this version id exists already
func (d *SqlDb) isMigrationApplied(version *Version) (bool, error) {
	exists, err := d.sql.SelectInt(d.prepareQuery("select count(1) as ex from migrations where version=?"), version.VersionString())

	if err != nil {
		fmt.Println("Creating migrations table")
		query := d.prepareMigration(initialSQL)
		if _, err = d.exec(query); err != nil {
			panic(err)
		}

		return d.isMigrationApplied(version)
	}

	return exists > 0, nil
}

// Run executes a database migration
func (d *SqlDb) applyMigration(version *Version) error {
	fmt.Printf("Executing migration %s (at %v)...\n", version.HumanoidVersion(), time.Now())

	tx, err := d.sql.Begin()
	if err != nil {
		return err
	}

	query := version.GetSQL(version.GetPath())
	for i, query := range query {
		fmt.Printf("\r [%d/%d]", i+1, len(query))

		if len(query) == 0 {
			continue
		}

		q := d.prepareMigration(query)
		if _, err := tx.Exec(q); err != nil {
			handleRollbackError(tx.Rollback())
			log.Warnf("\n ERR! Query: %v\n\n", q)
			return err
		}
	}

	if _, err := tx.Exec(d.prepareQuery("insert into migrations(version, upgraded_date) values (?, ?)"), version.VersionString(), time.Now()); err != nil {
		handleRollbackError(tx.Rollback())
		return err
	}

	fmt.Println()

	return tx.Commit()
}

// TryRollback attempts to rollback the database to an earlier version if a rollback exists
func (d *SqlDb) tryRollbackMigration(version *Version) {
	fmt.Printf("Rolling back %s (time: %v)...\n", version.HumanoidVersion(), time.Now())

	data := dbAssets.Bytes(version.GetErrPath())
	if len(data) == 0 {
		fmt.Println("Rollback SQL does not exist.")
		fmt.Println()
		return
	}

	query := version.GetSQL(version.GetErrPath())
	for _, query := range query {
		fmt.Printf(" [ROLLBACK] > %v\n", query)

		if _, err := d.exec(d.prepareMigration(query)); err != nil {
			fmt.Println(" [ROLLBACK] - Stopping")
			return
		}
	}
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

	dialect := cfg.Dialect.String()
	return sql.Open(dialect, connectionString)
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

func (d *SqlDb) getObject(projectID int, props db.ObjectProperties, objectID int, object interface{}) (err error) {
	q := squirrel.Select("*").
		From(props.TableName).
		Where("id=?", objectID)

	if props.IsGlobal {
		q = q.Where("project_id is null")
	} else {
		q = q.Where("project_id=?", projectID)
	}

	query, args, err := q.ToSql()

	if err != nil {
		return
	}

	err = d.selectOne(object, query, args...)

	if err == sql.ErrNoRows {
		err = db.ErrNotFound
	}

	return
}

func (d *SqlDb) getObjects(projectID int, props db.ObjectProperties, params db.RetrieveQueryParams, objects interface{}) (err error) {
	q := squirrel.Select("*").
		From(props.TableName + " pe").
		Where("pe.project_id=?", projectID)

	orderDirection := "ASC"
	if params.SortInverted {
		orderDirection = "DESC"
	}

	orderColumn := "name"
	if containsStr(props.SortableColumns, params.SortBy) {
		orderColumn = params.SortBy
	}

	q = q.OrderBy("pe." + orderColumn + " " + orderDirection)

	query, args, err := q.ToSql()

	if err != nil {
		return
	}

	_, err = d.selectAll(objects, query, args...)

	return
}

func (d *SqlDb) isObjectInUse(projectID int, props db.ObjectProperties, objectID int) (bool, error) {
	templatesC, err := d.sql.SelectInt(
		"select count(1) from project__template where project_id=? and " + props.ForeignColumnName+ "=?",
		projectID,
		objectID)

	if err != nil {
		return false, err
	}

	return templatesC > 0, nil
}

func (d *SqlDb) deleteObject(projectID int, props db.ObjectProperties, objectID int) error {
	inUse, err := d.isObjectInUse(projectID, props, objectID)

	if err != nil {
		return err
	}

	if inUse {
		return db.ErrInvalidOperation
	}

	return validateMutationResult(
		d.exec(
			"delete from " + props.TableName + " where project_id=? and id=?",
			projectID,
			objectID))
}

func (d *SqlDb) deleteObjectSoft(projectID int, props db.ObjectProperties, objectID int) error {
	return validateMutationResult(
		d.exec(
			"update " + props.TableName + " set removed=1 where project_id=? and id=?",
			projectID,
			objectID))
}

func (d *SqlDb) Migrate() error {
	fmt.Println("Checking DB migrations")
	didRun := false

	// go from beginning to the end
	for _, version := range Versions {
		if exists, err := d.isMigrationApplied(version); err != nil || exists {
			if exists {
				continue
			}

			return err
		}

		didRun = true
		if err := d.applyMigration(version); err != nil {
			d.tryRollbackMigration(version)

			return err
		}
	}

	if didRun {
		fmt.Println("Migrations Finished")
	}

	return nil
}

func (d *SqlDb) Close() error {
	return d.sql.Db.Close()
}

func (d *SqlDb) Connect() error {
	sqlDb, err := connect()
	if err != nil {
		return err
	}

	if err := sqlDb.Ping(); err != nil {
		if err = createDb(); err != nil {
			return err
		}

		sqlDb, err = connect()
		if err != nil {
			return err
		}

		if err = sqlDb.Ping(); err != nil {
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
	case util.DbDriverPostgres:
		dialect = gorp.PostgresDialect{}
	}

	d.sql = &gorp.DbMap{Db: sqlDb, Dialect: dialect}

	d.sql.AddTableWithName(db.APIToken{}, "user__token").SetKeys(false, "id")
	d.sql.AddTableWithName(db.AccessKey{}, "access_key").SetKeys(true, "id")
	d.sql.AddTableWithName(db.Environment{}, "project__environment").SetKeys(true, "id")
	d.sql.AddTableWithName(db.Inventory{}, "project__inventory").SetKeys(true, "id")
	d.sql.AddTableWithName(db.Project{}, "project").SetKeys(true, "id")
	d.sql.AddTableWithName(db.Repository{}, "project__repository").SetKeys(true, "id")
	d.sql.AddTableWithName(db.Task{}, "task").SetKeys(true, "id")
	d.sql.AddTableWithName(db.TaskOutput{}, "task__output").SetUniqueTogether("task_id", "time")
	d.sql.AddTableWithName(db.Template{}, "project__template").SetKeys(true, "id")
	d.sql.AddTableWithName(db.User{}, "user").SetKeys(true, "id")
	d.sql.AddTableWithName(db.Session{}, "session").SetKeys(true, "id")

	return nil
}

func getSqlForTable(tableName string, p db.RetrieveQueryParams) (string, []interface{}, error) {
	if p.Count > 0 && p.Offset <= 0 {
		return "", nil, fmt.Errorf("offset cannot be without limit")
	}

	q := squirrel.Select("*").
		From("`" + tableName + "`")

	if p.SortBy != "" {
		sortDirection := "ASC"
		if p.SortInverted {
			sortDirection = "DESC"
		}

		q = q.OrderBy(p.SortBy + " " + sortDirection)
	}

	if p.Offset > 0 {
		q = q.Offset(uint64(p.Offset))
	}

	if p.Count > 0 {
		q = q.Limit(uint64(p.Count))
	}

	return q.ToSql()
}


func (d *SqlDb) Sql() *gorp.DbMap {
	return d.sql
}


