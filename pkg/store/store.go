package store

import (
	"context"
	"fmt"

	model "github.com/ansible-semaphore/semaphore/db"
)

var (
	// ErrUnknownDriver defines a named error for unknown store drivers.
	ErrUnknownDriver = fmt.Errorf("unknown database driver")

	// ErrRecordNotFound is returned when a record was not found.
	ErrRecordNotFound = fmt.Errorf("record not found")
)

// Store provides the interface for the store implementations.
type Store interface {
	Info() map[string]interface{}
	Prepare() error
	Open() error
	Close() error
	Ping() error
	Migrate() error
	Admin(string, string, string) error

	GetOption(ctx context.Context, key string) error
	SetOption(ctx context.Context, key, val string) error

	UserEvents(ctx context.Context, params model.EventParams) ([]*model.Event, error)
	ProjectEvents(ctx context.Context, params model.EventParams) ([]*model.Event, error)
	CreateEvent(ctx context.Context, record *model.Event) error

	AdminUsers(ctx context.Context, params model.UserParams) ([]*model.User, error)
	ListUsers(Usersctx context.Context, params model.UserParams) ([]*model.User, error)
	ShowUser(ctx context.Context, params model.UserParams) (*model.User, error)
	DeleteUser(ctx context.Context, params model.UserParams) error
	CreateUser(ctx context.Context, record *model.User) error
	UpdateUser(ctx context.Context, record *model.User) error
	UpdatePassword(ctx context.Context, params model.UserParams) error

	ListTokens(ctx context.Context, params model.TokenParams) ([]*model.APIToken, error)
	ShowToken(ctx context.Context, params model.TokenParams) (*model.APIToken, error)
	DeleteToken(ctx context.Context, params model.TokenParams) error
	ExpireToken(ctx context.Context, params model.TokenParams) error
	CreateToken(ctx context.Context, record *model.APIToken) error

	ShowSession(ctx context.Context, params model.SessionParams) (*model.Session, error)
	ExpireSession(ctx context.Context, params model.SessionParams) error
	TouchSession(ctx context.Context, params model.SessionParams) error
	CreateSession(ctx context.Context, record *model.Session) error

	ListGlobalRunners(ctx context.Context, params model.RunnerParams) ([]*model.Runner, error)
	ShowGlobalRunner(ctx context.Context, params model.RunnerParams) (*model.Runner, error)
	DeleteGlobalRunner(ctx context.Context, params model.RunnerParams) error
	CreateGlobalRunner(ctx context.Context, record *model.Runner) error
	UpdateGlobalRunner(ctx context.Context, record *model.Runner) error

	ListProjectRunner(ctx context.Context, params model.RunnerParams) ([]*model.Runner, error)
	ShowProjectRunner(ctx context.Context, params model.RunnerParams) (*model.Runner, error)
	DeleteProjectRunner(ctx context.Context, params model.RunnerParams) error
	CreateProjectRunner(ctx context.Context, record *model.Runner) error
	UpdateProjectRunner(ctx context.Context, record *model.Runner) error

	ListProjects(ctx context.Context, params model.ProjectParams) ([]*model.Project, error)
	ShowProject(ctx context.Context, params model.ProjectParams) (*model.Project, error)
	DeleteProject(ctx context.Context, params model.ProjectParams) error
	CreateProject(ctx context.Context, record *model.Project) error
	UpdateProject(ctx context.Context, record *model.Project) error

	ListMembers(ctx context.Context, params model.MemberParams) ([]*model.UserWithProjectRole, error)
	ShowMember(ctx context.Context, params model.MemberParams) (*model.ProjectUser, error)
	DeleteMember(ctx context.Context, params model.MemberParams) error
	CreateMember(ctx context.Context, record *model.ProjectUser) error
	UpdateMember(ctx context.Context, record *model.ProjectUser) error

	TemplateRefs(ctx context.Context, params model.TemplateParams) (*model.ObjectReferrers, error)
	ListTemplates(ctx context.Context, params model.TemplateParams) ([]*model.Template, error)
	ShowTemplate(ctx context.Context, params model.TemplateParams) (*model.Template, error)
	DeleteTemplate(ctx context.Context, params model.TemplateParams) error
	CreateTemplate(ctx context.Context, record *model.Template) error
	UpdateTemplate(ctx context.Context, record *model.Template) error

	AccessKeyRefs(ctx context.Context, params model.AccessKeyParams) (*model.ObjectReferrers, error)
	ListAccessKeys(ctx context.Context, params model.AccessKeyParams) ([]*model.AccessKey, error)
	ShowAccessKey(ctx context.Context, params model.AccessKeyParams) (*model.AccessKey, error)
	DeleteAccessKey(ctx context.Context, params model.AccessKeyParams) error
	RekeyAccessKey(ctx context.Context, params model.AccessKeyParams) error
	CreateAccessKey(ctx context.Context, record *model.AccessKey) error
	UpdateAccessKey(ctx context.Context, record *model.AccessKey) error

	EnvRefs(ctx context.Context, params model.EnvParams) (*model.ObjectReferrers, error)
	ListEnvs(ctx context.Context, params model.EnvParams) ([]*model.Environment, error)
	ShowEnv(ctx context.Context, params model.EnvParams) (*model.Environment, error)
	DeleteEnv(ctx context.Context, params model.EnvParams) error
	CreateEnv(ctx context.Context, record *model.Environment) error
	UpdateEnv(ctx context.Context, record *model.Environment) error

	InventoryRefs(ctx context.Context, params model.InventoryParams) (*model.ObjectReferrers, error)
	ListInventories(ctx context.Context, params model.InventoryParams) ([]*model.Inventory, error)
	ShowInventory(ctx context.Context, params model.InventoryParams) (*model.Inventory, error)
	DeleteInventory(ctx context.Context, params model.InventoryParams) error
	CreateInventory(ctx context.Context, record *model.Inventory) error
	UpdateInventory(ctx context.Context, record *model.Inventory) error

	RepoRefs(ctx context.Context, params model.RepoParams) (*model.ObjectReferrers, error)
	ListRepos(ctx context.Context, params model.RepoParams) ([]*model.Repository, error)
	ShowRepo(ctx context.Context, params model.RepoParams) (*model.Repository, error)
	DeleteRepo(ctx context.Context, params model.RepoParams) error
	CreateRepo(ctx context.Context, record *model.Repository) error
	UpdateRepo(ctx context.Context, record *model.Repository) error

	ListViews(ctx context.Context, params model.ViewParams) ([]*model.View, error)
	ShowView(ctx context.Context, params model.ViewParams) (*model.View, error)
	PositionView(ctx context.Context, params model.ViewParams) error
	DeleteView(ctx context.Context, params model.ViewParams) error
	CreateView(ctx context.Context, record *model.View) error
	UpdateView(ctx context.Context, record *model.View) error

	TemplateSchedules(ctx context.Context, params model.ScheduleParams) ([]*model.Schedule, error)
	ListSchedules(ctx context.Context, params model.ScheduleParams) ([]*model.Schedule, error)
	ShowSchedule(ctx context.Context, params model.ScheduleParams) (*model.Schedule, error)
	HashSchedule(ctx context.Context, params model.ScheduleParams) error
	DeleteSchedule(ctx context.Context, params model.ScheduleParams) error
	CreateSchedule(ctx context.Context, record *model.Schedule) error
	UpdateSchedule(ctx context.Context, record *model.Schedule) error

	TemplateTasks(ctx context.Context, params model.TaskParams) ([]*model.TaskWithTpl, error)
	ProjectTasks(ctx context.Context, params model.TaskParams) ([]*model.TaskWithTpl, error)
	ShowTask(ctx context.Context, params model.TaskParams) (*model.Task, error)
	DeleteTask(ctx context.Context, params model.TaskParams) error
	CreateTask(ctx context.Context, record *model.Task) error
	UpdateTask(ctx context.Context, record *model.Task) error

	PushOutput(ctx context.Context, record *model.TaskOutput) error
	GetOutputs(ctx context.Context, params model.TaskParams) ([]*model.TaskOutput, error)

	IntegrationRefs(ctx context.Context, params model.IntegrationParams) (*model.IntegrationReferrers, error)
	SearchableIntegrations(ctx context.Context, params model.IntegrationParams) ([]*model.Integration, error)
	AliasedIntegrations(ctx context.Context, params model.IntegrationParams) ([]*model.Integration, error)
	ListIntegrations(ctx context.Context, params model.IntegrationParams) ([]*model.Integration, error)
	ShowIntegration(ctx context.Context, params model.IntegrationParams) (*model.Integration, error)
	DeleteIntegration(ctx context.Context, params model.IntegrationParams) error
	CreateIntegration(ctx context.Context, record *model.Integration) error
	UpdateIntegration(ctx context.Context, record *model.Integration) error

	IntegrationExtractValueRefs(ctx context.Context, params model.IntegrationExtractValueParams) (*model.IntegrationExtractorChildReferrers, error)
	ListIntegrationExtractValues(ctx context.Context, params model.IntegrationExtractValueParams) ([]*model.IntegrationExtractValue, error)
	ShowIntegrationExtractValue(ctx context.Context, params model.IntegrationExtractValueParams) (*model.IntegrationExtractValue, error)
	DeleteIntegrationExtractValue(ctx context.Context, params model.IntegrationExtractValueParams) error
	CreateIntegrationExtractValue(ctx context.Context, params model.IntegrationExtractValueParams, record *model.IntegrationExtractValue) error
	UpdateIntegrationExtractValue(ctx context.Context, params model.IntegrationExtractValueParams, record *model.IntegrationExtractValue) error

	IntegrationMatcherRefs(ctx context.Context, params model.IntegrationMatcherParams) (*model.IntegrationExtractorChildReferrers, error)
	ListIntegrationMatchers(ctx context.Context, params model.IntegrationMatcherParams) ([]*model.IntegrationMatcher, error)
	ShowIntegrationMatcher(ctx context.Context, params model.IntegrationMatcherParams) (*model.IntegrationMatcher, error)
	DeleteIntegrationMatcher(ctx context.Context, params model.IntegrationMatcherParams) error
	CreateIntegrationMatcher(ctx context.Context, params model.IntegrationMatcherParams, record *model.IntegrationMatcher) error
	UpdateIntegrationMatcher(ctx context.Context, params model.IntegrationMatcherParams, record *model.IntegrationMatcher) error

	ListIntegrationAliases(ctx context.Context, params model.IntegrationAliasParams) ([]*model.IntegrationAlias, error)
	DeleteIntegrationAlias(ctx context.Context, params model.IntegrationAliasParams) error
	CreateIntegrationAlias(ctx context.Context, record *model.IntegrationAlias) error
}
