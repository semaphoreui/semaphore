package project

import (
	"encoding/json"
	"fmt"
	"reflect"

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
			vaultKey, _ = findNameByID[db.AccessKey](vault.VaultKeyID, b.keys)
			vaults = append(vaults, BackupTemplateVault{
				TemplateVault: vault,
				VaultKey:      *vaultKey,
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
	return &BackupFormat{
		Meta: BackupMeta{
			b.meta,
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

func marshalValue(v reflect.Value) (interface{}, error) {
	// Handle pointers
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, nil
		}
		return marshalValue(v.Elem())
	}

	// Handle structs
	if v.Kind() == reflect.Struct {
		typeOfV := v.Type()
		result := make(map[string]interface{})

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			fieldType := typeOfV.Field(i)

			// Handle anonymous fields (embedded structs)
			if fieldType.Anonymous {
				embeddedValue, err := marshalValue(fieldValue)
				if err != nil {
					return nil, err
				}
				if embeddedMap, ok := embeddedValue.(map[string]interface{}); ok {
					// Merge embedded struct fields into parent result map
					for k, v := range embeddedMap {
						result[k] = v
					}
				}
				continue
			}

			tag := fieldType.Tag.Get("backup")

			// Check if the field should be backed up
			if tag == "-" {
				continue // Skip fields with backup:"-"
			} else if tag == "" {
				// Get the field name from the "db" tag
				tag = fieldType.Tag.Get("db")
				if tag == "" || tag == "-" {
					continue // Skip if "db" tag is empty or "-"
				}
			}

			// Recursively process the field value
			value, err := marshalValue(fieldValue)
			if err != nil {
				return nil, err
			}

			result[tag] = value
		}
		return result, nil
	}

	// Handle slices and arrays
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		if v.IsNil() {
			return nil, nil
		}
		var result []interface{}
		for i := 0; i < v.Len(); i++ {
			elemValue, err := marshalValue(v.Index(i))
			if err != nil {
				return nil, err
			}
			result = append(result, elemValue)
		}
		return result, nil
	}

	// Handle maps
	if v.Kind() == reflect.Map {
		if v.IsNil() {
			return nil, nil
		}
		result := make(map[string]interface{})
		for _, key := range v.MapKeys() {
			// Assuming the key is a string
			mapKey := fmt.Sprintf("%v", key.Interface())
			mapValue, err := marshalValue(v.MapIndex(key))
			if err != nil {
				return nil, err
			}
			result[mapKey] = mapValue
		}
		return result, nil
	}

	// Handle other types (int, string, etc.)
	return v.Interface(), nil
}

// UnmarshalStruct deserializes JSON data into a struct,
// using the "db" tag for field names and excluding fields with backup:"-".
func UnmarshalStruct(data []byte, v interface{}) error {
	// Parse the JSON data into an interface{}
	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return err
	}
	// Start the recursive unmarshaling process
	return unmarshalValue(jsonData, reflect.ValueOf(v))
}

func unmarshalValue(data interface{}, v reflect.Value) error {
	// Handle pointers
	if v.Kind() == reflect.Ptr {
		// Initialize pointer if it's nil
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		return unmarshalValue(data, v.Elem())
	}

	// Handle structs
	if v.Kind() == reflect.Struct {
		// Data should be a map
		m, ok := data.(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected object for struct, got %T", data)
		}
		return unmarshalStruct(m, v)
	}

	// Handle slices and arrays
	if v.Kind() == reflect.Slice {
		// Data should be an array
		dataSlice, ok := data.([]interface{})
		if !ok {
			return fmt.Errorf("expected array for slice, got %T", data)
		}
		// Create a new slice
		slice := reflect.MakeSlice(v.Type(), len(dataSlice), len(dataSlice))
		for i := 0; i < len(dataSlice); i++ {
			elem := slice.Index(i)
			if err := unmarshalValue(dataSlice[i], elem); err != nil {
				return err
			}
		}
		v.Set(slice)
		return nil
	}

	// Handle maps
	if v.Kind() == reflect.Map {
		// Data should be a map
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected object for map, got %T", data)
		}
		mapType := v.Type()
		mapValue := reflect.MakeMap(mapType)
		for key, value := range dataMap {
			keyVal := reflect.ValueOf(key).Convert(mapType.Key())
			valVal := reflect.New(mapType.Elem()).Elem()
			if err := unmarshalValue(value, valVal); err != nil {
				return err
			}
			mapValue.SetMapIndex(keyVal, valVal)
		}
		v.Set(mapValue)
		return nil
	}

	// Handle basic types
	if err := setBasicType(data, v); err != nil {
		return err
	}

	return nil
}

func unmarshalStruct(data map[string]interface{}, v reflect.Value) error {
	t := v.Type()

	// Build a map of db tags to field indices
	fieldMap := make(map[string]int)
	for i := 0; i < t.NumField(); i++ {
		fieldType := t.Field(i)

		// Skip fields with backup:"-"
		if backupTag := fieldType.Tag.Get("backup"); backupTag == "-" {
			continue
		}

		// Get the field name from the "db" tag
		dbTag := fieldType.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}

		fieldMap[dbTag] = i
	}

	// Iterate over the JSON data and set struct fields
	for key, value := range data {
		if index, ok := fieldMap[key]; ok {
			field := v.Field(index)
			if !field.CanSet() {
				continue // Skip unexportable fields
			}
			if err := unmarshalValue(value, field); err != nil {
				return err
			}
		}
	}

	return nil
}

func setBasicType(data interface{}, v reflect.Value) error {
	if !v.CanSet() {
		return fmt.Errorf("cannot set value of type %v", v.Type())
	}

	switch v.Kind() {
	case reflect.Bool:
		b, ok := data.(bool)
		if !ok {
			return fmt.Errorf("expected bool for field, got %T", data)
		}
		v.SetBool(b)
	case reflect.String:
		s, ok := data.(string)
		if !ok {
			return fmt.Errorf("expected string for field, got %T", data)
		}
		v.SetString(s)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, ok := toFloat64(data)
		if !ok {
			return fmt.Errorf("expected number for field, got %T", data)
		}
		v.SetInt(int64(n))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, ok := toFloat64(data)
		if !ok {
			return fmt.Errorf("expected number for field, got %T", data)
		}
		v.SetUint(uint64(n))
	case reflect.Float32, reflect.Float64:
		n, ok := toFloat64(data)
		if !ok {
			return fmt.Errorf("expected number for field, got %T", data)
		}
		v.SetFloat(n)
	default:
		return fmt.Errorf("unsupported kind %v", v.Kind())
	}
	return nil
}

func toFloat64(data interface{}) (float64, bool) {
	switch n := data.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case int32:
		return float64(n), true
	case int16:
		return float64(n), true
	case int8:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint64:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint8:
		return float64(n), true
	default:
		return 0, false
	}
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
	err = UnmarshalStruct([]byte(res), reflect.ValueOf(b))

	return
}
