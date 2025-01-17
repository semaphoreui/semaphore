package project

import (
	"fmt"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/schedules"
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
	env := e.Environment
	env.ProjectID = b.meta.ID
	newEnv, err := store.CreateEnvironment(env)
	if err != nil {
		return err
	}
	b.environments = append(b.environments, newEnv)
	return nil
}

func (e BackupView) Verify(backup *BackupFormat) error {
	return verifyDuplicate[BackupView](e.Title, backup.Views)
}

func (e BackupView) Restore(store db.Store, b *BackupDB) error {
	v := e.View
	v.ProjectID = b.meta.ID
	newView, err := store.CreateView(v)
	if err != nil {
		return err
	}
	b.views = append(b.views, newView)
	return nil
}

func (e BackupAccessKey) Verify(backup *BackupFormat) error {
	return verifyDuplicate[BackupAccessKey](e.Name, backup.Keys)
}

func (e BackupAccessKey) Restore(store db.Store, b *BackupDB) error {

	key := e.AccessKey
	key.ProjectID = &b.meta.ID

	newKey, err := store.CreateAccessKey(key)

	if err != nil {
		return err
	}
	b.keys = append(b.keys, newKey)
	return nil
}

func (e BackupInventory) Verify(backup *BackupFormat) error {
	if err := verifyDuplicate[BackupInventory](e.Name, backup.Inventories); err != nil {
		return err
	}
	if e.SSHKey != nil && getEntryByName[BackupAccessKey](e.SSHKey, backup.Keys) == nil {
		return fmt.Errorf("SSHKey does not exist in keys[].Name")
	}
	if e.BecomeKey != nil && getEntryByName[BackupAccessKey](e.BecomeKey, backup.Keys) == nil {
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

	inv := e.Inventory
	inv.ProjectID = b.meta.ID
	inv.SSHKeyID = SSHKeyID
	inv.BecomeKeyID = BecomeKeyID

	newInventory, err := store.CreateInventory(inv)
	if err != nil {
		return err
	}
	b.inventories = append(b.inventories, newInventory)
	return nil
}

func (e BackupRepository) Verify(backup *BackupFormat) error {
	if err := verifyDuplicate[BackupRepository](e.Name, backup.Repositories); err != nil {
		return err
	}
	if e.SSHKey != nil && getEntryByName[BackupAccessKey](e.SSHKey, backup.Keys) == nil {
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

	repo := e.Repository
	repo.ProjectID = b.meta.ID
	repo.SSHKeyID = SSHKeyID

	newRepo, err := store.CreateRepository(repo)
	if err != nil {
		return err
	}
	b.repositories = append(b.repositories, newRepo)
	return nil
}

func (e BackupTemplate) Verify(backup *BackupFormat) error {
	if err := verifyDuplicate[BackupTemplate](e.Name, backup.Templates); err != nil {
		return err
	}

	if getEntryByName[BackupRepository](&e.Repository, backup.Repositories) == nil {
		return fmt.Errorf("repository does not exist in repositories[].name")
	}

	if e.Inventory != nil && getEntryByName[BackupInventory](e.Inventory, backup.Inventories) == nil {
		return fmt.Errorf("inventory does not exist in inventories[].name")
	}

	if e.VaultKey != nil && getEntryByName[BackupAccessKey](e.VaultKey, backup.Keys) == nil {
		return fmt.Errorf("vault_key does not exist in keys[].name")
	}

	if e.Vaults != nil {
		for _, vault := range e.Vaults {
			if vault.VaultKey != nil {
				if getEntryByName[BackupAccessKey](vault.VaultKey, backup.Keys) == nil {
					return fmt.Errorf("vaults[].vaultKey does not exist in keys[].name")
				}
			}
		}
	}

	if e.View != nil && getEntryByName[BackupView](e.View, backup.Views) == nil {
		return fmt.Errorf("view does not exist in views[].name")
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
	var InventoryID *int
	if e.Inventory != nil {
		if k := findEntityByName[db.Inventory](e.Inventory, b.inventories); k == nil {
			return fmt.Errorf("inventory does not exist in inventories[].name")
		} else {
			id := k.GetID()
			InventoryID = &id
		}
	}

	var EnvironmentID *int
	if e.Environment != nil {
		if k := findEntityByName[db.Environment](e.Environment, b.environments); k == nil {
			return fmt.Errorf("environment does not exist in environments[].name")
		} else {
			id := k.GetID()
			EnvironmentID = &id
		}
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

	template := e.Template
	template.ProjectID = b.meta.ID
	template.RepositoryID = RepositoryID
	template.EnvironmentID = EnvironmentID
	template.InventoryID = InventoryID
	template.ViewID = ViewID
	template.BuildTemplateID = BuildTemplateID

	newTemplate, err := store.CreateTemplate(template)
	if err != nil {
		return err
	}
	b.templates = append(b.templates, newTemplate)
	if e.Cron != nil {
		_, err := store.CreateSchedule(
			db.Schedule{
				ProjectID:    b.meta.ID,
				TemplateID:   newTemplate.ID,
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
			if vault.VaultKeyID != nil {
				if k := findEntityByName[db.AccessKey](vault.VaultKey, b.keys); k == nil {
					return fmt.Errorf("vaults[].vaultKey does not exist in keys[].name")
				} else {
					VaultKeyID = k.ID
				}
			}

			tplVault := vault.TemplateVault
			tplVault.ProjectID = b.meta.ID
			tplVault.TemplateID = newTemplate.ID
			tplVault.VaultKeyID = &VaultKeyID

			_, err := store.CreateTemplateVault(tplVault)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (e BackupIntegration) Restore(store db.Store, b *BackupDB) error {
	var authSecretID *int

	if e.AuthSecret == nil {
		authSecretID = nil
	} else if k := findEntityByName[db.AccessKey](e.AuthSecret, b.keys); k == nil {
		authSecretID = nil
	} else {
		authSecretID = &((*k).ID)
	}

	tpl := findEntityByName[db.Template](&e.Template, b.templates)
	if tpl == nil {
		return fmt.Errorf("template does not exist in templates[].name")
	}

	integration := e.Integration
	integration.ProjectID = b.meta.ID
	integration.AuthSecretID = authSecretID
	integration.TemplateID = tpl.ID

	newIntegration, err := store.CreateIntegration(integration)
	if err != nil {
		return err
	}
	b.integrations = append(b.integrations, newIntegration)

	for _, m := range e.Matchers {
		m.IntegrationID = newIntegration.ID
		_, _ = store.CreateIntegrationMatcher(b.meta.ID, m)
	}

	for _, v := range e.ExtractValues {
		v.IntegrationID = newIntegration.ID
		_, _ = store.CreateIntegrationExtractValue(b.meta.ID, v)
	}

	for _, a := range e.Aliases {
		alias := db.IntegrationAlias{
			Alias:         a,
			ProjectID:     b.meta.ID,
			IntegrationID: &newIntegration.ID,
		}
		_, _ = store.CreateIntegrationAlias(alias)
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
	project := backup.Meta.Project

	newProject, err := store.CreateProject(project)

	if err != nil {
		return nil, err
	}

	if _, err = store.CreateProjectUser(db.ProjectUser{
		ProjectID: newProject.ID,
		UserID:    user.ID,
		Role:      db.ProjectOwner,
	}); err != nil {
		return nil, err
	}

	b.meta = newProject

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

	for i, o := range backup.Integration {
		if err := o.Restore(store, &b); err != nil {
			return nil, fmt.Errorf("error at integrations[%d]: %s", i, err.Error())
		}
	}

	for _, o := range backup.IntegrationAliases {
		alias := db.IntegrationAlias{
			Alias:     o,
			ProjectID: b.meta.ID,
		}
		_, _ = store.CreateIntegrationAlias(alias)
	}

	return &newProject, nil
}
