package db

type InventoryType string

const (
	InventoryStatic     InventoryType = "static"
	InventoryStaticYaml InventoryType = "static-yaml"
	// InventoryFile means that it is path to the Ansible inventory file
	InventoryFile                InventoryType = "file"
	InventoryTerraformWorkspace  InventoryType = "terraform-workspace"
	InventoryTofuWorkspace       InventoryType = "tofu-workspace"
	InventoryTerragruntWorkspace InventoryType = "terragrunt-workspace"
)

func (i InventoryType) IsStatic() bool {
	return i == InventoryStatic || i == InventoryStaticYaml
}

// Inventory is the model of an ansible inventory file
type Inventory struct {
	ID        int    `db:"id" json:"id" backup:"-"`
	Name      string `db:"name" json:"name" binding:"required"`
	ProjectID int    `db:"project_id" json:"project_id" backup:"-"`
	Inventory string `db:"inventory" json:"inventory"`

	// accesses hosts in inventory
	SSHKeyID *int      `db:"ssh_key_id" json:"ssh_key_id" backup:"-"`
	SSHKey   AccessKey `db:"-" json:"-" backup:"-"`

	BecomeKeyID *int      `db:"become_key_id" json:"become_key_id" backup:"-"`
	BecomeKey   AccessKey `db:"-" json:"-" backup:"-"`

	// static/file
	Type InventoryType `db:"type" json:"type"`

	// TemplateID is an ID of template which holds the inventory
	// It is not used now but can be used in feature for
	// inventories which can not be used more than one template
	// at once.
	TemplateID *int `db:"template_id" json:"template_id" backup:"-"`

	// RepositoryID is an ID of repo where inventory stored.
	// If null than inventory will be got from template repository.
	RepositoryID *int        `db:"repository_id" json:"repository_id" backup:"-"`
	Repository   *Repository `db:"-" json:"-" backup:"-"`

	// RunnerTag is a tag which allow join inventory to the runner.
	RunnerTag *string `db:"runner_tag" json:"runner_tag,omitempty"`
}

func (e Inventory) GetFilename() string {
	if e.Type != InventoryFile {
		return ""
	}

	return e.Inventory

	//return strings.TrimPrefix(e.Inventory, "/")
}

func (e Inventory) Validate() error {
	if e.RunnerTag == nil && *e.RunnerTag == "" {
		return &ValidationError{"template runner tag can not be empty"}
	}

	return nil
}
