package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/semaphoreui/semaphore/cli/cmd/client/config"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/apiclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// resourceServer serves canned lists and single objects for project 7 ("Homelab")
// and records every path it was asked for.
type resourceServer struct {
	mu    sync.Mutex
	paths []string
}

func (s *resourceServer) handler(t *testing.T) http.Handler {
	str := func(v string) *string { return &v }
	templates := []db.Template{
		{ID: 1, ProjectID: 7, Name: "Deploy", App: "ansible", Playbook: "site.yml", RepositoryID: 2},
		{ID: 2, ProjectID: 7, Name: "deploy", App: "ansible", Playbook: "other.yml", RepositoryID: 2},
		{ID: 3, ProjectID: 7, Name: "Backup", App: "bash", Playbook: "backup.sh", RepositoryID: 2},
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		s.paths = append(s.paths, r.URL.Path)
		s.mu.Unlock()
		if r.Header.Get("Authorization") != "Bearer tok" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		enc := func(v any) { _ = json.NewEncoder(w).Encode(v) }
		switch r.URL.Path {
		case "/api/projects":
			enc([]db.Project{{ID: 7, Name: "Homelab", Type: ""}, {ID: 8, Name: "Other"}, {ID: 9, Name: "other"}})
		case "/api/project/7":
			enc(db.Project{ID: 7, Name: "Homelab"})
		case "/api/project/7/templates":
			enc(templates)
		case "/api/project/7/templates/3":
			t := templates[2]
			t.Vaults = []db.TemplateVault{{ID: 5, Name: str("vault1")}}
			enc(t)
		case "/api/project/7/templates/3/tasks/last":
			assert.Equal(t, "50", r.URL.Query().Get("count"))
			enc([]db.TaskWithTpl{{Task: db.Task{ID: 101, Status: "success"}, TemplateAlias: "Backup", UserName: str("admin")}})
		case "/api/project/7/inventory":
			enc([]db.Inventory{{ID: 4, Name: "Prod hosts", Type: "static"}})
		case "/api/project/7/environment":
			enc([]db.Environment{{ID: 5, Name: "Vars", JSON: "{}", Password: str("pw"), Secrets: []db.EnvironmentSecret{{Name: "DB_PASS", Secret: "s3cret"}, {Name: "API_KEY", Secret: "k3y"}}}})
		case "/api/project/7/environment/5":
			enc(db.Environment{ID: 5, Name: "Vars", JSON: "{}", Password: str("pw"), Secrets: []db.EnvironmentSecret{{Name: "DB_PASS", Secret: "s3cret"}}})
		case "/api/project/7/keys":
			k := db.AccessKey{ID: 6, Name: "deploy key", Type: db.AccessKeyLoginPassword}
			k.LoginPassword.Login = "root"
			k.LoginPassword.Password = "hunter2"
			enc([]db.AccessKey{k})
		case "/api/project/7/repositories":
			enc([]db.Repository{{ID: 2, Name: "infra", GitURL: "git@x:y.git", GitBranch: "main", SSHKeyID: 6}})
		case "/api/project/7/repositories/2":
			enc(db.Repository{ID: 2, Name: "infra", GitURL: "git@x:y.git", GitBranch: "main", SSHKeyID: 6})
		case "/api/project/7/schedules":
			enc([]db.ScheduleWithTpl{{Schedule: db.Schedule{ID: 9, Name: "nightly", TemplateID: 3, CronFormat: "0 3 * * *", Active: true}, TemplateName: "Backup"}})
		case "/api/project/7/views":
			enc([]db.View{{ID: 11, Title: "Ops", Position: 1}})
		case "/api/project/7/views/11":
			enc(db.View{ID: 11, Title: "Ops", Position: 1})
		case "/api/project/7/views/11/templates":
			enc(templates[:1])
		case "/api/project/7/users":
			enc([]apiclient.ProjectMember{{ID: 1, Username: "admin", Name: "Admin", Role: "owner", Roles: []string{"owner"}}})
		case "/api/users":
			enc([]db.User{{ID: 1, Username: "admin", Name: "Admin", Email: "a@x", Admin: true}, {ID: 2, Username: "bob", Email: "bob@x"}})
		case "/api/users/2":
			enc(db.User{ID: 2, Username: "bob", Email: "bob@x"})
		case "/api/runners":
			enc([]db.Runner{{ID: 1, Name: "global-1", Tags: []string{"a", "b"}, Active: true}})
		case "/api/project/7/runners":
			enc([]db.Runner{{ID: 2, Name: "proj-1", Active: false}})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
}

func resourceEnv(t *testing.T, srv *httptest.Server, project string) map[string]string {
	t.Helper()
	env := map[string]string{config.EnvURL: srv.URL, config.EnvToken: "tok"}
	if project != "" {
		env[config.EnvProject] = project
	}
	return env
}

func TestResourceList_TableJSONQuietFields(t *testing.T) {
	s := &resourceServer{}
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()
	env := resourceEnv(t, srv, "7")

	out, _, err := run(t, env, "", "template", "list")
	require.NoError(t, err)
	assert.Equal(t, "ID\tNAME\tAPP\tPLAYBOOK\tINVENTORY ID\tREPOSITORY ID\tENVIRONMENT ID\n"+
		"1\tDeploy\tansible\tsite.yml\t\t2\t0\n2\tdeploy\tansible\tother.yml\t\t2\t0\n3\tBackup\tbash\tbackup.sh\t\t2\t0\n", out)

	out, _, err = run(t, env, "", "templates", "ls", "-q")
	require.NoError(t, err)
	assert.Equal(t, "1\n2\n3\n", out)

	out, _, err = run(t, env, "", "tpl", "list", "--fields", "id,playbook", "--no-headers")
	require.NoError(t, err)
	assert.Equal(t, "1\tsite.yml\n2\tother.yml\n3\tbackup.sh\n", out)

	out, _, err = run(t, env, "", "template", "list", "-o", "json")
	require.NoError(t, err)
	var rows []map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &rows))
	assert.Len(t, rows, 3)
	assert.Equal(t, "Deploy", rows[0]["name"])
}

func TestResourceGet_ByIDAndName(t *testing.T) {
	s := &resourceServer{}
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()
	env := resourceEnv(t, srv, "7")

	out, _, err := run(t, env, "", "template", "get", "3", "-o", "json")
	require.NoError(t, err)
	var tpl map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &tpl))
	assert.Equal(t, "Backup", tpl["name"])
	assert.Len(t, tpl["vaults"], 1, "get re-reads by id so vaults are present")
	assert.NotContains(t, s.paths, "/api/project/7/templates", "an id needs no list call")

	out, _, err = run(t, env, "", "template", "show", "backup", "--fields", "id,name")
	require.NoError(t, err)
	assert.Equal(t, "ID\t3\nNAME\tBackup\n", out, "name match is case-insensitive")

	_, _, err = run(t, env, "", "template", "get", "deploy")
	var amb *AmbiguousError
	require.ErrorAs(t, err, &amb)
	assert.Equal(t, []string{"1 (Deploy)", "2 (deploy)"}, amb.Candidates)
	assert.Equal(t, ExitNotFound, ExitCode(err))

	_, _, err = run(t, env, "", "template", "get", "nothing")
	assert.ErrorIs(t, err, apiclient.ErrNotFound)
	assert.ErrorContains(t, err, `template "nothing" not found`)

	_, _, err = run(t, env, "", "template", "get", "99")
	assert.ErrorIs(t, err, apiclient.ErrNotFound, "a 404 from the server")

	_, _, err = run(t, env, "", "template", "get")
	assert.Error(t, err, "argument required")
}

func TestProjectResolution(t *testing.T) {
	s := &resourceServer{}
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()

	// By name through the env.
	out, _, err := run(t, resourceEnv(t, srv, "homelab"), "", "inventory", "list")
	require.NoError(t, err)
	assert.Equal(t, "ID\tNAME\tTYPE\tSSH KEY ID\tREPOSITORY ID\n4\tProd hosts\tstatic\t\t\n", out)

	// By flag, which beats the env.
	out, _, err = run(t, resourceEnv(t, srv, "nope"), "", "inventory", "list", "-p", "7", "-q")
	require.NoError(t, err)
	assert.Equal(t, "4\n", out)

	// Missing.
	_, _, err = run(t, resourceEnv(t, srv, ""), "", "inventory", "list")
	var u *UsageError
	require.ErrorAs(t, err, &u)
	assert.Contains(t, err.Error(), "--project")

	// Unknown and ambiguous names.
	_, _, err = run(t, resourceEnv(t, srv, "missing"), "", "inventory", "list")
	assert.ErrorIs(t, err, apiclient.ErrNotFound)
	_, _, err = run(t, resourceEnv(t, srv, "other"), "", "inventory", "list")
	var amb *AmbiguousError
	require.ErrorAs(t, err, &amb)
	assert.Equal(t, "project", amb.Kind)

	// project get: default from the configured project, or an argument.
	out, _, err = run(t, resourceEnv(t, srv, "Homelab"), "", "project", "get", "--fields", "id")
	require.NoError(t, err)
	assert.Equal(t, "ID\t7\n", out)
	out, _, err = run(t, resourceEnv(t, srv, ""), "", "project", "get", "7", "-q")
	require.NoError(t, err)
	assert.Equal(t, "7\n", out)
	_, _, err = run(t, resourceEnv(t, srv, ""), "", "project", "get")
	require.ErrorAs(t, err, &u)

	out, _, err = run(t, resourceEnv(t, srv, ""), "", "project", "list", "--fields", "name")
	require.NoError(t, err)
	assert.Equal(t, "NAME\nHomelab\nOther\nother\n", out)
}

func TestSecretsAreHidden(t *testing.T) {
	s := &resourceServer{}
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()
	env := resourceEnv(t, srv, "7")

	out, _, err := run(t, env, "", "environment", "list")
	require.NoError(t, err)
	assert.Equal(t, "ID\tNAME\tSECRET NAMES\n5\tVars\tDB_PASS,API_KEY\n", out)
	assert.NotContains(t, out, "s3cret")

	out, _, err = run(t, env, "", "env", "get", "Vars", "-o", "json")
	require.NoError(t, err)
	assert.NotContains(t, out, "s3cret")
	assert.NotContains(t, out, `"pw"`)
	assert.Contains(t, out, `"has_password": true`)
	assert.Contains(t, out, `"secret_names"`)

	out, _, err = run(t, env, "", "key", "list", "-o", "json")
	require.NoError(t, err)
	assert.NotContains(t, out, "hunter2")
	assert.Contains(t, out, `"login": "root"`)
	assert.Contains(t, out, `"type": "login_password"`)
}

func TestOtherResources(t *testing.T) {
	s := &resourceServer{}
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()
	env := resourceEnv(t, srv, "7")

	tests := []struct {
		args []string
		want string
	}{
		{[]string{"repository", "list"}, "ID\tNAME\tGIT URL\tGIT BRANCH\tSSH KEY ID\n2\tinfra\tgit@x:y.git\tmain\t6\n"},
		{[]string{"repo", "get", "infra", "--fields", "id"}, "ID\t2\n"},
		{[]string{"schedule", "list"}, "ID\tNAME\tTPL NAME\tCRON FORMAT\tACTIVE\n9\tnightly\tBackup\t0 3 * * *\ttrue\n"},
		{[]string{"view", "list"}, "ID\tTITLE\tPOSITION\n11\tOps\t1\n"},
		{[]string{"view", "templates", "Ops", "-q"}, "1\n"},
		{[]string{"template", "tasks", "Backup"}, "ID\tSTATUS\tTPL ALIAS\tUSER NAME\tCREATED\tSTART\tEND\n101\tsuccess\tBackup\tadmin\t0001-01-01T00:00:00Z\t\t\n"},
		{[]string{"project", "member", "list"}, "ID\tUSERNAME\tNAME\tROLES\n1\tadmin\tAdmin\towner\n"},
		{[]string{"project", "members", "get", "ADMIN", "--fields", "id,roles"}, "ID\t1\nROLES\towner\n"},
		{[]string{"user", "list", "--fields", "username,admin"}, "USERNAME\tADMIN\nadmin\ttrue\nbob\tfalse\n"},
		{[]string{"user", "get", "bob@x", "--fields", "id,username"}, "ID\t2\nUSERNAME\tbob\n"},
		{[]string{"user", "get", "BOB", "-q"}, "2\n"},
		{[]string{"runner", "list"}, "ID\tNAME\tTAGS\tACTIVE\tSTATUS\tTOUCHED\n1\tglobal-1\ta,b\ttrue\t\t\n"},
		{[]string{"runner", "list", "-p", "7", "-q"}, "2\n"},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			out, _, err := run(t, env, "", tt.args...)
			require.NoError(t, err)
			assert.Equal(t, tt.want, out)
		})
	}

	// Runners stay global even when the context names a default project.
	out, _, err := run(t, env, "", "runner", "list", "-q")
	require.NoError(t, err)
	assert.Equal(t, "1\n", out)
}

func TestResource_AuthAndProjectScope(t *testing.T) {
	s := &resourceServer{}
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()

	_, _, err := run(t, map[string]string{config.EnvURL: srv.URL, config.EnvToken: "bad", config.EnvProject: "7"}, "", "template", "list")
	assert.Equal(t, ExitAuth, ExitCode(err))

	_, _, err = run(t, map[string]string{config.EnvURL: srv.URL, config.EnvProject: "7"}, "", "template", "list")
	assert.Equal(t, ExitUsage, ExitCode(err), "no token")

	_, _, err = run(t, map[string]string{config.EnvURL: srv.URL, config.EnvToken: "tok"}, "", "user", "list")
	require.NoError(t, err, "global resources need no project")
}

func TestPickByIDOrName(t *testing.T) {
	items := []db.Project{{ID: 1, Name: "A"}, {ID: 2, Name: "a"}, {ID: 3, Name: "B"}}
	id := func(p db.Project) int { return p.ID }
	match := func(p db.Project, s string) bool { return strings.EqualFold(p.Name, s) }
	label := func(p db.Project) string { return idName(p.ID, p.Name) }

	p, err := pickByIDOrName("project", "3", items, id, match, label)
	require.NoError(t, err)
	assert.Equal(t, 3, p.ID)

	p, err = pickByIDOrName("project", "b", items, id, match, label)
	require.NoError(t, err)
	assert.Equal(t, 3, p.ID)

	_, err = pickByIDOrName("project", "a", items, id, match, label)
	var amb *AmbiguousError
	require.ErrorAs(t, err, &amb)

	_, err = pickByIDOrName("project", "4", items, id, match, label)
	assert.ErrorIs(t, err, apiclient.ErrNotFound)

	_, err = pickByIDOrName("project", "-1", items, id, match, label)
	assert.ErrorIs(t, err, apiclient.ErrNotFound, "a negative number is treated as a name")
}

func TestPlural(t *testing.T) {
	assert.Equal(t, "templates", plural("template"))
	assert.Equal(t, "inventories", plural("inventory"))
	assert.Equal(t, "repositories", plural("repository"))
}
