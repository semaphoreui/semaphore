package export

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

const (
	User                    = "User"
	Project                 = "Project"
	AccessKey               = "AccessKey"
	Environment             = "Environment"
	Template                = "Template"
	TemplateVault           = "TemplateVault"
	TemplateRole            = "TemplateRole"
	SecretStorage           = "SecretStorage"
	Inventory               = "Inventory"
	Repository              = "Repository"
	View                    = "View"
	Role                    = "Role"
	TaskParams              = "TaskParams"
	Integration             = "Integration"
	IntegrationAlias        = "IntegrationAlias"
	IntegrationExtractValue = "IntegrationExtractValue"
	IntegrationMatcher      = "IntegrationMatcher"
	Schedule                = "Schedule"
	Task                    = "Task"
	TaskStage               = "TaskStage"
	TaskStageResult         = "TaskStageResult"
	TaskOutput              = "TaskOutput"
	ProjectUser             = "ProjectUser"
	Option                  = "Option"
	Event                   = "Event"
	Runner                  = "Runner"
)

type EntityKey = string

func NewKeyFromInt(key int) EntityKey {
	return strconv.Itoa(key)
}

func NewKey(key string) EntityKey {
	return key
}

type KeyMapper interface {
	getNewKey(name string, scope string, oldKey EntityKey, errHandler ErrorHandler) (EntityKey, error)

	getNewKeyInt(name string, scope string, oldKey int, errHandler ErrorHandler) (int, error)

	getNewKeyIntRef(name string, scope string, oldKey *int, errHandler ErrorHandler) (*int, error)

	mapKeys(name string, scope string, oldKey EntityKey, newKey EntityKey) error

	mapIntKeys(name string, scope string, oldKey int, newKey int) error

	ignoreKeyNotFound() bool
}

type DataExporter interface {
	KeyMapper

	getTypeExporter(name string) TypeExporter

	getLoadedKeys(name string, scope string) ([]EntityKey, error)

	getLoadedKeysInt(name string, scope string) ([]int, error)
}

type Progress interface {
	update(progress float32)
}

type ErrorHandler interface {
	onError(err string)
}

type TypeExporter interface {
	load(store db.Store, exporter DataExporter, progress Progress) error

	restore(store db.Store, exporter DataExporter, progress Progress) error

	getLoadedKeys(scope string) ([]EntityKey, error)

	getLoadedValues(scope string) ([]EntityType, error)

	getName() string

	exportDependsOn() []string

	importDependsOn() []string

	getErrors() []string

	clear()
}

var KeyNotFound = -1
var GlobalScope = ""

type EntityType interface {
	GetDbKey() EntityKey
}

type TypeKeyMapper struct {
	Keys                 map[string]map[string]map[EntityKey]EntityKey
	IgnoreKeyNotFoundErr bool
}

func (d *TypeKeyMapper) getNewKeyInt(name string, scope string, oldKey int, errHandler ErrorHandler) (int, error) {
	key, err := d.getNewKey(name, scope, NewKeyFromInt(oldKey), errHandler)

	if err != nil {
		return KeyNotFound, err
	}

	if key == "" && d.ignoreKeyNotFound() {
		return 0, nil
	}

	newKey, err := strconv.Atoi(key)
	if err != nil {
		return KeyNotFound, err
	}

	return newKey, nil
}

func (d *TypeKeyMapper) getNewKeyIntRef(name string, scope string, oldKey *int, errHandler ErrorHandler) (*int, error) {
	if oldKey == nil {
		return nil, nil
	}

	key, err := d.getNewKey(name, scope, NewKeyFromInt(*oldKey), errHandler)

	if err != nil {
		return nil, err
	}

	if key == "" && d.ignoreKeyNotFound() {
		return nil, nil
	}

	newKey, err := strconv.Atoi(key)
	if err != nil {
		return nil, err
	}

	return &newKey, nil
}

func (d *TypeKeyMapper) getCheckAndNewKey(name string, scope string, oldKey EntityKey, errHandler ErrorHandler) (EntityKey, error) {
	newKey, ok := d.Keys[name][scope][oldKey]
	if !ok {
		msg := fmt.Sprintf("%s key %s not found", name, oldKey)
		errHandler.onError(msg)
		return "", errors.New(msg)
	}

	return newKey, nil
}

func (d *TypeKeyMapper) getNewKey(name string, scope string, oldKey EntityKey, errHandler ErrorHandler) (EntityKey, error) {
	newKey, err := d.getCheckAndNewKey(name, scope, oldKey, errHandler)
	if err != nil {
		if d.ignoreKeyNotFound() {
			return "", nil
		}
		return "", err
	}
	return newKey, nil
}

func (d *TypeKeyMapper) mapKeys(name string, scope string, oldKey EntityKey, newKey EntityKey) error {
	_, ok := d.Keys[name]
	if !ok {
		d.Keys[name] = make(map[string]map[EntityKey]EntityKey)
	}

	_, ok = d.Keys[name][scope]
	if !ok {
		d.Keys[name][scope] = make(map[EntityKey]EntityKey)
	}

	d.Keys[name][scope][oldKey] = newKey
	return nil
}

func (d *TypeKeyMapper) mapIntKeys(name string, scope string, oldKey int, newKey int) error {
	newStrKey := strconv.Itoa(newKey)
	oldStrKey := strconv.Itoa(oldKey)
	return d.mapKeys(name, scope, oldStrKey, newStrKey)
}

//func (d *TypeKeyMapper) hasKey(name string, scope string, oldKey EntityKey) bool {
//	_, ok := d.Keys[name][scope][oldKey]
//	return ok
//}

func (d *TypeKeyMapper) ignoreKeyNotFound() bool {
	return d.IgnoreKeyNotFoundErr
}

type EntityObject[T EntityType] struct {
	value T
	scope string
}

type ValueMap[T EntityType] struct {
	values      []EntityObject[T]
	keyScopeMap map[string]bool
	errs        []string
}

func (t *ValueMap[T]) getLoadedKeys(scope string) ([]EntityKey, error) {

	if t.values == nil {
		return nil, fmt.Errorf("values not loaded")
	}

	keys := make([]EntityKey, 0, len(t.values))
	for _, v := range t.values {
		if v.scope == scope {
			keys = append(keys, v.value.GetDbKey())
		}
	}

	return keys, nil
}

func (t *ValueMap[T]) getLoadedKeysInt(scope string) ([]int, error) {
	keys, err := t.getLoadedKeys(scope)
	if err != nil {
		return nil, err
	}
	keysInt := make([]int, 0)
	for _, k := range keys {
		intKey, err := strconv.Atoi(k)
		if err != nil {
			return nil, err
		}
		keysInt = append(keysInt, intKey)
	}
	return keysInt, nil
}

func (t *ValueMap[T]) getLoadedValues(scope string) ([]EntityType, error) {
	keys := make([]EntityType, 0)
	for _, v := range t.values {
		if v.scope == scope {
			keys = append(keys, v.value)
		}
	}
	return keys, nil
}

func (t *ValueMap[T]) appendValues(values []T, scope string) error {
	return t.appendValuesAndCheck(values, scope, true)
}

func (t *ValueMap[T]) appendValuesAndCheck(values []T, scope string, checkDuplicates bool) error {
	if t.values == nil {
		t.keyScopeMap = make(map[string]bool)
		t.values = make([]EntityObject[T], 0)
	}
	for _, v := range values {
		if checkDuplicates {
			_, ok := t.keyScopeMap[scope+v.GetDbKey()]
			if ok {
				return fmt.Errorf("duplicate key %s", v.GetDbKey())
			}
			t.keyScopeMap[scope+v.GetDbKey()] = true
		}
		t.values = append(t.values, EntityObject[T]{value: v, scope: scope})
	}
	return nil
}

func (t *ValueMap[T]) exportDependsOn() []string {
	return []string{}
}

func (t *ValueMap[T]) importDependsOn() []string {
	return []string{}
}

func (t *ValueMap[T]) onError(err string) {
	if t.errs == nil {
		t.errs = []string{err}
	} else {
		t.errs = append(t.errs, err)
	}
}

func (t *ValueMap[T]) getErrors() []string {
	return t.errs
}

func (t *ValueMap[T]) clear() {
	t.keyScopeMap = nil
	t.values = nil
	t.errs = nil
}

type ExporterChain struct {
	exporters map[string]TypeExporter
	KeyMapper
}

func (p *ExporterChain) getTypeExporter(name string) TypeExporter {
	return p.exporters[name]
}

func (p *ExporterChain) getLoadedKeys(name string, scope string) ([]EntityKey, error) {
	exporter, ok := p.exporters[name]
	if !ok {
		return nil, fmt.Errorf("type %s not found", name)
	}

	return exporter.getLoadedKeys(scope)
}

func (p *ExporterChain) getLoadedKeysInt(name string, scope string) ([]int, error) {
	exporter, ok := p.exporters[name]
	if !ok {
		return nil, fmt.Errorf("type %s not found", name)
	}

	keys, err := exporter.getLoadedKeys(scope)
	if err != nil {
		return nil, err
	}

	out := make([]int, len(keys))
	for i, v := range keys {
		n, err := strconv.Atoi(v)

		if err != nil {
			return nil, err
		}
		out[i] = n
	}
	return out, nil
}

func getSortedKeys(exporters map[string]TypeExporter, dependsOn func(t TypeExporter) []string) ([]string, error) {
	var sorted []string
	visited := make(map[string]bool)
	visiting := make(map[string]bool)

	var visit func(name string) error
	visit = func(name string) error {
		if visiting[name] {
			return fmt.Errorf("cyclic dependency detected involving %s", name)
		}
		if visited[name] {
			return nil
		}

		visiting[name] = true

		if exporter, ok := exporters[name]; ok {

			order := dependsOn(exporter)

			for _, dep := range order {
				if _, exists := exporters[dep]; exists {
					if err := visit(dep); err != nil {
						return err
					}
				}
			}
		}

		visiting[name] = false
		visited[name] = true
		sorted = append(sorted, name)
		return nil
	}

	for name := range exporters {
		if err := visit(name); err != nil {
			return nil, err
		}
	}

	return sorted, nil
}

func InitProjectExporters(mapper KeyMapper, skipTaskOutput bool) *ExporterChain {

	exporters := map[string]TypeExporter{
		User:                    &UserExporter{},
		Project:                 &ProjectExporter{},
		Template:                &TemplateExporter{},
		TemplateVault:           &TemplateVaultExporter{},
		TemplateRole:            &TemplateRoleExporter{},
		AccessKey:               &AccessKeyExporter{},
		Environment:             &EnvironmentExporter{},
		Repository:              &RepositoryExporter{},
		SecretStorage:           &SecretStorageExporter{},
		Inventory:               &InventoryExporter{},
		View:                    &ViewExporter{},
		Role:                    &RoleExporter{},
		Schedule:                &ScheduleExporter{},
		ProjectUser:             &ProjectUserExporter{},
		Integration:             &IntegrationExporter{},
		IntegrationExtractValue: &IntegrationExtractValueExporter{},
		IntegrationMatcher:      &IntegrationMatcherExporter{},
		IntegrationAlias:        &IntegrationAliasExporter{},
		Task:                    &TaskExporter{},
		TaskStage:               &TaskStageExporter{},
		Option:                  &OptionExporter{},
		Event:                   &EventExporter{},
		Runner:                  &RunnerExporter{},
	}

	if !skipTaskOutput {
		exporters[TaskOutput] = &TaskOutputExporter{}
	}

	return &ExporterChain{exporters: exporters, KeyMapper: mapper}
}

func NewKeyMapper() *TypeKeyMapper {
	return &TypeKeyMapper{Keys: make(map[string]map[string]map[EntityKey]EntityKey), IgnoreKeyNotFoundErr: true}
}

type ProgressBar struct {
	progress float32
	printer  func(float32)
}

func (p *ProgressBar) update(progress float32) {
	if progress-p.progress > 0.01 {
		p.updateForce(progress)
	}
}

func (p *ProgressBar) updateForce(progress float32) {
	p.printer(progress)
	p.progress = progress
}

func (p *ExporterChain) Load(store db.Store) (err error) {
	keys, err := getSortedKeys(p.exporters, func(t TypeExporter) []string {
		return t.exportDependsOn()
	})
	if err != nil {
		return
	}

	for _, name := range keys {
		progress := &ProgressBar{printer: func(progress float32) {
			strLen := len(name)
			spaces := fmt.Sprintf("%*s", 36-strLen, " ")

			fmt.Printf("\rExporting %s%s %d%%", name, spaces, int(progress*100))
		}, progress: 0}

		progress.updateForce(0)
		exporter := p.exporters[name]
		err = exporter.load(store, p, progress)
		if err != nil {
			return fmt.Errorf("failed to export %s: %s", name, err.Error())
		}
		progress.updateForce(1)
		fmt.Println()
	}
	return
}

func (p *ExporterChain) Restore(store db.Store, errLogSize int) error {
	keys, err := getSortedKeys(p.exporters, func(t TypeExporter) []string {
		return t.importDependsOn()
	})
	if err != nil {
		return err
	}

	for _, name := range keys {

		progress := &ProgressBar{printer: func(progress float32) {
			strLen := len(name)
			spaces := fmt.Sprintf("%*s", 36-strLen, " ")

			fmt.Printf("\rImporting %s%s %d%%", name, spaces, int(progress*100))
		}, progress: 0}

		progress.updateForce(0)
		exporter := p.exporters[name]
		err := exporter.restore(store, p, progress)
		if err != nil {
			return fmt.Errorf("failed to import %s: %s", name, err.Error())
		}
		progress.updateForce(1)
		fmt.Println()

		errCount := len(exporter.getErrors())
		if errCount > 0 {
			fmt.Printf("    Errors: %d\n", errCount)

			if errLogSize > 0 {
				for i, err := range exporter.getErrors() {
					if i > errLogSize {
						break
					}
					fmt.Println("      ", err)
				}
			}
		}
		exporter.clear()
	}

	return nil
}
