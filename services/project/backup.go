package project

import (
	"fmt"

	"github.com/ansible-semaphore/semaphore/db"
)

func findNameByID[T db.BackupEntity](ID int, items []T) (string, error) {
	for _, o := range items {
		if o.GetID() == ID {
			return o.GetName(), nil
		}
	}
	return "", fmt.Errorf("item %d does not exist", ID)
}

type Backup struct {
	templates    []db.Template
	repositories []db.Repository
	keys         []db.AccessKey
	views        []db.View
	inventories  []db.Inventory
	environments []db.Environment
}

func (b *Backup) new(projectID int, store db.Store) (*Backup, error) {
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

	return b, nil
}

func (b *Backup) format() (interface{}, error) {
	var err error
	keys := make([]map[string]interface{}, len(b.keys))
	for i, o := range b.keys {
		key := make(map[string]interface{})
		key["Name"] = o.Name
		key["Type"] = o.Type
		keys[i] = key
	}

	environments := make([]map[string]interface{}, len(b.environments))
	for i, o := range b.environments {
		environment := make(map[string]interface{})
		environment["Name"] = o.Name
		environment["ENV"] = o.ENV
		environment["JSON"] = o.JSON
		environment["Password"] = o.Password
		environments[i] = environment
	}

	inventories := make([]map[string]interface{}, len(b.inventories))
	for i, o := range b.inventories {
		inventory := make(map[string]interface{})
		inventory["Name"] = o.Name
		if o.BecomeKeyID != nil {
			if inventory["BecomeKey"], err = findNameByID[db.AccessKey](*o.BecomeKeyID, b.keys); err != nil {
				return nil, err
			}
		} else {
			inventory["BecomeKey"] = nil
		}
		inventory["Inventory"] = o.Inventory
		if o.BecomeKeyID != nil {
			if inventory["SSHKey"], err = findNameByID[db.AccessKey](*o.SSHKeyID, b.keys); err != nil {
				return nil, err
			}
		} else {
			inventory["SSHKey"] = nil
		}
		inventory["Type"] = o.Type
		inventories[i] = inventory
	}

	views := make([]map[string]interface{}, len(b.views))
	for i, o := range b.views {
		view := make(map[string]interface{})
		view["Title"] = o.Title
		view["Position"] = o.Position
		views[i] = view
	}

	repositories := make([]map[string]interface{}, len(b.repositories))
	for i, o := range b.repositories {
		repository := make(map[string]interface{})
		repository["Name"] = o.Name
		if repository["SSHKey"], err = findNameByID[db.AccessKey](o.SSHKeyID, b.keys); err != nil {
			return nil, err
		}
		repository["GitBranch"] = o.GitBranch
		repository["GitURL"] = o.GitURL
		repositories[i] = repository
	}

	templates := make([]map[string]interface{}, len(b.templates))
	for i, o := range b.templates {
		template := make(map[string]interface{})
		if o.BuildTemplateID != nil {
			if template["BuildTemplate"], err = findNameByID[db.Template](*o.BuildTemplateID, b.templates); err != nil {
				return nil, err
			}
		} else {
			template["BuildTemplate"] = nil
		}

		if o.EnvironmentID != nil {
			if template["Environment"], err = findNameByID[db.Environment](*o.EnvironmentID, b.environments); err != nil {
				return nil, err
			}
		} else {
			template["Environment"] = nil
		}

		if template["Inventory"], err = findNameByID[db.Inventory](o.InventoryID, b.inventories); err != nil {
			return nil, err
		}

		if template["Repository"], err = findNameByID[db.Repository](o.RepositoryID, b.repositories); err != nil {
			return nil, err
		}

		if o.VaultKeyID != nil {
			if template["VaultKey"], err = findNameByID[db.AccessKey](*o.VaultKeyID, b.keys); err != nil {
				return nil, err
			}
		} else {
			template["VaultKey"] = nil
		}

		if o.ViewID != nil {
			if template["View"], err = findNameByID[db.View](*o.ViewID, b.views); err != nil {
				return nil, err
			}
		} else {
			template["View"] = nil
		}
		template["Name"] = o.Name
		template["AllowOverrideArgsInTask"] = o.AllowOverrideArgsInTask
		template["Arguments"] = o.Arguments
		template["Autorun"] = o.Autorun
		template["Description"] = o.Description
		template["Playbook"] = o.Playbook
		template["StartVersion"] = o.StartVersion
		template["SuppressSuccessAlerts"] = o.SuppressSuccessAlerts
		template["SurveyVars"] = o.SurveyVars
		template["SurveyVarsJSON"] = o.SurveyVarsJSON
		template["Type"] = o.Type
		template["VaultKey"] = o.VaultKey.Name
		templates[i] = template
	}

	return map[string][]map[string]interface{}{
		"keys":         keys,
		"environments": environments,
		"inventories":  inventories,
		"views":        views,
		"repositories": repositories,
		"templates":    templates,
	}, nil
}

func GetBackup(projectID int, store db.Store) (interface{}, error) {
	backup := Backup{}
	if _, err := backup.new(projectID, store); err != nil {
		return nil, err
	}

	return backup.format()
}
