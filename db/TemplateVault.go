package db

type TemplateVault struct {
	ID         int     `db:"id" json:"id" backup:"-"`
	ProjectID  int     `db:"project_id" json:"project_id" backup:"-"`
	TemplateID int     `db:"template_id" json:"template_id" backup:"-"`
	VaultKeyID int     `db:"vault_key_id" json:"vault_key_id" backup:"-"`
	Name       *string `db:"name" json:"name"`

	Vault *AccessKey `db:"-" json:"-"`
}

func FillTemplateVault(d Store, projectID int, templateVault *TemplateVault) (err error) {
	var vault AccessKey
	vault, err = d.GetAccessKey(projectID, templateVault.VaultKeyID)
	if err != nil {
		return
	}
	templateVault.Vault = &vault
	return
}
