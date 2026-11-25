package client

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/semaphoreui/semaphore/cli/cmd/client/config"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/apiclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mutation is one recorded write request.
type mutation struct {
	method string
	path   string
	body   map[string]any
}

// mutateServer answers reads for project 7 and records every write.
type mutateServer struct {
	mu   sync.Mutex
	muts []mutation
}

func (s *mutateServer) last() mutation {
	s.mu.Lock()
	defer s.mu.Unlock()
	require.NotEmpty(nil, s.muts)
	return s.muts[len(s.muts)-1]
}

func (s *mutateServer) handler(t *testing.T) http.Handler {
	str := func(v string) *string { return &v }
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		enc := func(v any) { _ = json.NewEncoder(w).Encode(v) }
		p := r.URL.Path

		if r.Method != http.MethodGet {
			raw, _ := io.ReadAll(r.Body)
			var body map[string]any
			_ = json.Unmarshal(raw, &body)
			s.mu.Lock()
			s.muts = append(s.muts, mutation{method: r.Method, path: p, body: body})
			s.mu.Unlock()
			switch {
			case r.Method == http.MethodPost && strings.HasSuffix(p, "/templates"):
				w.WriteHeader(http.StatusCreated)
				enc(db.Template{ID: 99, Name: body["name"].(string)})
			case r.Method == http.MethodPost && strings.HasSuffix(p, "/keys"):
				w.WriteHeader(http.StatusCreated)
				k := db.AccessKey{ID: 77, Name: body["name"].(string), Type: db.AccessKeyType(body["type"].(string))}
				enc(k)
			case r.Method == http.MethodPost && strings.HasSuffix(p, "/environment"):
				w.WriteHeader(http.StatusCreated)
				enc(db.Environment{ID: 55, Name: body["name"].(string)})
			case r.Method == http.MethodPost && p == "/api/projects":
				w.WriteHeader(http.StatusCreated)
				enc(db.Project{ID: 8, Name: body["name"].(string)})
			case r.Method == http.MethodPost && p == "/api/users":
				w.WriteHeader(http.StatusCreated)
				enc(db.User{ID: 5, Username: body["username"].(string)})
			case r.Method == http.MethodPost && (p == "/api/runners" || strings.HasSuffix(p, "/runners")):
				w.WriteHeader(http.StatusCreated)
				enc(apiclient.RunnerWithToken{Runner: db.Runner{ID: 6, Name: body["name"].(string)}, Token: "reg-token-once"})
			case r.Method == http.MethodPost && strings.HasSuffix(p, "/registration-token"):
				enc(apiclient.RegistrationToken{RegistrationToken: "reg-token-2", RunnerID: 1})
			case r.Method == http.MethodPost && strings.HasSuffix(p, "/schedules"):
				w.WriteHeader(http.StatusCreated)
				enc(db.Schedule{ID: 66, Name: "nightly"})
			case r.Method == http.MethodPost && strings.HasSuffix(p, "/views"):
				w.WriteHeader(http.StatusCreated)
				enc(db.View{ID: 44, Title: body["title"].(string)})
			case r.Method == http.MethodPost && strings.HasSuffix(p, "/inventory"):
				w.WriteHeader(http.StatusCreated)
				enc(db.Inventory{ID: 33, Name: body["name"].(string)})
			case r.Method == http.MethodPost && strings.HasSuffix(p, "/repositories"):
				w.WriteHeader(http.StatusCreated)
				enc(db.Repository{ID: 22, Name: body["name"].(string)})
			case r.Method == http.MethodPost && strings.HasSuffix(p, "/schedules/validate"):
				if body["cron_format"] == "bad" {
					w.WriteHeader(http.StatusBadRequest)
					_, _ = w.Write([]byte(`{"error":"cron: expected 5 fields"}`))
					return
				}
				w.WriteHeader(http.StatusNoContent)
			default:
				w.WriteHeader(http.StatusNoContent)
			}
			return
		}

		switch p {
		case "/api/projects":
			enc([]db.Project{{ID: 7, Name: "Homelab"}})
		case "/api/project/7":
			enc(db.Project{ID: 7, Name: "Homelab", MaxParallelTasks: 2})
		case "/api/project/7/templates":
			enc([]db.Template{{ID: 3, Name: "Deploy", Playbook: "site.yml", RepositoryID: 2}})
		case "/api/project/7/templates/3":
			enc(db.Template{ID: 3, Name: "Deploy", Playbook: "site.yml", RepositoryID: 2, Vaults: []db.TemplateVault{{ID: 5, Name: str("vault1")}}, SurveyVars: []db.SurveyVar{{Name: "v"}}})
		case "/api/project/7/inventory":
			enc([]db.Inventory{{ID: 4, Name: "prod"}})
		case "/api/project/7/inventory/4":
			enc(db.Inventory{ID: 4, Name: "prod"})
		case "/api/project/7/environment":
			enc([]db.Environment{{ID: 5, Name: "default"}})
		case "/api/project/7/environment/5":
			enc(db.Environment{ID: 5, Name: "default", JSON: "{}", Secrets: []db.EnvironmentSecret{{ID: 70, Type: db.EnvironmentSecretVar, Name: "OLD"}, {ID: 71, Type: db.EnvironmentSecretEnv, Name: "GONE"}}})
		case "/api/project/7/keys":
			enc([]db.AccessKey{{ID: 6, Name: "deploy", Type: db.AccessKeySSH}})
		case "/api/project/7/keys/6":
			enc(db.AccessKey{ID: 6, Name: "deploy", Type: db.AccessKeySSH})
		case "/api/project/7/repositories":
			enc([]db.Repository{{ID: 2, Name: "infra"}})
		case "/api/project/7/repositories/2":
			enc(db.Repository{ID: 2, Name: "infra", GitURL: "git@x", GitBranch: "main", SSHKeyID: 6})
		case "/api/project/7/schedules":
			enc([]db.ScheduleWithTpl{{Schedule: db.Schedule{ID: 9, Name: "nightly", TemplateID: 3, CronFormat: "0 3 * * *", Active: true}}})
		case "/api/project/7/schedules/9":
			enc(db.Schedule{ID: 9, Name: "nightly", TemplateID: 3, CronFormat: "0 3 * * *", Active: true})
		case "/api/project/7/views":
			enc([]db.View{{ID: 11, Title: "Ops"}})
		case "/api/project/7/views/11":
			enc(db.View{ID: 11, Title: "Ops", Position: 1})
		case "/api/project/7/users":
			enc([]apiclient.ProjectMember{{ID: 1, Username: "admin", Roles: []string{"owner"}}})
		case "/api/users":
			enc([]db.User{{ID: 1, Username: "admin", Email: "a@x"}, {ID: 2, Username: "bob", Name: "Bob", Email: "bob@x"}})
		case "/api/users/2":
			enc(db.User{ID: 2, Username: "bob", Name: "Bob", Email: "bob@x"})
		case "/api/runners":
			enc([]db.Runner{{ID: 1, Name: "builder", Tags: []string{"a"}}})
		case "/api/runners/1":
			enc(db.Runner{ID: 1, Name: "builder", Tags: []string{"a"}, MaxParallelTasks: 2})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
}

func mutateEnv(srv *httptest.Server) map[string]string {
	return map[string]string{config.EnvURL: srv.URL, config.EnvToken: "tok", config.EnvProject: "7"}
}

func TestTemplateCreate(t *testing.T) {
	s := &mutateServer{}
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()
	env := mutateEnv(srv)

	out, errOut, err := run(t, env, "", "template", "create", "--name", "deploy", "--app", "ansible", "--playbook", "site.yml",
		"--repository", "infra", "--inventory", "prod", "--environment", "default", "--view", "Ops", "--description", "d", "-q")
	require.NoError(t, err)
	assert.Equal(t, "99\n", out)
	assert.Contains(t, errOut, `Template "deploy" created with id 99.`)
	m := s.last()
	assert.Equal(t, "POST /api/project/7/templates", m.method+" "+m.path)
	assert.Equal(t, "deploy", m.body["name"])
	assert.Equal(t, "ansible", m.body["app"])
	assert.EqualValues(t, 2, m.body["repository_id"], "repository resolved by name")
	assert.EqualValues(t, 4, m.body["inventory_id"])
	assert.EqualValues(t, 5, m.body["environment_id"])
	assert.Equal(t, []any{float64(5)}, m.body["environment_ids"])
	assert.EqualValues(t, 11, m.body["view_id"])
	assert.EqualValues(t, 7, m.body["project_id"])

	// From a YAML file, flags override.
	f := filepath.Join(t.TempDir(), "tpl.yml")
	require.NoError(t, os.WriteFile(f, []byte("name: from-file\napp: bash\nplaybook: run.sh\nrepository_id: 2\n"), 0o600))
	_, _, err = run(t, env, "", "template", "create", "--from-file", f, "--name", "overridden")
	require.NoError(t, err)
	m = s.last()
	assert.Equal(t, "overridden", m.body["name"])
	assert.Equal(t, "bash", m.body["app"])
	assert.Equal(t, "run.sh", m.body["playbook"])

	// Required fields.
	_, _, err = run(t, env, "", "template", "create", "--name", "x")
	var u *UsageError
	require.ErrorAs(t, err, &u)
	assert.Contains(t, err.Error(), "--playbook")
	assert.Contains(t, err.Error(), "--repository")

	// Bad file.
	require.NoError(t, os.WriteFile(f, []byte("- a list\n"), 0o600))
	_, _, err = run(t, env, "", "template", "create", "--from-file", f)
	require.ErrorAs(t, err, &u)
	assert.Contains(t, err.Error(), "one object")
	_, _, err = run(t, env, "", "template", "create", "--from-file", filepath.Join(t.TempDir(), "nope.yml"))
	assert.ErrorContains(t, err, "read")
}

func TestTemplateUpdate_KeepsVaults(t *testing.T) {
	s := &mutateServer{}
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()

	_, errOut, err := run(t, mutateEnv(srv), "", "template", "update", "Deploy", "--description", "new text", "--inventory", "4")
	require.NoError(t, err)
	assert.Contains(t, errOut, `Template "Deploy" (id 3) updated.`)
	m := s.last()
	assert.Equal(t, "PUT /api/project/7/templates/3", m.method+" "+m.path)
	assert.Equal(t, "new text", m.body["description"])
	assert.Equal(t, "site.yml", m.body["playbook"], "untouched fields are sent back")
	assert.Len(t, m.body["vaults"], 1, "vaults read by id travel with the update")
	assert.Len(t, m.body["survey_vars"], 1)
	assert.EqualValues(t, 4, m.body["inventory_id"])
}

func TestDelete_ConfirmationRules(t *testing.T) {
	s := &mutateServer{}
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()
	env := mutateEnv(srv)

	_, _, err := run(t, env, "", "template", "delete", "Deploy")
	assert.Equal(t, ExitUsage, ExitCode(err), "no TTY, no --yes")
	assert.Empty(t, s.muts)

	_, _, err = runWith(t, env, "n\n", true, "template", "delete", "Deploy")
	assert.ErrorIs(t, err, ErrAborted)
	assert.Empty(t, s.muts)

	_, errOut, err := run(t, env, "", "template", "rm", "3", "-y")
	require.NoError(t, err)
	assert.Equal(t, "DELETE /api/project/7/templates/3", s.last().method+" "+s.last().path)
	assert.Contains(t, errOut, `Template "Deploy" (id 3) deleted.`)

	_, _, err = runWith(t, env, "y\n", true, "key", "delete", "deploy")
	require.NoError(t, err)
	assert.Equal(t, "DELETE /api/project/7/keys/6", s.last().method+" "+s.last().path)

	_, _, err = run(t, env, "", "template", "delete", "missing", "-y")
	assert.Equal(t, ExitNotFound, ExitCode(err))
}

func TestKeyCreateAndUpdate_Secrets(t *testing.T) {
	s := &mutateServer{}
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()
	env := mutateEnv(srv)

	out, _, err := run(t, env, "hunter2\n", "key", "create", "--name", "vault", "--type", "login_password", "--login", "admin", "--password-stdin", "-o", "json")
	require.NoError(t, err)
	m := s.last()
	assert.Equal(t, "POST /api/project/7/keys", m.method+" "+m.path)
	lp := m.body["login_password"].(map[string]any)
	assert.Equal(t, "admin", lp["login"])
	assert.Equal(t, "hunter2", lp["password"])
	assert.Equal(t, true, m.body["override_secret"])
	assert.NotContains(t, out, "hunter2", "the created key is printed without its secret")

	keyFile := filepath.Join(t.TempDir(), "id_ed25519")
	require.NoError(t, os.WriteFile(keyFile, []byte("-----BEGIN KEY-----\nabc\n-----END KEY-----\n"), 0o600))
	_, _, err = run(t, env, "", "key", "create", "--name", "deploy2", "--type", "ssh", "--login", "git", "--ssh-key-file", keyFile)
	require.NoError(t, err)
	ssh := s.last().body["ssh"].(map[string]any)
	assert.Equal(t, "-----BEGIN KEY-----\nabc\n-----END KEY-----", ssh["private_key"], "internal newlines kept, trailing one trimmed")
	assert.Equal(t, "git", ssh["login"])

	// Update without a secret keeps the stored one.
	_, _, err = run(t, env, "", "key", "update", "deploy", "--name", "deploy (old)")
	require.NoError(t, err)
	m = s.last()
	assert.Equal(t, "PUT /api/project/7/keys/6", m.method+" "+m.path)
	assert.Nil(t, m.body["override_secret"])
	assert.Equal(t, "deploy (old)", m.body["name"])

	_, _, err = run(t, env, "", "key", "update", "deploy", "--ssh-key-file", keyFile)
	require.NoError(t, err)
	assert.Equal(t, true, s.last().body["override_secret"])

	// Mistakes.
	_, _, err = run(t, env, "", "key", "create", "--name", "x", "--type", "login_password", "--password-stdin", "--password-file", keyFile)
	assert.Equal(t, ExitUsage, ExitCode(err))
	_, _, err = run(t, env, "", "key", "create", "--name", "x", "--type", "login_password", "--password-stdin")
	assert.Equal(t, ExitUsage, ExitCode(err), "empty stdin")
	_, _, err = run(t, env, "", "key", "create", "--name", "x")
	assert.Equal(t, ExitUsage, ExitCode(err), "type required")
}

func TestEnvironmentCreate_Secrets(t *testing.T) {
	s := &mutateServer{}
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()
	env := mutateEnv(srv)

	secrets := filepath.Join(t.TempDir(), "secrets.yml")
	require.NoError(t, os.WriteFile(secrets, []byte("vars:\n  DB_PASS: s3cret\nenv:\n  TOKEN: t0k\n"), 0o600))
	vars := filepath.Join(t.TempDir(), "vars.json")
	require.NoError(t, os.WriteFile(vars, []byte(`{"region":"eu"}`), 0o600))

	out, errOut, err := run(t, env, "", "environment", "create", "--name", "prod", "--json", vars, "--env", `{"A":"1"}`, "--secrets-file", secrets, "--secret", "API=k")
	require.NoError(t, err)
	m := s.last()
	assert.Equal(t, "POST /api/project/7/environment", m.method+" "+m.path)
	assert.JSONEq(t, `{"region":"eu"}`, m.body["json"].(string), "file content")
	assert.Equal(t, `{"A":"1"}`, m.body["env"], "inline JSON")
	secs := m.body["secrets"].([]any)
	require.Len(t, secs, 3)
	names := map[string]string{}
	for _, sec := range secs {
		x := sec.(map[string]any)
		names[x["name"].(string)] = x["type"].(string) + ":" + x["secret"].(string) + ":" + x["operation"].(string)
	}
	assert.Equal(t, "var:s3cret:create", names["DB_PASS"])
	assert.Equal(t, "env:t0k:create", names["TOKEN"])
	assert.Equal(t, "var:k:create", names["API"])
	assert.Contains(t, errOut, "visible to other users", "a --secret on the command line warns")
	assert.NotContains(t, out, "s3cret")

	_, _, err = run(t, env, "", "environment", "create", "--name", "e")
	require.NoError(t, err)
	assert.Equal(t, "{}", s.last().body["json"], "json defaults to an empty object")

	// Update sends only secret operations: existing secrets are not re-sent,
	// a new one is created, --delete-secret deletes by the stored id.
	_, _, err = run(t, env, "", "environment", "update", "default", "--secret", "NEW=v", "--delete-secret", "GONE")
	require.NoError(t, err)
	m = s.last()
	assert.Equal(t, "PUT /api/project/7/environment/5", m.method+" "+m.path)
	ops := m.body["secrets"].([]any)
	require.Len(t, ops, 2)
	del := ops[0].(map[string]any)
	assert.Equal(t, "delete", del["operation"])
	assert.EqualValues(t, 71, del["id"])
	assert.Equal(t, "GONE", del["name"])
	assert.Equal(t, "create", ops[1].(map[string]any)["operation"])

	_, _, err = run(t, env, "", "environment", "update", "default", "--name", "renamed")
	require.NoError(t, err)
	assert.Nil(t, s.last().body["secrets"], "no secret flags → no secret operations")

	_, _, err = run(t, env, "", "environment", "update", "default", "--delete-secret", "NOPE")
	assert.Equal(t, ExitNotFound, ExitCode(err))
}

func TestProjectCreateUpdateDelete(t *testing.T) {
	s := &mutateServer{}
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()
	env := map[string]string{config.EnvURL: srv.URL, config.EnvToken: "tok"}

	out, _, err := run(t, env, "", "project", "create", "--name", "Playground", "--demo", "-q")
	require.NoError(t, err)
	assert.Equal(t, "8\n", out)
	m := s.last()
	assert.Equal(t, "POST /api/projects", m.method+" "+m.path)
	assert.Equal(t, true, m.body["demo"])

	_, _, err = run(t, env, "", "project", "update", "Homelab", "--alert", "--max-parallel-tasks", "4")
	require.NoError(t, err)
	m = s.last()
	assert.Equal(t, "PUT /api/project/7", m.method+" "+m.path)
	assert.Equal(t, "Homelab", m.body["name"])
	assert.Equal(t, true, m.body["alert"])
	assert.EqualValues(t, 4, m.body["max_parallel_tasks"])

	_, _, err = run(t, env, "", "project", "delete", "Homelab", "-y")
	assert.Equal(t, ExitUsage, ExitCode(err), "--yes is not enough for a project")
	_, _, err = run(t, env, "", "project", "delete", "Homelab", "--confirm=homelab")
	assert.Equal(t, ExitUsage, ExitCode(err), "name must match exactly")
	_, _, err = run(t, env, "", "project", "delete", "7", "--confirm=Homelab")
	require.NoError(t, err)
	assert.Equal(t, "DELETE /api/project/7", s.last().method+" "+s.last().path)

	_, _, err = runWith(t, env, "Homelab\n", true, "project", "delete", "Homelab")
	require.NoError(t, err)
}

func TestMembers(t *testing.T) {
	s := &mutateServer{}
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()
	env := mutateEnv(srv)

	_, errOut, err := run(t, env, "", "project", "member", "add", "bob@x", "--role", "manager")
	require.NoError(t, err)
	m := s.last()
	assert.Equal(t, "POST /api/project/7/users", m.method+" "+m.path)
	assert.EqualValues(t, 2, m.body["user_id"], "user resolved by e-mail")
	assert.Equal(t, "manager", m.body["role"])
	assert.Contains(t, errOut, `User "bob" added to project 7 as manager.`)

	_, _, err = run(t, env, "", "project", "member", "update", "admin", "--role", "guest")
	require.NoError(t, err)
	assert.Equal(t, "PUT /api/project/7/users/1", s.last().method+" "+s.last().path)

	_, _, err = run(t, env, "", "project", "member", "remove", "admin")
	assert.Equal(t, ExitUsage, ExitCode(err))
	_, _, err = run(t, env, "", "project", "member", "remove", "admin", "-y")
	require.NoError(t, err)
	assert.Equal(t, "DELETE /api/project/7/users/1", s.last().method+" "+s.last().path)

	_, _, err = run(t, env, "", "project", "member", "add", "bob")
	assert.Equal(t, ExitUsage, ExitCode(err), "--role required")
}

func TestUsers(t *testing.T) {
	s := &mutateServer{}
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()
	env := map[string]string{config.EnvURL: srv.URL, config.EnvToken: "tok"}

	out, _, err := run(t, env, "pw123\n", "user", "create", "--login", "carol", "--name", "Carol", "--email", "c@x", "--admin", "--password-stdin", "-q")
	require.NoError(t, err)
	assert.Equal(t, "5\n", out)
	m := s.last()
	assert.Equal(t, "POST /api/users", m.method+" "+m.path)
	assert.Equal(t, "pw123", m.body["password"])
	assert.Equal(t, true, m.body["admin"])

	_, _, err = run(t, env, "", "user", "create", "--login", "dave", "--name", "Dave", "--email", "d@x")
	assert.Equal(t, ExitUsage, ExitCode(err), "a local user needs a password")

	_, _, err = run(t, env, "", "user", "create", "--login", "erin", "--name", "Erin", "--email", "e@x", "--external")
	require.NoError(t, err)
	assert.Equal(t, true, s.last().body["external"])

	_, _, err = run(t, env, "", "user", "update", "bob", "--admin=false", "--email", "new@x")
	require.NoError(t, err)
	m = s.last()
	assert.Equal(t, "PUT /api/users/2", m.method+" "+m.path)
	assert.Equal(t, "new@x", m.body["email"])
	assert.Equal(t, "Bob", m.body["name"], "unchanged fields are preserved")

	_, _, err = run(t, env, "newpw\n", "user", "set-password", "bob", "--password-stdin")
	require.NoError(t, err)
	m = s.last()
	assert.Equal(t, "POST /api/users/2/password", m.method+" "+m.path)
	assert.Equal(t, "newpw", m.body["password"])

	_, _, err = run(t, env, "", "user", "set-password", "bob")
	assert.Equal(t, ExitUsage, ExitCode(err))

	_, _, err = run(t, env, "", "user", "delete", "bob", "-y")
	require.NoError(t, err)
	assert.Equal(t, "DELETE /api/users/2", s.last().method+" "+s.last().path)
}

func TestRunners(t *testing.T) {
	s := &mutateServer{}
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()
	env := map[string]string{config.EnvURL: srv.URL, config.EnvToken: "tok"}

	out, errOut, err := run(t, env, "", "runner", "create", "--name", "builder-2", "--tags", "docker,linux")
	require.NoError(t, err)
	assert.Equal(t, "reg-token-once\n", out, "the token alone on stdout")
	assert.Contains(t, errOut, "shown once")
	m := s.last()
	assert.Equal(t, "POST /api/runners", m.method+" "+m.path)
	assert.Equal(t, []any{"docker", "linux"}, m.body["tags"])

	out, _, err = run(t, env, "", "runner", "create", "--name", "p", "-p", "7", "-o", "json")
	require.NoError(t, err)
	assert.Equal(t, "POST /api/project/7/runners", s.last().method+" "+s.last().path)
	assert.Contains(t, out, `"registration_token": "reg-token-once"`)

	_, _, err = run(t, env, "", "runner", "update", "builder", "--max-parallel-tasks", "4")
	require.NoError(t, err)
	m = s.last()
	assert.Equal(t, "PUT /api/runners/1", m.method+" "+m.path)
	assert.EqualValues(t, 4, m.body["max_parallel_tasks"])
	assert.Equal(t, []any{"a"}, m.body["tags"], "preserved")

	_, _, err = run(t, env, "", "runner", "disable", "builder")
	require.NoError(t, err)
	m = s.last()
	assert.Equal(t, "POST /api/runners/1/active", m.method+" "+m.path)
	assert.Equal(t, false, m.body["active"])

	out, _, err = run(t, env, "", "runner", "registration-token", "1")
	require.NoError(t, err)
	assert.Equal(t, "reg-token-2\n", out)

	_, _, err = run(t, env, "", "runner", "clear-cache", "builder")
	require.NoError(t, err)
	assert.Equal(t, "DELETE /api/runners/1/cache", s.last().method+" "+s.last().path)

	_, _, err = run(t, env, "", "runner", "delete", "1", "-y")
	require.NoError(t, err)
	assert.Equal(t, "DELETE /api/runners/1", s.last().method+" "+s.last().path)
}

func TestSchedulesViewsRepositoriesInventories(t *testing.T) {
	s := &mutateServer{}
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()
	env := mutateEnv(srv)

	_, _, err := run(t, env, "", "schedule", "create", "--template", "Deploy", "--cron", "0 3 * * *", "--name", "nightly")
	require.NoError(t, err)
	m := s.last()
	assert.Equal(t, "POST /api/project/7/schedules", m.method+" "+m.path)
	assert.EqualValues(t, 3, m.body["template_id"])
	assert.Equal(t, true, m.body["active"], "active by default")

	_, _, err = run(t, env, "", "schedule", "create", "--template", "3", "--cron", "@daily", "--inactive")
	require.NoError(t, err)
	assert.Equal(t, false, s.last().body["active"])

	_, _, err = run(t, env, "", "schedule", "update", "nightly", "--cron", "0 4 * * *")
	require.NoError(t, err)
	m = s.last()
	assert.Equal(t, "PUT /api/project/7/schedules/9", m.method+" "+m.path)
	assert.Equal(t, "0 4 * * *", m.body["cron_format"])
	assert.EqualValues(t, 3, m.body["template_id"], "preserved")

	_, _, err = run(t, env, "", "schedule", "disable", "nightly")
	require.NoError(t, err)
	m = s.last()
	assert.Equal(t, "PUT /api/project/7/schedules/9/active", m.method+" "+m.path)
	assert.Equal(t, false, m.body["active"])

	_, _, err = run(t, env, "", "schedule", "validate", "--cron", "bad")
	assert.ErrorContains(t, err, "expected 5 fields")
	_, _, err = run(t, env, "", "schedule", "validate", "--cron", "@hourly")
	require.NoError(t, err)

	_, _, err = run(t, env, "", "view", "create", "--title", "Infra", "--position", "3")
	require.NoError(t, err)
	assert.EqualValues(t, 3, s.last().body["position"])

	_, _, err = run(t, env, "", "repository", "create", "--name", "app", "--url", "git@x:app.git", "--branch", "main", "--ssh-key", "deploy")
	require.NoError(t, err)
	m = s.last()
	assert.EqualValues(t, 6, m.body["ssh_key_id"], "key resolved by name")
	_, _, err = run(t, env, "", "repository", "create", "--name", "app", "--url", "git@x:app.git")
	assert.Equal(t, ExitUsage, ExitCode(err))

	hosts := filepath.Join(t.TempDir(), "hosts.ini")
	require.NoError(t, os.WriteFile(hosts, []byte("[web]\nweb1\n"), 0o600))
	_, _, err = run(t, env, "", "inventory", "create", "--name", "web", "--type", "static", "--inventory-file", hosts, "--ssh-key", "6", "--become-key", "deploy")
	require.NoError(t, err)
	m = s.last()
	assert.Equal(t, "[web]\nweb1\n", m.body["inventory"])
	assert.EqualValues(t, 6, m.body["ssh_key_id"])
	assert.EqualValues(t, 6, m.body["become_key_id"])
	assert.Equal(t, "static", m.body["type"])

	_, _, err = run(t, env, "", "template", "stop-all", "Deploy", "-y")
	require.NoError(t, err)
	assert.Equal(t, "POST /api/project/7/templates/3/stop_all_tasks", s.last().method+" "+s.last().path)
}

func TestProjectBackupRestoreCommands(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method + " " + r.URL.Path {
		case "GET /api/projects":
			_ = json.NewEncoder(w).Encode([]db.Project{{ID: 7, Name: "Homelab"}})
		case "GET /api/project/7":
			_ = json.NewEncoder(w).Encode(db.Project{ID: 7, Name: "Homelab"})
		case "GET /api/project/7/backup":
			_, _ = w.Write([]byte(`{"meta":{"name":"Homelab"}}`))
		case "POST /api/projects/restore":
			raw, _ := io.ReadAll(r.Body)
			assert.JSONEq(t, `{"meta":{"name":"Homelab"}}`, string(raw))
			_ = json.NewEncoder(w).Encode(db.Project{ID: 9, Name: "Homelab"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	env := map[string]string{config.EnvURL: srv.URL, config.EnvToken: "tok"}

	out, _, err := run(t, env, "", "project", "backup", "Homelab")
	require.NoError(t, err)
	assert.Equal(t, `{"meta":{"name":"Homelab"}}`, out)

	file := filepath.Join(t.TempDir(), "b.json")
	_, _, err = run(t, env, "", "project", "backup", "7", "--file", file)
	require.NoError(t, err)
	info, err := os.Stat(file)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	out, _, err = run(t, env, "", "project", "restore", "--file", file, "-q")
	require.NoError(t, err)
	assert.Equal(t, "9\n", out)

	_, _, err = run(t, env, "", "project", "restore")
	assert.Equal(t, ExitUsage, ExitCode(err))
}
