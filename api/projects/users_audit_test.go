package projects

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

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
}

func newMembershipFixture(t *testing.T) membershipFixture {
	t.Helper()
	store := sql.InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "p"})
	require.NoError(t, err)
	create := func(name string, role db.ProjectUserRole) db.User {
		user, err := store.CreateUser(db.UserWithPwd{Pwd: "verystrongpassword1", User: db.User{Username: name, Name: name, Email: name + "@example.com"}})
		require.NoError(t, err)
		if role != db.ProjectNone {
			_, err = store.CreateProjectUser(db.ProjectUser{ProjectID: project.ID, UserID: user.ID, Role: role})
			require.NoError(t, err)
		}
		return user
	}
	return membershipFixture{store: store, project: project, owner: create("owner", db.ProjectOwner), member: create("member", db.ProjectManager)}
}

func (f membershipFixture) request(method string, body string, me db.User, myRole db.ProjectUserRole, target *db.User) (*http.Request, *audittest.Recorder) {
	rec := &audittest.Recorder{}
	r := httptest.NewRequest(method, "/api/project/1/users", bytes.NewBufferString(body))
	r = helpers.SetContextValue(r, "store", f.store)
	r = helpers.SetContextValue(r, "user", &me)
	r = helpers.SetContextValue(r, "project", f.project)
	r = helpers.SetContextValue(r, "projectUserRole", myRole)
	r = helpers.SetContextValue(r, "log_writer", nopLogWriter{})
	r = helpers.SetContextValue(r, "audit", rec)
	if target != nil {
		membership, _ := f.store.GetProjectUser(f.project.ID, target.ID)
		r = helpers.SetContextValue(r, "projectUser", *target)
		r = helpers.SetContextValue(r, "projectMembership", membership)
	}
	return r.WithContext(audit.WithActor(r.Context(), audit.UserActor(me.ID, me.Username, audit.AuthSession, ""))), rec
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
	r, rec := f.request(http.MethodPost, fmt.Sprintf(`{"user_id":%d,"role":"task_runner"}`, newcomer.ID), f.owner, db.ProjectOwner, nil)

	AddUser(httptest.NewRecorder(), r)

	got := only(t, rec, audit.IAMMembershipAdd)
	assert.Equal(t, audit.UserTarget(newcomer.ID, "new"), got.Event.Target)
	assert.Equal(t, f.project.ID, got.Event.ProjectID)
	assert.Equal(t, audit.MembershipMetadata{Role: "task_runner"}, got.Event.Metadata)
}

func TestRemoveUser_IsRecorded(t *testing.T) {
	f := newMembershipFixture(t)
	r, rec := f.request(http.MethodDelete, "", f.owner, db.ProjectOwner, &f.member)

	RemoveUser(httptest.NewRecorder(), r)

	assert.Equal(t, audit.MembershipMetadata{Role: "manager"}, only(t, rec, audit.IAMMembershipRemove).Event.Metadata)
}

func TestLeftProject_IsRecorded(t *testing.T) {
	f := newMembershipFixture(t)
	r, rec := f.request(http.MethodDelete, "", f.member, db.ProjectManager, nil)

	LeftProject(httptest.NewRecorder(), r)

	assert.Equal(t, audit.MembershipMetadata{Role: "manager", SelfRemoval: true}, only(t, rec, audit.IAMMembershipRemove).Event.Metadata)
}

func TestLeftProject_OwnerRefusalIsRecorded(t *testing.T) {
	f := newMembershipFixture(t)
	r, rec := f.request(http.MethodDelete, "", f.owner, db.ProjectOwner, nil)

	LeftProject(httptest.NewRecorder(), r)

	got := only(t, rec, audit.IAMMembershipRemove)
	assert.Equal(t, audit.OutcomeFailure, got.Event.Outcome)
	assert.Equal(t, audit.ReasonOwnerSelfChange, got.Event.Reason)
}

func TestUpdateUser_RoleChangeIsRecorded(t *testing.T) {
	f := newMembershipFixture(t)
	r, rec := f.request(http.MethodPut, `{"role":"task_runner"}`, f.owner, db.ProjectOwner, &f.member)

	UpdateUser(httptest.NewRecorder(), r)

	assert.Equal(t, audit.ProjectRoleMetadata{OldRole: "manager", NewRole: "task_runner"}, only(t, rec, audit.IAMProjectRoleChange).Event.Metadata)
}

func TestUpdateUser_OwnerSelfChangeIsRecorded(t *testing.T) {
	f := newMembershipFixture(t)
	r, rec := f.request(http.MethodPut, `{"role":"guest"}`, f.owner, db.ProjectOwner, &f.owner)

	UpdateUser(httptest.NewRecorder(), r)

	got := only(t, rec, audit.IAMProjectRoleChange)
	assert.Equal(t, audit.ReasonOwnerSelfChange, got.Event.Reason)
	assert.Equal(t, audit.ProjectRoleMetadata{OldRole: "owner"}, got.Event.Metadata)
}
