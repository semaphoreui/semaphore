package db

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"time"

	"github.com/semaphoreui/semaphore/pkg/task_logger"

	log "github.com/sirupsen/logrus"
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

func ObjectToJSON(obj any) *string {
	if obj == nil ||
		(reflect.ValueOf(obj).Kind() == reflect.Ptr && reflect.ValueOf(obj).IsNil()) ||
		(reflect.ValueOf(obj).Kind() == reflect.Slice && reflect.ValueOf(obj).IsZero()) {
		return nil
	}
	bytes, err := json.Marshal(obj)
	if err != nil {
		return nil
	}
	str := string(bytes)
	return &str
}

type OwnershipFilter struct {
	WithoutOwnerOnly bool
	TemplateID       *int
	EnvironmentID    *int
}

type RetrieveQueryParams struct {
	Offset       int
	Count        int
	SortBy       string
	SortInverted bool
	Filter       string
	Ownership    OwnershipFilter
}

type ObjectReferrer struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ObjectReferrers struct {
	Templates    []ObjectReferrer `json:"templates"`
	Inventories  []ObjectReferrer `json:"inventories"`
	Repositories []ObjectReferrer `json:"repositories"`
	Integrations []ObjectReferrer `json:"integrations"`
	Schedules    []ObjectReferrer `json:"schedules"`
	AccessKeys   []ObjectReferrer `json:"access_keys"`
}

type IntegrationReferrers struct {
	IntegrationMatchers      []ObjectReferrer `json:"matchers"`
	IntegrationExtractValues []ObjectReferrer `json:"values"`
}

type IntegrationExtractorChildReferrers struct {
	Integrations []ObjectReferrer `json:"integrations"`
}

func containsStr(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func (p *RetrieveQueryParams) Validate(props ObjectProps) (res RetrieveQueryParams, err error) {

	if p.Offset > 0 && p.Count <= 0 {
		err = &ValidationError{"offset cannot be without limit"}
		return
	}

	if p.Count < 0 {
		err = &ValidationError{"count must be positive"}
		return
	}

	if p.Offset < 0 {
		err = &ValidationError{"offset must be positive"}
		return
	}

	if p.SortBy != "" {
		if !containsStr(props.SortableColumns, p.SortBy) {
			err = &ValidationError{"invalid sort column"}
			return
		}
	}

	res = *p
	return
}

func (f *OwnershipFilter) GetOwnerID(ownership ObjectProps) *int {
	switch ownership.ReferringColumnSuffix {
	case "template_id":
		return f.TemplateID
	case "environment_id":
		return f.EnvironmentID
	default:
		return nil
	}
}

func (f *OwnershipFilter) SetOwnerID(ownership ObjectProps, ownerID int) {
	switch ownership.ReferringColumnSuffix {
	case "template_id":
		f.TemplateID = &ownerID
	case "environment_id":
		f.EnvironmentID = &ownerID
	}
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
	Ownerships            []*ObjectProps
	SelectColumns         []string
}

var ErrNotFound = errors.New("no rows in result set")
var ErrInvalidOperation = errors.New("invalid operation")

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

type TaskStatUnit string

const TaskStatUnitDay TaskStatUnit = "day"
const TaskStatUnitWeek TaskStatUnit = "week"
const TaskStatUnitMonth TaskStatUnit = "month"

type TaskFilter struct {
	Start  *time.Time `json:"start"`
	End    *time.Time `json:"end"`
	UserID *int       `json:"user_id"`
}

type TaskStat struct {
	Date          string                         `json:"date"`
	CountByStatus map[task_logger.TaskStatus]int `json:"count_by_status"`
	AvgDuration   int                            `json:"avg_duration"`
}

// ConnectionManager handles database connection lifecycle
type ConnectionManager interface {
	// Connect connects to the database.
	// Token parameter used if PermanentConnection returns false.
	// Token used for debugging of session connections.
	Connect(token string)
	Close(token string)

	// PermanentConnection returns true if connection should be kept from start to finish of the app.
	// This mode is suitable for MySQL and Postgres but not for BoltDB.
	// For BoltDB we should reconnect for each request because BoltDB support only one connection at time.
	PermanentConnection() bool
}

// MigrationManager handles database migrations
type MigrationManager interface {
	GetDialect() string
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
}

// OptionsManager handles system options
type OptionsManager interface {
	GetOptions(params RetrieveQueryParams) (map[string]string, error)
	GetOption(key string) (string, error)
	SetOption(key string, value string) error
	DeleteOption(key string) error
	DeleteOptions(filter string) error
}

// UserManager handles user-related operations
type UserManager interface {
	GetProUserCount() (int, error)
	GetUserCount() (int, error)
	GetUsers(params RetrieveQueryParams) ([]User, error)
	CreateUserWithoutPassword(user User) (User, error)
	CreateUser(user UserWithPwd) (User, error)
	DeleteUser(userID int) error
	UpdateUser(user UserWithPwd) error
	SetUserPassword(userID int, password string) error
	AddTotpVerification(userID int, url string, recoveryHash string) (UserTotp, error)
	DeleteTotpVerification(userID int, totpID int) error
	AddEmailOtpVerification(userID int, code string) (UserEmailOtp, error)
	DeleteEmailOtpVerification(userID int, totpID int) error
	GetUser(userID int) (User, error)
	GetUserByLoginOrEmail(login string, email string) (User, error)
	GetAllAdmins() ([]User, error)
}

// ProjectStore handles project-related operations
type ProjectStore interface {
	GetProject(projectID int) (Project, error)
	GetAllProjects() ([]Project, error)
	GetProjects(userID int) ([]Project, error)
	CreateProject(project Project) (Project, error)
	DeleteProject(projectID int) error
	UpdateProject(project Project) error
	GetProjectUsers(projectID int, params RetrieveQueryParams) ([]UserWithProjectRole, error)
	CreateProjectUser(projectUser ProjectUser) (ProjectUser, error)
	DeleteProjectUser(projectID int, userID int) error
	GetProjectUser(projectID int, userID int) (ProjectUser, error)
	UpdateProjectUser(projectUser ProjectUser) error
}

type ProjectInviteRepository interface {
	// Project invites
	GetProjectInvites(projectID int, params RetrieveQueryParams) ([]ProjectInviteWithUser, error)
	CreateProjectInvite(invite ProjectInvite) (ProjectInvite, error)
	GetProjectInvite(projectID int, inviteID int) (ProjectInvite, error)
	GetProjectInviteByToken(token string) (ProjectInvite, error)
	UpdateProjectInvite(invite ProjectInvite) error
	DeleteProjectInvite(projectID int, inviteID int) error
}

// TemplateManager handles template-related operations
type TemplateManager interface {
	GetTemplates(projectID int, filter TemplateFilter, params RetrieveQueryParams) ([]Template, error)
	GetTemplateRefs(projectID int, templateID int) (ObjectReferrers, error)
	CreateTemplate(template Template) (Template, error)
	UpdateTemplate(template Template) error
	GetTemplate(projectID int, templateID int) (Template, error)
	DeleteTemplate(projectID int, templateID int) error
	SetTemplateDescription(projectID int, templateID int, description string) error
	GetTemplateVaults(projectID int, templateID int) ([]TemplateVault, error)
	CreateTemplateVault(vault TemplateVault) (TemplateVault, error)
	UpdateTemplateVaults(projectID int, templateID int, vaults []TemplateVault) error
}

// InventoryManager handles inventory-related operations
type InventoryManager interface {
	GetInventory(projectID int, inventoryID int) (Inventory, error)
	GetInventoryRefs(projectID int, inventoryID int) (ObjectReferrers, error)
	GetInventories(projectID int, params RetrieveQueryParams, types []InventoryType) ([]Inventory, error)
	UpdateInventory(inventory Inventory) error
	CreateInventory(inventory Inventory) (Inventory, error)
	DeleteInventory(projectID int, inventoryID int) error
}

// RepositoryManager handles repository-related operations
type RepositoryManager interface {
	GetRepository(projectID int, repositoryID int) (Repository, error)
	GetRepositoryRefs(projectID int, repositoryID int) (ObjectReferrers, error)
	GetRepositories(projectID int, params RetrieveQueryParams) ([]Repository, error)
	UpdateRepository(repository Repository) error
	CreateRepository(repository Repository) (Repository, error)
	DeleteRepository(projectID int, repositoryID int) error
}

// EnvironmentManager handles environment-related operations
type EnvironmentManager interface {
	GetEnvironment(projectID int, environmentID int) (Environment, error)
	GetEnvironmentRefs(projectID int, environmentID int) (ObjectReferrers, error)
	GetEnvironments(projectID int, params RetrieveQueryParams) ([]Environment, error)
	UpdateEnvironment(env Environment) error
	CreateEnvironment(env Environment) (Environment, error)
	DeleteEnvironment(projectID int, templateID int) error
	GetEnvironmentSecrets(projectID int, environmentID int) ([]AccessKey, error)
}

type GetAccessKeyOptions struct {
	Owner         AccessKeyOwner
	IgnoreOwner   bool
	EnvironmentID *int
	StorageID     *int
}

// AccessKeyManager handles access key-related operations
type AccessKeyManager interface {
	GetAccessKey(projectID int, accessKeyID int) (AccessKey, error)
	GetAccessKeyRefs(projectID int, accessKeyID int) (ObjectReferrers, error)
	GetAccessKeys(projectID int, options GetAccessKeyOptions, params RetrieveQueryParams) ([]AccessKey, error)
	RekeyAccessKeys(oldKey string) error
	UpdateAccessKey(accessKey AccessKey) error
	CreateAccessKey(accessKey AccessKey) (AccessKey, error)
	DeleteAccessKey(projectID int, accessKeyID int) error
}

// IntegrationManager handles integration-related operations
type IntegrationManager interface {
	CreateIntegration(integration Integration) (newIntegration Integration, err error)
	GetIntegrations(projectID int, params RetrieveQueryParams, includeTaskParams bool) ([]Integration, error)
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
	GetIntegrationsByAlias(alias string) ([]Integration, IntegrationAliasLevel, error)
	DeleteIntegrationAlias(projectID int, aliasID int) error
}

// SessionManager handles session-related operations
type SessionManager interface {
	GetSession(userID int, sessionID int) (Session, error)
	CreateSession(session Session) (Session, error)
	ExpireSession(userID int, sessionID int) error
	TouchSession(userID int, sessionID int) error
	SetSessionVerificationMethod(userID int, sessionID int, verificationMethod SessionVerificationMethod) error
	VerifySession(userID int, sessionID int) error
}

// TokenManager handles token-related operations
type TokenManager interface {
	GetAPITokens(userID int) ([]APIToken, error)
	CreateAPIToken(token APIToken) (APIToken, error)
	GetAPIToken(tokenID string) (APIToken, error)
	ExpireAPIToken(userID int, tokenID string) error
	DeleteAPIToken(userID int, tokenID string) error
}

// TaskManager handles task-related operations
type TaskManager interface {
	CreateTask(task Task, maxTasks int) (Task, error)
	UpdateTask(task Task) error
	GetTemplateTasks(projectID int, templateID int, params RetrieveQueryParams) ([]TaskWithTpl, error)
	GetProjectTasks(projectID int, params RetrieveQueryParams) ([]TaskWithTpl, error)
	GetTask(projectID int, taskID int) (Task, error)
	DeleteTaskWithOutputs(projectID int, taskID int) error
	GetTaskOutputs(projectID int, taskID int, params RetrieveQueryParams) ([]TaskOutput, error)
	CreateTaskOutput(output TaskOutput) (TaskOutput, error)
	CreateTaskStage(stage TaskStage) (TaskStage, error)
	EndTaskStage(taskID int, stageID int, end time.Time, endOutputID int) error
	CreateTaskStageResult(taskID int, stageID int, result map[string]any) error
	GetTaskStages(projectID int, taskID int) ([]TaskStageWithResult, error)
	GetTaskStageResult(projectID int, taskID int, stageID int) (TaskStageResult, error)
	GetTaskStageOutputs(projectID int, taskID int, stageID int) ([]TaskOutput, error)
	GetTaskStats(projectID int, templateID *int, unit TaskStatUnit, filter TaskFilter) ([]TaskStat, error)
}

type AnsibleTaskRepository interface {
	CreateAnsibleTaskHost(host AnsibleTaskHost) error
	CreateAnsibleTaskError(error AnsibleTaskError) error
	GetAnsibleTaskHosts(projectID int, taskID int) ([]AnsibleTaskHost, error)
	GetAnsibleTaskErrors(projectID int, taskID int) ([]AnsibleTaskError, error)
}

// ScheduleManager handles schedule-related operations
type ScheduleManager interface {
	GetSchedules() ([]Schedule, error)
	GetProjectSchedules(projectID int, includeTaskParams bool) ([]ScheduleWithTpl, error)
	GetTemplateSchedules(projectID int, templateID int, onlyCommitCheckers bool) ([]Schedule, error)
	CreateSchedule(schedule Schedule) (Schedule, error)
	UpdateSchedule(schedule Schedule) error
	SetScheduleCommitHash(projectID int, scheduleID int, hash string) error
	SetScheduleActive(projectID int, scheduleID int, active bool) error
	GetSchedule(projectID int, scheduleID int) (Schedule, error)
	DeleteSchedule(projectID int, scheduleID int) error
}

// ViewManager handles view-related operations
type ViewManager interface {
	GetView(projectID int, viewID int) (View, error)
	GetViews(projectID int) ([]View, error)
	UpdateView(view View) error
	CreateView(view View) (View, error)
	DeleteView(projectID int, viewID int) error
	SetViewPositions(projectID int, viewPositions map[int]int) error
}

// RunnerManager handles runner-related operations
type RunnerManager interface {
	GetRunner(projectID int, runnerID int) (Runner, error)
	GetRunners(projectID int, activeOnly bool, tag *string) ([]Runner, error)
	DeleteRunner(projectID int, runnerID int) error
	GetRunnerByToken(token string) (Runner, error)
	GetGlobalRunner(runnerID int) (Runner, error)
	GetAllRunners(activeOnly bool, globalOnly bool) ([]Runner, error)
	DeleteGlobalRunner(runnerID int) error
	UpdateRunner(runner Runner) error
	CreateRunner(runner Runner) (Runner, error)
	TouchRunner(runner Runner) (err error)
	ClearRunnerCache(runner Runner) (err error)
	GetRunnerTags(projectID int) ([]RunnerTag, error)
}

// EventManager handles event-related operations
type EventManager interface {
	CreateEvent(event Event) (Event, error)
	GetUserEvents(userID int, params RetrieveQueryParams) ([]Event, error)
	GetEvents(projectID int, params RetrieveQueryParams) ([]Event, error)
}

type SecretStorageRepository interface {
	GetSecretStorages(projectID int) ([]SecretStorage, error)
	CreateSecretStorage(storage SecretStorage) (SecretStorage, error)
	GetSecretStorage(projectID int, storageID int) (SecretStorage, error)
	UpdateSecretStorage(storage SecretStorage) error
	GetSecretStorageRefs(projectID int, storageID int) (ObjectReferrers, error)
	DeleteSecretStorage(projectID int, storageID int) error
}

// Store is the main interface that aggregates all specialized interfaces
type Store interface {
	ConnectionManager
	MigrationManager
	OptionsManager
	UserManager
	ProjectStore
	ProjectInviteRepository
	TemplateManager
	InventoryManager
	RepositoryManager
	EnvironmentManager
	AccessKeyManager
	IntegrationManager
	SessionManager
	TokenManager
	TaskManager
	ScheduleManager
	ViewManager
	RunnerManager
	EventManager
	SecretStorageRepository
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

var TaskParamsProps = ObjectProps{
	TableName:             "project__task_params",
	Type:                  reflect.TypeOf(TaskParams{}),
	PrimaryColumnName:     "id",
	ReferringColumnSuffix: "params_id",
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
	Ownerships:            []*ObjectProps{&TemplateProps},
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
	SortableColumns:       []string{"name", "playbook", "inventory", "environment", "repository"},
	DefaultSortingColumn:  "name",
}

var ProjectUserProps = ObjectProps{
	TableName:         "project__user",
	Type:              reflect.TypeOf(ProjectUser{}),
	PrimaryColumnName: "user_id",
}

var ProjectInviteProps = ObjectProps{
	TableName:             "project__invite",
	Type:                  reflect.TypeOf(ProjectInvite{}),
	PrimaryColumnName:     "id",
	ReferringColumnSuffix: "invite_id",
	SortableColumns:       []string{"created", "status", "role"},
	DefaultSortingColumn:  "created",
}

var ProjectProps = ObjectProps{
	TableName:             "project",
	Type:                  reflect.TypeOf(Project{}),
	PrimaryColumnName:     "id",
	ReferringColumnSuffix: "project_id",
	DefaultSortingColumn:  "name",
	IsGlobal:              true,
}

var ScheduleProps = ObjectProps{
	TableName:         "project__schedule",
	Type:              reflect.TypeOf(Schedule{}),
	PrimaryColumnName: "id",
	Ownerships:        []*ObjectProps{&ProjectProps},
}

var SecretStorageProps = ObjectProps{
	TableName:             "project__secret_storage",
	ReferringColumnSuffix: "storage_id",
	Type:                  reflect.TypeOf(SecretStorage{}),
	PrimaryColumnName:     "id",
	Ownerships:            []*ObjectProps{&ProjectProps},
}

var UserProps = ObjectProps{
	TableName:         "user",
	Type:              reflect.TypeOf(User{}),
	PrimaryColumnName: "id",
	IsGlobal:          true,
	SortableColumns:   []string{"name", "username", "email", "role"},
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

var TaskStageResultProps = ObjectProps{
	TableName: "task__stage_result",
	Type:      reflect.TypeOf(TaskStageResult{}),
}

var ViewProps = ObjectProps{
	TableName:            "project__view",
	Type:                 reflect.TypeOf(View{}),
	PrimaryColumnName:    "id",
	DefaultSortingColumn: "position",
}

var GlobalRunnerProps = ObjectProps{
	TableName:            "runner",
	Type:                 reflect.TypeOf(Runner{}),
	PrimaryColumnName:    "id",
	DefaultSortingColumn: "id",
	SortInverted:         true,
	IsGlobal:             true,
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

var UserTotpProps = ObjectProps{
	TableName:         "user__totp",
	Type:              reflect.TypeOf(UserTotp{}),
	PrimaryColumnName: "id",
}

func (p ObjectProps) GetReferringFieldsFrom(t reflect.Type) (fields []string, err error) {
	if p.ReferringColumnSuffix == "" {
		err = errors.New("referring column suffix is not set")
		return
	}

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

	if inventory.TemplateID != nil {
		_, err = store.GetTemplate(inventory.ProjectID, *inventory.TemplateID)
	}

	return
}

type StringArrayField []string

func (m *StringArrayField) Scan(value any) error {
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
func (m *StringArrayField) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

type MapStringAnyField map[string]any

func (m *MapStringAnyField) Scan(value any) error {
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
// DO NOT ADD *, It breaks method call
func (m MapStringAnyField) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}
