package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/semaphoreui/semaphore/pkg/git"
	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
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
	SurveyVarStr    SurveyVarType = ""
	SurveyVarInt    SurveyVarType = "int"
	SurveyVarEnum   SurveyVarType = "enum"
	SurveyVarText   SurveyVarType = "text"
	SurveyVarSelect SurveyVarType = "select"
)

type SurveyVarTarget string

const (
	// SurveyVarTargetDefault passes the variable the app-specific way:
	// --extra-vars for Ansible, -var for Terraform apps, CLI argument for shell apps.
	SurveyVarTargetDefault SurveyVarTarget = ""
	// SurveyVarTargetEnv passes the variable as a process environment variable.
	SurveyVarTargetEnv SurveyVarTarget = "env"
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

// IsArray reports whether the value was decoded from a JSON array.
// Used by ValidateSurveyVar to enforce type/default_value compatibility.
func (d SurveyVarDefaultValue) IsArray() bool {
	return d.originalWasArray
}

// ValidateSurveyVar enforces compatibility between a SurveyVar's Type and
// its DefaultValue. The custom SurveyVarDefaultValue codec preserves the
// original JSON shape (string vs []string), which means a client can submit
// a default_value that does not match the declared type (e.g. an array for
// an "int" var, or a scalar for a "select" var). Without this check, bad
// data lands in the DB and surfaces as UI glitches or runtime errors much
// later.
//
// Rules:
//   - For SurveyVarSelect: default_value must be array-shaped (or nil).
//     A single scalar string is accepted and normalised to [scalar] to
//     stay backward-compatible with clients that predate the select type.
//   - For all other types: default_value must be scalar-shaped (or nil).
//     An array with exactly one element is accepted and the caller is
//     expected to read it via .String(); an array with >1 element is
//     rejected.
//   - For SurveyVarEnum and SurveyVarSelect: every value in default_value
//     must be present in the var's Values list (matched by Value field).
func ValidateSurveyVar(v SurveyVar) error {
	switch v.Type {
	case SurveyVarSelect:
		if v.DefaultValue != nil {
			if !v.DefaultValue.IsArray() {
				// Accept legacy scalar string for backward compat;
				// it is normalised into [scalar] by the caller on save.
				if len(v.DefaultValue.Values) > 1 {
					return common_errors.NewValidationError(
						"survey variable \"" + v.Name + "\": default_value must be an array for select type")
				}
			}
			// Verify every default value is present in Values.
			allowed := make(map[string]struct{}, len(v.Values))
			for _, ev := range v.Values {
				allowed[ev.Value] = struct{}{}
			}
			for _, dv := range v.DefaultValue.Values {
				if _, ok := allowed[dv]; !ok {
					return common_errors.NewValidationError(
						"survey variable \"" + v.Name + "\": default_value \"" + dv + "\" is not in values list")
				}
			}
		}
	case SurveyVarEnum:
		if v.DefaultValue != nil {
			if v.DefaultValue.IsArray() && len(v.DefaultValue.Values) > 1 {
				return common_errors.NewValidationError(
					"survey variable \"" + v.Name + "\": default_value must be a string for enum type")
			}
			if len(v.DefaultValue.Values) > 0 {
				allowed := make(map[string]struct{}, len(v.Values))
				for _, ev := range v.Values {
					allowed[ev.Value] = struct{}{}
				}
				dv := v.DefaultValue.Values[0]
				if _, ok := allowed[dv]; !ok {
					return common_errors.NewValidationError(
						"survey variable \"" + v.Name + "\": default_value \"" + dv + "\" is not in values list")
				}
			}
		}
	default:
		// String, int, text, secret: scalar only.
		if v.DefaultValue != nil && v.DefaultValue.IsArray() && len(v.DefaultValue.Values) > 1 {
			return common_errors.NewValidationError(
				"survey variable \"" + v.Name + "\": default_value must be a string for type \"" + string(v.Type) + "\"")
		}
	}
	return nil
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

	// SkipGalaxyInstall skips the Galaxy install step (role and collection
	// requirements) before running the playbook.
	SkipGalaxyInstall bool `json:"skip_galaxy_install"`
	// AllowOverrideSkipGalaxyInstall lets the user toggle SkipGalaxyInstall when
	// launching a task.
	AllowOverrideSkipGalaxyInstall bool `json:"allow_override_skip_galaxy_install"`
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
	Target       SurveyVarTarget        `json:"target,omitempty" backup:"target"`
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

	ProjectID    int  `db:"project_id" json:"project_id" backup:"-"`
	InventoryID  *int `db:"inventory_id" json:"inventory_id,omitempty" backup:"-"`
	RepositoryID int  `db:"repository_id" json:"repository_id" backup:"-"`

	// EnvironmentIDs is the list of Variable Groups (environments) used by the
	// template. At task run time their JSON, ENV vars, and secrets are merged
	// into a single environment, with later entries overriding earlier ones.
	// Persisted via the project__template_environment junction table in SQL,
	// and serialized inline on the template object in BoltDB.
	EnvironmentIDs []int `db:"-" json:"environment_ids" backup:"-"`

	// EnvironmentID is the ID of the environment associated with the template.
	// Deprecated: Use EnvironmentIDs instead.
	EnvironmentID int `db:"-" json:"environment_id" backup:"-"`

	// Name as described in https://github.com/semaphoreui/semaphore/issues/188
	Name string `db:"name" json:"name"`
	// playbook name in the form of "some_play.yml"
	Playbook string `db:"playbook" json:"playbook"`
	// WorkingDirectory is the repository-relative current directory for Ansible
	// commands. It is valid only for Ansible templates.
	WorkingDirectory *string `db:"working_directory" json:"working_directory,omitempty"`
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

	// ExecutorImage overrides the container image the runner uses to run this
	// template's tasks. Only the container-based executors (Docker, Kubernetes)
	// honour it; the local executor ignores it. Empty/nil means "use the image
	// from the runner configuration".
	ExecutorImage *string `db:"executor_image" json:"executor_image,omitempty"`

	AllowOverrideBranchInTask bool `db:"allow_override_branch_in_task" json:"allow_override_branch_in_task,omitempty"`
	//AllowOverrideEnvInTask    bool `db:"allow_override_env_in_task" json:"allow_override_env_in_task,omitempty"`
	AllowParallelTasks bool `db:"allow_parallel_tasks" json:"allow_parallel_tasks,omitempty"`

	JWTParams *TemplateJWTParams `db:"jwt_params" json:"jwt_params,omitempty"`
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

// NormalizedExecutorImage returns the value to persist. Clearing the field in the
// WebUI sends an empty string; storing it as NULL keeps "no override" a single
// representation in the database.
func (tpl *Template) NormalizedExecutorImage() *string {
	if tpl.ExecutorImage == nil {
		return nil
	}

	img := strings.TrimSpace(*tpl.ExecutorImage)

	if img == "" {
		return nil
	}

	return &img
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
		return common_errors.NewValidationError("template runner tag can not be empty")
	}

	// Reject apps that are not in the administrator-configured whitelist, otherwise
	// an unknown app becomes a ShellApp that executes string(App) as a system binary
	// (arbitrary command execution). util.Config is nil in some unit tests; an empty
	// app is the legacy default and runs no command, so both are skipped.
	if tpl.App != "" {
		if _, ok := util.Config.Apps[string(tpl.App)]; !ok {
			return common_errors.NewValidationError("invalid app: " + string(tpl.App))
		}
	}

	switch tpl.App {
	case AppAnsible:
		if tpl.InventoryID == nil {
			return common_errors.NewValidationError("template inventory can not be empty")
		}
	}

	if tpl.Name == "" {
		return common_errors.NewValidationError("template name can not be empty")
	}

	if !tpl.App.IsTerraform() && tpl.Playbook == "" {
		return common_errors.NewValidationError("template playbook can not be empty")
	}

	if err := ValidatePlaybookPath(tpl.Playbook, "template"); err != nil {
		return err
	}

	if tpl.WorkingDirectory != nil {
		if tpl.App != AppAnsible {
			return common_errors.NewValidationError("template working directory is supported only for Ansible templates")
		}
		if strings.TrimSpace(*tpl.WorkingDirectory) == "" {
			return common_errors.NewValidationError("template working directory can not be empty")
		}
		if err := ValidateWorkingDirectoryLexically(*tpl.WorkingDirectory); err != nil {
			return err
		}
	}

	if tpl.Arguments != nil {
		if !json.Valid([]byte(*tpl.Arguments)) {
			return common_errors.NewValidationError("template arguments must be valid JSON")
		}
	}

	if tpl.GitBranch != nil {
		if err := git.ValidateGitBranch(*tpl.GitBranch, "template"); err != nil {
			return err
		}
	}

	if err := tpl.JWTParams.Validate(); err != nil {
		return err
	}

	for _, v := range tpl.SurveyVars {
		switch v.Target {
		case SurveyVarTargetDefault, SurveyVarTargetEnv:
		default:
			return &common_errors.ValidationError{Message: "invalid survey variable target: " + string(v.Target)}
		}

		if err := ValidateSurveyVar(v); err != nil {
			return err
		}
	}

	return nil
}

// ApplyLegacyEnvironmentField copies deprecated environment_id into environment_ids when
// the client omitted environment_ids (nil). An explicit empty JSON array unmarshals as a
// non-nil empty slice and is left unchanged so clients can clear all variable groups.
func (tpl *Template) ApplyLegacyEnvironmentField() {
	if tpl.EnvironmentIDs == nil && tpl.EnvironmentID > 0 {
		tpl.EnvironmentIDs = []int{tpl.EnvironmentID}
	}
}

func FillTemplate(d Store, template *Template) (err error) {
	var vaults []TemplateVault
	vaults, err = d.GetTemplateVaults(template.ProjectID, template.ID)
	if err != nil {
		return
	}
	template.Vaults = vaults

	var envIDs []int
	envIDs, err = d.GetTemplateEnvironments(template.ProjectID, template.ID)
	if err != nil {
		return
	}
	template.EnvironmentIDs = envIDs

	var tasks []TaskWithTpl
	tasks, err = d.GetTemplateTasks(template.ProjectID, template.ID, RetrieveQueryParams{Count: 1})
	if err != nil {
		return
	}
	if len(tasks) > 0 {
		template.LastTask = &tasks[0]
	}

	if template.SurveyVarsJSON != nil {
		if err2 := json.Unmarshal([]byte(*template.SurveyVarsJSON), &template.SurveyVars); err2 != nil {
			log.WithFields(log.Fields{
				"context":     common_errors.GetErrorContext(),
				"project_id":  &template.ProjectID,
				"template_id": template.ID,
				"hint":        "validate JSON array in project__template.survey_vars",
			}).Error("failed to unmarshal template survey vars")
		}
	}

	// For backward compatibility
	if len(template.EnvironmentIDs) > 0 {
		template.EnvironmentID = template.EnvironmentIDs[0]
	}

	return
}
