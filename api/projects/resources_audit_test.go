package projects

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	proFactory "github.com/semaphoreui/semaphore/pro/db/factory"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/semaphoreui/semaphore/services/audit/audittest"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type resourceFixture struct {
	store   *sql.SqlDb
	project db.Project
	user    db.User
	key     db.AccessKey
}

func newResourceFixture(t *testing.T) resourceFixture {
	t.Helper()
	config := util.Config
	t.Cleanup(func() { util.Config = config })
	store := sql.InitConfigCreateTestStore()
	// Template validation and the demo project need the apps whitelisted.
	util.Config.Apps = map[string]util.App{}
	for _, app := range []db.TemplateApp{db.AppAnsible, db.AppTofu, db.AppTerragrunt, db.AppBash, db.AppPowerShell} {
		util.Config.Apps[string(app)] = util.App{Active: true}
	}
	project, err := store.CreateProject(db.Project{Name: "p"})
	require.NoError(t, err)
	user, err := store.CreateUser(db.UserWithPwd{Pwd: "verystrongpassword1", User: db.User{Username: "alice", Name: "alice", Email: "alice@example.com", Admin: true}})
	require.NoError(t, err)
	key, err := store.CreateAccessKey(db.AccessKey{Name: "none", Type: db.AccessKeyNone, ProjectID: &project.ID})
	require.NoError(t, err)
	return resourceFixture{store: store, project: project, user: user, key: key}
}

// request sets what the router middlewares would set, plus values such as "inventory".
func (f resourceFixture) request(method, body string, values map[string]any, vars map[string]string) (*http.Request, *audittest.Recorder) {
	rec := &audittest.Recorder{}
	r := httptest.NewRequest(method, "/api/project/"+strconv.Itoa(f.project.ID)+"/x", strings.NewReader(body))
	r = helpers.SetContextValue(r, "store", f.store)
	r = helpers.SetContextValue(r, "user", &f.user)
	r = helpers.SetContextValue(r, "project", f.project)
	r = helpers.SetContextValue(r, "log_writer", nopLogWriter{})
	r = helpers.SetContextValue(r, "audit", rec)
	for key, value := range values {
		r = helpers.SetContextValue(r, key, value)
	}
	if vars != nil {
		r = mux.SetURLVars(r, vars)
	}
	return r.WithContext(audit.WithActor(r.Context(), audit.UserActor(f.user.ID, f.user.Username, audit.AuthSession, ""))), rec
}

func TestAddProject_IsRecorded(t *testing.T) {
	tests := []struct {
		name string
		body string
		demo bool
	}{
		{"plain", `{"name":"ops"}`, false},
		{"demo", `{"name":"ops","demo":true}`, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newResourceFixture(t)
			r, rec := f.request(http.MethodPost, tt.body, nil, nil)
			w := httptest.NewRecorder()

			NewProjectsController(storeKeyService{&mockAccessKeyService{}, f.store}).AddProject(w, r)

			require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
			got := only(t, rec, audit.ResourceProjectCreate)
			assert.Equal(t, audit.TargetProject, got.Event.Target.Type)
			assert.Equal(t, "ops", got.Event.Target.Name)
			assert.Equal(t, audit.ProjectCreateMetadata{Demo: tt.demo}, got.Event.Metadata)
		})
	}
}

func TestAddProject_DeniedIsRecorded(t *testing.T) {
	f := newResourceFixture(t)
	f.user.Admin = false
	util.Config.NonAdminCanCreateProject = false
	r, rec := f.request(http.MethodPost, `{"name":"ops"}`, nil, nil)
	w := httptest.NewRecorder()

	NewProjectsController(storeKeyService{&mockAccessKeyService{}, f.store}).AddProject(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, audit.DenyMetadata{Method: http.MethodPost, Permission: "admin"}, only(t, rec, audit.AuthAuthorizationDeny).Event.Metadata)
}

// storeKeyService stores keys for real, because the demo project refers to the None key.
type storeKeyService struct {
	*mockAccessKeyService
	store db.Store
}

func (s storeKeyService) Create(key db.AccessKey) (db.AccessKey, error) {
	return s.store.CreateAccessKey(key)
}

type fakeProjectService struct{}

func (fakeProjectService) UpdateProject(db.Project) error { return nil }
func (fakeProjectService) DeleteProject(int) error        { return nil }

func TestUpdateAndDeleteProject_AreRecorded(t *testing.T) {
	f := newResourceFixture(t)
	controller := &ProjectController{ProjectService: fakeProjectService{}}

	r, rec := f.request(http.MethodPut, `{"id":`+strconv.Itoa(f.project.ID)+`,"name":"renamed"}`, nil, nil)
	controller.UpdateProject(httptest.NewRecorder(), r)
	assert.Equal(t, audit.ResourceTarget(audit.TargetProject, f.project.ID, "renamed"), only(t, rec, audit.ResourceProjectUpdate).Event.Target)

	r, rec = f.request(http.MethodDelete, "", nil, nil)
	controller.DeleteProject(httptest.NewRecorder(), r)
	got := only(t, rec, audit.ResourceProjectDelete)
	assert.Equal(t, audit.ResourceTarget(audit.TargetProject, f.project.ID, "p"), got.Event.Target)
	assert.Equal(t, f.project.ID, got.Event.ProjectID)
}

func TestBackupExportAndRestore_AreRecorded(t *testing.T) {
	f := newResourceFixture(t)
	controller := NewBackupController(proFactory.NewWorkflowStore(f.store))

	r, rec := f.request(http.MethodGet, "", nil, nil)
	w := httptest.NewRecorder()
	controller.GetBackup(w, r)
	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, audit.ResourceTarget(audit.TargetProject, f.project.ID, "p"), only(t, rec, audit.ResourceProjectBackupExport).Event.Target)

	backup := strings.Replace(w.Body.String(), `"name": "p"`, `"name": "restored"`, 1)
	r, rec = f.request(http.MethodPost, backup, nil, nil)
	w = httptest.NewRecorder()
	controller.Restore(w, r)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	got := only(t, rec, audit.ResourceProjectBackupRestore)
	assert.Equal(t, "restored", got.Event.Target.Name)
	assert.Equal(t, 1, got.Event.Metadata.(audit.BackupRestoreMetadata).Objects["keys"])
}

func TestInventoryEvents(t *testing.T) {
	f := newResourceFixture(t)
	body := `{"name":"prod","project_id":` + strconv.Itoa(f.project.ID) + `,"type":"static","inventory":"[all]"}`

	r, rec := f.request(http.MethodPost, body, nil, nil)
	w := httptest.NewRecorder()
	AddInventory(w, r)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	created := only(t, rec, audit.ResourceInventoryCreate)
	assert.Equal(t, "prod", created.Event.Target.Name)
	assert.Equal(t, f.project.ID, created.Event.ProjectID)

	inventory, err := f.store.GetInventory(f.project.ID, mustAtoi(t, created.Event.Target.ID))
	require.NoError(t, err)

	r, rec = f.request(http.MethodPut, `{"id":`+strconv.Itoa(inventory.ID)+`,"name":"prod2","project_id":`+strconv.Itoa(f.project.ID)+`,"type":"static","inventory":"[all]"}`, map[string]any{"inventory": inventory}, nil)
	UpdateInventory(httptest.NewRecorder(), r)
	assert.Equal(t, audit.ResourceTarget(audit.TargetInventory, inventory.ID, "prod2"), only(t, rec, audit.ResourceInventoryUpdate).Event.Target)

	r, rec = f.request(http.MethodDelete, "", map[string]any{"inventory": inventory}, nil)
	RemoveInventory(httptest.NewRecorder(), r)
	assert.Equal(t, audit.ResourceTarget(audit.TargetInventory, inventory.ID, "prod"), only(t, rec, audit.ResourceInventoryDelete).Event.Target)
}

func TestRepositoryEvents_NeverRecordTheURL(t *testing.T) {
	f := newResourceFixture(t)
	url := "https://user:secret-token@git.example.com/ops.git"
	body := `{"name":"ops","project_id":` + strconv.Itoa(f.project.ID) + `,"git_url":"` + url + `","git_branch":"main","ssh_key_id":` + strconv.Itoa(f.key.ID) + `}`

	r, rec := f.request(http.MethodPost, body, nil, nil)
	w := httptest.NewRecorder()
	AddRepository(w, r)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	created := only(t, rec, audit.ResourceRepositoryCreate)
	assert.Equal(t, "ops", created.Event.Target.Name)
	assert.NotContains(t, created.Event.Target.ID+created.Event.Target.Name, "secret-token")

	repo, err := f.store.GetRepository(f.project.ID, mustAtoi(t, created.Event.Target.ID))
	require.NoError(t, err)

	r, rec = f.request(http.MethodDelete, "", map[string]any{"repository": repo}, nil)
	RemoveRepository(httptest.NewRecorder(), r)
	assert.Equal(t, audit.ResourceTarget(audit.TargetRepository, repo.ID, "ops"), only(t, rec, audit.ResourceRepositoryDelete).Event.Target)
}

func TestHostConfigEvents_CarryNoName(t *testing.T) {
	f := newResourceFixture(t)
	sshKey, err := f.store.CreateAccessKey(db.AccessKey{Name: "ssh", Type: db.AccessKeySSH, ProjectID: &f.project.ID})
	require.NoError(t, err)
	body := `{"project_id":` + strconv.Itoa(f.project.ID) + `,"type":"host","host":"github.com","ssh_key_id":` + strconv.Itoa(sshKey.ID) + `}`

	r, rec := f.request(http.MethodPost, body, nil, nil)
	w := httptest.NewRecorder()
	AddHostConfig(w, r)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	created := only(t, rec, audit.ResourceHostConfigCreate)
	assert.Empty(t, created.Event.Target.Name, "a URL mapping keeps its URL in name")
	assert.Equal(t, audit.HostConfigMetadata{Type: "host"}, created.Event.Metadata)

	hostConfig := db.HostConfig{ID: mustAtoi(t, created.Event.Target.ID), ProjectID: f.project.ID, Type: db.HostConfigHost, Name: "github.com", SSHKeyID: sshKey.ID}
	r, rec = f.request(http.MethodDelete, "", map[string]any{"host_config": hostConfig}, nil)
	RemoveHostConfig(httptest.NewRecorder(), r)
	assert.Equal(t, &audit.Target{Type: audit.TargetHostConfig, ID: created.Event.Target.ID}, only(t, rec, audit.ResourceHostConfigDelete).Event.Target)
}

func mustAtoi(t *testing.T, s string) int {
	t.Helper()
	n, err := strconv.Atoi(s)
	require.NoError(t, err)
	return n
}
