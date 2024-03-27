package db

type InventoryType string

const (
	//InventoryNone       InventoryType = "none"
	InventoryStatic     InventoryType = "static"
	InventoryStaticYaml InventoryType = "static-yaml"
	// InventoryFile means that it is path to the Ansible inventory file
	InventoryFile               InventoryType = "file"
	InventoryTerraformWorkspace InventoryType = "terraform-workspace"
)

// Inventory is the model of an ansible inventory file
type Inventory struct {
	ID        int    `db:"id" json:"id"`
	Name      string `db:"name" json:"name" binding:"required"`
	ProjectID int    `db:"project_id" json:"project_id"`
	Inventory string `db:"inventory" json:"inventory"`

	// accesses hosts in inventory
	SSHKeyID *int      `db:"ssh_key_id" json:"ssh_key_id"`
	SSHKey   AccessKey `db:"-" json:"-"`

	BecomeKeyID *int      `db:"become_key_id" json:"become_key_id"`
	BecomeKey   AccessKey `db:"-" json:"-"`

	// static/file
	Type InventoryType `db:"type" json:"type"`

	// HolderID is an ID of template which holds the inventory
	HolderID *int `db:"holder_id" json:"holder_id"`
}

func FillInventory(d Store, inventory *Inventory) (err error) {
	if inventory.SSHKeyID != nil {
		inventory.SSHKey, err = d.GetAccessKey(inventory.ProjectID, *inventory.SSHKeyID)
	}

	if err != nil {
		return
	}

	if inventory.BecomeKeyID != nil {
		inventory.BecomeKey, err = d.GetAccessKey(inventory.ProjectID, *inventory.BecomeKeyID)
	}

	return
}
