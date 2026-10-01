package projects

import (
	"encoding/json"
	"errors"
	"fmt"
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
	"github.com/semaphoreui/semaphore/services/schedules"
	"github.com/semaphoreui/semaphore/services/server"
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

func TestAddProject_SetupFailureIsPartial(t *testing.T) {
	f := newResourceFixture(t)
	util.Config.Apps = map[string]util.App{}
	r, rec := f.request(http.MethodPost, `{"name":"ops","demo":true}`, nil, nil)
	w := httptest.NewRecorder()

	NewProjectsController(storeKeyService{&mockAccessKeyService{}, f.store}).AddProject(w, r)

	assert.GreaterOrEqual(t, w.Code, http.StatusBadRequest, w.Body.String())
	got := only(t, rec, audit.ResourceProjectCreate)
	assert.Equal(t, audit.OutcomeSuccess, got.Event.Outcome, "the project is stored")
	assert.Equal(t, audit.ReasonSetupFailed, got.Event.Reason)
	assert.Equal(t, audit.ProjectCreateMetadata{Demo: true, Partial: true}, got.Event.Metadata)
	assert.Equal(t, "ops", got.Event.Target.Name)
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

func (f resourceFixture) template(t *testing.T) db.Template {
	t.Helper()
	repo, err := f.store.CreateRepository(db.Repository{Name: "r", ProjectID: f.project.ID, GitURL: "git@example.com:r.git", GitBranch: "main", SSHKeyID: f.key.ID})
	require.NoError(t, err)
	inventory, err := f.store.CreateInventory(db.Inventory{Name: "i", ProjectID: f.project.ID, Type: db.InventoryStatic})
	require.NoError(t, err)
	template, err := f.store.CreateTemplate(db.Template{Name: "deploy", ProjectID: f.project.ID, RepositoryID: repo.ID, InventoryID: &inventory.ID, Playbook: "site.yml", App: db.AppAnsible})
	require.NoError(t, err)
	return template
}

func TestTemplateEvents(t *testing.T) {
	f := newResourceFixture(t)
	util.Config.Apps = map[string]util.App{"ansible": {}}
	template := f.template(t)

	r, rec := f.request(http.MethodPut, `{"id":`+strconv.Itoa(template.ID)+`,"project_id":`+strconv.Itoa(f.project.ID)+`,"name":"deploy2","repository_id":`+strconv.Itoa(template.RepositoryID)+`,"inventory_id":`+strconv.Itoa(*template.InventoryID)+`,"playbook":"site.yml","app":"ansible"}`, map[string]any{"template": template}, nil)
	w := httptest.NewRecorder()
	UpdateTemplate(w, r)
	require.Equal(t, http.StatusNoContent, w.Code, w.Body.String())
	updated := only(t, rec, audit.ResourceTemplateUpdate)
	assert.Equal(t, audit.ResourceTarget(audit.TargetTemplate, template.ID, "deploy2"), updated.Event.Target)
	assert.Equal(t, audit.TemplateMetadata{App: "ansible"}, updated.Event.Metadata)

	inventory, err := f.store.CreateInventory(db.Inventory{Name: "extra", ProjectID: f.project.ID, Type: db.InventoryStatic})
	require.NoError(t, err)
	r, rec = f.request(http.MethodPost, "", map[string]any{"template": template, "inventory": inventory}, nil)
	AttachInventory(httptest.NewRecorder(), r)
	assert.Equal(t, audit.TemplateInventoryMetadata{InventoryID: inventory.ID}, only(t, rec, audit.ResourceTemplateAttachInventory).Event.Metadata)

	r, rec = f.request(http.MethodDelete, "", map[string]any{"template": template}, nil)
	RemoveTemplate(httptest.NewRecorder(), r)
	assert.Equal(t, audit.ResourceTarget(audit.TargetTemplate, template.ID, "deploy"), only(t, rec, audit.ResourceTemplateDelete).Event.Target)
}

func TestAddTemplate_RecordsTheCreatedInventory(t *testing.T) {
	f := newResourceFixture(t)
	util.Config.Apps = map[string]util.App{"terraform": {}}
	repo, err := f.store.CreateRepository(db.Repository{Name: "r", ProjectID: f.project.ID, GitURL: "git@example.com:r.git", GitBranch: "main", SSHKeyID: f.key.ID})
	require.NoError(t, err)

	r, rec := f.request(http.MethodPost, `{"name":"infra","repository_id":`+strconv.Itoa(repo.ID)+`,"app":"terraform"}`, nil, nil)
	w := httptest.NewRecorder()
	AddTemplate(w, r)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())

	meta := only(t, rec, audit.ResourceTemplateCreate).Event.Metadata.(audit.TemplateMetadata)
	assert.Equal(t, "terraform", meta.App)
	assert.NotZero(t, meta.CreatedInventoryID)
}

func TestAddTemplate_InventoryFailureIsPartial(t *testing.T) {
	f := newResourceFixture(t)
	util.Config.Apps = map[string]util.App{"terraform": {}}
	repo, err := f.store.CreateRepository(db.Repository{Name: "r", ProjectID: f.project.ID, GitURL: "git@example.com:r.git", GitBranch: "main", SSHKeyID: f.key.ID})
	require.NoError(t, err)

	other, err := f.store.CreateProject(db.Project{Name: "other"})
	require.NoError(t, err)
	foreign, err := f.store.CreateInventory(db.Inventory{Name: "foreign", ProjectID: other.ID, Type: db.InventoryTerraformWorkspace})
	require.NoError(t, err)

	r, rec := f.request(http.MethodPost, `{"name":"infra","repository_id":`+strconv.Itoa(repo.ID)+`,"inventory_id":`+strconv.Itoa(foreign.ID)+`,"app":"terraform"}`, nil, nil)
	w := httptest.NewRecorder()
	AddTemplate(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	got := only(t, rec, audit.ResourceTemplateCreate)
	assert.Equal(t, audit.OutcomeSuccess, got.Event.Outcome, "the template is stored")
	assert.Equal(t, audit.ReasonInventoryFailed, got.Event.Reason)
	assert.Equal(t, audit.TemplateMetadata{App: "terraform", Partial: true}, got.Event.Metadata)
	stored, err := f.store.GetTemplates(f.project.ID, db.TemplateFilter{}, db.RetrieveQueryParams{})
	require.NoError(t, err)
	require.Len(t, stored, 1)
	assert.Equal(t, audit.ResourceTarget(audit.TargetTemplate, stored[0].ID, "infra"), got.Event.Target)
}

func TestScheduleEvents(t *testing.T) {
	f := newResourceFixture(t)
	template := f.template(t)
	schedule, err := f.store.CreateSchedule(db.Schedule{Name: "nightly", ProjectID: f.project.ID, TemplateID: template.ID, CronFormat: "0 0 * * *"})
	require.NoError(t, err)
	util.Config.Schedule = &util.ScheduleConfig{Timezone: "UTC"}
	pool := schedules.CreateSchedulePool(f.store, nil, nil, nil)
	t.Cleanup(pool.Destroy)

	tests := []struct {
		body string
		kind audit.Kind
	}{
		{`{"active":true}`, audit.ResourceScheduleActivate},
		{`{"active":false}`, audit.ResourceScheduleDeactivate},
	}
	for _, tt := range tests {
		t.Run(string(tt.kind), func(t *testing.T) {
			r, rec := f.request(http.MethodPut, tt.body, map[string]any{"schedule": schedule, "schedule_pool": pool}, nil)
			SetScheduleActive(httptest.NewRecorder(), r)
			got := only(t, rec, tt.kind)
			assert.Equal(t, audit.ResourceTarget(audit.TargetSchedule, schedule.ID, "nightly"), got.Event.Target)
			assert.Equal(t, audit.ScheduleMetadata{TemplateID: template.ID}, got.Event.Metadata)
		})
	}

	r, rec := f.request(http.MethodDelete, "", map[string]any{"schedule": schedule, "schedule_pool": pool}, nil)
	RemoveSchedule(httptest.NewRecorder(), r)
	assert.Nil(t, only(t, rec, audit.ResourceScheduleDelete).Event.Metadata)
}

func TestUpdateSchedule_ActiveChangeAlsoRecordsSwitch(t *testing.T) {
	f := newResourceFixture(t)
	template := f.template(t)
	schedule, err := f.store.CreateSchedule(db.Schedule{Name: "nightly", ProjectID: f.project.ID, TemplateID: template.ID, CronFormat: "0 0 * * *", Active: false})
	require.NoError(t, err)
	util.Config.Schedule = &util.ScheduleConfig{Timezone: "UTC"}
	pool := schedules.CreateSchedulePool(f.store, nil, nil, nil)
	t.Cleanup(pool.Destroy)

	tests := []struct {
		name   string
		active bool
		want   []audit.Kind
	}{
		{"switched on", true, []audit.Kind{audit.ResourceScheduleUpdate, audit.ResourceScheduleActivate}},
		{"unchanged", true, []audit.Kind{audit.ResourceScheduleUpdate}},
		{"switched off", false, []audit.Kind{audit.ResourceScheduleUpdate, audit.ResourceScheduleDeactivate}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stored, err := f.store.GetSchedule(f.project.ID, schedule.ID)
			require.NoError(t, err)
			body := `{"id":` + strconv.Itoa(schedule.ID) + `,"project_id":` + strconv.Itoa(f.project.ID) + `,"template_id":` + strconv.Itoa(template.ID) + `,"name":"nightly","cron_format":"0 0 * * *","active":` + strconv.FormatBool(tt.active) + `}`
			r, rec := f.request(http.MethodPut, body, map[string]any{"schedule": stored, "schedule_pool": pool}, nil)
			w := httptest.NewRecorder()
			UpdateSchedule(w, r)
			require.Equal(t, http.StatusNoContent, w.Code, w.Body.String())

			kinds, err := rec.Kinds()
			require.NoError(t, err)
			assert.Equal(t, tt.want, kinds)
			for _, recorded := range rec.All() {
				assert.Equal(t, audit.ResourceTarget(audit.TargetSchedule, schedule.ID, "nightly"), recorded.Event.Target)
				assert.Equal(t, audit.ScheduleMetadata{TemplateID: template.ID}, recorded.Event.Metadata)
			}
		})
	}
}

func TestUpdateKey_RecordsTheStoredType(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{"rename only keeps the old type", `"type":"ssh"`, "none"},
		{"override stores the new type", `"type":"ssh","override_secret":true`, "ssh"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newResourceFixture(t)
			controller := NewKeyController(storeKeyService{&mockAccessKeyService{}, f.store})
			body := `{"id":` + strconv.Itoa(f.key.ID) + `,"name":"renamed","project_id":` + strconv.Itoa(f.project.ID) + `,` + tt.body + `}`
			r, rec := f.request(http.MethodPut, body, map[string]any{"accessKey": f.key}, nil)
			w := httptest.NewRecorder()
			controller.UpdateKey(w, r)
			require.Equal(t, http.StatusNoContent, w.Code, w.Body.String())
			assert.Equal(t, audit.CredentialMetadata{Type: tt.want}, only(t, rec, audit.SecretCredentialUpdate).Event.Metadata)
		})
	}
}

func TestIntegrationEvents(t *testing.T) {
	f := newResourceFixture(t)
	template := f.template(t)

	r, rec := f.request(http.MethodPost, `{"name":"hook","project_id":`+strconv.Itoa(f.project.ID)+`,"template_id":`+strconv.Itoa(template.ID)+`,"auth_method":"none"}`, nil, nil)
	w := httptest.NewRecorder()
	AddIntegration(w, r)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	created := only(t, rec, audit.ResourceIntegrationCreate)
	assert.Equal(t, audit.IntegrationMetadata{TemplateID: template.ID, AuthMethod: "none"}, created.Event.Metadata)
	integrationID := mustAtoi(t, created.Event.Target.ID)
	integration, err := f.store.GetIntegration(f.project.ID, integrationID)
	require.NoError(t, err)

	r, rec = f.request(http.MethodPost, "", map[string]any{"integration": integration}, nil)
	AddIntegrationAlias(httptest.NewRecorder(), r)
	alias := only(t, rec, audit.ResourceIntegrationAliasCreate)
	assert.Empty(t, alias.Event.Target.Name, "the alias value is a bearer secret")
	assert.Equal(t, audit.IntegrationPartMetadata{IntegrationID: integrationID}, alias.Event.Metadata)

	r, rec = f.request(http.MethodDelete, "", nil, map[string]string{"integration_id": strconv.Itoa(integrationID)})
	DeleteIntegration(httptest.NewRecorder(), r)
	assert.Equal(t, audit.ResourceTarget(audit.TargetIntegration, integrationID, ""), only(t, rec, audit.ResourceIntegrationDelete).Event.Target)
}

func TestDeleteIntegration_FailedDeleteIsNotRecorded(t *testing.T) {
	f := newResourceFixture(t)
	r, rec := f.request(http.MethodDelete, "", nil, map[string]string{"integration_id": "999999"})
	DeleteIntegration(httptest.NewRecorder(), r)
	assert.Empty(t, rec.All())
}

func TestIntegrationMatcherAndExtractorEvents(t *testing.T) {
	f := newResourceFixture(t)
	template := f.template(t)
	integration, err := f.store.CreateIntegration(db.Integration{Name: "hook", ProjectID: f.project.ID, TemplateID: template.ID})
	require.NoError(t, err)
	values := map[string]any{"integration": integration}
	id := strconv.Itoa(integration.ID)

	r, rec := f.request(http.MethodPost, `{"name":"branch","integration_id":`+id+`,"match_type":"body","method":"equals","body_data_type":"json","key":"ref","value":"main"}`, values, nil)
	w := httptest.NewRecorder()
	AddIntegrationMatcher(w, r)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	matcher := only(t, rec, audit.ResourceIntegrationMatcherCreate)
	assert.Equal(t, "branch", matcher.Event.Target.Name)
	assert.Equal(t, audit.IntegrationPartMetadata{IntegrationID: integration.ID}, matcher.Event.Metadata)

	r, rec = f.request(http.MethodDelete, "", values, map[string]string{"matcher_id": matcher.Event.Target.ID})
	DeleteIntegrationMatcher(httptest.NewRecorder(), r)
	assert.Equal(t, matcher.Event.Target.ID, only(t, rec, audit.ResourceIntegrationMatcherDelete).Event.Target.ID)

	r, rec = f.request(http.MethodPost, `{"name":"version","integration_id":`+id+`,"value_source":"body","body_data_type":"json","key":"v","variable":"VERSION","variable_type":"environment"}`, values, nil)
	w = httptest.NewRecorder()
	AddIntegrationExtractValue(w, r)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	assert.Equal(t, "version", only(t, rec, audit.ResourceIntegrationExtractorCreate).Event.Target.Name)
}

func TestCredentialEvents_NeverRecordTheSecret(t *testing.T) {
	f := newResourceFixture(t)
	controller := NewKeyController(storeKeyService{&mockAccessKeyService{}, f.store})
	body := `{"name":"deploy","type":"login_password","project_id":` + strconv.Itoa(f.project.ID) + `,"login_password":{"login":"bob","password":"hunter2-secret"}}`

	r, rec := f.request(http.MethodPost, body, nil, nil)
	controller.AddKey(httptest.NewRecorder(), r)
	created := only(t, rec, audit.SecretCredentialCreate)
	assert.Equal(t, "deploy", created.Event.Target.Name)
	assert.Equal(t, audit.CredentialMetadata{Type: "login_password"}, created.Event.Metadata)
	assert.NotContains(t, eventText(created), "hunter2-secret")

	r, rec = f.request(http.MethodDelete, "", map[string]any{"accessKey": f.key}, nil)
	controller.RemoveKey(httptest.NewRecorder(), r)
	assert.Equal(t, audit.ResourceTarget(audit.TargetCredential, f.key.ID, "none"), only(t, rec, audit.SecretCredentialDelete).Event.Target)
}

type fakeSecretStorageService struct {
	server.SecretStorageService
	deleteErr error
}

func (fakeSecretStorageService) Create(s db.SecretStorage) (db.SecretStorage, error) {
	s.ID = 3
	return s, nil
}
func (f fakeSecretStorageService) Delete(int, int) error         { return f.deleteErr }
func (fakeSecretStorageService) SyncSecrets(db.SecretSync) error { return nil }

func TestSecretStorageEvents(t *testing.T) {
	f := newResourceFixture(t)
	controller := NewSecretStorageController(f.store, fakeSecretStorageService{})

	r, rec := f.request(http.MethodPost, `{"name":"vault","type":"vault","project_id":`+strconv.Itoa(f.project.ID)+`}`, nil, nil)
	controller.Add(httptest.NewRecorder(), r)
	created := only(t, rec, audit.SecretStorageCreate)
	assert.Equal(t, audit.ResourceTarget(audit.TargetSecretStorage, 3, "vault"), created.Event.Target)
	assert.Equal(t, audit.SecretStorageMetadata{Type: "vault"}, created.Event.Metadata)

	r, rec = f.request(http.MethodDelete, "", map[string]any{"secretStorage": db.SecretStorage{ID: 3, Name: "vault", ProjectID: f.project.ID}}, map[string]string{"storage_id": "3"})
	controller.Remove(httptest.NewRecorder(), r)
	assert.Equal(t, audit.ResourceTarget(audit.TargetSecretStorage, 3, "vault"), only(t, rec, audit.SecretStorageDelete).Event.Target)
}

func TestRemoveSecretStorage_LeftKeysIsPartial(t *testing.T) {
	f := newResourceFixture(t)
	controller := NewSecretStorageController(f.store, fakeSecretStorageService{deleteErr: fmt.Errorf("%w: boom", server.ErrSecretsLeftBehind)})

	r, rec := f.request(http.MethodDelete, "", map[string]any{"secretStorage": db.SecretStorage{ID: 3, Name: "vault", ProjectID: f.project.ID}}, map[string]string{"storage_id": "3"})
	w := httptest.NewRecorder()
	controller.Remove(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code, "the API answer is unchanged")
	got := only(t, rec, audit.SecretStorageDelete)
	assert.Equal(t, audit.OutcomeSuccess, got.Event.Outcome, "the storage is deleted")
	assert.Equal(t, audit.ReasonKeyFailed, got.Event.Reason)
	assert.Equal(t, audit.DeleteMetadata{Partial: true}, got.Event.Metadata)
}

func TestRemoveSecretStorage_FailedDeleteIsNotRecorded(t *testing.T) {
	f := newResourceFixture(t)
	controller := NewSecretStorageController(f.store, fakeSecretStorageService{deleteErr: errors.New("boom")})

	r, rec := f.request(http.MethodDelete, "", map[string]any{"secretStorage": db.SecretStorage{ID: 3, Name: "vault", ProjectID: f.project.ID}}, map[string]string{"storage_id": "3"})
	controller.Remove(httptest.NewRecorder(), r)

	assert.Empty(t, rec.All())
}

func eventText(recorded audittest.Recorded) string {
	return fmt.Sprintf("%+v", recorded.Event)
}

func TestAddEnvironment_Events(t *testing.T) {
	const secretValue = "s3cr3t-7f3a9c1e-env-value"
	tests := []struct {
		name        string
		secrets     string
		wantCode    int
		wantMeta    audit.EnvironmentMetadata
		wantPartial bool
	}{
		{"all secrets saved", `[{"name":"A","secret":"` + secretValue + `","type":"env","operation":"create"}]`, http.StatusCreated, audit.EnvironmentMetadata{SecretsCreated: 1}, false},
		{"one secret refused", `[{"name":"A","secret":"` + secretValue + `","type":"env","operation":"create"},{"id":42,"type":"env","operation":"delete"}]`, http.StatusNotFound, audit.EnvironmentMetadata{SecretsCreated: 1, Partial: true}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newResourceFixture(t)
			controller := &EnvironmentController{accessKeyRepo: &mockAccessKeyRepo{keys: map[int]db.AccessKey{}}, accessKeyService: &mockAccessKeyService{}}
			body := `{"name":"prod","project_id":` + strconv.Itoa(f.project.ID) + `,"json":"{}","secrets":` + tt.secrets + `}`

			r, rec := f.request(http.MethodPost, body, nil, nil)
			w := httptest.NewRecorder()
			controller.AddEnvironment(w, r)

			assert.Equal(t, tt.wantCode, w.Code, "the API answer is unchanged")
			got := only(t, rec, audit.ResourceEnvironmentCreate)
			assert.Equal(t, audit.OutcomeSuccess, got.Event.Outcome, "the environment is stored")
			assert.Equal(t, tt.wantMeta, got.Event.Metadata)
			if tt.wantPartial {
				assert.Equal(t, audit.ReasonSecretFailed, got.Event.Reason)
			} else {
				assert.Equal(t, audit.ReasonNone, got.Event.Reason)
			}
			recorded, err := json.Marshal(got.Event)
			require.NoError(t, err)
			assert.NotContains(t, string(recorded), secretValue)
		})
	}
}

func TestRemoveEnvironment_IsRecorded(t *testing.T) {
	f := newResourceFixture(t)
	env, err := f.store.CreateEnvironment(db.Environment{Name: "prod", ProjectID: f.project.ID, JSON: "{}"})
	require.NoError(t, err)
	controller := &EnvironmentController{environmentService: fakeEnvironmentService{}}

	r, rec := f.request(http.MethodDelete, "", map[string]any{"environment": env}, nil)
	controller.RemoveEnvironment(httptest.NewRecorder(), r)
	got := only(t, rec, audit.ResourceEnvironmentDelete)
	assert.Equal(t, audit.ResourceTarget(audit.TargetEnvironment, env.ID, "prod"), got.Event.Target)
	assert.Equal(t, audit.ReasonNone, got.Event.Reason)
}

func TestRemoveEnvironment_LeftSecretsIsPartial(t *testing.T) {
	f := newResourceFixture(t)
	env, err := f.store.CreateEnvironment(db.Environment{Name: "prod", ProjectID: f.project.ID, JSON: "{}"})
	require.NoError(t, err)
	controller := &EnvironmentController{environmentService: fakeEnvironmentService{deleteErr: fmt.Errorf("%w: boom", server.ErrSecretsLeftBehind)}}

	r, rec := f.request(http.MethodDelete, "", map[string]any{"environment": env}, nil)
	w := httptest.NewRecorder()
	controller.RemoveEnvironment(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code, "the API answer is unchanged")
	got := only(t, rec, audit.ResourceEnvironmentDelete)
	assert.Equal(t, audit.OutcomeSuccess, got.Event.Outcome, "the environment is deleted")
	assert.Equal(t, audit.ReasonSecretFailed, got.Event.Reason)
	assert.Equal(t, audit.DeleteMetadata{Partial: true}, got.Event.Metadata)
}

type fakeEnvironmentService struct {
	server.EnvironmentService
	deleteErr error
}

func (f fakeEnvironmentService) Delete(int, int) error { return f.deleteErr }
