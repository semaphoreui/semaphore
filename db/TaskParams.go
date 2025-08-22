package db

type TaskParams struct {
	ID        int `db:"id" json:"-" backup:"-"`
	ProjectID int `db:"project_id" json:"-" backup:"-"`

	Environment string  `db:"environment" json:"environment,omitempty"`
	Arguments   *string `db:"arguments" json:"arguments,omitempty"`
	GitBranch   *string `db:"git_branch" json:"git_branch,omitempty"`

	Message string `db:"message" json:"message,omitempty"`

	// Version is a build version.
	// This field available only for Build tasks.
	Version *string `db:"version" json:"version,omitempty"`

	InventoryID   *int    `db:"inventory_id" json:"inventory_id,omitempty" backup:"-"`
	InventoryName *string `db:"-" json:"-" backup:"inventory_name"`

	Params MapStringAnyField `db:"params" json:"params,omitempty"`
}

func (p TaskParams) CreateTask(templateID int) (task Task) {
	task = Task{
		ProjectID:   p.ProjectID,
		Environment: p.Environment,
		Arguments:   p.Arguments,
		GitBranch:   p.GitBranch,
		Message:     p.Message,
		Version:     p.Version,
		InventoryID: p.InventoryID,
		Params:      p.Params,
		TemplateID:  templateID,
	}

	return
}
