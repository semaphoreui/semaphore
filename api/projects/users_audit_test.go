package projects

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/semaphoreui/semaphore/services/audit/audittest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type nopLogWriter struct{}

func (nopLogWriter) WriteEventLog(pro_interfaces.EventLogRecord) error { return nil }
func (nopLogWriter) WriteTaskLog(pro_interfaces.TaskLogRecord) error   { return nil }
func (nopLogWriter) WriteResult(any) error                             { return nil }

type membershipFixture struct {
	store   *sql.SqlDb
	project db.Project
	owner   db.User
	member  db.User
	roles   map[db.ProjectUserRole]*db.Role
}

func newMembershipFixture(t *testing.T) membershipFixture {
	t.Helper()
	store := sql.InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "p"})
	require.NoError(t, err)
	roles := map[db.ProjectUserRole]*db.Role{}
	for _, key := range []db.ProjectUserRole{db.ProjectOwner, db.ProjectManager, db.ProjectTaskRunner, db.ProjectGuest} {
		role, err := store.GetSharedRoleBySlug(string(key))
		require.NoError(t, err)
		roles[key] = &role
	}
	create := func(name string, role *db.Role) db.User {
		user, err := store.CreateUser(db.UserWithPwd{Pwd: "verystrongpassword1", User: db.User{Username: name, Name: name, Email: name + "@example.com"}})
		require.NoError(t, err)
		err = store.CreateProjectUser(project.ID, user.ID, []int{role.ID})
		require.NoError(t, err)
		return user
	}
	return membershipFixture{
		store:   store,
		project: project,
		owner:   create("owner", roles[db.ProjectOwner]),
		member:  create("member", roles[db.ProjectManager]),
		roles:   roles,
	}
}

func (f membershipFixture) ref(key db.ProjectUserRole) audit.RoleRef {
	return audit.NewRoleRef(*f.roles[key])
}

func (f membershipFixture) request(method string, body string, me db.User, myRole *db.Role, target *db.User) (*http.Request, *audittest.Recorder) {
	rec := &audittest.Recorder{}
	r := httptest.NewRequest(method, "/api/project/1/users", bytes.NewBufferString(body))
	r = helpers.SetContextValue(r, "store", f.store)
	r = helpers.SetContextValue(r, "user", &me)
	r = helpers.SetContextValue(r, "project", f.project)
	var myMember db.ProjectMember
	if myRole != nil {
		myMember.Roles = []db.Role{*myRole}
	}
	r = helpers.SetContextValue(r, "projectMember", myMember)
	r = helpers.SetContextValue(r, "log_writer", nopLogWriter{})
	r = helpers.SetContextValue(r, "audit", rec)
	if target != nil {
		member, _ := f.store.GetProjectMember(f.project.ID, target.ID)
		r = helpers.SetContextValue(r, "projectUser", *target)
		r = helpers.SetContextValue(r, "projectUserMember", member)
	}
	return r.WithContext(audit.WithActor(r.Context(), audit.UserActor(me.ID, me.Username, audit.AuthSession, ""))), rec
}

// newMappedUser creates a user whose only access comes from a group mapping of the okta provider.
func (f membershipFixture) newMappedUser(t *testing.T, group string, role *db.Role) db.User {
	t.Helper()
	user, err := f.store.CreateUser(db.UserWithPwd{Pwd: "verystrongpassword1", User: db.User{Username: group, Name: group, Email: group + "@example.com"}})
	require.NoError(t, err)
	res, err := f.store.Sql().Exec(`
		insert into group_role_mapping (provider, group_name, project_id, role_id, created)
		values ('okta', ?, ?, ?, ?)
	`, group, f.project.ID, role.ID, time.Now().UTC())
	require.NoError(t, err)
	mappingID, err := res.LastInsertId()
	require.NoError(t, err)
	_, err = f.store.Sql().Exec("insert into project__user_mapped_role (project_id, user_id, mapping_id) values (?, ?, ?)", f.project.ID, user.ID, mappingID)
	require.NoError(t, err)
	return user
}

func only(t *testing.T, rec *audittest.Recorder, kind audit.Kind) audittest.Recorded {
	t.Helper()
	got, err := rec.Only(kind)
	require.NoError(t, err)
	return got
}

func TestAddUser_IsRecorded(t *testing.T) {
	f := newMembershipFixture(t)
	newcomer, err := f.store.CreateUser(db.UserWithPwd{Pwd: "verystrongpassword1", User: db.User{Username: "new", Name: "new", Email: "new@example.com"}})
	require.NoError(t, err)
	r, rec := f.request(http.MethodPost, fmt.Sprintf(`{"user_id":%d,"role":"task_runner"}`, newcomer.ID), f.owner, f.roles[db.ProjectOwner], nil)

	AddUser(httptest.NewRecorder(), r)

	got := only(t, rec, audit.IAMMembershipAdd)
	assert.Equal(t, audit.UserTarget(newcomer.ID, "new"), got.Event.Target)
	assert.Equal(t, f.project.ID, got.Event.ProjectID)
	assert.Equal(t, audit.MembershipMetadata{Roles: []audit.RoleRef{f.ref(db.ProjectTaskRunner)}}, got.Event.Metadata)
}

func TestRemoveUser_IsRecorded(t *testing.T) {
	f := newMembershipFixture(t)
	r, rec := f.request(http.MethodDelete, "", f.owner, f.roles[db.ProjectOwner], &f.member)

	RemoveUser(httptest.NewRecorder(), r)

	assert.Equal(t, audit.MembershipMetadata{Roles: []audit.RoleRef{f.ref(db.ProjectManager)}}, only(t, rec, audit.IAMMembershipRemove).Event.Metadata)
}

func TestAddUser_MappedMemberIsRoleChange(t *testing.T) {
	f := newMembershipFixture(t)
	user := f.newMappedUser(t, "m", f.roles[db.ProjectGuest])
	r, rec := f.request(http.MethodPost, fmt.Sprintf(`{"user_id":%d,"role":"manager"}`, user.ID), f.owner, f.roles[db.ProjectOwner], nil)

	AddUser(httptest.NewRecorder(), r)

	got := only(t, rec, audit.IAMProjectRoleChange)
	encoded, err := json.Marshal(got.Event.Metadata)
	require.NoError(t, err)
	assert.JSONEq(t, fmt.Sprintf(`{"old_roles":[],"new_roles":[{"id":%d,"slug":"manager"}]}`, f.roles[db.ProjectManager].ID), string(encoded))
}

type concurrentMemberRemoval struct{ *sql.SqlDb }

func (concurrentMemberRemoval) DeleteProjectUser(int, int) error { return db.ErrNotFound }

func TestRemoveUser_RemovedByAnotherRequestRecordsNothing(t *testing.T) {
	f := newMembershipFixture(t)
	r, rec := f.request(http.MethodDelete, "", f.owner, f.roles[db.ProjectOwner], &f.member)
	r = helpers.SetContextValue(r, "store", concurrentMemberRemoval{f.store})
	w := httptest.NewRecorder()

	RemoveUser(w, r)

	assert.Equal(t, http.StatusNoContent, w.Code, "the response does not change")
	assert.Empty(t, rec.All())
}

func TestLeftProject_IsRecorded(t *testing.T) {
	f := newMembershipFixture(t)
	r, rec := f.request(http.MethodDelete, "", f.member, f.roles[db.ProjectManager], nil)

	LeftProject(httptest.NewRecorder(), r)

	assert.Equal(t, audit.MembershipMetadata{Roles: []audit.RoleRef{f.ref(db.ProjectManager)}, SelfRemoval: true}, only(t, rec, audit.IAMMembershipRemove).Event.Metadata)
}

func TestLeftProject_AdminWithoutMembershipRecordsNothing(t *testing.T) {
	f := newMembershipFixture(t)
	admin, err := f.store.CreateUser(db.UserWithPwd{Pwd: "verystrongpassword1", User: db.User{Username: "admin", Name: "admin", Email: "admin@example.com", Admin: true}})
	require.NoError(t, err)
	r, rec := f.request(http.MethodDelete, "", admin, nil, nil)
	w := httptest.NewRecorder()

	LeftProject(w, r)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, rec.All())
}

func TestLeftProject_OwnerRefusalIsRecorded(t *testing.T) {
	f := newMembershipFixture(t)
	r, rec := f.request(http.MethodDelete, "", f.owner, f.roles[db.ProjectOwner], nil)
	w := httptest.NewRecorder()

	LeftProject(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	_, err := f.store.GetProjectUser(f.project.ID, f.owner.ID)
	assert.NoError(t, err, "the owner stays in the project")
	got := only(t, rec, audit.IAMMembershipRemove)
	assert.Equal(t, audit.OutcomeFailure, got.Event.Outcome)
	assert.Equal(t, audit.ReasonLastOwner, got.Event.Reason)
}

func TestUpdateUser_RoleChangeIsRecorded(t *testing.T) {
	f := newMembershipFixture(t)
	r, rec := f.request(http.MethodPut, `{"role":"task_runner"}`, f.owner, f.roles[db.ProjectOwner], &f.member)

	UpdateUser(httptest.NewRecorder(), r)

	assert.Equal(t, audit.ProjectRoleMetadata{OldRoles: []audit.RoleRef{f.ref(db.ProjectManager)}, NewRoles: []audit.RoleRef{f.ref(db.ProjectTaskRunner)}}, only(t, rec, audit.IAMProjectRoleChange).Event.Metadata)
}

func TestUpdateUser_LastOwnerRefusalIsRecorded(t *testing.T) {
	f := newMembershipFixture(t)
	r, rec := f.request(http.MethodPut, `{"role":"guest"}`, f.owner, f.roles[db.ProjectOwner], &f.owner)
	w := httptest.NewRecorder()

	UpdateUser(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	member, err := f.store.GetProjectMember(f.project.ID, f.owner.ID)
	require.NoError(t, err)
	assert.Equal(t, []string{"owner"}, slugsOf(member.Roles), "the owner keeps the role")
	got := only(t, rec, audit.IAMProjectRoleChange)
	assert.Equal(t, audit.ReasonLastOwner, got.Event.Reason)
	assert.Equal(t, audit.ProjectRoleMetadata{OldRoles: []audit.RoleRef{f.ref(db.ProjectOwner)}}, got.Event.Metadata)

	encoded, err := json.Marshal(got.Event.Metadata)
	require.NoError(t, err)
	assert.JSONEq(t, fmt.Sprintf(`{"old_roles":[{"id":%d,"slug":"owner"}]}`, f.roles[db.ProjectOwner].ID), string(encoded))
}
