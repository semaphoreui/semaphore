package sql

import (
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The store writes projects, users and templates on the 2.20.11 schema, the role rows below use its columns.
func newRoleMigrationTestStore() *SqlDb {
	before := "2.20.11"
	return InitConfigCreateTestStoreAt(&before)
}

func migrateRoles(t *testing.T, store *SqlDb) {
	t.Helper()
	// 2.20.13 adds the unique slug index once 2.20.12 has renamed the collisions.
	target := "2.20.13"
	require.NoError(t, db.Migrate(store, &target))
}

func insertLegacyRole(t *testing.T, store *SqlDb, slug string, name string, projectID *int) {
	t.Helper()
	_, err := store.exec("insert into `role` (slug, name, permissions, project_id) values (?, ?, 1, ?)", slug, name, projectID)
	require.NoError(t, err)
}

func insertLegacyMember(t *testing.T, store *SqlDb, projectID int, userID int, role string) {
	t.Helper()
	_, err := store.exec("insert into project__user (project_id, user_id, `role`) values (?, ?, ?)", projectID, userID, role)
	require.NoError(t, err)
}

func insertLegacyGrant(t *testing.T, store *SqlDb, projectID int, templateID int, role string) {
	t.Helper()
	_, err := store.exec(`
		insert into project__template_role (project_id, template_id, role_slug, permissions)
		values (?, ?, ?, 2)
	`, projectID, templateID, role)
	require.NoError(t, err)
}

func createLegacyUser(t *testing.T, store *SqlDb, name string) int {
	t.Helper()
	user, err := store.CreateUserWithoutPassword(db.User{Username: name, Name: name, Email: name + "@example.com"})
	require.NoError(t, err)
	return user.ID
}

func createLegacyTemplate(t *testing.T, store *SqlDb) (int, int) {
	t.Helper()
	projectID, repositoryID := newTemplateTestProject(t, store)
	template, err := store.CreateTemplate(db.Template{ProjectID: projectID, RepositoryID: repositoryID, Name: "t", Playbook: "p.yml"})
	require.NoError(t, err)
	return projectID, template.ID
}

func roleIDOf(t *testing.T, store *SqlDb, query string, args ...any) int {
	t.Helper()
	id, err := store.Sql().SelectInt(store.PrepareQuery(query), args...)
	require.NoError(t, err)
	return int(id)
}

func TestMigration_2_20_12_SeedsBuiltinsAndKeepsSlugs(t *testing.T) {
	store := newRoleMigrationTestStore()
	project, err := store.CreateProject(db.Project{Name: "p"})
	require.NoError(t, err)
	insertLegacyRole(t, store, "deployer", "Deployer", nil)
	insertLegacyRole(t, store, "auditor", "Auditor", &project.ID)

	migrateRoles(t, store)

	var roles []db.Role
	_, err = store.selectAll(&roles, "select * from `role` order by id")
	require.NoError(t, err)
	got := map[string]db.Role{}
	for _, r := range roles {
		got[r.Slug] = r
	}
	for slug, mask := range map[string]db.ProjectUserPermission{"owner": 15, "manager": 5, "task_runner": 1, "guest": 0} {
		assert.True(t, got[slug].Builtin, slug)
		assert.Equal(t, mask, got[slug].Permissions, slug)
	}
	assert.Equal(t, 0, got["deployer"].Scope)
	assert.Equal(t, project.ID, got["auditor"].Scope)

	leftovers, err := store.Sql().SelectInt(`
		select
		  count(*)
		from sqlite_master
		where type = 'table'
		  and name in ('project__user', 'project__user_old', 'role_old', 'project__template_role_old', 'role_new', 'project__template_role_new')
	`)
	require.NoError(t, err)
	assert.Zero(t, leftovers, "the membership table and the old tables are gone")
}

func TestMigration_2_20_12_MovesMembersAndGrantsToRoleIDs(t *testing.T) {
	store := newRoleMigrationTestStore()
	projectID, templateID := createLegacyTemplate(t, store)
	owner, deployer := createLegacyUser(t, store, "owner"), createLegacyUser(t, store, "deployer")
	insertLegacyRole(t, store, "deployer", "Deployer", nil)
	insertLegacyMember(t, store, projectID, owner, "owner")
	insertLegacyMember(t, store, projectID, deployer, "deployer")
	insertLegacyGrant(t, store, projectID, templateID, "deployer")

	migrateRoles(t, store)

	ownerRole := roleIDOf(t, store, "select id from `role` where slug='owner'")
	deployerRole := roleIDOf(t, store, "select id from `role` where slug='deployer'")
	assert.Equal(t, ownerRole, roleIDOf(t, store, "select role_id from project__user_role where user_id=?", owner))
	assert.Equal(t, deployerRole, roleIDOf(t, store, "select role_id from project__user_role where user_id=?", deployer))
	assert.Equal(t, deployerRole, roleIDOf(t, store, "select role_id from project__template_role where template_id=?", templateID))
}

func TestMigration_2_20_12_Orphans(t *testing.T) {
	// PRO-69: the role was deleted, PRO-70: the role belongs to another project.
	store := newRoleMigrationTestStore()
	projectID, templateID := createLegacyTemplate(t, store)
	other, err := store.CreateProject(db.Project{Name: "other"})
	require.NoError(t, err)
	ghost, foreign := createLegacyUser(t, store, "ghost"), createLegacyUser(t, store, "foreign")
	insertLegacyRole(t, store, "other_only", "Other", &other.ID)
	insertLegacyRole(t, store, "g", "G", nil)
	insertLegacyMember(t, store, projectID, ghost, "deleted-role")
	insertLegacyMember(t, store, projectID, foreign, "other_only")
	insertLegacyMember(t, store, other.ID, foreign, "other_only")
	insertLegacyGrant(t, store, projectID, templateID, "other_only")
	insertLegacyGrant(t, store, other.ID, templateID, "g")

	migrateRoles(t, store)

	guest := roleIDOf(t, store, "select id from role where slug='guest' and builtin=true")
	assert.Equal(t, guest, roleIDOf(t, store, "select role_id from project__user_role where project_id=? and user_id=?", projectID, ghost))
	assert.Equal(t, guest, roleIDOf(t, store, "select role_id from project__user_role where project_id=? and user_id=?", projectID, foreign))
	assert.Equal(t, 1, roleIDOf(t, store, "select count(*) from project__user_role where project_id=?", other.ID))
	assert.Zero(t, roleIDOf(t, store, "select count(*) from project__template_role"))
}

func TestMigration_2_20_12_RenamesCollidingSlugs(t *testing.T) {
	type role struct {
		slug    string
		project int // 0 global, 1 or 2 a project
	}
	long := strings.Repeat("é", 49)
	tests := []struct {
		name  string
		roles []role
		want  []string // final slugs in the order of roles
	}{
		{"global keeps the slug over a project", []role{{"deployer", 0}, {"Deployer", 1}}, []string{"deployer", "Deployer-2"}},
		{"smallest slug in one scope keeps it", []role{{"zed", 0}, {"Zed", 0}}, []string{"zed-2", "Zed"}},
		{"case variants in two projects both stay", []role{{"Dev", 1}, {"dev", 2}}, []string{"Dev", "dev"}},
		{"case variants in one project", []role{{"dev", 1}, {"Dev", 1}}, []string{"dev-2", "Dev"}},
		{"custom role named like a built-in", []role{{"manager", 0}}, []string{"manager-2"}},
		{"case variant of a built-in", []role{{"Manager", 1}}, []string{"Manager-2"}},
		{"multibyte slug is cut on a character", []role{{"a" + long, 0}, {"A" + long, 1}}, []string{"a" + long, "A" + strings.Repeat("é", 48) + "-2"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newRoleMigrationTestStore()
			projects := []*int{nil}
			for _, name := range []string{"p1", "p2"} {
				p, err := store.CreateProject(db.Project{Name: name})
				require.NoError(t, err)
				projects = append(projects, &p.ID)
			}
			for i, r := range tt.roles {
				insertLegacyRole(t, store, r.slug, string(rune('a'+i)), projects[r.project])
			}

			migrateRoles(t, store)

			for i := range tt.roles {
				slug, err := store.Sql().SelectStr(store.PrepareQuery("select slug from `role` where name=?"), string(rune('a'+i)))
				require.NoError(t, err)
				assert.Equal(t, tt.want[i], slug)
			}
		})
	}
}

func TestMigration_2_20_12_MembersAndGrantsFollowTheirRole(t *testing.T) {
	store := newRoleMigrationTestStore()
	projectID, templateID := createLegacyTemplate(t, store)
	viaBuiltin, viaVariant := createLegacyUser(t, store, "builtin"), createLegacyUser(t, store, "variant")
	// An exact built-in slug meant the built-in role, a case variant was an ordinary custom role.
	insertLegacyRole(t, store, "manager", "Shadow", nil)
	insertLegacyRole(t, store, "Manager", "Variant", &projectID)
	insertLegacyMember(t, store, projectID, viaBuiltin, "manager")
	insertLegacyMember(t, store, projectID, viaVariant, "Manager")
	insertLegacyGrant(t, store, projectID, templateID, "manager")
	insertLegacyGrant(t, store, projectID, templateID, "Manager")

	migrateRoles(t, store)

	builtin := roleIDOf(t, store, "select id from `role` where builtin=true and slug='manager'")
	variant := roleIDOf(t, store, "select id from `role` where name='Variant'")
	assert.Equal(t, builtin, roleIDOf(t, store, "select role_id from project__user_role where user_id=?", viaBuiltin))
	assert.Equal(t, variant, roleIDOf(t, store, "select role_id from project__user_role where user_id=?", viaVariant))
	assert.Equal(t, 1, roleIDOf(t, store, "select count(*) from project__template_role where role_id=?", builtin))
	assert.Equal(t, 1, roleIDOf(t, store, "select count(*) from project__template_role where role_id=?", variant))
}

func TestMigration_2_20_12_EnforcesScopeRules(t *testing.T) {
	store := newRoleMigrationTestStore()
	p1, err := store.CreateProject(db.Project{Name: "p1"})
	require.NoError(t, err)
	p2, err := store.CreateProject(db.Project{Name: "p2"})
	require.NoError(t, err)
	migrateRoles(t, store)

	insert := func(slug string, projectID *int, scope int) error {
		_, err := store.exec(`
			insert into role (slug, name, permissions, project_id, scope, builtin)
			values (?, ?, 0, ?, ?, false)
		`, slug, slug, projectID, scope)
		return err
	}
	require.NoError(t, insert("dev", &p1.ID, p1.ID))
	require.NoError(t, insert("dev", &p2.ID, p2.ID))
	assert.Error(t, insert("DEV", &p1.ID, p1.ID), "slugs are unique ignoring case in a scope")
	assert.Error(t, insert("OWNER", nil, 0), "built-in slugs are taken")
	assert.Error(t, insert("x", nil, p1.ID), "a global role has scope 0")
	assert.Error(t, insert("y", &p1.ID, p2.ID), "scope equals project_id")
}
