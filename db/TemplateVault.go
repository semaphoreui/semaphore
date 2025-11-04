package db

import (
	"errors"

	log "github.com/sirupsen/logrus"
)

type TemplateVaultType string

const (
	TemplateVaultPassword TemplateVaultType = "password"
	TemplateVaultScript   TemplateVaultType = "script"
)

type TemplateVault struct {
	ID         int               `db:"id" json:"id" backup:"-"`
	ProjectID  int               `db:"project_id" json:"project_id" backup:"-"`
	TemplateID int               `db:"template_id" json:"template_id" backup:"-"`
	VaultKeyID *int              `db:"vault_key_id" json:"vault_key_id,omitempty" backup:"-"`
	Name       *string           `db:"name" json:"name,omitempty"`
	Type       TemplateVaultType `db:"type" json:"type"`
	Script     *string           `db:"script" json:"script,omitempty"`

	Vault *AccessKey `db:"-" json:"-"`
}

// FillTemplateVault populates the Vault field of a TemplateVault by fetching
// the associated access key from the database.
//
// If the vault type is TemplateVaultPassword and a VaultKeyID is set, this function
// attempts to retrieve the corresponding AccessKey. If the key is not found
// (returns ErrNotFound), a warning is logged and the function returns successfully
// with the Vault field remaining nil. This allows templates to load even when
// vault keys have been deleted.
//
// For other types of errors, the error is returned to the caller.
func FillTemplateVault(d Store, projectID int, templateVault *TemplateVault) (err error) {
	if templateVault.Type == TemplateVaultPassword && templateVault.VaultKeyID != nil {
		var vault AccessKey
		vault, err = d.GetAccessKey(projectID, *templateVault.VaultKeyID)
		if err != nil {
			// If the vault key is not found, log a warning but don't fail the entire request.
			// This allows templates to load even when vault keys have been deleted.
			if errors.Is(err, ErrNotFound) {
				log.WithFields(log.Fields{
					"vault_id":     templateVault.ID,
					"vault_key_id": *templateVault.VaultKeyID,
					"project_id":   projectID,
				}).Warn("Template vault references non-existent access key")
				return nil
			}
			return
		}
		templateVault.Vault = &vault
	}
	return
}
