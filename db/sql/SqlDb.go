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

// validateMutationResult checks the success of the update query
func validateMutationResult(res sql.Result, err error) error {
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()

	if affected == 0 {
		return db.ErrNotFound
	}

	return nil
}

// getParsedTime returns the timestamp as it will retrieved from the database
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
	res, err := d.sql.Exec("delete from `user` where id=?", userID)

	return validateMutationResult(res, err)
}

func (d *SqlDb) UpdateUser(user models.User) error {
	res, err := d.sql.Exec("update `user` set name=?, username=?, email=?, alert=?, admin=? where id=?",
		user.Name,
		user.Username,
		user.Email,
		user.Alert,
		user.Admin,
		user.ID)

	return validateMutationResult(res, err)
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
func (d *SqlDb) GetUser(userID int) (models.User, error) {
	var user models.User

	err := d.sql.SelectOne(&user, "select * from `user` where id=?", userID)

	if err == sql.ErrNoRows {
		err = db.ErrNotFound
	}

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

func (d *SqlDb) Sql() *gorp.DbMap {
	return d.sql
}


func (d *SqlDb) CreateAPIToken(token models.APIToken) (models.APIToken, error) {
	token.Created = getParsedTime(time.Now())
	err := d.sql.Insert(&token)
	return token, err
}


func (d *SqlDb) GetAPIToken(tokenID string) (token models.APIToken, err error) {
	err = d.sql.SelectOne(&token, "select * from user__token where id=? and expired=0", tokenID)

	if err == sql.ErrNoRows {
		err = db.ErrNotFound
	}

	return
}

func (d *SqlDb) ExpireAPIToken(userID int, tokenID string) (err error) {
	res, err := d.sql.Exec("update user__token set expired=1 where id=? and user_id=?", tokenID, userID)

	return validateMutationResult(res, err)
}

func (d *SqlDb) GetSession(userID int, sessionID int) (session models.Session, err error) {
	err = d.sql.SelectOne(&session, "select * from session where id=? and user_id=? and expired=0", sessionID, userID)

	if err == sql.ErrNoRows {
		err = db.ErrNotFound
	}

	return
}

func (d *SqlDb) ExpireSession(userID int, sessionID int) error {
	res, err := d.sql.Exec("update session set expired=1 where id=? and user_id=?", sessionID, userID)

	return validateMutationResult(res, err)
}

func (d *SqlDb) TouchSession(userID int, sessionID int) error{
	res, err := d.sql.Exec("update session set last_active=? where id=? and user_id=?", time.Now(), sessionID, userID)

	return validateMutationResult(res, err)
}

func (d *SqlDb) GetAPITokens(userID int) (tokens []models.APIToken, err error) {
	_, err = d.sql.Select(&tokens, "select * from user__token where user_id=?", userID)

	if err == sql.ErrNoRows {
		err = db.ErrNotFound
	}

	return
}

func (d *SqlDb) GetEnvironment(projectID int, environmentID int) (env models.Environment, err error) {
	query, args, err := squirrel.Select("*").
		From("project__environment").
		Where("project_id=?", projectID).
		Where("id=?", environmentID).
		ToSql()

	if err != nil {
		return
	}

	err = d.sql.SelectOne(&env, query, args...)

	if err != nil {
		if err == sql.ErrNoRows {
			err = db.ErrNotFound
			return
		}

		return
	}

	return
}

func (d *SqlDb) GetEnvironments(projectID int, params db.RetrieveQueryParams) (environments []models.Environment, err error) {
	q := squirrel.Select("*").
		From("project__environment pe").
		Where("project_id=?", projectID)

	order := "ASC"
	if params.SortInverted {
		order = "DESC"
	}

	switch params.SortBy {
	case "name":
		q = q.Where("pe.project_id=?", projectID).
			OrderBy("pe." + params.SortBy + " " + order)
	default:
		q = q.Where("pe.project_id=?", projectID).
			OrderBy("pe.name " + order)
	}

	query, args, err := q.ToSql()

	if err != nil {
		return
	}

	_, err = d.sql.Select(&environments, query, args...)

	return
}

func (d *SqlDb) UpdateEnvironment(env models.Environment) error {
	res, err := d.sql.Exec("update project__environment set name=?, json=? where id=?", env.Name, env.JSON, env.ID)
	return validateMutationResult(res, err)
}

func (d *SqlDb) CreateTemplate(template models.Template) (newTemplate models.Template, err error) {
	res, err := d.sql.Exec("insert into project__template set ssh_key_id=?, project_id=?, inventory_id=?, repository_id=?, environment_id=?, alias=?, playbook=?, arguments=?, override_args=?",
		template.SSHKeyID,
		template.ProjectID,
		template.InventoryID,
		template.RepositoryID,
		template.EnvironmentID,
		template.Alias,
		template.Playbook,
		template.Arguments,
		template.OverrideArguments)
	if err != nil {
		return
	}

	insertID, err := res.LastInsertId()
	if err != nil {
		return
	}

	newTemplate = template
	newTemplate.ID = int(insertID)
	return
}

func (d *SqlDb) UpdateTemplate(template models.Template) error {
	res, err := d.sql.Exec("update project__template set ssh_key_id=?, inventory_id=?, repository_id=?, environment_id=?, alias=?, playbook=?, arguments=?, override_args=? where id=?",
		template.SSHKeyID,
		template.InventoryID,
		template.RepositoryID,
		template.EnvironmentID,
		template.Alias,
		template.Playbook,
		template.Arguments,
		template.OverrideArguments,
		template.ID)

	return validateMutationResult(res, err)
}

func (d *SqlDb) GetTemplates(projectID int, params db.RetrieveQueryParams) (templates []models.Template, err error) {
	q := squirrel.Select("pt.id",
		"pt.ssh_key_id",
		"pt.project_id",
		"pt.inventory_id",
		"pt.repository_id",
		"pt.environment_id",
		"pt.alias",
		"pt.playbook",
		"pt.arguments",
		"pt.override_args").
		From("project__template pt")

	order := "ASC"
	if params.SortInverted {
		order = "DESC"
	}

	switch params.SortBy {
	case "alias", "playbook":
		q = q.Where("pt.project_id=?", projectID).
			OrderBy("pt." + params.SortBy + " " + order)
	case "ssh_key":
		q = q.LeftJoin("access_key ak ON (pt.ssh_key_id = ak.id)").
			Where("pt.project_id=?", projectID).
			OrderBy("ak.name " + order)
	case "inventory":
		q = q.LeftJoin("project__inventory pi ON (pt.inventory_id = pi.id)").
			Where("pt.project_id=?", projectID).
			OrderBy("pi.name " + order)
	case "environment":
		q = q.LeftJoin("project__environment pe ON (pt.environment_id = pe.id)").
			Where("pt.project_id=?", projectID).
			OrderBy("pe.name " + order)
	case "repository":
		q = q.LeftJoin("project__repository pr ON (pt.repository_id = pr.id)").
			Where("pt.project_id=?", projectID).
			OrderBy("pr.name " + order)
	default:
		q = q.Where("pt.project_id=?", projectID).
			OrderBy("pt.alias " + order)
	}

	query, args, err := q.ToSql()

	if err != nil {
		return
	}

	_, err = d.sql.Select(&templates, query, args...)
	return
}

func (d *SqlDb) GetTemplate(projectID int, templateID int) (template models.Template, err error) {
	err = d.sql.SelectOne(
		&template,
		"select * from project__template where project_id=? and id=?",
		projectID,
		templateID)

	if err == sql.ErrNoRows {
		err = db.ErrNotFound
	}

	return
}

func (d *SqlDb) DeleteTemplate(projectID int, templateID int) error {
	res, err := d.sql.Exec(
		"delete from project__template where project_id=? and id=?",
		projectID,
		templateID)

	return validateMutationResult(res, err)
}
