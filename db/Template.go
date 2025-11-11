package db

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type TemplateType string

const (
	TemplateTask   TemplateType = ""
	TemplateBuild  TemplateType = "build"
	TemplateDeploy TemplateType = "deploy"
)

type TemplateApp string

const (
	AppAnsible    TemplateApp = "ansible"
	AppTerraform  TemplateApp = "terraform"
	AppTofu       TemplateApp = "tofu"
	AppTerragrunt TemplateApp = "terragrunt"
	AppBash       TemplateApp = "bash"
	AppPowerShell TemplateApp = "powershell"
	AppPython     TemplateApp = "python"
	AppPulumi     TemplateApp = "pulumi"
)

func (t TemplateApp) InventoryTypes() []InventoryType {
	switch t {
	case AppAnsible:
		return []InventoryType{InventoryStatic, InventoryStaticYaml, InventoryFile}
	case AppTerraform:
		return []InventoryType{InventoryTerraformWorkspace}
	case AppTofu:
		return []InventoryType{InventoryTofuWorkspace}
	case AppTerragrunt:
		return []InventoryType{InventoryTerragruntWorkspace}
	default:
		return []InventoryType{}
	}
}

func (t TemplateApp) HasInventoryType(inventoryType InventoryType) bool {
	types := t.InventoryTypes()

	for _, typ := range types {
		if typ == inventoryType {
			return true
		}
	}

	return false
}

func (t TemplateApp) IsTerraform() bool {
	return t == AppTerraform || t == AppTofu || t == AppTerragrunt
}

type SurveyVarType string

const (
	SurveyVarStr    SurveyVarType = "str"
	SurveyVarInt    SurveyVarType = "int"
	SurveyVarEnum   SurveyVarType = "enum"
	SurveyVarSelect SurveyVarType = "select"
)

// SurveyVarDefaultValue supports both a single string or an array of strings in JSON.
// It preserves whether the original JSON was an array so encoding will keep the
// original shape when possible (single value -> string, multiple -> array).
type SurveyVarDefaultValue struct {
	Values           []string `json:"-"`
	originalWasArray bool     `json:"-"`
}

func (d *SurveyVarDefaultValue) UnmarshalJSON(b []byte) error {
	if len(bytes.TrimSpace(b)) == 0 || bytes.Equal(bytes.TrimSpace(b), []byte("null")) {
		d.Values = nil
		d.originalWasArray = false
		return nil
	}

	// try string
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		d.Values = []string{s}
		d.originalWasArray = false
		return nil
	}

	// try []string
	var arr []string
	if err := json.Unmarshal(b, &arr); err == nil {
		d.Values = arr
		d.originalWasArray = true
		return nil
	}

	return fmt.Errorf("invalid default_value: must be string or []string")
}

func (d SurveyVarDefaultValue) MarshalJSON() ([]byte, error) {
	if d.Values == nil {
		return []byte("null"), nil
	}
	if len(d.Values) == 1 && !d.originalWasArray {
		return json.Marshal(d.Values[0])
	}
	return json.Marshal(d.Values)
}

func (d SurveyVarDefaultValue) String() string {
	if len(d.Values) == 0 {
		return ""
	}
	return d.Values[0]
}

type AnsibleTemplateParams struct {
	AllowDebug             bool     `json:"allow_debug"`
	AllowOverrideInventory bool     `json:"allow_override_inventory"`
	AllowOverrideLimit     bool     `json:"allow_override_limit"`
	AllowOverrideTags      bool     `json:"allow_override_tags"`
	AllowOverrideSkipTags  bool     `json:"allow_override_skip_tags"`
	Limit                  []string `json:"limit"`
	Tags                   []string `json:"tags"`
	SkipTags               []string `json:"skip_tags"`
}

type TerraformTemplateParams struct {
	AllowDestroy     bool   `json:"allow_destroy,omitempty"`
	AllowAutoApprove bool   `json:"allow_auto_approve,omitempty"`
	AutoApprove      bool   `json:"auto_approve,omitempty"`
	OverrideBackend  bool   `json:"override_backend,omitempty"` // override backend if internal backend is used
	BackendFilename  string `json:"backend_filename,omitempty"`
}

type SurveyVarEnumValue struct {
	Name  string `json:"name" backup:"name"`
	Value string `json:"value" backup:"value"`
}

type SurveyVar struct {
	Name         string                 `json:"name" backup:"name"`
	Title        string                 `json:"title" backup:"title"`
	Required     bool                   `json:"required,omitempty" backup:"required"`
	Type         SurveyVarType          `json:"type,omitempty" backup:"type"`
	Description  string                 `json:"description,omitempty" backup:"description"`
	Values       []SurveyVarEnumValue   `json:"values,omitempty" backup:"values"`
	DefaultValue *SurveyVarDefaultValue `json:"default_value,omitempty" backup:"default_value"`
}

type TemplateFilter struct {
	ViewID          *int
	BuildTemplateID *int
	AutorunOnly     bool
	App             *TemplateApp
}

// Template is a user defined model that is used to run a task
type Template struct {
	ID int `db:"id" json:"id" backup:"-"`

	ProjectID     int  `db:"project_id" json:"project_id" backup:"-"`
	InventoryID   *int `db:"inventory_id" json:"inventory_id,omitempty" backup:"-"`
	RepositoryID  int  `db:"repository_id" json:"repository_id" backup:"-"`
	EnvironmentID *int `db:"environment_id" json:"environment_id,omitempty" backup:"-"`

	// Name as described in https://github.com/semaphoreui/semaphore/issues/188
	Name string `db:"name" json:"name"`
	// playbook name in the form of "some_play.yml"
	Playbook string `db:"playbook" json:"playbook"`
	// to fit into []string
	Arguments *string `db:"arguments" json:"arguments,omitempty"`
	// if true, semaphore will not prepend any arguments to `arguments` like inventory, etc
	AllowOverrideArgsInTask bool `db:"allow_override_args_in_task" json:"allow_override_args_in_task,omitempty"`

	Description *string `db:"description" json:"description,omitempty"`

	Vaults []TemplateVault `db:"-" json:"vaults,omitempty" backup:"-"`

	Type            TemplateType `db:"type" json:"type,omitempty"`
	StartVersion    *string      `db:"start_version" json:"start_version,omitempty"`
	BuildTemplateID *int         `db:"build_template_id" json:"build_template_id,omitempty" backup:"-"`

	ViewID *int `db:"view_id" json:"view_id,omitempty" backup:"-"`

	LastTask *TaskWithTpl `db:"-" json:"last_task,omitempty" backup:"-"`

	Autorun bool `db:"autorun" json:"autorun,omitempty"`

	// override variables
	GitBranch *string `db:"git_branch" json:"git_branch,omitempty"`

	// SurveyVarsJSON used internally for read from database.
	// It is not used for store survey vars to database.
	// Do not use it in your code. Use SurveyVars instead.
	SurveyVarsJSON *string     `db:"survey_vars" json:"-" backup:"-"`
	SurveyVars     []SurveyVar `db:"-" json:"survey_vars,omitempty" backup:"survey_vars"`

	SuppressSuccessAlerts bool `db:"suppress_success_alerts" json:"suppress_success_alerts,omitempty"`

	App TemplateApp `db:"app" json:"app,omitempty"`

	Tasks int `db:"tasks" json:"tasks" backup:"-"`

	TaskParams MapStringAnyField `db:"task_params" json:"task_params,omitempty"`

	RunnerTag *string `db:"runner_tag" json:"runner_tag,omitempty"`

	AllowOverrideBranchInTask bool `db:"allow_override_branch_in_task" json:"allow_override_branch_in_task,omitempty"`
	AllowParallelTasks        bool `db:"allow_parallel_tasks" json:"allow_parallel_tasks,omitempty"`
}

type TemplateWithPerms struct {
	Template
	Permissions *ProjectUserPermission `db:"permissions" json:"permissions"`
}

func (tpl *Template) FillParams(target any) error {
	content, err := json.Marshal(tpl.TaskParams)
	if err != nil {
		return nil
	}
	err = json.Unmarshal(content, target)
	return err
}

func (tpl *Template) CanOverrideInventory() (ok bool, err error) {
	switch tpl.App {
	case AppAnsible, "":
		var params AnsibleTemplateParams
		err = tpl.FillParams(&params)
		if err != nil {
			return
		}
		ok = params.AllowOverrideInventory
	}

	return
}

func (tpl *Template) Validate() error {
	if tpl.RunnerTag != nil && *tpl.RunnerTag == "" {
		return &ValidationError{"template runner tag can not be empty"}
	}
	switch tpl.App {
	case AppAnsible:
		if tpl.InventoryID == nil {
			return &ValidationError{"template inventory can not be empty"}
		}
	}

	if tpl.Name == "" {
		return &ValidationError{"template name can not be empty"}
	}

	if !tpl.App.IsTerraform() && tpl.Playbook == "" {
		return &ValidationError{"template playbook can not be empty"}
	}

	if tpl.Arguments != nil {
		if !json.Valid([]byte(*tpl.Arguments)) {
			return &ValidationError{"template arguments must be valid JSON"}
		}
	}

	return nil
}

func FillTemplate(d Store, template *Template) (err error) {
	var vaults []TemplateVault
	vaults, err = d.GetTemplateVaults(template.ProjectID, template.ID)
	if err != nil {
		return
	}
	template.Vaults = vaults

	var tasks []TaskWithTpl
	tasks, err = d.GetTemplateTasks(template.ProjectID, template.ID, RetrieveQueryParams{Count: 1})
	if err != nil {
		return
	}
	if len(tasks) > 0 {
		template.LastTask = &tasks[0]
	}

	if template.SurveyVarsJSON != nil {
		err = json.Unmarshal([]byte(*template.SurveyVarsJSON), &template.SurveyVars)
	}

	return
}
