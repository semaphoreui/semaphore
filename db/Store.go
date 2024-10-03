package db

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
	"reflect"
	"strings"
	"time"
)

const databaseTimeFormat = "2006-01-02T15:04:05:99Z"

// GetParsedTime returns the timestamp as it will retrieved from the database
// This allows us to create timestamp consistency on return values from create requests
func GetParsedTime(t time.Time) time.Time {
	parsedTime, err := time.Parse(databaseTimeFormat, t.Format(databaseTimeFormat))
	if err != nil {
		log.Error(err)
	}
	return parsedTime
}

func ObjectToJSON(obj interface{}) *string {
	if obj == nil || (reflect.ValueOf(obj).Kind() == reflect.Ptr && reflect.ValueOf(obj).IsNil()) {
		return nil
	}
	bytes, err := json.Marshal(obj)
	if err != nil {
		return nil
	}
	str := string(bytes)
	return &str
}

type RetrieveQueryParams struct {
	Offset       int
	Count        int
	SortBy       string
	SortInverted bool
	Filter       string
}

type ObjectReferrer struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ObjectReferrers struct {
	Templates    []ObjectReferrer `json:"templates"`
	Inventories  []ObjectReferrer `json:"inventories"`
	Repositories []ObjectReferrer `json:"repositories"`
}

type IntegrationReferrers struct {
	IntegrationMatchers      []ObjectReferrer `json:"matchers"`
	IntegrationExtractValues []ObjectReferrer `json:"values"`
}

type IntegrationExtractorChildReferrers struct {
	Integrations []ObjectReferrer `json:"integrations"`
}

// ObjectProps describe database entities.
// It mainly used for NoSQL implementations (currently BoltDB) to preserve same
// data structure of different implementations and easy change it if required.
type ObjectProps struct {
	TableName             string
	Type                  reflect.Type // to which type the table bust be mapped.
	IsGlobal              bool         // doesn't belong to other table, for example to project or user.
	ReferringColumnSuffix string
	PrimaryColumnName     string
	SortableColumns       []string
	DefaultSortingColumn  string
	SortInverted          bool // sort from high to low object ID by default. It is useful for some NoSQL implementations.
}

var ErrNotFound = errors.New("no rows in result set")
var ErrInvalidOperation = errors.New("invalid operation")

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

type Store interface {
	// Connect connects to the database.
	// Token parameter used if PermanentConnection returns false.
	// Token used for debugging of session connections.
	Connect(token string)
	Close(token string)

	// PermanentConnection returns true if connection should be kept from start to finish of the app.
	// This mode is suitable for MySQL and Postgres but not for BoltDB.
	// For BoltDB we should reconnect for each request because BoltDB support only one connection at time.
	PermanentConnection() bool

	// IsInitialized indicates is database already initialized, or it is empty.
	// The method is useful for creating required entities in database during first run.
	IsInitialized() (bool, error)
	// IsMigrationApplied queries the database to see if a migration table with
	// this version id exists already
	IsMigrationApplied(version Migration) (bool, error)
	// ApplyMigration runs executes a database migration
	ApplyMigration(version Migration) error
	// TryRollbackMigration attempts to roll back the database to an earlier version
	// if a rollback exists
	TryRollbackMigration(version Migration)

	GetOptions(params RetrieveQueryParams) (map[string]string, error)
	GetOption(key string) (string, error)
	SetOption(key string, value string) error
	DeleteOption(key string) error
	DeleteOptions(filter string) error

	GetEnvironment(projectID int, environmentID int) (Environment, error)
	GetEnvironmentRefs(projectID int, environmentID int) (ObjectReferrers, error)
	GetEnvironments(projectID int, params RetrieveQueryParams) ([]Environment, error)
	UpdateEnvironment(env Environment) error
	CreateEnvironment(env Environment) (Environment, error)
	DeleteEnvironment(projectID int, templateID int) error
	GetEnvironmentSecrets(projectID int, environmentID int) ([]AccessKey, error)

	GetInventory(projectID int, inventoryID int) (Inventory, error)
	GetInventoryRefs(projectID int, inventoryID int) (ObjectReferrers, error)
	GetInventories(projectID int, params RetrieveQueryParams) ([]Inventory, error)
	UpdateInventory(inventory Inventory) error
	CreateInventory(inventory Inventory) (Inventory, error)
	DeleteInventory(projectID int, inventoryID int) error

	GetRepository(projectID int, repositoryID int) (Repository, error)
	GetRepositoryRefs(projectID int, repositoryID int) (ObjectReferrers, error)
	GetRepositories(projectID int, params RetrieveQueryParams) ([]Repository, error)
	UpdateRepository(repository Repository) error
	CreateRepository(repository Repository) (Repository, error)
	DeleteRepository(projectID int, repositoryID int) error

	GetAccessKey(projectID int, accessKeyID int) (AccessKey, error)
	GetAccessKeyRefs(projectID int, accessKeyID int) (ObjectReferrers, error)
	GetAccessKeys(projectID int, params RetrieveQueryParams) ([]AccessKey, error)
	RekeyAccessKeys(oldKey string) error

	CreateIntegration(integration Integration) (newIntegration Integration, err error)
	GetIntegrations(projectID int, params RetrieveQueryParams) ([]Integration, error)
	GetIntegration(projectID int, integrationID int) (integration Integration, err error)
	UpdateIntegration(integration Integration) error
	GetIntegrationRefs(projectID int, integrationID int) (IntegrationReferrers, error)
	DeleteIntegration(projectID int, integrationID int) error

	CreateIntegrationExtractValue(projectId int, value IntegrationExtractValue) (newValue IntegrationExtractValue, err error)
	GetIntegrationExtractValues(projectID int, params RetrieveQueryParams, integrationID int) ([]IntegrationExtractValue, error)
	GetIntegrationExtractValue(projectID int, valueID int, integrationID int) (value IntegrationExtractValue, err error)
	UpdateIntegrationExtractValue(projectID int, integrationExtractValue IntegrationExtractValue) error
	GetIntegrationExtractValueRefs(projectID int, valueID int, integrationID int) (IntegrationExtractorChildReferrers, error)
	DeleteIntegrationExtractValue(projectID int, valueID int, integrationID int) error

	CreateIntegrationMatcher(projectID int, matcher IntegrationMatcher) (newMatcher IntegrationMatcher, err error)
	GetIntegrationMatchers(projectID int, params RetrieveQueryParams, integrationID int) ([]IntegrationMatcher, error)
	GetIntegrationMatcher(projectID int, matcherID int, integrationID int) (matcher IntegrationMatcher, err error)
	UpdateIntegrationMatcher(projectID int, integrationMatcher IntegrationMatcher) error
	GetIntegrationMatcherRefs(projectID int, matcherID int, integrationID int) (IntegrationExtractorChildReferrers, error)
	DeleteIntegrationMatcher(projectID int, matcherID int, integrationID int) error

	CreateIntegrationAlias(alias IntegrationAlias) (IntegrationAlias, error)
	GetIntegrationAliases(projectID int, integrationID *int) ([]IntegrationAlias, error)
	GetIntegrationsByAlias(alias string) ([]Integration, error)
	DeleteIntegrationAlias(projectID int, aliasID int) error
	GetAllSearchableIntegrations() ([]Integration, error)

	UpdateAccessKey(accessKey AccessKey) error
	CreateAccessKey(accessKey AccessKey) (AccessKey, error)
	DeleteAccessKey(projectID int, accessKeyID int) error

	GetUserCount() (int, error)
	GetUsers(params RetrieveQueryParams) ([]User, error)
	CreateUserWithoutPassword(user User) (User, error)
	CreateUser(user UserWithPwd) (User, error)
	DeleteUser(userID int) error

	// UpdateUser updates all fields of the entity except Pwd.
	// Pwd should be present of you want update user password. Empty Pwd ignored.
	UpdateUser(user UserWithPwd) error
	SetUserPassword(userID int, password string) error
	GetUser(userID int) (User, error)
	GetUserByLoginOrEmail(login string, email string) (User, error)

	GetProject(projectID int) (Project, error)
	GetAllProjects() ([]Project, error)
	GetProjects(userID int) ([]Project, error)
	CreateProject(project Project) (Project, error)
	DeleteProject(projectID int) error
	UpdateProject(project Project) error

	GetTemplates(projectID int, filter TemplateFilter, params RetrieveQueryParams) ([]Template, error)
	GetTemplateRefs(projectID int, templateID int) (ObjectReferrers, error)
	CreateTemplate(template Template) (Template, error)
	UpdateTemplate(template Template) error
	GetTemplate(projectID int, templateID int) (Template, error)
	DeleteTemplate(projectID int, templateID int) error

	GetSchedules() ([]Schedule, error)
	GetProjectSchedules(projectID int) ([]ScheduleWithTpl, error)
	GetTemplateSchedules(projectID int, templateID int) ([]Schedule, error)
	CreateSchedule(schedule Schedule) (Schedule, error)
	UpdateSchedule(schedule Schedule) error
	SetScheduleCommitHash(projectID int, scheduleID int, hash string) error
	SetScheduleActive(projectID int, scheduleID int, active bool) error
	GetSchedule(projectID int, scheduleID int) (Schedule, error)
	DeleteSchedule(projectID int, scheduleID int) error

	GetAllAdmins() ([]User, error)
	GetProjectUsers(projectID int, params RetrieveQueryParams) ([]UserWithProjectRole, error)
	CreateProjectUser(projectUser ProjectUser) (ProjectUser, error)
	DeleteProjectUser(projectID int, userID int) error
	GetProjectUser(projectID int, userID int) (ProjectUser, error)
	UpdateProjectUser(projectUser ProjectUser) error

	CreateEvent(event Event) (Event, error)
	GetUserEvents(userID int, params RetrieveQueryParams) ([]Event, error)
	GetEvents(projectID int, params RetrieveQueryParams) ([]Event, error)

	GetAPITokens(userID int) ([]APIToken, error)
	CreateAPIToken(token APIToken) (APIToken, error)
	GetAPIToken(tokenID string) (APIToken, error)
	ExpireAPIToken(userID int, tokenID string) error
	DeleteAPIToken(userID int, tokenID string) error

	GetSession(userID int, sessionID int) (Session, error)
	CreateSession(session Session) (Session, error)
	ExpireSession(userID int, sessionID int) error
	TouchSession(userID int, sessionID int) error

	CreateTask(task Task, maxTasks int) (Task, error)
	UpdateTask(task Task) error

	GetTemplateTasks(projectID int, templateID int, params RetrieveQueryParams) ([]TaskWithTpl, error)
	GetProjectTasks(projectID int, params RetrieveQueryParams) ([]TaskWithTpl, error)
	GetTask(projectID int, taskID int) (Task, error)
	DeleteTaskWithOutputs(projectID int, taskID int) error
	GetTaskOutputs(projectID int, taskID int) ([]TaskOutput, error)
	CreateTaskOutput(output TaskOutput) (TaskOutput, error)
	GetTaskStages(projectID int, taskID int) ([]TaskStage, error)
	CreateTaskStage(stage TaskStage) (TaskStage, error)

	GetView(projectID int, viewID int) (View, error)
	GetViews(projectID int) ([]View, error)
	UpdateView(view View) error
	CreateView(view View) (View, error)
	DeleteView(projectID int, viewID int) error
	SetViewPositions(projectID int, viewPositions map[int]int) error

	GetRunner(projectID int, runnerID int) (Runner, error)
	GetRunners(projectID int) ([]Runner, error)
	DeleteRunner(projectID int, runnerID int) error
	GetGlobalRunnerByToken(token string) (Runner, error)
	GetGlobalRunner(runnerID int) (Runner, error)
	GetGlobalRunners(activeOnly bool) ([]Runner, error)
	DeleteGlobalRunner(runnerID int) error
	UpdateRunner(runner Runner) error
	CreateRunner(runner Runner) (Runner, error)

	GetTemplateVaults(projectID int, templateID int) ([]TemplateVault, error)
	CreateTemplateVault(vault TemplateVault) (TemplateVault, error)
	UpdateTemplateVaults(projectID int, templateID int, vaults []TemplateVault) error
}

var AccessKeyProps = ObjectProps{
	TableName:             "access_key",
	Type:                  reflect.TypeOf(AccessKey{}),
	PrimaryColumnName:     "id",
	ReferringColumnSuffix: "key_id",
	SortableColumns:       []string{"name", "type"},
	DefaultSortingColumn:  "name",
}

var IntegrationProps = ObjectProps{
	TableName:             "project__integration",
	Type:                  reflect.TypeOf(Integration{}),
	PrimaryColumnName:     "id",
	ReferringColumnSuffix: "integration_id",
	SortableColumns:       []string{"name"},
	DefaultSortingColumn:  "name",
}

var IntegrationExtractValueProps = ObjectProps{
	TableName:            "project__integration_extract_value",
	Type:                 reflect.TypeOf(IntegrationExtractValue{}),
	PrimaryColumnName:    "id",
	SortableColumns:      []string{"name"},
	DefaultSortingColumn: "name",
}

var IntegrationMatcherProps = ObjectProps{
	TableName:            "project__integration_matcher",
	Type:                 reflect.TypeOf(IntegrationMatcher{}),
	PrimaryColumnName:    "id",
	SortableColumns:      []string{"name"},
	DefaultSortingColumn: "name",
}

var IntegrationAliasProps = ObjectProps{
	TableName:         "project__integration_alias",
	Type:              reflect.TypeOf(IntegrationAlias{}),
	PrimaryColumnName: "id",
}

var EnvironmentProps = ObjectProps{
	TableName:             "project__environment",
	Type:                  reflect.TypeOf(Environment{}),
	PrimaryColumnName:     "id",
	ReferringColumnSuffix: "environment_id",
	SortableColumns:       []string{"name"},
	DefaultSortingColumn:  "name",
}

var InventoryProps = ObjectProps{
	TableName:             "project__inventory",
	Type:                  reflect.TypeOf(Inventory{}),
	PrimaryColumnName:     "id",
	ReferringColumnSuffix: "inventory_id",
	SortableColumns:       []string{"name"},
	DefaultSortingColumn:  "name",
}

var RepositoryProps = ObjectProps{
	TableName:             "project__repository",
	Type:                  reflect.TypeOf(Repository{}),
	PrimaryColumnName:     "id",
	ReferringColumnSuffix: "repository_id",
	DefaultSortingColumn:  "name",
}

var TemplateProps = ObjectProps{
	TableName:             "project__template",
	Type:                  reflect.TypeOf(Template{}),
	PrimaryColumnName:     "id",
	ReferringColumnSuffix: "template_id",
	SortableColumns:       []string{"name"},
	DefaultSortingColumn:  "name",
}

var ScheduleProps = ObjectProps{
	TableName:         "project__schedule",
	Type:              reflect.TypeOf(Schedule{}),
	PrimaryColumnName: "id",
}

var ProjectUserProps = ObjectProps{
	TableName:         "project__user",
	Type:              reflect.TypeOf(ProjectUser{}),
	PrimaryColumnName: "user_id",
}

var ProjectProps = ObjectProps{
	TableName:             "project",
	Type:                  reflect.TypeOf(Project{}),
	PrimaryColumnName:     "id",
	ReferringColumnSuffix: "project_id",
	DefaultSortingColumn:  "name",
	IsGlobal:              true,
}

var UserProps = ObjectProps{
	TableName:         "user",
	Type:              reflect.TypeOf(User{}),
	PrimaryColumnName: "id",
	IsGlobal:          true,
}

var SessionProps = ObjectProps{
	TableName:         "session",
	Type:              reflect.TypeOf(Session{}),
	PrimaryColumnName: "id",
}

var TokenProps = ObjectProps{
	TableName:         "user__token",
	Type:              reflect.TypeOf(APIToken{}),
	PrimaryColumnName: "id",
}

var TaskProps = ObjectProps{
	TableName:         "task",
	Type:              reflect.TypeOf(Task{}),
	PrimaryColumnName: "id",
	IsGlobal:          true,
	SortInverted:      true,
}

var TaskOutputProps = ObjectProps{
	TableName: "task__output",
	Type:      reflect.TypeOf(TaskOutput{}),
}

var TaskStageProps = ObjectProps{
	TableName: "task__stage",
	Type:      reflect.TypeOf(TaskStage{}),
}

var ViewProps = ObjectProps{
	TableName:            "project__view",
	Type:                 reflect.TypeOf(View{}),
	PrimaryColumnName:    "id",
	DefaultSortingColumn: "position",
}

var GlobalRunnerProps = ObjectProps{
	TableName:         "runner",
	Type:              reflect.TypeOf(Runner{}),
	PrimaryColumnName: "id",
	IsGlobal:          true,
}

var OptionProps = ObjectProps{
	TableName:         "option",
	Type:              reflect.TypeOf(Option{}),
	PrimaryColumnName: "key",
	IsGlobal:          true,
}

var TemplateVaultProps = ObjectProps{
	TableName:             "project__template_vault",
	Type:                  reflect.TypeOf(TemplateVault{}),
	PrimaryColumnName:     "id",
	ReferringColumnSuffix: "template_id",
}

func (p ObjectProps) GetReferringFieldsFrom(t reflect.Type) (fields []string, err error) {
	n := t.NumField()
	for i := 0; i < n; i++ {
		if !strings.HasSuffix(t.Field(i).Tag.Get("db"), p.ReferringColumnSuffix) {
			continue
		}
		fields = append(fields, t.Field(i).Tag.Get("db"))
	}

	for i := 0; i < n; i++ {
		if t.Field(i).Tag != "" || t.Field(i).Type.Kind() != reflect.Struct {
			continue
		}
		var nested []string
		nested, err = p.GetReferringFieldsFrom(t.Field(i).Type)
		if err != nil {
			return
		}
		fields = append(fields, nested...)
	}

	return
}

func StoreSession(store Store, token string, callback func()) {
	if !store.PermanentConnection() {
		store.Connect(token)
	}

	callback()

	if !store.PermanentConnection() {
		store.Close(token)
	}
}

func ValidateRepository(store Store, repo *Repository) (err error) {
	_, err = store.GetAccessKey(repo.ProjectID, repo.SSHKeyID)

	return
}

func ValidateInventory(store Store, inventory *Inventory) (err error) {
	if inventory.SSHKeyID != nil {
		_, err = store.GetAccessKey(inventory.ProjectID, *inventory.SSHKeyID)
	}

	if err != nil {
		return
	}

	if inventory.BecomeKeyID != nil {
		_, err = store.GetAccessKey(inventory.ProjectID, *inventory.BecomeKeyID)
	}

	if err != nil {
		return
	}

	if inventory.HolderID != nil {
		_, err = store.GetTemplate(inventory.ProjectID, *inventory.HolderID)
	}

	return
}

type MapStringAnyField map[string]interface{}

func (m *MapStringAnyField) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, m)
	case string:
		return json.Unmarshal([]byte(v), m)
	default:
		return errors.New("unsupported type for MapStringAnyField")
	}
}

// Value implements the driver.Valuer interface for MapStringAnyField
func (m MapStringAnyField) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}
