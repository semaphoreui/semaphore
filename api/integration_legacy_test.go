package api

import (
	"bytes"
	"crypto/sha256"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func itoa(i int) string { return strconv.Itoa(i) }

// legacyWebhooks is a project configured the pre-2.20.19 way, seeded on the
// old schema by hand and then migrated: it is what an upgraded installation
// looks like, and the tests below send it the webhooks it used to receive.
type legacyWebhooks struct {
	store                           *templateProbeStore
	project                         db.Project
	key                             db.AccessKey
	searchableToken, searchableNone int // behind the project aliases
	ownAliases, workflowHook        int // behind their own aliases
	workflowID                      int
}

func seedLegacyWebhooks(t *testing.T) legacyWebhooks {
	t.Helper()
	before := "2.20.16"
	base := sql.InitConfigCreateTestStoreAt(&before)
	store := &templateProbeStore{Store: base, failTemplate: true}
	exec := func(query string, args ...any) {
		_, err := base.Sql().Exec(base.PrepareQuery(query), args...)
		require.NoError(t, err)
	}
	selectInt := func(query string, args ...any) int {
		v, err := base.Sql().SelectInt(base.PrepareQuery(query), args...)
		require.NoError(t, err)
		return int(v)
	}

	project, err := store.CreateProject(db.Project{Name: "legacy"})
	require.NoError(t, err)
	key, err := store.CreateAccessKey(db.AccessKey{Name: "hook secret", Type: db.AccessKeyString, String: "s3cret", ProjectID: &project.ID})
	require.NoError(t, err)
	noneKey, err := store.CreateAccessKey(db.AccessKey{Name: "none", Type: db.AccessKeyNone, ProjectID: &project.ID})
	require.NoError(t, err)
	repo, err := store.CreateRepository(db.Repository{Name: "r", ProjectID: project.ID, GitURL: "git@example.com:t/t.git", GitBranch: "main", SSHKeyID: noneKey.ID})
	require.NoError(t, err)
	template, err := store.CreateTemplate(db.Template{Name: "t", ProjectID: project.ID, RepositoryID: repo.ID, Playbook: "t.yml", App: db.AppBash})
	require.NoError(t, err)
	exec("insert into project__workflow_template (project_id, name) values (?, ?)", project.ID, "wf")
	workflowID := selectInt("select id from project__workflow_template where project_id = ? and name = ?", project.ID, "wf")

	seed := func(name string, templateID, workflowID *int, method string, secret *int, header string, searchable bool) int {
		exec("insert into project__integration (project_id, name, template_id, workflow_id, auth_method, auth_secret_id, auth_header, searchable) values (?, ?, ?, ?, ?, ?, ?, ?)",
			project.ID, name, templateID, workflowID, method, secret, header, searchable)
		return selectInt("select id from project__integration where project_id = ? and name = ?", project.ID, name)
	}
	f := legacyWebhooks{store: store, project: project, key: key, workflowID: workflowID}
	f.searchableToken = seed("searchable token", &template.ID, nil, "token", &key.ID, "X-Token", true)
	f.searchableNone = seed("searchable none", &template.ID, nil, "", nil, "", true)
	f.ownAliases = seed("own aliases", &template.ID, nil, "hmac", &key.ID, "X-Sig", false)
	f.workflowHook = seed("workflow hook", nil, &workflowID, "", nil, "", false)

	alias := func(integrationID *int, alias string) {
		exec("insert into project__integration_alias (project_id, integration_id, alias) values (?, ?, ?)", project.ID, integrationID, alias)
	}
	alias(nil, "project-alias")
	alias(&f.searchableToken, "dead-alias")
	alias(&f.ownAliases, "own-alias")
	alias(&f.workflowHook, "workflow-alias")
	exec("insert into project__integration_matcher (name, integration_id, match_type, `method`, body_data_type, `key`, `value`) values ('m', ?, 'header', 'equals', '', 'X-Event', 'push')", f.searchableToken)

	require.NoError(t, db.Migrate(base, nil))
	return f
}

func (f legacyWebhooks) post(t *testing.T, ctrl *IntegrationController, alias string, headers map[string]string) []int {
	t.Helper()
	f.store.ran = nil
	payload := []byte(`{"event":"deploy"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/integrations/"+alias, bytes.NewReader(payload))
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	req = mux.SetURLVars(req, map[string]string{"integration_alias": alias})
	w := httptest.NewRecorder()
	ctrl.ReceiveIntegration(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
	return f.store.ran
}

// TestReceiveIntegration_MigratedConfigurationBehavesAsBefore sends an
// upgraded installation the webhooks of the old configuration. Each
// expectation is what the pre-redesign receiver did with the same request,
// except where noted.
func TestReceiveIntegration_MigratedConfigurationBehavesAsBefore(t *testing.T) {
	f := seedLegacyWebhooks(t)
	payload := []byte(`{"event":"deploy"}`)
	svc := &keyedIntegrationService{keys: map[int]db.AccessKey{f.key.ID: f.key}}
	wfSvc := &stubWorkflowService{run: db.WorkflowRun{ID: 1, ProjectID: f.project.ID, WorkflowTemplateID: f.workflowID}}
	wfStore := &stubWorkflowStore{workflows: map[int]db.WorkflowTemplate{f.workflowID: {ID: f.workflowID, ProjectID: f.project.ID, Name: "wf"}}}
	ctrl := NewIntegrationController(f.store, svc, wfStore, wfSvc)

	t.Run("project alias with token and matching event runs the token integration", func(t *testing.T) {
		ran := f.post(t, ctrl, "project-alias", map[string]string{"X-Token": "s3cret", "X-Event": "push"})
		// Before: the token integration only (its matcher matches, its token checks out).
		// Now also "searchable none": it has no matchers, and with the new rule an
		// integration without matchers runs on every authorized request, where
		// before a searchable integration without matchers never ran (JOURNAL
		// decision on zero matchers).
		assert.ElementsMatch(t, []int{f.searchableToken, f.searchableNone}, ran)
	})

	t.Run("project alias with the wrong token skips the token integration", func(t *testing.T) {
		ran := f.post(t, ctrl, "project-alias", map[string]string{"X-Token": "nope", "X-Event": "push"})
		assert.Equal(t, []int{f.searchableNone}, ran)
	})

	t.Run("project alias with a non-matching event skips the matcher integration", func(t *testing.T) {
		ran := f.post(t, ctrl, "project-alias", map[string]string{"X-Token": "s3cret", "X-Event": "ping"})
		assert.Equal(t, []int{f.searchableNone}, ran)
	})

	t.Run("own alias with a valid hmac runs its integration, matchers aside", func(t *testing.T) {
		ran := f.post(t, ctrl, "own-alias", map[string]string{"X-Sig": hmacHashPayload("s3cret", payload, sha256.New)})
		assert.Equal(t, []int{f.ownAliases}, ran)
	})

	t.Run("own alias with a bad hmac runs nothing", func(t *testing.T) {
		ran := f.post(t, ctrl, "own-alias", map[string]string{"X-Sig": "deadbeef"})
		assert.Empty(t, ran)
	})

	t.Run("the alias of a searchable integration still matches nothing", func(t *testing.T) {
		ran := f.post(t, ctrl, "dead-alias", map[string]string{"X-Token": "s3cret", "X-Event": "push"})
		assert.Empty(t, ran)
	})

	t.Run("the workflow alias starts the workflow", func(t *testing.T) {
		ran := f.post(t, ctrl, "workflow-alias", nil)
		assert.Equal(t, []int{f.workflowHook}, ran)
		assert.Len(t, wfSvc.calls, 1)
	})
}
