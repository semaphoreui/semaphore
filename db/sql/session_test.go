package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPITokens_PrefixLookupAndExactDelete(t *testing.T) {
	store := InitConfigCreateTestStore()
	user, err := store.CreateUser(db.UserWithPwd{Pwd: "verystrongpassword1", User: db.User{Username: "alice", Name: "alice", Email: "alice@example.com"}})
	require.NoError(t, err)
	for _, id := range []string{"abcdefgh1111", "abcdefgh2222", "zzzzzzzz3333"} {
		_, err = store.CreateAPIToken(db.APIToken{ID: id, UserID: user.ID, Name: id})
		require.NoError(t, err)
	}

	matched, err := store.GetAPITokensByPrefix(user.ID, "abcdefgh")
	require.NoError(t, err)
	var ids []string
	for _, token := range matched {
		ids = append(ids, token.ID)
	}
	assert.ElementsMatch(t, []string{"abcdefgh1111", "abcdefgh2222"}, ids)

	_, err = store.GetAPITokensByPrefix(user.ID, "short")
	assert.Error(t, err, "a prefix shorter than 8 characters is refused")

	require.NoError(t, store.DeleteAPIToken(user.ID, "abcdefgh1111"))
	assert.ErrorIs(t, store.DeleteAPIToken(user.ID, "abcdefgh1111"), db.ErrNotFound, "a token deleted earlier is not reported again")
	assert.ErrorIs(t, store.DeleteAPIToken(user.ID, "abcdefgh"), db.ErrNotFound, "delete takes an exact ID, not a prefix")
}

// The audit records a delete only when the store confirms it removed a row.
func TestAuditedDeletes_ReportMissingRow(t *testing.T) {
	store := InitConfigCreateTestStore()
	user, err := store.CreateUser(db.UserWithPwd{Pwd: "verystrongpassword1", User: db.User{Username: "bob", Name: "bob", Email: "bob@example.com"}})
	require.NoError(t, err)
	projectID, repositoryID := newTemplateTestProject(t, store)
	template, err := store.CreateTemplate(db.Template{ProjectID: projectID, RepositoryID: repositoryID, Name: "t", Playbook: "p.yml"})
	require.NoError(t, err)

	_, err = store.CreateProjectUser(db.ProjectUser{ProjectID: projectID, UserID: user.ID, Role: db.ProjectManager})
	require.NoError(t, err)
	_, err = store.CreateRole(db.Role{Slug: "ops", Name: "ops", ProjectID: &projectID})
	require.NoError(t, err)
	perm, err := store.CreateTemplateRole(db.TemplateRolePerm{ProjectID: projectID, TemplateID: template.ID, RoleSlug: "ops", Permissions: db.CanRunProjectTasks})
	require.NoError(t, err)
	totp, err := store.AddTotpVerification(user.ID, "otpauth://totp/x", "")
	require.NoError(t, err)
	_, err = store.CreateExternalIdentity(db.UserExternalIdentity{UserID: user.ID, Type: db.IdentityTypeOidc, Provider: "corp", ExternalUID: "sub-1"})
	require.NoError(t, err)
	inventory, err := store.CreateInventory(db.Inventory{ProjectID: projectID, Name: "i", Type: db.InventoryStatic})
	require.NoError(t, err)
	env, err := store.CreateEnvironment(db.Environment{ProjectID: projectID, Name: "e", JSON: "{}"})
	require.NoError(t, err)
	other, err := store.CreateProject(db.Project{Name: "other"})
	require.NoError(t, err)

	deletes := map[string]func() error{
		"project user":  func() error { return store.DeleteProjectUser(projectID, user.ID) },
		"template role": func() error { return store.DeleteTemplateRole(projectID, template.ID, perm.ID) },
		"totp":          func() error { return store.DeleteTotpVerification(user.ID, totp.ID) },
		"identity":      func() error { return store.DeleteExternalIdentity(user.ID, db.IdentityTypeOidc, "corp") },
		"inventory":     func() error { return store.DeleteInventory(projectID, inventory.ID) },
		"environment":   func() error { return store.DeleteEnvironment(projectID, env.ID) },
		"project":       func() error { return store.DeleteProject(other.ID) },
	}
	for name, del := range deletes {
		t.Run(name, func(t *testing.T) {
			require.NoError(t, del())
			assert.ErrorIs(t, del(), db.ErrNotFound)
		})
	}

	t.Run("template", func(t *testing.T) {
		require.NoError(t, store.DeleteTemplate(projectID, template.ID))
		assert.ErrorIs(t, store.DeleteTemplate(projectID, template.ID), db.ErrNotFound)
	})

	t.Run("option stays idempotent", func(t *testing.T) {
		require.NoError(t, store.SetOption("audit_test", "1"))
		require.NoError(t, store.DeleteOption("audit_test"))
		assert.NoError(t, store.DeleteOption("audit_test"))
	})
}
