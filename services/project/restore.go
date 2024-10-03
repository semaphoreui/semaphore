package project

import (
	"fmt"

	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/services/schedules"
)

func getEntryByName[T BackupEntry](name *string, items []T) *T {
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

func verifyDuplicate[T BackupEntry](name string, items []T) error {
	n := 0
	for _, o := range items {
		if o.GetName() == name {
			n++
		}
		if n > 2 {
			return fmt.Errorf("%s is duplicate", name)
		}
	}
	return nil
}

func (e BackupEnvironment) Verify(backup *BackupFormat) error {
	return verifyDuplicate[BackupEnvironment](e.Name, backup.Environments)
}

func (e BackupEnvironment) Restore(store db.Store, b *BackupDB) error {
	environment, err := store.CreateEnvironment(
		db.Environment{
			Name:      e.Name,
			Password:  e.Password,
			ProjectID: b.meta.ID,
			JSON:      e.JSON,
			ENV:       e.ENV,
		},
	)
	if err != nil {
		return err
	}
	b.environments = append(b.environments, environment)
	return nil
}

func (e BackupView) Verify(backup *BackupFormat) error {
	return verifyDuplicate[BackupView](e.Name, backup.Views)
}

func (e BackupView) Restore(store db.Store, b *BackupDB) error {
	view, err := store.CreateView(
		db.View{
			Title:     e.Name,
			ProjectID: b.meta.ID,
			Position:  e.Position,
		},
	)
	if err != nil {
		return err
	}
	b.views = append(b.views, view)
	return nil
}

func (e BackupKey) Verify(backup *BackupFormat) error {
	return verifyDuplicate[BackupKey](e.Name, backup.Keys)
}

func (e BackupKey) Restore(store db.Store, b *BackupDB) error {
	key, err := store.CreateAccessKey(
		db.AccessKey{
			Name:      e.Name,
			ProjectID: &b.meta.ID,
			Type:      e.Type,
		},
	)
	if err != nil {
		return err
	}
	b.keys = append(b.keys, key)
	return nil
}

func (e BackupInventory) Verify(backup *BackupFormat) error {
	if err := verifyDuplicate[BackupInventory](e.Name, backup.Inventories); err != nil {
		return err
	}
	if e.SSHKey != nil && getEntryByName[BackupKey](e.SSHKey, backup.Keys) == nil {
		return fmt.Errorf("SSHKey does not exist in keys[].Name")
	}
	if e.BecomeKey != nil && getEntryByName[BackupKey](e.BecomeKey, backup.Keys) == nil {
		return fmt.Errorf("BecomeKey does not exist in keys[].Name")
	}
	return nil
}

func (e BackupInventory) Restore(store db.Store, b *BackupDB) error {
	var SSHKeyID *int
	if e.SSHKey == nil {
		SSHKeyID = nil
	} else if k := findEntityByName[db.AccessKey](e.SSHKey, b.keys); k == nil {
		SSHKeyID = nil
	} else {
		SSHKeyID = &((*k).ID)
	}
	var BecomeKeyID *int
	if e.BecomeKey == nil {
		BecomeKeyID = nil
	} else if k := findEntityByName[db.AccessKey](e.BecomeKey, b.keys); k == nil {
		BecomeKeyID = nil
	} else {
		BecomeKeyID = &((*k).ID)
	}
	inventory, err := store.CreateInventory(
		db.Inventory{
			ProjectID:   b.meta.ID,
			Name:        e.Name,
			Type:        e.Type,
			SSHKeyID:    SSHKeyID,
			BecomeKeyID: BecomeKeyID,
			Inventory:   e.Inventory,
		},
	)
	if err != nil {
		return err
	}
	b.inventories = append(b.inventories, inventory)
	return nil
}

func (e BackupRepository) Verify(backup *BackupFormat) error {
	if err := verifyDuplicate[BackupRepository](e.Name, backup.Repositories); err != nil {
		return err
	}
	if e.SSHKey != nil && getEntryByName[BackupKey](e.SSHKey, backup.Keys) == nil {
		return fmt.Errorf("SSHKey does not exist in keys[].Name")
	}
	return nil
}

func (e BackupRepository) Restore(store db.Store, b *BackupDB) error {
	var SSHKeyID int
	if k := findEntityByName[db.AccessKey](e.SSHKey, b.keys); k == nil {
		return fmt.Errorf("SSHKey does not exist in keys[].Name")
	} else {
		SSHKeyID = (*k).ID
	}
	repository, err := store.CreateRepository(
		db.Repository{
			ProjectID: b.meta.ID,
			Name:      e.Name,
			GitBranch: e.GitBranch,
			GitURL:    e.GitURL,
			SSHKeyID:  SSHKeyID,
		},
	)
	if err != nil {
		return err
	}
	b.repositories = append(b.repositories, repository)
	return nil
}

func (e BackupTemplate) Verify(backup *BackupFormat) error {
	if err := verifyDuplicate[BackupTemplate](e.Name, backup.Templates); err != nil {
		return err
	}
	if getEntryByName[BackupRepository](&e.Repository, backup.Repositories) == nil {
		return fmt.Errorf("repository does not exist in repositories[].name")
	}
	if getEntryByName[BackupInventory](e.Inventory, backup.Inventories) == nil {
		return fmt.Errorf("inventory does not exist in inventories[].name")
	}
	if e.VaultKey != nil && getEntryByName[BackupKey](e.VaultKey, backup.Keys) == nil {
		return fmt.Errorf("vault_key does not exist in keys[].name")
	}
	if e.Vaults != nil {
		for _, vault := range e.Vaults {
			if getEntryByName[BackupKey](&vault.VaultKey, backup.Keys) == nil {
				return fmt.Errorf("vaults[].vaultKey does not exist in keys[].name")
			}
		}
	}
	if e.View != nil && getEntryByName[BackupView](e.View, backup.Views) == nil {
		return fmt.Errorf("view does not exist in views[].name")
	}
	if string(e.Type) == "deploy" && e.BuildTemplate == nil {
		return fmt.Errorf("type is deploy but build_template is null")
	}
	if string(e.Type) != "deploy" && e.BuildTemplate != nil {
		return fmt.Errorf("type is not deploy but build_template is not null")
	}
	if buildTemplate := getEntryByName[BackupTemplate](e.BuildTemplate, backup.Templates); string(e.Type) == "deploy" && buildTemplate == nil {
		return fmt.Errorf("deploy is build but build_template does not exist in templates[].name")
	}

	if e.Cron != nil {
		if err := schedules.ValidateCronFormat(*e.Cron); err != nil {
			return err
		}
	}

	return nil
}

func (e BackupTemplate) Restore(store db.Store, b *BackupDB) error {
	var InventoryID int
	if k := findEntityByName[db.Inventory](e.Inventory, b.inventories); k == nil {
		return fmt.Errorf("inventory does not exist in inventories[].name")
	} else {
		InventoryID = k.GetID()
	}
	var EnvironmentID int
	if k := findEntityByName[db.Environment](e.Environment, b.environments); k == nil {
		return fmt.Errorf("environment does not exist in environments[].name")
	} else {
		EnvironmentID = k.GetID()
	}
	var RepositoryID int
	if k := findEntityByName[db.Repository](&e.Repository, b.repositories); k == nil {
		return fmt.Errorf("repository does not exist in repositories[].name")
	} else {
		RepositoryID = k.GetID()
	}
	var BuildTemplateID *int
	if string(e.Type) != "deploy" {
		BuildTemplateID = nil
	} else if k := findEntityByName[db.Template](e.BuildTemplate, b.templates); k == nil {
		BuildTemplateID = nil
	} else {
		BuildTemplateID = &(k.ID)
	}
	var ViewID *int
	if k := findEntityByName[db.View](e.View, b.views); k == nil {
		ViewID = nil
	} else {
		ViewID = &k.ID
	}
	template, err := store.CreateTemplate(
		db.Template{
			ProjectID:               b.meta.ID,
			InventoryID:             &InventoryID,
			EnvironmentID:           &EnvironmentID,
			RepositoryID:            RepositoryID,
			ViewID:                  ViewID,
			Autorun:                 e.Autorun,
			AllowOverrideArgsInTask: e.AllowOverrideArgsInTask,
			SuppressSuccessAlerts:   e.SuppressSuccessAlerts,
			Name:                    e.Name,
			Playbook:                e.Playbook,
			Arguments:               e.Arguments,
			Type:                    e.Type,
			BuildTemplateID:         BuildTemplateID,
		},
	)
	if err != nil {
		return err
	}
	b.templates = append(b.templates, template)
	if e.Cron != nil {
		_, err := store.CreateSchedule(
			db.Schedule{
				ProjectID:    b.meta.ID,
				TemplateID:   template.ID,
				CronFormat:   *e.Cron,
				RepositoryID: &RepositoryID,
			},
		)
		if err != nil {
			return err
		}
	}
	if e.Vaults != nil {
		for _, vault := range e.Vaults {
			var VaultKeyID int
			if k := findEntityByName[db.AccessKey](&vault.VaultKey, b.keys); k == nil {
				return fmt.Errorf("vaults[].vaultKey does not exist in keys[].name")
			} else {
				VaultKeyID = k.ID
			}
			_, err := store.CreateTemplateVault(
				db.TemplateVault{
					ProjectID:  b.meta.ID,
					TemplateID: template.ID,
					VaultKeyID: VaultKeyID,
					Name:       vault.Name,
				},
			)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (backup *BackupFormat) Verify() error {
	for i, o := range backup.Environments {
		if err := o.Verify(backup); err != nil {
			return fmt.Errorf("error at environments[%d]: %s", i, err.Error())
		}
	}
	for i, o := range backup.Views {
		if err := o.Verify(backup); err != nil {
			return fmt.Errorf("error at views[%d]: %s", i, err.Error())
		}
	}
	for i, o := range backup.Keys {
		if err := o.Verify(backup); err != nil {
			return fmt.Errorf("error at keys[%d]: %s", i, err.Error())
		}
	}
	for i, o := range backup.Repositories {
		if err := o.Verify(backup); err != nil {
			return fmt.Errorf("error at repositories[%d]: %s", i, err.Error())
		}
	}
	for i, o := range backup.Inventories {
		if err := o.Verify(backup); err != nil {
			return fmt.Errorf("error at inventories[%d]: %s", i, err.Error())
		}
	}
	for i, o := range backup.Templates {
		if err := o.Verify(backup); err != nil {
			return fmt.Errorf("error at templates[%d]: %s", i, err.Error())
		}
	}
	return nil
}

func (backup *BackupFormat) Restore(user db.User, store db.Store) (*db.Project, error) {
	var b = BackupDB{}
	project, err := store.CreateProject(
		db.Project{
			Name:             backup.Meta.Name,
			Alert:            backup.Meta.Alert,
			MaxParallelTasks: backup.Meta.MaxParallelTasks,
			AlertChat:        backup.Meta.AlertChat,
		},
	)
	if err != nil {
		return nil, err
	}
	b.meta = project
	for i, o := range backup.Environments {
		if err := o.Restore(store, &b); err != nil {
			return nil, fmt.Errorf("error at environments[%d]: %s", i, err.Error())
		}
	}
	for i, o := range backup.Views {
		if err := o.Restore(store, &b); err != nil {
			return nil, fmt.Errorf("error at views[%d]: %s", i, err.Error())
		}
	}
	for i, o := range backup.Keys {
		if err := o.Restore(store, &b); err != nil {
			return nil, fmt.Errorf("error at keys[%d]: %s", i, err.Error())
		}
	}
	for i, o := range backup.Repositories {
		if err := o.Restore(store, &b); err != nil {
			return nil, fmt.Errorf("error at repositories[%d]: %s", i, err.Error())
		}
	}
	for i, o := range backup.Inventories {
		if err := o.Restore(store, &b); err != nil {
			return nil, fmt.Errorf("error at inventories[%d]: %s", i, err.Error())
		}
	}
	deployTemplates := make([]int, 0)
	for i, o := range backup.Templates {
		if string(o.Type) == "deploy" {
			deployTemplates = append(deployTemplates, i)
			continue
		}
		if err := o.Restore(store, &b); err != nil {
			return nil, fmt.Errorf("error at templates[%d]: %s", i, err.Error())
		}
	}
	for _, i := range deployTemplates {
		o := backup.Templates[i]
		if err := o.Restore(store, &b); err != nil {
			return nil, fmt.Errorf("error at templates[%d]: %s", i, err.Error())
		}
	}

	if _, err = store.CreateProjectUser(db.ProjectUser{
		ProjectID: project.ID,
		UserID:    user.ID,
		Role:      db.ProjectOwner,
	}); err != nil {
		return nil, err
	}

	return &project, nil
}
