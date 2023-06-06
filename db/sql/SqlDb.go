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
	"reflect"
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
	identifierQuoteRE = regexp.MustCompile("`")
)

// validateMutationResult checks the success of the update query
func validateMutationResult(res sql.Result, err error) error {
	if err != nil {
		if strings.Contains(err.Error(), "foreign key") {
			err = db.ErrInvalidOperation
		}
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
		for _, r := range query {
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

func (d *SqlDb) PrepareQuery(query string) string {
	return d.prepareQueryWithDialect(query, d.sql.Dialect)
}

func (d *SqlDb) insert(primaryKeyColumnName string, query string, args ...interface{}) (int, error) {
	var insertId int64

	switch d.sql.Dialect.(type) {
	case gorp.PostgresDialect:
		query += " returning " + primaryKeyColumnName

		err := d.sql.QueryRow(d.PrepareQuery(query), args...).Scan(&insertId)

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
	q := d.PrepareQuery(query)
	return d.sql.Exec(q, args...)
}

func (d *SqlDb) selectOne(holder interface{}, query string, args ...interface{}) error {
	return d.sql.SelectOne(holder, d.PrepareQuery(query), args...)
}

func (d *SqlDb) selectAll(i interface{}, query string, args ...interface{}) ([]interface{}, error) {
	q := d.PrepareQuery(query)
	return d.sql.Select(i, q, args...)
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

	conn, err := sql.Open(cfg.Dialect.String(), connectionString)
	if err != nil {
		return err
	}

	_, err = conn.Exec("create database " + cfg.GetDbName())

	if err != nil {
		log.Warn(err.Error())
	}

	return nil
}

func (d *SqlDb) getObject(projectID int, props db.ObjectProps, objectID int, object interface{}) (err error) {
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

func (d *SqlDb) getObjects(projectID int, props db.ObjectProps, params db.RetrieveQueryParams, objects interface{}) (err error) {
	q := squirrel.Select("*").
		From(props.TableName+" pe").
		Where("pe.project_id=?", projectID)

	orderDirection := "ASC"
	if params.SortInverted {
		orderDirection = "DESC"
	}

	orderColumn := props.DefaultSortingColumn
	if containsStr(props.SortableColumns, params.SortBy) {
		orderColumn = params.SortBy
	}

	if orderColumn != "" {
		q = q.OrderBy("pe." + orderColumn + " " + orderDirection)
	}

	query, args, err := q.ToSql()

	if err != nil {
		return
	}

	_, err = d.selectAll(objects, query, args...)

	return
}

func (d *SqlDb) deleteObject(projectID int, props db.ObjectProps, objectID int) error {
	return validateMutationResult(
		d.exec(
			"delete from "+props.TableName+" where project_id=? and id=?",
			projectID,
			objectID))
}

func (d *SqlDb) Close(token string) {
	err := d.sql.Db.Close()
	if err != nil {
		panic(err)
	}
}

func (d *SqlDb) PermanentConnection() bool {
	return true
}

func keepAlive(dbc *sql.DB) {
	for {
		err := dbc.Ping()
		if err != nil {
			panic(err)
			return
		}
		time.Sleep(60 * time.Second)
	}
}

func (d *SqlDb) Connect(token string) {
	sqlDb, err := connect()
	if err != nil {
		panic(err)
	}

	if err := sqlDb.Ping(); err != nil {
		if err = createDb(); err != nil {
			panic(err)
		}

		sqlDb, err = connect()
		if err != nil {
			panic(err)
		}

		if err = sqlDb.Ping(); err != nil {
			panic(err)
		}
	}
	if token == "root" {
		go keepAlive(sqlDb)
	}

	cfg, err := util.Config.GetDBConfig()
	if err != nil {
		panic(err)
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
}

func getSqlForTable(tableName string, p db.RetrieveQueryParams) (string, []interface{}, error) {
	if p.Offset > 0 && p.Count <= 0 {
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

	if p.Offset > 0 || p.Count > 0 {
		q = q.Offset(uint64(p.Offset))
	}

	if p.Count > 0 {
		q = q.Limit(uint64(p.Count))
	}

	return q.ToSql()
}

func (d *SqlDb) getObjectRefs(projectID int, objectProps db.ObjectProps, objectID int) (refs db.ObjectReferrers, err error) {
	refs.Templates, err = d.getObjectRefsFrom(projectID, objectProps, objectID, db.TemplateProps)
	if err != nil {
		return
	}

	refs.Repositories, err = d.getObjectRefsFrom(projectID, objectProps, objectID, db.RepositoryProps)
	if err != nil {
		return
	}

	refs.Inventories, err = d.getObjectRefsFrom(projectID, objectProps, objectID, db.InventoryProps)
	if err != nil {
		return
	}

	templates, err := d.getObjectRefsFrom(projectID, objectProps, objectID, db.ScheduleProps)
	if err != nil {
		return
	}

	for _, st := range templates {
		exists := false
		for _, tpl := range refs.Templates {
			if tpl.ID == st.ID {
				exists = true
				break
			}
		}
		if exists {
			continue
		}
		refs.Templates = append(refs.Templates, st)
	}

	return
}

func (d *SqlDb) getObjectRefsFrom(
	projectID int,
	objectProps db.ObjectProps,
	objectID int,
	referringObjectProps db.ObjectProps,
) (referringObjs []db.ObjectReferrer, err error) {
	referringObjs = make([]db.ObjectReferrer, 0)

	fields, err := objectProps.GetReferringFieldsFrom(referringObjectProps.Type)

	cond := ""
	vals := []interface{}{projectID}

	for _, f := range fields {
		if cond != "" {
			cond += " or "
		}

		cond += f + " = ?"

		vals = append(vals, objectID)
	}

	if cond == "" {
		return
	}

	var referringObjects reflect.Value

	if referringObjectProps.Type == db.ScheduleProps.Type {
		var referringSchedules []db.Schedule
		_, err = d.selectAll(&referringSchedules, "select template_id id from project__schedule where project_id = ? and ("+cond+")", vals...)

		if err != nil {
			return
		}

		if len(referringSchedules) == 0 {
			return
		}

		var ids []string
		for _, schedule := range referringSchedules {
			ids = append(ids, strconv.Itoa(schedule.ID))
		}

		referringObjects = reflect.New(reflect.SliceOf(db.TemplateProps.Type))
		_, err = d.selectAll(referringObjects.Interface(),
			"select id, name from project__template where id in ("+strings.Join(ids, ",")+")")
	} else {
		referringObjects = reflect.New(reflect.SliceOf(referringObjectProps.Type))
		_, err = d.selectAll(
			referringObjects.Interface(),
			"select id, name from "+referringObjectProps.TableName+" where project_id = ? and "+cond,
			vals...)
	}

	if err != nil {
		return
	}

	for i := 0; i < referringObjects.Elem().Len(); i++ {
		id := int(referringObjects.Elem().Index(i).FieldByName("ID").Int())
		name := referringObjects.Elem().Index(i).FieldByName("Name").String()
		referringObjs = append(referringObjs, db.ObjectReferrer{ID: id, Name: name})
	}

	return
}

func (d *SqlDb) Sql() *gorp.DbMap {
	return d.sql
}

func (d *SqlDb) IsInitialized() (bool, error) {
	_, err := d.sql.SelectInt(d.PrepareQuery("select count(1) from migrations"))
	return err == nil, nil
}
