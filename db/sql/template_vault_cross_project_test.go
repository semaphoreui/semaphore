package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestManagerCannotMoveForeignProjectVault reproduces GHSA-ffp6-w57q-758q.
//
// A user who can manage templates in project A creates a template there and, in
// the Vaults array, supplies the numeric ID of a vault binding that belongs to a
// template in project B. On the vulnerable code (v2.19.18) UpdateTemplateVaults
// ran `UPDATE project__template_vault SET project_id=?, template_id=? ... WHERE
// id=?`, selecting the row by id alone, so B's binding was reparented into A's
// new template: B silently lost its binding.
//
// The test asserts the secure outcome: B keeps its binding and A's template does
// not acquire B's row. It fails on the vulnerable code and passes once the UPDATE
// WHERE clause is scoped to the authorized project and template.
func TestManagerCannotMoveForeignProjectVault(t *testing.T) {
	store := InitConfigCreateTestStore()

	// Project B (the victim) with a template that owns a vault binding.
	victimProjectID, victimRepoID := newTemplateTestProject(t, store)
	victimScript := "victim-vault-client.py"
	victimTemplate, err := store.CreateTemplate(db.Template{
		ProjectID:    victimProjectID,
		RepositoryID: victimRepoID,
		Name:         "victim-template",
		Playbook:     "site.yml",
		Vaults: []db.TemplateVault{
			{Type: db.TemplateVaultScript, Script: &victimScript},
		},
	})
	require.NoError(t, err)

	// Read the persisted binding back from the store so we have its id on every
	// release (older CreateTemplate did not populate returned vault ids).
	victimVaultsBefore, err := store.GetTemplateVaults(victimProjectID, victimTemplate.ID)
	require.NoError(t, err)
	require.Len(t, victimVaultsBefore, 1)
	foreignVaultID := victimVaultsBefore[0].ID
	require.NotZero(t, foreignVaultID)

	// Project A (the attacker) — a separate project the attacker manages.
	attackerProjectID, attackerRepoID := newTemplateTestProject(t, store)

	// The attack: create a template in A whose Vaults array references B's
	// binding by id. type "script" clears vault_key_id, so no key from B is
	// needed in the body.
	attackScript := "attacker-vault-client.py"
	attackerTemplate, err := store.CreateTemplate(db.Template{
		ProjectID:    attackerProjectID,
		RepositoryID: attackerRepoID,
		Name:         "attacker-template",
		Playbook:     "site.yml",
		Vaults: []db.TemplateVault{
			{ID: foreignVaultID, Type: db.TemplateVaultScript, Script: &attackScript},
		},
	})
	require.NoError(t, err)

	// B must still own its binding.
	victimVaults, err := store.GetTemplateVaults(victimProjectID, victimTemplate.ID)
	require.NoError(t, err)
	assert.Len(t, victimVaults, 1,
		"victim project B lost its vault binding: it was reparented into attacker project A (GHSA-ffp6-w57q-758q)")

	// A must not have stolen B's row.
	attackerVaults, err := store.GetTemplateVaults(attackerProjectID, attackerTemplate.ID)
	require.NoError(t, err)
	for _, v := range attackerVaults {
		assert.NotEqual(t, foreignVaultID, v.ID,
			"attacker template acquired the victim's vault-binding row (GHSA-ffp6-w57q-758q)")
	}
}
