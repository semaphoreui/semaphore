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
	TemplateAnsible   = ""
	TemplateTerraform = "terraform"
	TemplateBash      = "bash"
)

type SurveyVarType string

const (
	SurveyVarStr TemplateType = ""
	SurveyVarInt TemplateType = "int"
)

type SurveyVar struct {
	Name        string        `json:"name"`
	Title       string        `json:"title"`
	Required    bool          `json:"required"`
	Type        SurveyVarType `json:"type"`
	Description string        `json:"description"`
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

	VaultKeyID *int      `db:"vault_key_id" json:"vault_key_id"`
	VaultKey   AccessKey `db:"-" json:"-"`

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
}

func (tpl *Template) Validate() error {
	if tpl.Name == "" {
		return &ValidationError{"template name can not be empty"}
	}

	if tpl.App != TemplateTerraform && tpl.Playbook == "" {
		return &ValidationError{"template playbook can not be empty"}
	}

	if tpl.Arguments != nil {
		if !json.Valid([]byte(*tpl.Arguments)) {
			return &ValidationError{"template arguments must be valid JSON"}
		}
	}

	return nil
}

func FillTemplates(d Store, templates []Template) (err error) {
	for i := range templates {
		tpl := &templates[i]
		var tasks []TaskWithTpl
		tasks, err = d.GetTemplateTasks(tpl.ProjectID, tpl.ID, RetrieveQueryParams{Count: 1})
		if err != nil {
			return
		}
		if len(tasks) > 0 {
			tpl.LastTask = &tasks[0]
		}
	}

	return
}

func FillTemplate(d Store, template *Template) (err error) {
	if template.VaultKeyID != nil {
		template.VaultKey, err = d.GetAccessKey(template.ProjectID, *template.VaultKeyID)
	}

	if err != nil {
		return
	}

	err = FillTemplates(d, []Template{*template})

	if err != nil {
		return
	}

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
