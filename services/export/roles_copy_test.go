package export

import (
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCopyKeepsRolesMembershipsAndGrants(t *testing.T) {
	src := sql.InitConfigCreateTestStore()

	user, err := src.CreateUser(db.UserWithPwd{Pwd: "3412341234123", User: db.User{Username: "u", Name: "u", Email: "u@example.com"}})
	require.NoError(t, err)
	proj, err := src.CreateProject(db.Project{Name: "P"})
	require.NoError(t, err)
	key, err := src.CreateAccessKey(db.AccessKey{ProjectID: &proj.ID, Type: db.AccessKeyNone})
	require.NoError(t, err)
	repo, err := src.CreateRepository(db.Repository{ProjectID: proj.ID, SSHKeyID: key.ID, Name: "Repo", GitURL: "git@example.com:t/t", GitBranch: "master"})
	require.NoError(t, err)
	inv, err := src.CreateInventory(db.Inventory{ProjectID: proj.ID})
	require.NoError(t, err)
	tpl, err := src.CreateTemplate(db.Template{Name: "Tpl", Playbook: "t.yml", ProjectID: proj.ID, RepositoryID: repo.ID, InventoryID: &inv.ID})
	require.NoError(t, err)

	local, err := src.CreateRole(db.Role{Slug: "local", Name: "Local", ProjectID: &proj.ID, Permissions: db.CanRunProjectTasks})
	require.NoError(t, err)
	global, err := src.CreateRole(db.Role{Slug: "deployer", Name: "Deployer"})
	require.NoError(t, err)
	legacy, err := src.CreateRole(db.Role{Slug: "legacy role", Name: "Legacy"})
	require.NoError(t, err)
	manager, err := src.GetGlobalRoleBySlug("manager")
	require.NoError(t, err)

	err = src.CreateProjectUser(proj.ID, user.ID, []int{manager.ID, local.ID, global.ID})
	require.NoError(t, err)
	for _, id := range []int{local.ID, legacy.ID, manager.ID} {
		_, err = src.CreateTemplateRole(db.TemplateRolePerm{ProjectID: proj.ID, TemplateID: tpl.ID, RoleID: id, Permissions: db.CanRunProjectTasks})
		require.NoError(t, err)
	}

	dst := sql.InitConfigCreateTestStore()
	chain := InitProjectExporters(NewKeyMapper(), true, false)
	require.NoError(t, chain.Load(src))
	require.NoError(t, chain.Restore(dst, 10))

	projects, err := dst.GetAllProjects()
	require.NoError(t, err)
	require.Len(t, projects, 1)
	users, err := dst.GetProjectUsers(projects[0].ID, db.RetrieveQueryParams{})
	require.NoError(t, err)
	require.Len(t, users, 1)
	var slugs []string
	for _, r := range users[0].Roles {
		slugs = append(slugs, r.Slug)
	}
	assert.ElementsMatch(t, []string{"manager", "local", "deployer"}, slugs)

	kept, err := dst.GetGlobalRoleBySlug("legacy role")
	require.NoError(t, err)
	assert.Equal(t, "Legacy", kept.Name)
	assert.False(t, kept.Builtin)

	templates, err := dst.GetTemplates(projects[0].ID, db.TemplateFilter{}, db.RetrieveQueryParams{})
	require.NoError(t, err)
	require.Len(t, templates, 1)
	grants, err := dst.GetTemplateRoles(projects[0].ID, templates[0].ID)
	require.NoError(t, err)
	var grantSlugs []string
	for _, g := range grants {
		grantSlugs = append(grantSlugs, g.RoleSlug)
	}
	assert.ElementsMatch(t, []string{"local", "legacy role", "manager"}, grantSlugs)
}

// addMappedRole gives the user a role through a new mapping, as a login sync would.
func addMappedRole(t *testing.T, store *sql.SqlDb, projectID int, userID int, roleID int, group string) {
	t.Helper()
	_, err := store.Sql().Exec(store.PrepareQuery(`
		insert into group_role_mapping (provider, group_name, project_id, role_id, created)
		values ('okta', ?, ?, ?, ?)
	`), group, projectID, roleID, time.Now().UTC())
	require.NoError(t, err)
	mappingID, err := store.Sql().SelectInt(store.PrepareQuery("select id from group_role_mapping where group_name=? and project_id=?"), group, projectID)
	require.NoError(t, err)
	_, err = store.Sql().Exec(store.PrepareQuery("insert into project__user_mapped_role (project_id, user_id, mapping_id) values (?, ?, ?)"), projectID, userID, mappingID)
	require.NoError(t, err)
}

func TestCopySkipsMemberWithoutManualRoles(t *testing.T) {
	src := sql.InitConfigCreateTestStore()

	proj, err := src.CreateProject(db.Project{Name: "P"})
	require.NoError(t, err)
	manager, err := src.GetGlobalRoleBySlug("manager")
	require.NoError(t, err)
	manual, err := src.CreateUser(db.UserWithPwd{Pwd: "3412341234123", User: db.User{Username: "manual", Name: "manual", Email: "manual@example.com"}})
	require.NoError(t, err)
	require.NoError(t, src.CreateProjectUser(proj.ID, manual.ID, []int{manager.ID}))
	mapped, err := src.CreateUser(db.UserWithPwd{Pwd: "3412341234123", User: db.User{Username: "mapped", Name: "mapped", Email: "mapped@example.com"}})
	require.NoError(t, err)
	addMappedRole(t, src, proj.ID, mapped.ID, manager.ID, "devops")

	dst := sql.InitConfigCreateTestStore()
	chain := InitProjectExporters(NewKeyMapper(), true, false)
	require.NoError(t, chain.Load(src))
	// Restore only logs a failed member, so the skip is checked on the loaded values.
	members := chain.getTypeExporter(ProjectUser).(*ProjectUserExporter).values
	require.Len(t, members, 1)
	assert.Equal(t, manual.ID, members[0].value.UserID)
	require.NoError(t, chain.Restore(dst, 10))

	projects, err := dst.GetAllProjects()
	require.NoError(t, err)
	require.Len(t, projects, 1)
	users, err := dst.GetProjectUsers(projects[0].ID, db.RetrieveQueryParams{})
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, "manual", users[0].Username)
	require.Len(t, users[0].Roles, 1)
	assert.Equal(t, "manager", users[0].Roles[0].Slug)

	copied, err := dst.GetUserByLoginOrEmail("mapped", "mapped@example.com")
	require.NoError(t, err)
	rows, err := dst.Sql().SelectInt(dst.PrepareQuery("select count(*) from project__user_role where user_id=?"), copied.ID)
	require.NoError(t, err)
	assert.Zero(t, rows)
}
