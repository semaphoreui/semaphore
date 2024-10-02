package db

import (
	"encoding/json"
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
	AppBash       TemplateApp = "bash"
	AppPowerShell TemplateApp = "powershell"
	AppPython     TemplateApp = "python"
	AppPulumi     TemplateApp = "pulumi"
)

func (t TemplateApp) IsTerraform() bool {
	return t == AppTerraform || t == AppTofu
}

type SurveyVarType string

const (
	SurveyVarStr  TemplateType = ""
	SurveyVarInt  TemplateType = "int"
	SurveyVarEnum TemplateType = "enum"
)

type SurveyVarEnumValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type SurveyVar struct {
	Name        string               `json:"name"`
	Title       string               `json:"title"`
	Required    bool                 `json:"required"`
	Type        SurveyVarType        `json:"type"`
	Description string               `json:"description"`
	Values      []SurveyVarEnumValue `json:"values"`
}

type TemplateFilter struct {
	ViewID          *int
	BuildTemplateID *int
	AutorunOnly     bool
}

// Template is a user defined model that is used to run a task
type Template struct {
	ID int `db:"id" json:"id"`

	ProjectID     int  `db:"project_id" json:"project_id"`
	InventoryID   *int `db:"inventory_id" json:"inventory_id"`
	RepositoryID  int  `db:"repository_id" json:"repository_id"`
	EnvironmentID *int `db:"environment_id" json:"environment_id"`

	// Name as described in https://github.com/ansible-semaphore/semaphore/issues/188
	Name string `db:"name" json:"name"`
	// playbook name in the form of "some_play.yml"
	Playbook string `db:"playbook" json:"playbook"`
	// to fit into []string
	Arguments *string `db:"arguments" json:"arguments"`
	// if true, semaphore will not prepend any arguments to `arguments` like inventory, etc
	AllowOverrideArgsInTask bool `db:"allow_override_args_in_task" json:"allow_override_args_in_task"`

	Description *string `db:"description" json:"description"`

	Vaults []TemplateVault `db:"-" json:"vaults"`

	Type            TemplateType `db:"type" json:"type"`
	StartVersion    *string      `db:"start_version" json:"start_version"`
	BuildTemplateID *int         `db:"build_template_id" json:"build_template_id"`

	ViewID *int `db:"view_id" json:"view_id"`

	LastTask *TaskWithTpl `db:"-" json:"last_task"`

	Autorun bool `db:"autorun" json:"autorun"`

	// SurveyVarsJSON used internally for read from database.
	// It is not used for store survey vars to database.
	// Do not use it in your code. Use SurveyVars instead.
	SurveyVarsJSON *string     `db:"survey_vars" json:"-"`
	SurveyVars     []SurveyVar `db:"-" json:"survey_vars"`

	SuppressSuccessAlerts bool `db:"suppress_success_alerts" json:"suppress_success_alerts"`

	App TemplateApp `db:"app" json:"app"`

	Tasks int `db:"tasks" json:"tasks"`
}

func (tpl *Template) Validate() error {
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
