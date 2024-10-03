package project

import (
	"github.com/ansible-semaphore/semaphore/db"
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
}

type BackupFormat struct {
	Meta         BackupMeta          `json:"meta"`
	Templates    []BackupTemplate    `json:"templates"`
	Repositories []BackupRepository  `json:"repositories"`
	Keys         []BackupKey         `json:"keys"`
	Views        []BackupView        `json:"views"`
	Inventories  []BackupInventory   `json:"inventories"`
	Environments []BackupEnvironment `json:"environments"`
}

type BackupMeta struct {
	Name             string  `json:"name"`
	Alert            bool    `json:"alert"`
	AlertChat        *string `json:"alert_chat"`
	MaxParallelTasks int     `json:"max_parallel_tasks"`
}

type BackupEnvironment struct {
	Name     string  `json:"name"`
	Password *string `json:"password"`
	JSON     string  `json:"json"`
	ENV      *string `json:"env"`
}

type BackupKey struct {
	Name string           `json:"name"`
	Type db.AccessKeyType `json:"type"`
}

type BackupView struct {
	Name     string `json:"name"`
	Position int    `json:"position"`
}

type BackupInventory struct {
	Name      string           `json:"name"`
	Inventory string           `json:"inventory"`
	SSHKey    *string          `json:"ssh_key"`
	BecomeKey *string          `json:"become_key"`
	Type      db.InventoryType `json:"type"`
}

type BackupRepository struct {
	Name      string  `json:"name"`
	GitURL    string  `json:"git_url"`
	GitBranch string  `json:"git_branch"`
	SSHKey    *string `json:"ssh_key"`
}

type BackupTemplate struct {
	Inventory               *string               `json:"inventory"`
	Repository              string                `json:"repository"`
	Environment             *string               `json:"environment"`
	Name                    string                `json:"name"`
	Playbook                string                `json:"playbook"`
	Arguments               *string               `json:"arguments"`
	AllowOverrideArgsInTask bool                  `json:"allow_override_args_in_task"`
	Description             *string               `json:"description"`
	Type                    db.TemplateType       `json:"type"`
	StartVersion            *string               `json:"start_version"`
	BuildTemplate           *string               `json:"build_template"`
	View                    *string               `json:"view"`
	Autorun                 bool                  `json:"autorun"`
	SurveyVars              *string               `json:"survey_vars"`
	SuppressSuccessAlerts   bool                  `json:"suppress_success_alerts"`
	Cron                    *string               `json:"cron"`
	Vaults                  []BackupTemplateVault `json:"vaults"`

	// Deprecated: Left here for compatibility with old backups
	VaultKey *string `json:"vault_key"`
}

type BackupTemplateVault struct {
	Name     *string `json:"name"`
	VaultKey string  `json:"vault_key"`
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

func (e BackupKey) GetName() string {
	return e.Name
}

func (e BackupRepository) GetName() string {
	return e.Name
}

func (e BackupView) GetName() string {
	return e.Name
}

func (e BackupTemplate) GetName() string {
	return e.Name
}
