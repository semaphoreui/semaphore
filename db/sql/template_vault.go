package sql

import (
	"github.com/ansible-semaphore/semaphore/db"
	"strconv"
	"strings"
)

func (d *SqlDb) GetTemplateVaults(projectID int, templateID int) (vaults []db.TemplateVault, err error) {
	vaults = []db.TemplateVault{}

	var vlts []db.TemplateVault
	_, err = d.selectAll(&vlts, "select * from project__template_vault where project_id=? and template_id=?", projectID, templateID)
	if err != nil {
		return
	}
	for _, vault := range vlts {
		vault := vault
		err = db.FillTemplateVault(d, projectID, &vault)
		if err != nil {
			return
		}
		vaults = append(vaults, vault)
	}
	return
}

func (d *SqlDb) CreateTemplateVault(vault db.TemplateVault) (newVault db.TemplateVault, err error) {
	insertID, err := d.insert(
		"id",
		"insert into project__template_vault (project_id, template_id, vault_key_id, name) values (?, ?, ?, ?)",
		vault.ProjectID,
		vault.TemplateID,
		vault.VaultKeyID,
		vault.Name)
	if err != nil {
		return
	}

	newVault = vault
	newVault.ID = insertID
	return
}

func (d *SqlDb) UpdateTemplateVaults(projectID int, templateID int, vaults []db.TemplateVault) (err error) {
	if vaults == nil {
		vaults = []db.TemplateVault{}
	}

	var vaultIDs []string
	for _, vault := range vaults {
		if vault.ID == 0 {
			// Insert new vaults
			var vaultId int
			vaultId, err = d.insert("id", "insert into project__template_vault (project_id, template_id, vault_key_id, name) values (?, ?, ?, ?)", projectID, templateID, vault.VaultKeyID, vault.Name)
			if err != nil {
				return
			}
			vaultIDs = append(vaultIDs, strconv.Itoa(vaultId))
		} else {
			// Update existing vaults
			_, err = d.exec("update project__template_vault set project_id=?, template_id=?, vault_key_id=?, name=? where id=?", projectID, templateID, vault.VaultKeyID, vault.Name, vault.ID)
			vaultIDs = append(vaultIDs, strconv.Itoa(vault.ID))
		}
		if err != nil {
			return
		}
	}

	// Delete removed vaults
	if len(vaultIDs) == 0 {
		_, err = d.exec("delete from project__template_vault where project_id=? and template_id=?", projectID, templateID)
	} else {
		_, err = d.exec("delete from project__template_vault where project_id=? and template_id=? and id not in ("+strings.Join(vaultIDs, ",")+")", projectID, templateID)
	}

	return
}
