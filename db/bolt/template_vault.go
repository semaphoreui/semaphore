package bolt

import (
	"github.com/semaphoreui/semaphore/db"
	"go.etcd.io/bbolt"
)

func (d *BoltDb) GetTemplateVaults(projectID int, templateID int) (vaults []db.TemplateVault, err error) {
	err = d.getObjects(projectID, db.TemplateVaultProps, db.RetrieveQueryParams{}, func(referringObj interface{}) bool {
		return referringObj.(db.TemplateVault).TemplateID == templateID
	}, &vaults)
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

	err = d.db.Update(func(tx *bbolt.Tx) error {
		for _, vault := range oldVaults {
			err = d.deleteObject(projectID, db.TemplateVaultProps, intObjectID(vault.ID), tx)
			if err != nil {
				return err
			}
		}

		for _, vault := range vaults {
			vault.ProjectID = projectID
			vault.TemplateID = templateID

			switch vault.Type {
			case "password":
				vault.Script = nil
			case "script":
				vault.VaultKeyID = nil
			}

			_, err = d.createObjectTx(tx, projectID, db.TemplateVaultProps, vault)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return
}
