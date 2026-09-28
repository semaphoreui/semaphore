package sql

import (
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type migratedRole struct {
	ID          int                      `db:"id"`
	Name        string                   `db:"name"`
	Permissions db.ProjectUserPermission `db:"permissions"`
	ProjectID   *int                     `db:"project_id"`
	BuiltinKey  *string                  `db:"builtin_key"`
}

func newRoleIDMigrationTestStore() *SqlDb {
	before := "2.20.7"
	return InitConfigCreateTestStoreAt(&before)
}

func migrateRolesToIDs(t *testing.T, store *SqlDb) {
	t.Helper()
	target := "2.20.8"
	require.NoError(t, db.Migrate(store, &target))
}

func insertRole(
	t *testing.T,
	store *SqlDb,
	slug string,
	name string,
	permissions db.ProjectUserPermission,
	projectID *int,
) {
	t.Helper()
	_, err := store.exec(`
		insert into "role" (
		    slug,
		    name,
		    permissions,
		    project_id
		)
		values (?, ?, ?, ?)`,
		slug,
		name,
		permissions,
		projectID)
	require.NoError(t, err)
}

func insertProjectUser(
	t *testing.T,
	store *SqlDb,
	projectID int,
	userID int,
	role string,
) {
	t.Helper()
	_, err := store.exec(`
		insert into project__user (
		    project_id,
		    user_id,
		    role
		)
		values (?, ?, ?)`,
		projectID,
		userID,
		role)
	require.NoError(t, err)
}

func insertTemplateGrant(
	t *testing.T,
	store *SqlDb,
	templateID int,
	roleSlug string,
	projectID int,
	permissions db.ProjectUserPermission,
) int64 {
	t.Helper()
	result, err := store.exec(`
		insert into project__template_role (
		    template_id,
		    role_slug,
		    project_id,
		    permissions
		)
		values (?, ?, ?, ?)`,
		templateID,
		roleSlug,
		projectID,
		permissions)
	require.NoError(t, err)
	id, err := result.LastInsertId()
	require.NoError(t, err)
	return id
}

func insertInvite(
	t *testing.T,
	store *SqlDb,
	projectID int,
	email string,
	role string,
	token string,
	inviterUserID int,
) int64 {
	t.Helper()
	result, err := store.exec(`
		insert into project__invite (
		    project_id,
		    email,
		    role,
		    status,
		    token,
		    inviter_user_id,
		    created
		)
		values (?, ?, ?, ?, ?, ?, ?)`,
		projectID,
		email,
		role,
		db.ProjectInvitePending,
		token,
		inviterUserID,
		time.Now())
	require.NoError(t, err)
	id, err := result.LastInsertId()
	require.NoError(t, err)
	return id
}

func TestMigration_2_20_8_CreatesIDBasedRoles(t *testing.T) {
	store := newRoleIDMigrationTestStore()

	project, err := store.CreateProject(db.Project{
		Name: "project",
	})
	require.NoError(t, err)

	insertRole(t, store, "global_one", "Duplicate", db.CanRunProjectTasks, nil)
	insertRole(t, store, "global_two", "Duplicate", db.CanUpdateProject, nil)
	insertRole(t, store, "project_one", "Duplicate", db.CanManageProjectResources, &project.ID)

	migrateRolesToIDs(t, store)

	var roles []migratedRole
	_, err = store.selectAll(&roles, `
		select *
		from "role"
		order by id`)
	require.NoError(t, err)
	require.Len(t, roles, 7)

	expectedBuiltins := map[string]struct {
		name        string
		permissions db.ProjectUserPermission
	}{
		"owner": {
			name:        "Owner",
			permissions: 15,
		},
		"manager": {
			name:        "Manager",
			permissions: 5,
		},
		"task_runner": {
			name:        "Task Runner",
			permissions: 1,
		},
		"guest": {
			name:        "Guest",
			permissions: 0,
		},
	}

	customCount := 0
	seenIDs := map[int]bool{}
	for _, role := range roles {
		require.Positive(t, role.ID)
		require.False(t, seenIDs[role.ID], "duplicate role ID %d", role.ID)
		seenIDs[role.ID] = true

		if role.BuiltinKey == nil {
			customCount++
			assert.Equal(t, "Duplicate", role.Name)
			continue
		}

		want, ok := expectedBuiltins[*role.BuiltinKey]
		require.True(t, ok, "unexpected built-in key %q", *role.BuiltinKey)
		assert.Equal(t, want.name, role.Name)
		assert.Equal(t, want.permissions, role.Permissions)
		assert.Nil(t, role.ProjectID)
	}
	assert.Equal(t, 3, customCount)

	slugColumns, err := store.Sql().SelectInt(`
		select count(*)
		from pragma_table_info('role')
		where name = 'slug'`)
	require.NoError(t, err)
	assert.Zero(t, slugColumns)

	isBuiltinColumns, err := store.Sql().SelectInt(`
		select count(*)
		from pragma_table_info('role')
		where name = 'is_builtin'`)
	require.NoError(t, err)
	assert.Zero(t, isBuiltinColumns)

	leftoverTables, err := store.Sql().SelectInt(`
		select count(*)
		from sqlite_master
		where type = 'table'
		  and name in (
		      'role_new',
		      'project__user_new',
		      'project__template_role_new',
		      'project__invite_new',
		      'role_old',
		      'project__user_old',
		      'project__template_role_old',
		      'project__invite_old'
		  )`)
	require.NoError(t, err)
	assert.Zero(t, leftoverTables)

	foreignKeyViolations, err := store.Sql().SelectInt(`
		select count(*)
		from pragma_foreign_key_check`)
	require.NoError(t, err)
	assert.Zero(t, foreignKeyViolations)
}

func createRoleIDMigrationTestUser(
	t *testing.T,
	store *SqlDb,
	username string,
) db.User {
	t.Helper()
	user, err := store.CreateUserWithoutPassword(db.User{
		Username: username,
		Name:     username,
		Email:    username + "@example.com",
	})
	require.NoError(t, err)
	return user
}

func TestMigration_2_20_8_MapsValidMembershipsToRoleIDs(t *testing.T) {
	store := newRoleIDMigrationTestStore()
	project, err := store.CreateProject(db.Project{
		Name: "project",
	})
	require.NoError(t, err)

	insertRole(t, store, "operator", "Operator", db.CanManageProjectResources, &project.ID)

	owner := createRoleIDMigrationTestUser(t, store, "owner")
	operator := createRoleIDMigrationTestUser(t, store, "operator")
	insertProjectUser(t, store, project.ID, owner.ID, "owner")
	insertProjectUser(t, store, project.ID, operator.ID, "operator")

	migrateRolesToIDs(t, store)

	ownerRoleID, err := store.Sql().SelectInt(`
		select id
		from "role"
		where builtin_key = 'owner'`)
	require.NoError(t, err)
	operatorRoleID, err := store.Sql().SelectInt(`
		select id
		from "role"
		where builtin_key is null
		  and name = 'Operator'
		  and project_id = ?`,
		project.ID)
	require.NoError(t, err)

	actualOwnerRoleID, err := store.Sql().SelectInt(`
		select role_id
		from project__user
		where project_id = ?
		  and user_id = ?`,
		project.ID,
		owner.ID)
	require.NoError(t, err)
	assert.Equal(t, ownerRoleID, actualOwnerRoleID)

	actualOperatorRoleID, err := store.Sql().SelectInt(`
		select role_id
		from project__user
		where project_id = ?
		  and user_id = ?`,
		project.ID,
		operator.ID)
	require.NoError(t, err)
	assert.Equal(t, operatorRoleID, actualOperatorRoleID)
}

func TestMigration_2_20_8_MapsValidTemplateGrantToRoleID(t *testing.T) {
	store := newRoleIDMigrationTestStore()
	projectID, repositoryID := newTemplateTestProject(t, store)

	insertRole(t, store, "operator", "Operator", db.CanManageProjectResources, &projectID)
	template, err := store.CreateTemplate(db.Template{
		ProjectID:    projectID,
		RepositoryID: repositoryID,
		Name:         "template",
		Playbook:     "site.yml",
	})
	require.NoError(t, err)
	grantID := insertTemplateGrant(t, store, template.ID, "operator", projectID, db.CanUpdateProject)

	migrateRolesToIDs(t, store)

	operatorRoleID, err := store.Sql().SelectInt(`
		select id
		from "role"
		where builtin_key is null
		  and name = 'Operator'
		  and project_id = ?`,
		projectID)
	require.NoError(t, err)
	var migratedGrant struct {
		ID     int64 `db:"id"`
		RoleID int64 `db:"role_id"`
	}
	err = store.selectOne(&migratedGrant, `
		select id, role_id
		from project__template_role
		where template_id = ?`, template.ID)
	require.NoError(t, err)
	assert.Equal(t, grantID, migratedGrant.ID)
	assert.Equal(t, operatorRoleID, migratedGrant.RoleID)
}

func TestMigration_2_20_8_MapsValidInvitationToRoleID(t *testing.T) {
	store := newRoleIDMigrationTestStore()
	project, err := store.CreateProject(db.Project{
		Name: "project",
	})
	require.NoError(t, err)
	inviter := createRoleIDMigrationTestUser(t, store, "inviter")
	inviteID := insertInvite(t, store, project.ID, "invitee@example.com", "manager", "token", inviter.ID)

	migrateRolesToIDs(t, store)

	managerRoleID, err := store.Sql().SelectInt(`
		select id
		from "role"
		where builtin_key = 'manager'`)
	require.NoError(t, err)
	var migratedInvite struct {
		ID     int64 `db:"id"`
		RoleID int64 `db:"role_id"`
	}
	err = store.selectOne(&migratedInvite, `
		select id, role_id
		from project__invite
		where token = 'token'`)
	require.NoError(t, err)
	assert.Equal(t, inviteID, migratedInvite.ID)
	assert.Equal(t, managerRoleID, migratedInvite.RoleID)
}

func TestMigration_2_20_8_OmitsInvalidMemberships(t *testing.T) {
	store := newRoleIDMigrationTestStore()

	roleProject, err := store.CreateProject(db.Project{
		Name: "role project",
	})
	require.NoError(t, err)
	memberProject, err := store.CreateProject(db.Project{
		Name: "member project",
	})
	require.NoError(t, err)

	insertRole(t, store, "scoped", "Scoped", db.CanRunProjectTasks, &roleProject.ID)

	wrongScopeUser, err := store.CreateUserWithoutPassword(db.User{
		Username: "wrong-scope",
		Name:     "Wrong Scope",
		Email:    "wrong-scope@example.com",
	})
	require.NoError(t, err)
	orphanedUser, err := store.CreateUserWithoutPassword(db.User{
		Username: "orphaned",
		Name:     "Orphaned",
		Email:    "orphaned@example.com",
	})
	require.NoError(t, err)

	insertProjectUser(t, store, memberProject.ID, wrongScopeUser.ID, "scoped")
	insertProjectUser(t, store, memberProject.ID, orphanedUser.ID, "missing")

	migrateRolesToIDs(t, store)

	count, err := store.Sql().SelectInt(`
		select count(*)
		from project__user`)
	require.NoError(t, err)
	assert.Zero(t, count)
}

func TestMigration_2_20_8_OmitsWrongScopeTemplateGrant(t *testing.T) {
	store := newRoleIDMigrationTestStore()
	roleProject, err := store.CreateProject(db.Project{
		Name: "role project",
	})
	require.NoError(t, err)
	templateProjectID, repositoryID := newTemplateTestProject(t, store)

	insertRole(t, store, "scoped", "Scoped", db.CanRunProjectTasks, &roleProject.ID)
	template, err := store.CreateTemplate(db.Template{
		ProjectID:    templateProjectID,
		RepositoryID: repositoryID,
		Name:         "template",
		Playbook:     "site.yml",
	})
	require.NoError(t, err)
	insertTemplateGrant(t, store, template.ID, "scoped", templateProjectID, db.CanRunProjectTasks)

	migrateRolesToIDs(t, store)

	count, err := store.Sql().SelectInt(`
		select count(*)
		from project__template_role`)
	require.NoError(t, err)
	assert.Zero(t, count)
}

func TestMigration_2_20_8_OmitsInvalidInvitation(t *testing.T) {
	store := newRoleIDMigrationTestStore()
	project, err := store.CreateProject(db.Project{
		Name: "project",
	})
	require.NoError(t, err)
	inviter, err := store.CreateUserWithoutPassword(db.User{
		Username: "inviter",
		Name:     "Inviter",
		Email:    "inviter@example.com",
	})
	require.NoError(t, err)
	insertRole(t, store, "custom_invite", "Custom invite", db.CanRunProjectTasks, &project.ID)
	insertInvite(t, store, project.ID, "invitee@example.com", "custom_invite", "token", inviter.ID)

	migrateRolesToIDs(t, store)

	count, err := store.Sql().SelectInt(`
		select count(*)
		from project__invite`)
	require.NoError(t, err)
	assert.Zero(t, count)
}

func TestMigration_2_20_8_RestrictsAssignedRoleDeletion(t *testing.T) {
	store := newRoleIDMigrationTestStore()
	project, err := store.CreateProject(db.Project{
		Name: "project",
	})
	require.NoError(t, err)
	user, err := store.CreateUserWithoutPassword(db.User{
		Username: "member",
		Name:     "Member",
		Email:    "member@example.com",
	})
	require.NoError(t, err)
	insertRole(t, store, "scoped", "Scoped", db.CanRunProjectTasks, &project.ID)
	insertProjectUser(t, store, project.ID, user.ID, "scoped")

	migrateRolesToIDs(t, store)

	roleID, err := store.Sql().SelectInt(`
		select id
		from "role"
		where name = 'Scoped'`)
	require.NoError(t, err)
	_, err = store.exec(`
		delete from "role"
		where id = ?`,
		roleID)
	require.Error(t, err)
}

func TestMigration_2_20_8_RejectsBuiltinSlugCollision(t *testing.T) {
	store := newRoleIDMigrationTestStore()
	target := "2.20.8"

	insertRole(t, store, "owner", "Conflicting owner", 0, nil)

	err := store.ApplyMigration(db.Migration{
		Version: target,
	})
	require.Error(t, err)

	applied, err := store.IsMigrationApplied(db.Migration{
		Version: target,
	})
	require.NoError(t, err)
	assert.False(t, applied)

	count, err := store.Sql().SelectInt(`
		select count(*)
		from "role"
		where slug = 'owner'`)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	newTables, err := store.Sql().SelectInt(`
		select count(*)
		from sqlite_master
		where type = 'table'
		  and name in (
		      'role_new',
		      'project__user_new',
		      'project__template_role_new',
		      'project__invite_new'
		  )`)
	require.NoError(t, err)
	assert.Zero(t, newTables)
}
