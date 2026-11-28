package api

import (
	"bytes"
	"crypto/sha256"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerifyAuth(t *testing.T) {
	payload := []byte(`{"event":"deploy"}`)
	stringKey := db.AccessKey{Type: db.AccessKeyString, String: "s3cret"}
	loginKey := db.AccessKey{Type: db.AccessKeyLoginPassword, LoginPassword: db.LoginPassword{Login: "bob", Password: "pw"}}
	sig256 := hmacHashPayload("s3cret", payload, sha256.New)

	tests := []struct {
		name    string
		auth    db.IntegrationAuth
		headers map[string]string
		basic   *[2]string
		wantOk  bool
		reason  string
	}{
		{"none accepts anything", db.IntegrationAuth{AuthMethod: db.IntegrationAuthNone}, nil, nil, true, ""},
		{"token matches", db.IntegrationAuth{AuthMethod: db.IntegrationAuthToken, AuthHeader: "X-Token", AuthSecret: stringKey}, map[string]string{"X-Token": "s3cret"}, nil, true, ""},
		{"token differs", db.IntegrationAuth{AuthMethod: db.IntegrationAuthToken, AuthHeader: "X-Token", AuthSecret: stringKey}, map[string]string{"X-Token": "wrong"}, nil, false, "Invalid verification token"},
		{"token header missing", db.IntegrationAuth{AuthMethod: db.IntegrationAuthToken, AuthHeader: "X-Token", AuthSecret: stringKey}, nil, nil, false, "Invalid verification token"},
		{"token without a usable key", db.IntegrationAuth{AuthMethod: db.IntegrationAuthToken, AuthHeader: "X-Token"}, map[string]string{"X-Token": ""}, nil, false, "Auth secret is empty or its key type does not fit the auth method"},
		{"hmac sha256 matches", db.IntegrationAuth{AuthMethod: db.IntegrationAuthHmac, AuthHeader: "X-Sig", AuthSecret: stringKey}, map[string]string{"X-Sig": sig256}, nil, true, ""},
		{"hmac sha256 differs", db.IntegrationAuth{AuthMethod: db.IntegrationAuthHmac, AuthHeader: "X-Sig", AuthSecret: stringKey}, map[string]string{"X-Sig": "deadbeef"}, nil, false, "Invalid HMAC-SHA256 signature"},
		{"github matches", db.IntegrationAuth{AuthMethod: db.IntegrationAuthGitHub, AuthSecret: stringKey}, map[string]string{"X-Hub-Signature-256": "sha256=" + sig256}, nil, true, ""},
		{"github without prefix", db.IntegrationAuth{AuthMethod: db.IntegrationAuthGitHub, AuthSecret: stringKey}, map[string]string{"X-Hub-Signature-256": sig256}, nil, false, "Invalid GitHub signature"},
		{"bitbucket matches", db.IntegrationAuth{AuthMethod: db.IntegrationAuthBitbucket, AuthSecret: stringKey}, map[string]string{"x-hub-signature": "sha256=" + sig256}, nil, true, ""},
		{"basic matches", db.IntegrationAuth{AuthMethod: db.IntegrationAuthBasic, AuthSecret: loginKey}, nil, &[2]string{"bob", "pw"}, true, ""},
		{"basic wrong login", db.IntegrationAuth{AuthMethod: db.IntegrationAuthBasic, AuthSecret: loginKey}, nil, &[2]string{"alice", "pw"}, false, "Invalid BasicAuth: incorrect login or password"},
		{"basic wrong password", db.IntegrationAuth{AuthMethod: db.IntegrationAuthBasic, AuthSecret: loginKey}, nil, &[2]string{"bob", "nope"}, false, "Invalid BasicAuth: incorrect login or password"},
		{"basic without credentials", db.IntegrationAuth{AuthMethod: db.IntegrationAuthBasic, AuthSecret: loginKey}, nil, nil, false, "No BasicAuth credentials in the request"},
		{"basic with a string key", db.IntegrationAuth{AuthMethod: db.IntegrationAuthBasic, AuthSecret: stringKey}, nil, &[2]string{"", "s3cret"}, false, "BasicAuth needs a login/password credential"},
		{"basic with an empty stored password", db.IntegrationAuth{AuthMethod: db.IntegrationAuthBasic, AuthSecret: db.AccessKey{Type: db.AccessKeyLoginPassword, LoginPassword: db.LoginPassword{Login: "bob"}}}, nil, &[2]string{"bob", ""}, false, "BasicAuth credential has no password"},
		{"unknown method", db.IntegrationAuth{AuthMethod: "jwt"}, nil, nil, false, "Unknown verification method: jwt"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/integrations/x", bytes.NewReader(payload))
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			if tt.basic != nil {
				req.SetBasicAuth(tt.basic[0], tt.basic[1])
			}

			ok, reason := verifyAuth(req, payload, tt.auth)

			assert.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.reason, reason)
		})
	}
}

func TestMatchAll(t *testing.T) {
	header := http.Header{}
	header.Set("X-Event", "push")
	push := db.IntegrationMatcher{MatchType: db.IntegrationMatchHeader, Method: db.IntegrationMatchMethodEquals, Key: "X-Event", Value: "push"}
	ping := db.IntegrationMatcher{MatchType: db.IntegrationMatchHeader, Method: db.IntegrationMatchMethodEquals, Key: "X-Event", Value: "ping"}

	assert.True(t, matchAll(nil, header, nil), "no matchers accept every request")
	assert.True(t, matchAll([]db.IntegrationMatcher{push}, header, nil))
	assert.False(t, matchAll([]db.IntegrationMatcher{ping}, header, nil))
	assert.False(t, matchAll([]db.IntegrationMatcher{push, ping}, header, nil), "every matcher must accept")
}

// post sends payload to the alias with headers and returns the recorder and
// the integrations that ran, in order.
func (f receiveFixture) post(t *testing.T, ctrl *IntegrationController, alias string, payload []byte, headers map[string]string) (*httptest.ResponseRecorder, []int) {
	t.Helper()
	f.store.ran = nil
	req := httptest.NewRequest(http.MethodPost, "/api/integrations/"+alias, bytes.NewReader(payload))
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	req = mux.SetURLVars(req, map[string]string{"integration_alias": alias})
	w := httptest.NewRecorder()
	ctrl.ReceiveIntegration(w, req)
	return w, f.store.ran
}

func TestReceiveIntegration_UnknownAlias(t *testing.T) {
	f := newReceiveFixture(t, db.IntegrationAuthNone, "", "known")
	group := f.addGroup(t, "group", nil)
	f.addIntegration(t, group, "hook", nil)
	svc := &stubIntegrationService{}
	ctrl := NewIntegrationController(f.store, svc, nil, nil)

	w, ran := f.post(t, ctrl, "unknown", []byte(`{}`), nil)

	assert.Equal(t, http.StatusNoContent, w.Code, "an unknown alias answers like a known one")
	assert.Empty(t, ran)
	assert.Zero(t, svc.calls, "nothing is loaded for an unknown alias")
	assert.Empty(t, w.Header(), "no header hints at what exists")
}

// An endpoint without auth, one group inheriting it and one overriding with a
// token: the token decides only for the overriding group.
func TestReceiveIntegration_GroupOverridesEndpointAuth(t *testing.T) {
	f := newReceiveFixture(t, db.IntegrationAuthNone, "", "hooks")
	inheriting := f.addGroup(t, "inherits", nil)
	open := f.addIntegration(t, inheriting, "open", nil)
	overriding := f.addGroup(t, "overrides", &db.IntegrationAuth{AuthMethod: db.IntegrationAuthToken, AuthHeader: "X-Token"})
	guarded := f.addIntegration(t, overriding, "guarded", nil)
	svc := &stubIntegrationService{key: &db.AccessKey{Type: db.AccessKeyString, String: "s3cret"}}
	ctrl := NewIntegrationController(f.store, svc, nil, nil)

	w, ran := f.post(t, ctrl, f.alias, []byte(`{}`), nil)
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, []int{open.ID}, ran, "without the token only the inheriting group runs")

	_, ran = f.post(t, ctrl, f.alias, []byte(`{}`), map[string]string{"X-Token": "s3cret"})
	assert.ElementsMatch(t, []int{open.ID, guarded.ID}, ran, "with the token both groups run")

	_, ran = f.post(t, ctrl, f.alias, []byte(`{}`), map[string]string{"X-Token": "wrong"})
	assert.Equal(t, []int{open.ID}, ran)
}

// Two groups on one token endpoint, one with its own token: each request runs
// only the group whose secret it carries.
func TestReceiveIntegration_TwoGroupsDifferentAuth(t *testing.T) {
	f := newReceiveFixture(t, db.IntegrationAuthToken, "X-Token", "hooks")
	inheriting := f.addGroup(t, "inherits", nil)
	a := f.addIntegration(t, inheriting, "a", nil)
	overriding := f.addGroup(t, "overrides", &db.IntegrationAuth{AuthMethod: db.IntegrationAuthToken, AuthHeader: "X-Token"})
	b := f.addIntegration(t, overriding, "b", nil)

	// The service answers by project key: the endpoint's key is nil here (no
	// AuthSecretID), so it hands out the secret of whichever auth asks by the
	// header the group uses. Two keys are told apart by storing them.
	endpointKey, err := f.store.CreateAccessKey(db.AccessKey{Name: "endpoint", Type: db.AccessKeyString, String: "endpoint-secret", ProjectID: &f.project.ID})
	require.NoError(t, err)
	groupKey, err := f.store.CreateAccessKey(db.AccessKey{Name: "group", Type: db.AccessKeyString, String: "group-secret", ProjectID: &f.project.ID})
	require.NoError(t, err)
	f.endpoint.AuthSecretID = &endpointKey.ID
	require.NoError(t, f.store.UpdateEndpoint(f.endpoint))
	overriding.AuthSecretID = &groupKey.ID
	require.NoError(t, f.store.UpdateIntegrationGroup(overriding))

	svc := &keyedIntegrationService{keys: map[int]db.AccessKey{endpointKey.ID: endpointKey, groupKey.ID: groupKey}}
	ctrl := NewIntegrationController(f.store, svc, nil, nil)

	_, ran := f.post(t, ctrl, f.alias, []byte(`{}`), map[string]string{"X-Token": "endpoint-secret"})
	assert.Equal(t, []int{a.ID}, ran)

	_, ran = f.post(t, ctrl, f.alias, []byte(`{}`), map[string]string{"X-Token": "group-secret"})
	assert.Equal(t, []int{b.ID}, ran)

	_, ran = f.post(t, ctrl, f.alias, []byte(`{}`), map[string]string{"X-Token": "neither"})
	assert.Empty(t, ran)
}

// keyedIntegrationService answers FillAuthSecret from stored keys by id, the
// way the real service does without the encryption.
type keyedIntegrationService struct {
	keys map[int]db.AccessKey
}

func (s *keyedIntegrationService) FillAuthSecret(projectID int, auth *db.IntegrationAuth) error {
	if auth.AuthSecretID == nil {
		return nil
	}
	key, ok := s.keys[*auth.AuthSecretID]
	if !ok {
		return db.ErrNotFound
	}
	auth.AuthSecret = key
	return nil
}

func TestReceiveIntegration_MatchersFilterWithinAGroup(t *testing.T) {
	f := newReceiveFixture(t, db.IntegrationAuthNone, "", "hooks")
	group := f.addGroup(t, "group", nil)
	always := f.addIntegration(t, group, "always", nil)
	onPush := f.addIntegration(t, group, "on push", map[string]string{"X-Event": "push"})
	ctrl := NewIntegrationController(f.store, &stubIntegrationService{}, nil, nil)

	_, ran := f.post(t, ctrl, f.alias, []byte(`{}`), map[string]string{"X-Event": "push"})
	assert.ElementsMatch(t, []int{always.ID, onPush.ID}, ran)

	_, ran = f.post(t, ctrl, f.alias, []byte(`{}`), map[string]string{"X-Event": "ping"})
	assert.Equal(t, []int{always.ID}, ran, "an integration without matchers runs on every authorized request")
}

func TestReceiveIntegration_EndpointAuthCheckedOncePerRequest(t *testing.T) {
	f := newReceiveFixture(t, db.IntegrationAuthToken, "X-Token", "hooks")
	for _, name := range []string{"one", "two", "three"} {
		group := f.addGroup(t, name, nil)
		f.addIntegration(t, group, name, nil)
	}
	own := f.addGroup(t, "own", &db.IntegrationAuth{AuthMethod: db.IntegrationAuthToken, AuthHeader: "X-Token"})
	f.addIntegration(t, own, "own", nil)
	svc := &stubIntegrationService{key: &db.AccessKey{Type: db.AccessKeyString, String: "s3cret"}}
	ctrl := NewIntegrationController(f.store, svc, nil, nil)

	_, ran := f.post(t, ctrl, f.alias, []byte(`{}`), map[string]string{"X-Token": "s3cret"})

	assert.Len(t, ran, 4)
	assert.Equal(t, 2, svc.calls, "the endpoint's auth once, plus once for the overriding group")
}

func TestReceiveIntegration_RunHeadersNameEndpointAndGroup(t *testing.T) {
	f := newReceiveFixture(t, db.IntegrationAuthNone, "", "hooks")
	f.store.failTemplate = false
	group := f.addGroup(t, "group", nil)
	integration := f.addIntegration(t, group, "hook", nil)

	// The run stops at the task pool, which the test has no context for;
	// headers are written by the workflow path instead.
	base := f.store.Store.(*sql.SqlDb)
	_, err := base.Sql().Exec(base.PrepareQuery(
		"insert into project__workflow_template (project_id, name) values (?, ?)"), f.project.ID, "wf")
	require.NoError(t, err)
	id64, err := base.Sql().SelectInt(base.PrepareQuery(
		"select id from project__workflow_template where project_id = ? and name = ?"), f.project.ID, "wf")
	require.NoError(t, err)
	workflowID := int(id64)
	integration.TemplateID = nil
	integration.WorkflowID = &workflowID
	require.NoError(t, f.store.UpdateIntegration(integration))

	svc := &stubWorkflowService{run: db.WorkflowRun{ID: 42, ProjectID: f.project.ID, WorkflowTemplateID: workflowID}}
	wfStore := &stubWorkflowStore{workflows: map[int]db.WorkflowTemplate{workflowID: {ID: workflowID, ProjectID: f.project.ID, Name: "wf"}}}
	ctrl := NewIntegrationController(f.store, &stubIntegrationService{}, wfStore, svc)

	w, ran := f.post(t, ctrl, f.alias, []byte(`{}`), nil)

	assert.Equal(t, []int{integration.ID}, ran)
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, []string{"42"}, w.Header().Values("X-Semaphore-Workflow-Run-ID"))
	assert.Equal(t, []string{itoa(f.endpoint.ID)}, w.Header().Values("X-Semaphore-Endpoint-ID"))
	assert.Equal(t, []string{itoa(group.ID)}, w.Header().Values("X-Semaphore-Integration-Group-ID"))
	assert.Equal(t, []string{itoa(integration.ID)}, w.Header().Values("X-Semaphore-Integration-ID"))
}
