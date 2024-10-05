package bolt

import (
	"github.com/ansible-semaphore/semaphore/db"
	"go.etcd.io/bbolt"
	"slices"
)

func (d *BoltDb) GetTemplateVaults(projectID int, templateID int) (vaults []db.TemplateVault, err error) {
	err = d.getObjects(projectID, db.TemplateVaultProps, db.RetrieveQueryParams{}, func(referringObj interface{}) bool {
		return referringObj.(db.TemplateVault).TemplateID == templateID
	}, &vaults)
	if err != nil {
		return
	}
	for _, vault := range vaults {
		err = db.FillTemplateVault(d, projectID, &vault)
		if err != nil {
			return
		}
	}
	return
}

func (d *BoltDb) CreateTemplateVault(vault db.TemplateVault) (newVault db.TemplateVault, err error) {
	var newTpl interface{}
	newTpl, err = d.createObject(vault.ProjectID, db.TemplateVaultProps, vault)
	if err != nil {
		return
	}
	newVault = newTpl.(db.TemplateVault)
	return
}

func (d *BoltDb) UpdateTemplateVaults(projectID int, templateID int, vaults []db.TemplateVault) (err error) {
	if vaults == nil {
		vaults = []db.TemplateVault{}
	}

	var oldVaults []db.TemplateVault
	oldVaults, err = d.GetTemplateVaults(projectID, templateID)

	var vaultIDs []int
	for _, vault := range vaults {
		vault.ProjectID = projectID
		vault.TemplateID = templateID
		if vault.ID == 0 {
			// Insert new vaults
			var newTpl interface{}
			newTpl, err = d.createObject(projectID, db.TemplateVaultProps, vault)
			if err != nil {
				return
			}
			vaultIDs = append(vaultIDs, newTpl.(db.TemplateVault).ID)
		} else {
			// Update existing vaults
			err = d.updateObject(projectID, db.TemplateVaultProps, vault)
			vaultIDs = append(vaultIDs, vault.ID)
		}
		if err != nil {
			return
		}
	}

	// Delete missing vaults
	for _, vault := range oldVaults {
		if !slices.Contains(vaultIDs, vault.ID) {
			err = d.db.Update(func(tx *bbolt.Tx) error {
				return d.deleteObject(projectID, db.TemplateVaultProps, intObjectID(vault.ID), tx)
			})
			if err != nil {
				return
			}
		}
	}

	return
}
