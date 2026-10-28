package project

import (
	"fmt"

	"github.com/semaphoreui/semaphore/db"
)

type BackupDB struct {
	meta         db.Project
	templates    []db.Template
	repositories []db.Repository
	hostConfigs  []db.HostConfig
	keys         []db.AccessKey
	views        []db.View
	inventories  []db.Inventory
	environments []db.Environment
	schedules    []db.Schedule

	integrationProjAliases   []db.IntegrationAlias
	integrations             []db.Integration
	integrationAliases       map[int][]db.IntegrationAlias
	integrationMatchers      map[int][]db.IntegrationMatcher
	integrationExtractValues map[int][]db.IntegrationExtractValue

	secretStorages []db.SecretStorage
	roles          []db.Role
	unassigned     []UnassignedRole
	templateRoles  map[int][]db.TemplateRolePerm
	workflowRoles  map[int][]db.WorkflowRolePerm
	runners        []db.Runner
	workflows      []db.WorkflowTemplate
	// version is the format version of the file being restored.
	version int

	// store is the main store every entity restores into. Held here so Restore
	// implementations read it from BackupDB instead of taking it as a parameter.
	store db.Store
	// workflowStore persists workflow templates. Workflows are a Pro feature
	// living outside db.Store (see db.WorkflowManager), so it is injected
	// separately rather than reached through store.
	workflowStore db.WorkflowManager
}

// BackupFormatVersion is 3: version 2 gave roles slugs, 3 added the workflow verbs and grants on workflows.
const BackupFormatVersion = 3

// CleanupError says the project stayed behind after a failed restore.
type CleanupError struct {
	ProjectID int
	Err       error
	DeleteErr error
}

func (e *CleanupError) Error() string {
	return fmt.Sprintf("%v (project %d could not be removed: %v)", e.Err, e.ProjectID, e.DeleteErr)
}

func (e *CleanupError) Unwrap() error { return e.Err }

// UnassignedTemplate is a template whose grant to the role was not restored.
type UnassignedTemplate struct {
	ID          int                      `json:"id"`
	Name        string                   `json:"name"`
	Permissions db.ProjectUserPermission `json:"permissions"`
}

// UnassignedWorkflow is a workflow whose grant to the role was not restored.
type UnassignedWorkflow struct {
	ID          int                      `json:"id"`
	Name        string                   `json:"name"`
	Permissions db.ProjectUserPermission `json:"permissions"`
}

// Why a role is in the report.
const (
	ReasonNotInFile = "not_in_file"
	ReasonSlugTaken = "slug_taken"
)

// UnassignedRole is a role the restore left to the user: not in the file, or its slug is held by a shared role.
type UnassignedRole struct {
	Slug        string                   `json:"slug"`
	Reason      string                   `json:"reason"`
	Name        string                   `json:"name,omitempty"`
	Permissions db.ProjectUserPermission `json:"permissions,omitempty"`
	Templates   []UnassignedTemplate     `json:"templates"`
	Workflows   []UnassignedWorkflow     `json:"workflows"`
}

// RestoredProject is the restore result: the project and what it did not restore.
type RestoredProject struct {
	db.Project
	UnassignedRoles []UnassignedRole `json:"unassigned_roles"`
}

type BackupFormat struct {
	Version            int                   `backup:"version"`
	Meta               BackupMeta            `backup:"meta"`
	Templates          []BackupTemplate      `backup:"templates"`
	Repositories       []BackupRepository    `backup:"repositories"`
	HostConfigs        []BackupHostConfig    `backup:"host_configs"`
	Keys               []BackupAccessKey     `backup:"keys"`
	Views              []BackupView          `backup:"views"`
	Inventories        []BackupInventory     `backup:"inventories"`
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
}

type BackupRepository struct {
	db.Repository
	SSHKey *string `backup:"ssh_key"`
}

type BackupHostConfig struct {
	db.HostConfig
	SSHKey *string `backup:"ssh_key"`
}

type BackupTemplateRole struct {
	Slug        string                   `backup:"slug"`
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

// BackupIntegration wraps an integration with its target by name: exactly
// one of Template and Workflow is set. Backups written before workflow
// targets existed carry "template" only and restore unchanged.
type BackupIntegration struct {
	db.Integration
	Aliases       []string                     `backup:"aliases"`
	Matchers      []db.IntegrationMatcher      `backup:"matchers"`
	ExtractValues []db.IntegrationExtractValue `backup:"extract_values"`
	Template      *string                      `backup:"template"`
	Workflow      *string                      `backup:"workflow"`
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
	Roles []BackupWorkflowRole `backup:"roles"`
}

type BackupWorkflowRole struct {
	Slug        string                   `backup:"slug"`
	Permissions db.ProjectUserPermission `backup:"permissions"`
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
