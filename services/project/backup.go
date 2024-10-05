package project

import (
	"fmt"

	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/pkg/random"
)

func findNameByID[T db.BackupEntity](ID int, items []T) (*string, error) {
	for _, o := range items {
		if o.GetID() == ID {
			name := o.GetName()
			return &name, nil
		}
	}
	return nil, fmt.Errorf("item %d does not exist", ID)
}
func findEntityByName[T db.BackupEntity](name *string, items []T) *T {
	if name == nil {
		return nil
	}
	for _, o := range items {
		if o.GetName() == *name {
			return &o
		}
	}
	return nil
}

func getSchedulesByProject(projectID int, schedules []db.Schedule) []db.Schedule {
	result := make([]db.Schedule, 0)
	for _, o := range schedules {
		if o.ProjectID == projectID {
			result = append(result, o)
		}
	}
	return result
}

func getScheduleByTemplate(templateID int, schedules []db.Schedule) *string {
	for _, o := range schedules {
		if o.TemplateID == templateID {
			return &o.CronFormat
		}
	}
	return nil
}

func getRandomName(name string) string {
	return name + " - " + random.String(10)
}

func makeUniqueNames[T any](items []T, getter func(item *T) string, setter func(item *T, name string)) {
	for i := len(items) - 1; i >= 0; i-- {
		for k, other := range items {
			if k == i {
				break
			}

			name := getter(&items[i])

			if name == getter(&other) {
				randomName := getRandomName(name)
				setter(&items[i], randomName)
				break
			}
		}
	}
}

func (b *BackupDB) makeUniqueNames() {

	makeUniqueNames(b.templates, func(item *db.Template) string {
		return item.Name
	}, func(item *db.Template, name string) {
		item.Name = name
	})

	makeUniqueNames(b.repositories, func(item *db.Repository) string {
		return item.Name
	}, func(item *db.Repository, name string) {
		item.Name = name
	})

	makeUniqueNames(b.inventories, func(item *db.Inventory) string {
		return item.Name
	}, func(item *db.Inventory, name string) {
		item.Name = name
	})

	makeUniqueNames(b.environments, func(item *db.Environment) string {
		return item.Name
	}, func(item *db.Environment, name string) {
		item.Name = name
	})

	makeUniqueNames(b.keys, func(item *db.AccessKey) string {
		return item.Name
	}, func(item *db.AccessKey, name string) {
		item.Name = name
	})

	makeUniqueNames(b.views, func(item *db.View) string {
		return item.Title
	}, func(item *db.View, name string) {
		item.Title = name
	})
}

func (b *BackupDB) new(projectID int, store db.Store) (*BackupDB, error) {
	var err error

	b.templates, err = store.GetTemplates(projectID, db.TemplateFilter{}, db.RetrieveQueryParams{})
	if err != nil {
		return nil, err
	}

	b.repositories, err = store.GetRepositories(projectID, db.RetrieveQueryParams{})
	if err != nil {
		return nil, err
	}

	b.keys, err = store.GetAccessKeys(projectID, db.RetrieveQueryParams{})
	if err != nil {
		return nil, err
	}

	b.views, err = store.GetViews(projectID)
	if err != nil {
		return nil, err
	}

	b.inventories, err = store.GetInventories(projectID, db.RetrieveQueryParams{})
	if err != nil {
		return nil, err
	}

	b.environments, err = store.GetEnvironments(projectID, db.RetrieveQueryParams{})
	if err != nil {
		return nil, err
	}
	schedules, err := store.GetSchedules()
	if err != nil {
		return nil, err
	}
	b.schedules = getSchedulesByProject(projectID, schedules)
	b.meta, err = store.GetProject(projectID)
	if err != nil {
		return nil, err
	}

	b.makeUniqueNames()

	return b, nil
}

func (b *BackupDB) format() (*BackupFormat, error) {
	keys := make([]BackupKey, len(b.keys))
	for i, o := range b.keys {
		keys[i] = BackupKey{
			Name: o.Name,
			Type: o.Type,
		}
	}

	environments := make([]BackupEnvironment, len(b.environments))
	for i, o := range b.environments {
		environments[i] = BackupEnvironment{
			Name:     o.Name,
			ENV:      o.ENV,
			JSON:     o.JSON,
			Password: o.Password,
		}
	}

	inventories := make([]BackupInventory, len(b.inventories))
	for i, o := range b.inventories {
		var SSHKey *string = nil
		if o.SSHKeyID != nil {
			SSHKey, _ = findNameByID[db.AccessKey](*o.SSHKeyID, b.keys)
		}
		var BecomeKey *string = nil
		if o.BecomeKeyID != nil {
			BecomeKey, _ = findNameByID[db.AccessKey](*o.BecomeKeyID, b.keys)
		}
		inventories[i] = BackupInventory{
			Name:      o.Name,
			Inventory: o.Inventory,
			Type:      o.Type,
			SSHKey:    SSHKey,
			BecomeKey: BecomeKey,
		}
	}

	views := make([]BackupView, len(b.views))
	for i, o := range b.views {
		views[i] = BackupView{
			Name:     o.Title,
			Position: o.Position,
		}
	}

	repositories := make([]BackupRepository, len(b.repositories))
	for i, o := range b.repositories {
		SSHKey, _ := findNameByID[db.AccessKey](o.SSHKeyID, b.keys)
		repositories[i] = BackupRepository{
			Name:      o.Name,
			SSHKey:    SSHKey,
			GitURL:    o.GitURL,
			GitBranch: o.GitBranch,
		}
	}

	templates := make([]BackupTemplate, len(b.templates))
	for i, o := range b.templates {
		var View *string = nil
		if o.ViewID != nil {
			View, _ = findNameByID[db.View](*o.ViewID, b.views)
		}
		var vaults []BackupTemplateVault = nil
		for _, vault := range o.Vaults {
			var vaultKey *string = nil
			vaultKey, _ = findNameByID[db.AccessKey](vault.VaultKeyID, b.keys)
			vaults = append(vaults, BackupTemplateVault{
				Name:     vault.Name,
				VaultKey: *vaultKey,
			})

		}
		var Environment *string = nil
		if o.EnvironmentID != nil {
			Environment, _ = findNameByID[db.Environment](*o.EnvironmentID, b.environments)
		}
		var BuildTemplate *string = nil
		if o.BuildTemplateID != nil {
			BuildTemplate, _ = findNameByID[db.Template](*o.BuildTemplateID, b.templates)
		}
		Repository, _ := findNameByID[db.Repository](o.RepositoryID, b.repositories)

		var Inventory *string = nil
		if o.InventoryID != nil {
			Inventory, _ = findNameByID[db.Inventory](*o.InventoryID, b.inventories)
		}

		templates[i] = BackupTemplate{
			Name:                    o.Name,
			AllowOverrideArgsInTask: o.AllowOverrideArgsInTask,
			Arguments:               o.Arguments,
			Autorun:                 o.Autorun,
			Description:             o.Description,
			Playbook:                o.Playbook,
			StartVersion:            o.StartVersion,
			SuppressSuccessAlerts:   o.SuppressSuccessAlerts,
			SurveyVars:              o.SurveyVarsJSON,
			Type:                    o.Type,
			View:                    View,
			Repository:              *Repository,
			Inventory:               Inventory,
			Environment:             Environment,
			BuildTemplate:           BuildTemplate,
			Cron:                    getScheduleByTemplate(o.ID, b.schedules),
			Vaults:                  vaults,
		}
	}
	return &BackupFormat{
		Meta: BackupMeta{
			Name:             b.meta.Name,
			MaxParallelTasks: b.meta.MaxParallelTasks,
			Alert:            b.meta.Alert,
			AlertChat:        b.meta.AlertChat,
		},
		Inventories:  inventories,
		Environments: environments,
		Views:        views,
		Repositories: repositories,
		Keys:         keys,
		Templates:    templates,
	}, nil
}

func GetBackup(projectID int, store db.Store) (*BackupFormat, error) {
	backup := BackupDB{}
	if _, err := backup.new(projectID, store); err != nil {
		return nil, err
	}

	return backup.format()
}
