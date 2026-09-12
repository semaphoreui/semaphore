package sql

import (
	"strconv"
	"strings"

	"github.com/go-gorp/gorp/v3"
	"github.com/semaphoreui/semaphore/db"
)

func (d *SqlDb) GetTemplateVaults(projectID int, templateID int) (vaults []db.TemplateVault, err error) {
	vaults = []db.TemplateVault{}

	_, err = d.selectAll(&vaults, "select * from project__template_vault where project_id=? and template_id=?", projectID, templateID)
	if err != nil {
		return
	}
	for i := range vaults {
		err = db.FillTemplateVault(d, projectID, &vaults[i])
		if err != nil {
			return
		}
	}
	return
}

func (d *SqlDb) CreateTemplateVault(vault db.TemplateVault) (newVault db.TemplateVault, err error) {
	insertID, err := d.insert(
		"id",
		"insert into project__template_vault (project_id, template_id, vault_key_id, name, type, script) values (?, ?, ?, ?, ?, ?)",
		vault.ProjectID,
		vault.TemplateID,
		vault.VaultKeyID,
		vault.Name,
		vault.Type,
		vault.Script)
	if err != nil {
		return
	}

	newVault = vault
	newVault.ID = insertID
	return
}

func (d *SqlDb) UpdateTemplateVaults(projectID int, templateID int, vaults []db.TemplateVault) (err error) {
	tx, err := d.Sql().Begin()
	if err != nil {
		return
	}
	if err = d.updateTemplateVaultsInTx(tx, projectID, templateID, vaults); err != nil {
		_ = tx.Rollback()
		return
	}
	return tx.Commit()
}

func (d *SqlDb) updateTemplateVaultsInTx(tx *gorp.Transaction, projectID int, templateID int, vaults []db.TemplateVault) (err error) {
	if vaults == nil {
		vaults = []db.TemplateVault{}
	}

	var vaultIDs []string
	for _, vault := range vaults {
		switch vault.Type {
		case "password":
			vault.Script = nil
		case "script":
			vault.VaultKeyID = nil
		}
		if vault.ID == 0 {
			var vaultId int
			vaultId, err = d.insertTx(tx, "id", "insert into project__template_vault (project_id, template_id, vault_key_id, name, type, script) values (?, ?, ?, ?, ?, ?)", projectID, templateID, vault.VaultKeyID, vault.Name, vault.Type, vault.Script)
			if err != nil {
				return
			}
			vaultIDs = append(vaultIDs, strconv.Itoa(vaultId))
		} else {
			_, err = d.execTx(tx, "update project__template_vault set vault_key_id=?, name=?, type=?, script=? where id=? and project_id=? and template_id=?", vault.VaultKeyID, vault.Name, vault.Type, vault.Script, vault.ID, projectID, templateID)
			vaultIDs = append(vaultIDs, strconv.Itoa(vault.ID))
		}
		if err != nil {
			return
		}
	}

	if len(vaultIDs) == 0 {
		_, err = d.execTx(tx, "delete from project__template_vault where project_id=? and template_id=?", projectID, templateID)
	} else {
		_, err = d.execTx(tx, "delete from project__template_vault where project_id=? and template_id=? and id not in ("+strings.Join(vaultIDs, ",")+")", projectID, templateID)
	}

	return
}
