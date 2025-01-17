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
	environments []db.Environment
	schedules    []db.Schedule

	integrationProjAliases   []db.IntegrationAlias
	integrations             []db.Integration
	integrationAliases       map[int][]db.IntegrationAlias
	integrationMatchers      map[int][]db.IntegrationMatcher
	integrationExtractValues map[int][]db.IntegrationExtractValue
}

type BackupFormat struct {
	Meta               BackupMeta          `backup:"meta"`
	Templates          []BackupTemplate    `backup:"templates"`
	Repositories       []BackupRepository  `backup:"repositories"`
	Keys               []BackupAccessKey   `backup:"keys"`
	Views              []BackupView        `backup:"views"`
	Inventories        []BackupInventory   `backup:"inventories"`
	Environments       []BackupEnvironment `backup:"environments"`
	Integration        []BackupIntegration `backup:"integrations"`
	IntegrationAliases []string            `backup:"integration_aliases"`
}

type BackupMeta struct {
	db.Project
}

type BackupEnvironment struct {
	db.Environment
}

type BackupAccessKey struct {
	db.AccessKey
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

type BackupTemplate struct {
	db.Template

	Inventory     *string               `backup:"inventory"`
	Repository    string                `backup:"repository"`
	Environment   *string               `backup:"environment"`
	BuildTemplate *string               `backup:"build_template"`
	View          *string               `backup:"view"`
	Vaults        []BackupTemplateVault `backup:"vaults"`
	Cron          *string               `backup:"cron"`

	// Deprecated: Left here for compatibility with old backups
	VaultKey *string `json:"vault_key"`
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

type BackupEntry interface {
	GetName() string
	Verify(backup *BackupFormat) error
	Restore(store db.Store, b *BackupDB) error
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
