package project

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/random"
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

	makeUniqueNames(b.integrations, func(item *db.Integration) string {
		return item.Name
	}, func(item *db.Integration, name string) {
		item.Name = name
	})
}

func (b *BackupDB) load(projectID int, store db.Store) (err error) {

	b.templates, err = store.GetTemplates(projectID, db.TemplateFilter{}, db.RetrieveQueryParams{})
	if err != nil {
		return
	}

	for i := range b.templates {
		var vaults []db.TemplateVault
		vaults, err = store.GetTemplateVaults(b.templates[i].ProjectID, b.templates[i].ID)
		if err != nil {
			return
		}
		b.templates[i].Vaults = vaults
	}

	b.repositories, err = store.GetRepositories(projectID, db.RetrieveQueryParams{})
	if err != nil {
		return
	}

	b.keys, err = store.GetAccessKeys(projectID, db.RetrieveQueryParams{})
	if err != nil {
		return
	}

	b.views, err = store.GetViews(projectID)
	if err != nil {
		return
	}

	b.inventories, err = store.GetInventories(projectID, db.RetrieveQueryParams{})
	if err != nil {
		return
	}

	b.environments, err = store.GetEnvironments(projectID, db.RetrieveQueryParams{})
	if err != nil {
		return
	}

	schedules, err := store.GetSchedules()
	if err != nil {
		return
	}

	b.schedules = getSchedulesByProject(projectID, schedules)

	b.meta, err = store.GetProject(projectID)
	if err != nil {
		return
	}

	b.integrationProjAliases, err = store.GetIntegrationAliases(projectID, nil)
	if err != nil {
		return
	}

	b.integrations, err = store.GetIntegrations(projectID, db.RetrieveQueryParams{})
	if err != nil {
		return
	}

	b.integrationAliases = make(map[int][]db.IntegrationAlias)
	b.integrationMatchers = make(map[int][]db.IntegrationMatcher)
	b.integrationExtractValues = make(map[int][]db.IntegrationExtractValue)
	for _, o := range b.integrations {
		b.integrationAliases[o.ID], err = store.GetIntegrationAliases(projectID, &o.ID)
		if err != nil {
			return
		}
		b.integrationMatchers[o.ID], err = store.GetIntegrationMatchers(projectID, db.RetrieveQueryParams{}, o.ID)
		if err != nil {
			return
		}
		b.integrationExtractValues[o.ID], err = store.GetIntegrationExtractValues(projectID, db.RetrieveQueryParams{}, o.ID)
		if err != nil {
			return
		}
	}

	b.makeUniqueNames()

	return
}

func (b *BackupDB) format() (*BackupFormat, error) {
	keys := make([]BackupAccessKey, len(b.keys))
	for i, o := range b.keys {
		keys[i] = BackupAccessKey{
			o,
		}
	}

	environments := make([]BackupEnvironment, len(b.environments))
	for i, o := range b.environments {
		environments[i] = BackupEnvironment{
			o,
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
			Inventory: o,
			SSHKey:    SSHKey,
			BecomeKey: BecomeKey,
		}
	}

	views := make([]BackupView, len(b.views))
	for i, o := range b.views {
		views[i] = BackupView{
			o,
		}
	}

	repositories := make([]BackupRepository, len(b.repositories))
	for i, o := range b.repositories {
		SSHKey, _ := findNameByID[db.AccessKey](o.SSHKeyID, b.keys)
		repositories[i] = BackupRepository{
			Repository: o,
			SSHKey:     SSHKey,
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
			if vault.VaultKeyID != nil {
				vaultKey, _ = findNameByID[db.AccessKey](*vault.VaultKeyID, b.keys)
			}
			vaults = append(vaults, BackupTemplateVault{
				TemplateVault: vault,
				VaultKey:      vaultKey,
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
			Template:      o,
			View:          View,
			Repository:    *Repository,
			Inventory:     Inventory,
			Environment:   Environment,
			BuildTemplate: BuildTemplate,
			Cron:          getScheduleByTemplate(o.ID, b.schedules),
			Vaults:        vaults,
		}
	}

	integrations := make([]BackupIntegration, len(b.integrations))
	for i, o := range b.integrations {

		var aliases []string

		for _, a := range b.integrationAliases[o.ID] {
			aliases = append(aliases, a.Alias)
		}

		tplName, _ := findNameByID[db.Template](o.TemplateID, b.templates)

		if tplName == nil {
			continue
		}

		var keyName *string

		if o.AuthSecretID != nil {
			keyName, _ = findNameByID[db.AccessKey](*o.AuthSecretID, b.keys)
		}

		integrations[i] = BackupIntegration{
			Integration:   o,
			Aliases:       aliases,
			Matchers:      b.integrationMatchers[o.ID],
			ExtractValues: b.integrationExtractValues[o.ID],
			Template:      *tplName,
			AuthSecret:    keyName,
		}
	}

	var integrationAliases []string

	for _, alias := range b.integrationProjAliases {
		integrationAliases = append(integrationAliases, alias.Alias)
	}

	return &BackupFormat{
		Meta: BackupMeta{
			b.meta,
		},
		Inventories:        inventories,
		Environments:       environments,
		Views:              views,
		Repositories:       repositories,
		Keys:               keys,
		Templates:          templates,
		Integration:        integrations,
		IntegrationAliases: integrationAliases,
	}, nil
}

func GetBackup(projectID int, store db.Store) (*BackupFormat, error) {
	backup := BackupDB{}
	if err := backup.load(projectID, store); err != nil {
		return nil, err
	}
	return backup.format()
}

func (b *BackupFormat) Marshal() (res string, err error) {
	data, err := marshalValue(reflect.ValueOf(b))
	if err != nil {
		return
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return
	}

	res = string(bytes)

	return
}

func (b *BackupFormat) Unmarshal(res string) (err error) {
	// Parse the JSON data into a map
	var jsonData interface{}
	if err = json.Unmarshal([]byte(res), &jsonData); err != nil {
		return
	}

	// Start the recursive unmarshaling process
	err = unmarshalValueWithBackupTags(jsonData, reflect.ValueOf(b))
	return
}
