package project

import (
	"github.com/semaphoreui/semaphore/db"
)

type BackupDB struct {
	meta         db.Project
	templates    []db.Template
	repositories []db.Repository
	keys         []db.AccessKey
	views        []db.View
	inventories  []db.Inventory
	proxies      []db.Proxy
	environments []db.Environment
	schedules    []db.Schedule

	integrationProjAliases   []db.IntegrationAlias
	integrations             []db.Integration
	integrationAliases       map[int][]db.IntegrationAlias
	integrationMatchers      map[int][]db.IntegrationMatcher
	integrationExtractValues map[int][]db.IntegrationExtractValue

	secretStorages []db.SecretStorage
	globalRoles    []db.Role
	roles          []db.Role
	templateRoles  map[int][]db.TemplateRolePerm
	runners        []db.Runner
	workflows      []db.WorkflowTemplate

	// store is the main store every entity restores into. Held here so Restore
	// implementations read it from BackupDB instead of taking it as a parameter.
	store db.Store
	// workflowStore persists workflow templates. Workflows are a Pro feature
	// living outside db.Store (see db.WorkflowManager), so it is injected
	// separately rather than reached through store.
	workflowStore db.WorkflowManager
}

type BackupFormat struct {
	Meta               BackupMeta            `backup:"meta"`
	Templates          []BackupTemplate      `backup:"templates"`
	Repositories       []BackupRepository    `backup:"repositories"`
	Keys               []BackupAccessKey     `backup:"keys"`
	Views              []BackupView          `backup:"views"`
	Inventories        []BackupInventory     `backup:"inventories"`
	Proxies            []BackupProxy         `backup:"proxies"`
	Environments       []BackupEnvironment   `backup:"environments"`
	Integration        []BackupIntegration   `backup:"integrations"`
	IntegrationAliases []string              `backup:"integration_aliases"`
	Schedules          []BackupSchedule      `backup:"schedules"`
	SecretStorages     []BackupSecretStorage `backup:"secret_storages"`
	Roles              []BackupRole          `backup:"roles"`
	Runners            []BackupRunner        `backup:"runners"`
	Workflows          []BackupWorkflow      `backup:"workflows"`
}

type BackupMeta struct {
	db.Project
}

type BackupEnvironment struct {
	db.Environment
}

type BackupAccessKey struct {
	db.AccessKey
	SourceStorage *string `backup:"source_storage"`
	Storage       *string `backup:"storage"`
}

type BackupSchedule struct {
	db.Schedule
	Template            string  `backup:"template"`
	CheckableRepository *string `backup:"checkable_repository"`
}

type BackupView struct {
	db.View
}

type BackupInventory struct {
	db.Inventory
	SSHKey    *string `backup:"ssh_key"`
	BecomeKey *string `backup:"become_key"`
	Proxy     *string `backup:"proxy"`
}

type BackupProxy struct {
	db.Proxy
	SSHKey        *string `backup:"ssh_key"`
	RequiresProxy *string `backup:"requires_proxy"`
}

type BackupRepository struct {
	db.Repository
	SSHKey *string `backup:"ssh_key"`
	Proxy  *string `backup:"proxy"`
}

type BackupTemplateRole struct {
	Role        string                   `backup:"role"`
	IsGlobal    bool                     `backup:"is_global"`
	Permissions db.ProjectUserPermission `backup:"permissions"`
}

type BackupTemplate struct {
	db.Template

	Inventory     *string               `backup:"inventory"`
	Repository    string                `backup:"repository"`
	Environments  []string              `backup:"environments"`
	BuildTemplate *string               `backup:"build_template"`
	View          *string               `backup:"view"`
	Vaults        []BackupTemplateVault `backup:"vaults"`
	//Cron          *string               `backup:"cron"`

	// Deprecated: Left here for compatibility with old backups
	VaultKey *string `json:"vault_key"`

	Roles []BackupTemplateRole `backup:"roles"`
}

type BackupTemplateVault struct {
	db.TemplateVault
	VaultKey *string `backup:"vault_key"`
}

type BackupIntegration struct {
	db.Integration
	Aliases       []string                     `backup:"aliases"`
	Matchers      []db.IntegrationMatcher      `backup:"matchers"`
	ExtractValues []db.IntegrationExtractValue `backup:"extract_values"`
	Template      string                       `backup:"template"`
	AuthSecret    *string                      `backup:"auth_secret"`
}

type BackupSecretStorage struct {
	db.SecretStorage
}

type BackupRole struct {
	db.Role
}

type BackupRunner struct {
	db.Runner
}

// BackupWorkflow wraps a workflow template for export/import. Nodes are wrapped
// separately so their template reference can be stored by name instead of by
// project-scoped ID. Edges (carried by the embedded WorkflowTemplate) reference
// nodes by ID, which is preserved verbatim and remapped on restore.
type BackupWorkflow struct {
	db.WorkflowTemplate
	Nodes []BackupWorkflowNode `backup:"nodes"`
}

// BackupWorkflowNode wraps a workflow node, replacing its template_id with a
// name reference that is portable across projects. The task params inventory
// is carried by name via TaskParams.InventoryName, like schedules do.
type BackupWorkflowNode struct {
	db.WorkflowNode
	Template *string `backup:"template"`
}

type BackupEntry interface {
	GetName() string
	Verify(backup *BackupFormat) error
	Restore(b *BackupDB) error
}

func (e BackupEnvironment) GetName() string {
	return e.Name
}

func (e BackupInventory) GetName() string {
	return e.Name
}

func (e BackupAccessKey) GetName() string {
	return e.Name
}

func (e BackupRepository) GetName() string {
	return e.Name
}

func (e BackupView) GetName() string {
	return e.Title
}

func (e BackupTemplate) GetName() string {
	return e.Name
}

func (e BackupSecretStorage) GetName() string {
	return e.Name
}

func (e BackupRole) GetName() string {
	return e.Name
}

func (e BackupRunner) GetName() string {
	return e.Name
}

func (e BackupWorkflow) GetName() string {
	return e.Name
}
