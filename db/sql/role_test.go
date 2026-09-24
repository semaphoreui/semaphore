package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_RoleQueries(t *testing.T) {
	store := InitConfigCreateTestStore()

	project1, err := store.CreateProject(db.Project{Name: "project1"})
	require.NoError(t, err)

	// Seed another project to verify project-scoped and available-role queries
	// do not leak roles owned by unrelated projects.
	project2, err := store.CreateProject(db.Project{Name: "project2"})
	require.NoError(t, err)

	_, err = store.CreateRole(db.Role{
		Slug: "global_custom",
		Name: "Global custom",
	})
	require.NoError(t, err)

	_, err = store.CreateRole(db.Role{
		Slug:      "project_custom",
		Name:      "Project custom",
		ProjectID: &project1.ID,
	})
	require.NoError(t, err)

	_, err = store.CreateRole(db.Role{
		Slug:      "other_project_custom",
		Name:      "Other project custom",
		ProjectID: &project2.ID,
	})
	require.NoError(t, err)

	t.Run("global custom", func(t *testing.T) {
		roles, err := store.GetRoles(db.GlobalRoleQuery{Kinds: db.RoleKindCustom})
		require.NoError(t, err)
		require.Len(t, roles, 1)
		assert.Equal(t, "global_custom", roles[0].Slug)
	})

	t.Run("global built-in", func(t *testing.T) {
		roles, err := store.GetRoles(db.GlobalRoleQuery{Kinds: db.RoleKindBuiltin})
		require.NoError(t, err)
		require.Len(t, roles, 4)
		for _, role := range roles {
			assert.True(t, role.IsBuiltin)
		}
	})

	t.Run("project custom", func(t *testing.T) {
		roles, err := store.GetRoles(db.ProjectRoleQuery{ProjectID: project1.ID})
		require.NoError(t, err)
		require.Len(t, roles, 1)
		assert.Equal(t, "project_custom", roles[0].Slug)
	})

	t.Run("available custom", func(t *testing.T) {
		roles, err := store.GetRoles(db.AvailableRoleQuery{
			ProjectID: project1.ID,
			Kinds:     db.RoleKindCustom,
		})
		require.NoError(t, err)

		slugs := make([]string, 0, len(roles))
		for _, role := range roles {
			slugs = append(slugs, role.Slug)
		}
		assert.ElementsMatch(t, []string{"global_custom", "project_custom"}, slugs)
	})

	t.Run("available all", func(t *testing.T) {
		roles, err := store.GetRoles(db.AvailableRoleQuery{
			ProjectID: project1.ID,
			Kinds:     db.RoleKindAll,
		})
		require.NoError(t, err)

		slugs := make([]string, 0, len(roles))
		for _, role := range roles {
			slugs = append(slugs, role.Slug)
		}
		assert.ElementsMatch(t, []string{
			"guest",
			"global_custom",
			"manager",
			"owner",
			"project_custom",
			"task_runner",
		}, slugs)
	})
}

func Test_GetRoleBySlug_AppliesQueryScope(t *testing.T) {
	store := InitConfigCreateTestStore()

	project, err := store.CreateProject(db.Project{Name: "project"})
	require.NoError(t, err)

	_, err = store.CreateRole(db.Role{
		Slug: "global_custom",
		Name: "Global custom",
	})
	require.NoError(t, err)

	role, err := store.GetRoleBySlug("owner", db.GlobalRoleQuery{Kinds: db.RoleKindBuiltin})
	require.NoError(t, err)
	assert.True(t, role.IsBuiltin)

	_, err = store.GetRoleBySlug("owner", db.GlobalRoleQuery{Kinds: db.RoleKindCustom})
	assert.ErrorIs(t, err, db.ErrNotFound)

	role, err = store.GetRoleBySlug("global_custom", db.AvailableRoleQuery{
		ProjectID: project.ID,
		Kinds:     db.RoleKindCustom,
	})
	require.NoError(t, err)
	assert.Equal(t, "global_custom", role.Slug)
}

func Test_RoleQueries_RejectInvalidFilters(t *testing.T) {
	store := InitConfigCreateTestStore()

	_, err := store.GetRoles(nil)
	assert.Error(t, err)

	// Kinds is required; its zero value is intentionally invalid so callers
	// cannot accidentally include or exclude built-in roles.
	_, err = store.GetRoles(db.GlobalRoleQuery{})
	assert.Error(t, err)

	_, err = store.GetRoles(db.AvailableRoleQuery{Kinds: db.RoleKind(8)})
	assert.Error(t, err)
}
