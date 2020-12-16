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



func (d *SqlDb) getObject(projectID int, tableName string, objectID int, object interface{}) (err error) {
	query, args, err := squirrel.Select("*").
		From(tableName).
		Where("project_id=?", projectID).
		Where("id=?", objectID).
		ToSql()

	if err != nil {
		return
	}

	err = d.sql.SelectOne(object, query, args...)

	if err == sql.ErrNoRows {
		err = db.ErrNotFound
	}

	return
}

func (d *SqlDb) getObjects(projectID int, tableName string, sortableColumns []string, params db.RetrieveQueryParams, objects interface{}) (err error) {
	q := squirrel.Select("*").
		From(tableName + " pe").
		Where("pe.project_id=?", projectID)

	orderDirection := "ASC"
	if params.SortInverted {
		orderDirection = "DESC"
	}

	orderColumn := "name"
	if containsStr(sortableColumns, params.SortBy) {
		orderColumn = params.SortBy
	}

	q = q.OrderBy("pe." + orderColumn + " " + orderDirection)

	query, args, err := q.ToSql()

	if err != nil {
		return
	}

	_, err = d.sql.Select(objects, query, args...)

	return
}

func (d *SqlDb) isObjectInUse(projectID int, templateColumnName string, objectID int) (bool, error) {
	templatesC, err := d.sql.SelectInt(
		"select count(1) from project__template where project_id=? and " + templateColumnName + "=?",
		projectID,
		objectID)

	if err != nil {
		return false, err
	}

	return templatesC > 0, nil
}

func (d *SqlDb) deleteObject(projectID int, tableName string, templateColumnName string, objectID int) error {
	inUse, err := d.isObjectInUse(projectID, templateColumnName, objectID)

	if err != nil {
		return err
	}

	if inUse {
		return db.ErrInvalidOperation
	}

	return validateMutationResult(
		d.sql.Exec(
			"delete from " + tableName + " where project_id=? and id=?",
			projectID,
			objectID))
}

func (d *SqlDb) deleteObjectSoft(projectID int, tableName string, objectID int) error {
	return validateMutationResult(
		d.sql.Exec(
			"update " + tableName + " set removed=1 where project_id=? and id=?",
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

func (d *SqlDb) CreateProject(project db.Project) (newProject db.Project, err error) {
	project.Created = time.Now()

	res, err := d.sql.Exec("insert into project(name, created) values (?, ?)", project.Name, project.Created)
	if err != nil {
		return
	}

	insertId, err := res.LastInsertId()
	if err != nil {
		return
	}

	newProject = project
	newProject.ID = int(insertId)
	return
}

func (d *SqlDb) GetProjects(userID int) (projects []db.Project, err error) {
	query, args, err := squirrel.Select("p.*").
		From("project as p").
		Join("project__user as pu on pu.project_id=p.id").
		Where("pu.user_id=?", userID).
		OrderBy("p.name").
		ToSql()

	if err != nil {
		return
	}

	_, err = d.sql.Select(&projects, query, args...)

	return
}

func (d *SqlDb) GetProject(projectID int) (project db.Project, err error) {
	query, args, err := squirrel.Select("p.*").
		From("project as p").
		Where("p.id=?", projectID).
		ToSql()

	if err != nil {
		return
	}

	err = d.sql.SelectOne(&project, query, args...)

	return
}

func (d *SqlDb) DeleteProject(projectID int) error {
	tx, err := d.sql.Begin()

	if err != nil {
		return err
	}

	statements := []string{
		"delete from project__template where project_id=?",
		"delete from project__user where project_id=?",
		"delete from project__repository where project_id=?",
		"delete from project__inventory where project_id=?",
		"delete from access_key where project_id=?",
		"delete from project where id=?",
	}

	for _, statement := range statements {
		_, err = tx.Exec(statement, projectID)

		if err != nil {
			err = tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (d *SqlDb) UpdateProject(project db.Project) error {
	_, err := d.sql.Exec(
		"update project set name=?, alert=?, alert_chat=? where id=?",
		project.Name,
		project.Alert,
		project.AlertChat,
		project.ID)
	return err
}

func (d *SqlDb) CreateUser(user db.User) (newUser db.User, err error) {
	pwdHash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 11)

	if err != nil {
		return
	}

	created := db.GetParsedTime(time.Now())

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

	insertID, err := res.LastInsertId()

	if err != nil {
		return
	}

	newUser = user
	newUser.ID = int(insertID)
	newUser.Created = created
	return
}

func (d *SqlDb) DeleteUser(userID int) error {
	res, err := d.sql.Exec("delete from `user` where id=?", userID)

	return validateMutationResult(res, err)
}

func (d *SqlDb) UpdateUser(user db.User) error {
	_, err := d.sql.Exec("update `user` set name=?, username=?, email=?, alert=?, admin=? where id=?",
		user.Name,
		user.Username,
		user.Email,
		user.Alert,
		user.Admin,
		user.ID)

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

func (d *SqlDb) CreateProjectUser(projectUser db.ProjectUser) (newProjectUser db.ProjectUser, err error) {
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

func (d *SqlDb) GetProjectUser(projectID, userID int) (db.ProjectUser, error) {
	var user db.ProjectUser

	err := d.sql.SelectOne(&user,
		"select * from project__user where project_id=? and user_id=?",
		projectID,
		userID)

	if err == sql.ErrNoRows {
		err = db.ErrNotFound
	}

	return user, err
}

func (d *SqlDb) GetProjectUsers(projectID int) ([]db.ProjectUser, error) {
	var users []db.ProjectUser

	err := d.sql.SelectOne(
		&users,
		"select * from project__user where project_id=?",
		projectID)

	if err == sql.ErrNoRows {
		err = db.ErrNotFound
	}

	return users, err
}

func (d *SqlDb) CreateEvent(evt db.Event) (newEvent db.Event, err error) {
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
func (d *SqlDb) GetUser(userID int) (db.User, error) {
	var user db.User

	err := d.sql.SelectOne(&user, "select * from `user` where id=?", userID)

	if err == sql.ErrNoRows {
		err = db.ErrNotFound
	}

	return user, err
}

func getSqlForTable(tableName string, p db.RetrieveQueryParams) (string, []interface{}, error) {
	if p.Count > 0 && p.Offset <= 0 {
		return "", nil, fmt.Errorf("offset cannot be without limit")
	}

	q := squirrel.Select("*").
		From(tableName)

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

func (d *SqlDb) GetUsers(params db.RetrieveQueryParams) (users []db.User, err error) {
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

func (d *SqlDb) CreateAPIToken(token db.APIToken) (db.APIToken, error) {
	token.Created = db.GetParsedTime(time.Now())
	err := d.sql.Insert(&token)
	return token, err
}

func (d *SqlDb) GetAPIToken(tokenID string) (token db.APIToken, err error) {
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

func (d *SqlDb) GetSession(userID int, sessionID int) (session db.Session, err error) {
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

func (d *SqlDb) TouchSession(userID int, sessionID int) error {
	_, err := d.sql.Exec("update session set last_active=? where id=? and user_id=?", time.Now(), sessionID, userID)

	return err
}

func (d *SqlDb) GetAPITokens(userID int) (tokens []db.APIToken, err error) {
	_, err = d.sql.Select(&tokens, "select * from user__token where user_id=?", userID)

	if err == sql.ErrNoRows {
		err = db.ErrNotFound
	}

	return
}

func (d *SqlDb) GetEnvironment(projectID int, environmentID int) (db.Environment, error) {
	var environment db.Environment
	err := d.getObject(projectID, "project__environment", environmentID, &environment)
	return environment, err
}

func (d *SqlDb) GetEnvironments(projectID int, params db.RetrieveQueryParams) ([]db.Environment, error) {
	var environment []db.Environment
	err := d.getObjects(projectID, "project__environment", []string{"name"}, params, &environment)
	return environment, err
}

func (d *SqlDb) UpdateEnvironment(env db.Environment) error {
	_, err := d.sql.Exec(
		"update project__environment set name=?, json=? where id=?",
		env.Name,
		env.JSON,
		env.ID)
	return err
}

func (d *SqlDb) CreateEnvironment(env db.Environment) (newEnv db.Environment, err error) {
	res, err := d.sql.Exec(
		"insert into project__environment (project_id, name, json, password) values (?, ?, ?, ?)",
		env.ProjectID,
		env.Name,
		env.JSON,
		env.Password)

	if err != nil {
		return
	}

	insertID, err := res.LastInsertId()

	if err != nil {
		return
	}

	newEnv = env
	newEnv.ID = int(insertID)
	return
}

func (d *SqlDb) DeleteEnvironment(projectID int, environmentID int) error {
	return d.deleteObject(projectID, "project__environment", "environment_id", environmentID)
}

func (d *SqlDb) DeleteEnvironmentSoft(projectID int, environmentID int) error {
	return d.deleteObjectSoft(projectID, "project__environment", environmentID)
}

func (d *SqlDb) CreateTemplate(template db.Template) (newTemplate db.Template, err error) {
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

func (d *SqlDb) UpdateTemplate(template db.Template) error {
	_, err := d.sql.Exec("update project__template set ssh_key_id=?, inventory_id=?, repository_id=?, environment_id=?, alias=?, playbook=?, arguments=?, override_args=? where id=?",
		template.SSHKeyID,
		template.InventoryID,
		template.RepositoryID,
		template.EnvironmentID,
		template.Alias,
		template.Playbook,
		template.Arguments,
		template.OverrideArguments,
		template.ID)

	return err
}

func (d *SqlDb) GetTemplates(projectID int, params db.RetrieveQueryParams) (templates []db.Template, err error) {
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

func (d *SqlDb) GetTemplate(projectID int, templateID int) (db.Template, error) {
	var template db.Template

	err := d.sql.SelectOne(
		&template,
		"select * from project__template where project_id=? and id=?",
		projectID,
		templateID)

	if err == sql.ErrNoRows {
		err = db.ErrNotFound
	}

	return template, err
}

func (d *SqlDb) DeleteTemplate(projectID int, templateID int) error {
	res, err := d.sql.Exec(
		"delete from project__template where project_id=? and id=?",
		projectID,
		templateID)

	return validateMutationResult(res, err)
}

func (d *SqlDb) GetInventory(projectID int, inventoryID int) (db.Inventory, error) {
	var inventory db.Inventory
	err := d.getObject(projectID, "project__inventory", inventoryID, &inventory)
	return inventory, err
}

func (d *SqlDb) GetInventories(projectID int, params db.RetrieveQueryParams) ([]db.Inventory, error) {
	var inventories []db.Inventory
	err := d.getObjects(projectID, "project__inventory", []string{"name"}, params, &inventories)
	return inventories, err
}

func (d *SqlDb) DeleteInventory(projectID int, inventoryID int) error {
	return d.deleteObject(projectID, "project__inventory", "inventory_id", inventoryID);
}

func (d *SqlDb) DeleteInventorySoft(projectID int, inventoryID int) error {
	return d.deleteObjectSoft(projectID, "project__inventory",  inventoryID)
}


func (d *SqlDb) UpdateInventory(inventory db.Inventory) error {
	_, err := d.sql.Exec(
		"update project__inventory set name=?, type=?, key_id=?, ssh_key_id=?, inventory=? where id=?",
		inventory.Name,
		inventory.Type,
		inventory.KeyID,
		inventory.SSHKeyID,
		inventory.Inventory,
		inventory.ID)

	return err
}

func (d *SqlDb) CreateInventory(inventory db.Inventory) (newInventory db.Inventory, err error) {
	res, err := d.sql.Exec(
		"insert into project__inventory set project_id=?, name=?, type=?, key_id=?, ssh_key_id=?, inventory=?",
		inventory.ProjectID,
		inventory.Name,
		inventory.Type,
		inventory.KeyID,
		inventory.SSHKeyID,
		inventory.Inventory)

	if err != nil {
		return
	}

	insertID, err := res.LastInsertId()
	if err != nil {
		return
	}

	newInventory = inventory
	newInventory.ID = int(insertID)
	return
}



func (d *SqlDb) GetRepository(projectID int, repositoryID int) (db.Repository, error) {
	var repository db.Repository
	err := d.getObject(projectID, "project__repository", repositoryID, &repository)
	return repository, err
}

func (d *SqlDb) GetRepositories(projectID int, params db.RetrieveQueryParams) (repositories []db.Repository, err error) {
	q := squirrel.Select("*").
		From("project__repository pr")

	order := "ASC"
	if params.SortInverted {
		order = "DESC"
	}

	switch params.SortBy {
	case "name", "git_url":
		q = q.Where("pr.project_id=?", projectID).
			OrderBy("pr." + params.SortBy + " " + order)
	case "ssh_key":
		q = q.LeftJoin("access_key ak ON (pr.ssh_key_id = ak.id)").
			Where("pr.project_id=?", projectID).
			OrderBy("ak.name " + order)
	default:
		q = q.Where("pr.project_id=?", projectID).
			OrderBy("pr.name " + order)
	}

	query, args, err := q.ToSql()

	if err != nil {
		return
	}

	_, err = d.sql.Select(&repositories, query, args...)

	return
}

func (d *SqlDb) UpdateRepository(repository db.Repository) error {
	_, err := d.sql.Exec(
		"update project__repository set name=?, git_url=?, ssh_key_id=? where id=?",
		repository.Name,
		repository.GitURL,
		repository.SSHKeyID,
		repository.ID)

	return err
}

func (d *SqlDb) CreateRepository(repository db.Repository) (newRepo db.Repository, err error) {
	res, err := d.sql.Exec(
		"insert into project__repository(project_id, git_url, ssh_key_id, name) values (?, ?, ?, ?)",
		repository.ProjectID,
		repository.GitURL,
		repository.SSHKeyID,
		repository.Name)

	if err != nil {
		return
	}

	insertID, err := res.LastInsertId()
	if err != nil {
		return
	}

	newRepo = repository
	newRepo.ID = int(insertID)
	return
}

func (d *SqlDb) DeleteRepository(projectID int, repositoryId int) error {
	return d.deleteObject(projectID, "project__repository", "repository_id", repositoryId)
}

func (d *SqlDb) DeleteRepositorySoft(projectID int, repositoryId int) error {
	return d.deleteObjectSoft(projectID, "project__repository", repositoryId)
}


func (d *SqlDb) GetAccessKey(projectID int, accessKeyID int) (db.AccessKey, error) {
	var key db.AccessKey
	err := d.getObject(projectID, "access_key", accessKeyID, &key)
	return key, err
}

func (d *SqlDb) GetAccessKeys(projectID int, params db.RetrieveQueryParams) ([]db.AccessKey, error) {
	var keys []db.AccessKey
	err := d.getObjects(projectID, "access_key", []string{"name", "type"}, params, &keys)
	return keys, err
}

func (d *SqlDb) UpdateAccessKey(key db.AccessKey) error {
	res, err := d.sql.Exec(
		"update access_key set name=?, type=?, `key`=?, secret=? where id=?",
		key.Name,
		key.Type,
		key.Key,
		key.Secret,
		key.ID)

	return validateMutationResult(res, err)
}

func (d *SqlDb) CreateAccessKey(key db.AccessKey) (newKey db.AccessKey, err error) {
	res, err := d.sql.Exec(
		"insert into access_key (name, type, project_id, `key`, secret) values (?, ?, ?, ?, ?)",
		key.Name,
		key.Type,
		key.ProjectID,
		key.Key,
		key.Secret)

	if err != nil {
		return
	}

	insertID, err := res.LastInsertId()
	if err != nil {
		return
	}

	newKey = key
	newKey.ID = int(insertID)
	return
}

func (d *SqlDb) DeleteAccessKey(projectID int, accessKeyID int) error {
	return d.deleteObject(projectID, "access_key", "ssh_key_id", accessKeyID)
}

func (d *SqlDb) DeleteAccessKeySoft(projectID int, accessKeyID int) error {
	return d.deleteObjectSoft(projectID, "access_key", accessKeyID)
}
