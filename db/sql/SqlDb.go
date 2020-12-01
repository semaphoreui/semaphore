package sql

import (
	"database/sql"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/models"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/go-gorp/gorp/v3"
	_ "github.com/go-sql-driver/mysql" // imports mysql driver
	"github.com/gobuffalo/packr"
	"github.com/masterminds/squirrel"
	"golang.org/x/crypto/bcrypt"
	"regexp"
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

func handleRollbackError(err error) {
	if err != nil {
		log.Warn(err.Error())
	}
}

var (
	autoIncrementRE = regexp.MustCompile(`(?i)\bautoincrement\b`)
)

const databaseTimeFormat = "2006-01-02T15:04:05:99Z"

// GetParsedTime returns the timestamp as it will retrieved from the database
// This allows us to create timestamp consistency on return values from create requests
func getParsedTime(t time.Time) time.Time {
	parsedTime, err := time.Parse(databaseTimeFormat, t.Format(databaseTimeFormat))
	if err != nil {
		log.Error(err)
	}
	return parsedTime
}

func (d *SqlDb) prepareMigration(query string) string {
	switch d.sql.Dialect.(type) {
	case gorp.MySQLDialect:
		query = autoIncrementRE.ReplaceAllString(query, "auto_increment")
	}
	return query
}

// isMigrationApplied queries the database to see if a migration table with this version id exists already
func (d *SqlDb) isMigrationApplied(version *Version) (bool, error) {
	exists, err := d.sql.SelectInt("select count(1) as ex from migrations where version=?", version.VersionString())

	if err != nil {
		fmt.Println("Creating migrations table")
		if _, err = d.sql.Exec(d.prepareMigration(initialSQL)); err != nil {
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

		if _, err := tx.Exec(d.prepareMigration(query)); err != nil {
			handleRollbackError(tx.Rollback())
			log.Warnf("\n ERR! Query: %v\n\n", query)
			return err
		}
	}

	if _, err := tx.Exec("insert into migrations(version, upgraded_date) values (?, ?)", version.VersionString(), time.Now()); err != nil {
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

		if _, err := d.sql.Exec(d.prepareMigration(query)); err != nil {
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

	return sql.Open(cfg.Dialect.String(), connectionString)
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

	d.sql = &gorp.DbMap{Db: db, Dialect: dialect}

	d.sql.AddTableWithName(models.APIToken{}, "user__token").SetKeys(false, "id")
	d.sql.AddTableWithName(models.AccessKey{}, "access_key").SetKeys(true, "id")
	d.sql.AddTableWithName(models.Environment{}, "project__environment").SetKeys(true, "id")
	d.sql.AddTableWithName(models.Inventory{}, "project__inventory").SetKeys(true, "id")
	d.sql.AddTableWithName(models.Project{}, "project").SetKeys(true, "id")
	d.sql.AddTableWithName(models.Repository{}, "project__repository").SetKeys(true, "id")
	d.sql.AddTableWithName(models.Task{}, "task").SetKeys(true, "id")
	d.sql.AddTableWithName(models.TaskOutput{}, "task__output").SetUniqueTogether("task_id", "time")
	d.sql.AddTableWithName(models.Template{}, "project__template").SetKeys(true, "id")
	d.sql.AddTableWithName(models.User{}, "user").SetKeys(true, "id")
	d.sql.AddTableWithName(models.Session{}, "session").SetKeys(true, "id")

	return nil
}

func (d *SqlDb) GetDbMap() *gorp.DbMap {
	return d.sql
}

func (d *SqlDb) CreateProject(project models.Project) (newProject models.Project, err error) {
	project.Created = time.Now()

	res, err := d.sql.Exec("insert into project(name, created) values (?, ?)", project.Name, project.Created)
	if err != nil {
		return
	}

	projectID, err := res.LastInsertId()
	if err != nil {
		return
	}

	newProject = project
	newProject.ID = int(projectID)
	return
}

func (d *SqlDb) CreateUser(user models.User) (newUser models.User, err error) {
	pwdHash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 11)

	if err != nil {
		return
	}

	created := getParsedTime(time.Now())

	res, err := d.sql.Exec(
		"insert into `user`(name, username, email, password, admin, created) values (?, ?, ?, ?, true, ?)",
		user.Name,
		user.Username,
		user.Email,
		pwdHash,
		created)

	if err != nil {
		return
	}

	userID, err := res.LastInsertId()

	if err != nil {
		return
	}

	newUser = user
	newUser.ID = int(userID)
	newUser.Created = created
	return
}

func (d *SqlDb) DeleteUser(userID int) error {
	_, err := d.sql.Exec("delete from `user` where id=?", userID)
	return err
}

func (d *SqlDb) UpdateUser(userID int, user models.User) error {
	_, err := d.sql.Exec("update `user` set name=?, username=?, email=?, alert=?, admin=? where id=?",
		user.Name,
		user.Username,
		user.Email,
		user.Alert,
		user.Admin,
		userID)

	return err
}

func (d *SqlDb) SetUserPassword(userID int, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 11)
	if err != nil {
		return err
	}
	_, err = d.sql.Exec("update `user` set password=? where id=?", string(hash), userID)
	return err
}


func (d *SqlDb) CreateProjectUser(projectUser models.ProjectUser) (newProjectUser models.ProjectUser, err error) {
	_, err = d.sql.Exec("insert into project__user (project_id, user_id, `admin`) values (?, ?, ?)",
		projectUser.ProjectID,
		projectUser.UserID,
		projectUser.Admin)

	if err != nil {
		return
	}

	newProjectUser = projectUser
	return
}

func (d *SqlDb) CreateEvent(evt models.Event) (newEvent models.Event, err error) {
	var created = time.Now()

	_, err = d.sql.Exec(
		"insert into event(project_id, object_id, object_type, description, created) values (?, ?, ?, ?, ?)",
		evt.ProjectID,
		evt.ObjectID,
		evt.ObjectType,
		evt.Description,
		created)

	if err != nil {
		return
	}

	newEvent = evt
	newEvent.Created = created
	return
}


func (d *SqlDb) DeleteProjectUser(projectID, userID int) error {
	_, err := d.sql.Exec("delete from project__user where user_id=? and project_id=?", userID, projectID)
	return err
}

//FetchUser retrieves a user from the database by ID
func (d *SqlDb) GetUserById(userID int) (models.User, error) {
	var user models.User
	err := d.sql.SelectOne(&user, "select * from `user` where id=?", userID)
	return user, err
}

func getSqlForTable(tableName string, p db.RetrieveQueryParams) (string, []interface{}, error) {
	q := squirrel.Select("*").
		From(tableName)

	if p.SortBy != "" {
		sortDirection := "ASC"
		if p.SortInverted {
			sortDirection = "DESC"
		}

		q = q.OrderBy(p.SortBy + " " + sortDirection)
	}

	q = q.Offset(uint64(p.Offset))

	if p.Count > 0 {
		q = q.Limit(uint64(p.Count))
	}

	return q.ToSql()
}

func (d *SqlDb) GetUsers(params db.RetrieveQueryParams) (users []models.User, err error) {
	query, args, err := getSqlForTable("user", params)

	if err != nil {
		return
	}

	_, err = d.sql.Select(&users, query, args...)

	return
}

func (d *SqlDb) GetAllUsers() (users []models.User, err error) {
	_, err = d.sql.Select(&users, "select * from `user`")
	return
}


func (d *SqlDb) Sql() *gorp.DbMap {
	return d.sql
}


func (d *SqlDb) CreateAPIToken(token models.APIToken) (models.APIToken, error) {
	token.Created = getParsedTime(time.Now())
	err := d.sql.Insert(&token)
	return token, err
}
