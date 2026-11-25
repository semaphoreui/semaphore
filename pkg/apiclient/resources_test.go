package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// recorder is a server that records the last request and answers with a
// canned JSON body.
type recorder struct {
	method, path, query string
	body                map[string]any
	rawBody             []byte
	cookies             []*http.Cookie
	setCookies          []*http.Cookie
	answer              any
	status              int
	calls               int
}

func (rec *recorder) handler(t *testing.T) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.calls++
		rec.method, rec.path, rec.query = r.Method, r.URL.Path, r.URL.RawQuery
		rec.cookies = r.Cookies()
		rec.body = nil
		rec.rawBody = nil
		if r.Body != nil {
			buf := make([]byte, 1<<16)
			n, _ := r.Body.Read(buf)
			rec.rawBody = buf[:n]
			_ = json.Unmarshal(rec.rawBody, &rec.body)
		}
		for _, c := range rec.setCookies {
			http.SetCookie(w, c)
		}
		if rec.status != 0 {
			w.WriteHeader(rec.status)
		}
		if rec.answer != nil {
			require.NoError(t, json.NewEncoder(w).Encode(rec.answer))
		}
	})
}

func TestResourcePaths(t *testing.T) {
	rec := &recorder{}
	srv := httptest.NewServer(rec.handler(t))
	defer srv.Close()
	c := newTestClient(t, srv, nil)
	ctx := context.Background()
	p := 3

	tests := []struct {
		name string
		call func() error
		want string // METHOD path
		body map[string]any
	}{
		{"projects", func() error { _, e := c.ListProjects(ctx); return e }, "GET /api/projects", nil},
		{"project", func() error { _, e := c.GetProject(ctx, p); return e }, "GET /api/project/3", nil},
		{"create project", func() error { _, e := c.CreateProject(ctx, db.Project{Name: "n"}, true); return e }, "POST /api/projects", map[string]any{"name": "n", "demo": true}},
		{"update project", func() error { return c.UpdateProject(ctx, db.Project{ID: p, Name: "n"}) }, "PUT /api/project/3", nil},
		{"delete project", func() error { return c.DeleteProject(ctx, p) }, "DELETE /api/project/3", nil},
		{"members", func() error { _, e := c.ListProjectMembers(ctx, p); return e }, "GET /api/project/3/users", nil},
		{"add member", func() error { return c.AddProjectMember(ctx, p, 8, "manager") }, "POST /api/project/3/users", map[string]any{"user_id": float64(8), "role": "manager"}},
		{"update member", func() error { return c.UpdateProjectMember(ctx, p, 8, "guest") }, "PUT /api/project/3/users/8", map[string]any{"role": "guest"}},
		{"remove member", func() error { return c.RemoveProjectMember(ctx, p, 8) }, "DELETE /api/project/3/users/8", nil},
		{"users", func() error { _, e := c.ListUsers(ctx); return e }, "GET /api/users", nil},
		{"user", func() error { _, e := c.GetUser(ctx, 8); return e }, "GET /api/users/8", nil},
		{"delete user", func() error { return c.DeleteUser(ctx, 8) }, "DELETE /api/users/8", nil},
		{"set password", func() error { return c.SetUserPassword(ctx, 8, "new", "") }, "POST /api/users/8/password", map[string]any{"password": "new"}},
		{"templates", func() error { _, e := c.ListTemplates(ctx, p); return e }, "GET /api/project/3/templates", nil},
		{"template", func() error { _, e := c.GetTemplate(ctx, p, 4); return e }, "GET /api/project/3/templates/4", nil},
		{"update template", func() error { return c.UpdateTemplate(ctx, p, db.Template{ID: 4}) }, "PUT /api/project/3/templates/4", nil},
		{"delete template", func() error { return c.DeleteTemplate(ctx, p, 4) }, "DELETE /api/project/3/templates/4", nil},
		{"stop all", func() error { return c.StopAllTemplateTasks(ctx, p, 4) }, "POST /api/project/3/templates/4/stop_all_tasks", nil},
		{"inventory list", func() error { _, e := c.ListInventories(ctx, p); return e }, "GET /api/project/3/inventory", nil},
		{"inventory update", func() error { return c.UpdateInventory(ctx, p, db.Inventory{ID: 2}) }, "PUT /api/project/3/inventory/2", nil},
		{"environment", func() error { _, e := c.GetEnvironment(ctx, p, 2); return e }, "GET /api/project/3/environment/2", nil},
		{"keys", func() error { _, e := c.ListKeys(ctx, p); return e }, "GET /api/project/3/keys", nil},
		{"delete key", func() error { return c.DeleteKey(ctx, p, 2) }, "DELETE /api/project/3/keys/2", nil},
		{"repositories", func() error { _, e := c.ListRepositories(ctx, p); return e }, "GET /api/project/3/repositories", nil},
		{"schedules", func() error { _, e := c.ListSchedules(ctx, p); return e }, "GET /api/project/3/schedules", nil},
		{"template schedules", func() error { _, e := c.ListTemplateSchedules(ctx, p, 4); return e }, "GET /api/project/3/templates/4/schedules", nil},
		{"schedule active", func() error { return c.SetScheduleActive(ctx, p, 2, false) }, "PUT /api/project/3/schedules/2/active", map[string]any{"active": false}},
		{"validate cron", func() error { return c.ValidateCron(ctx, p, "* * * * *") }, "POST /api/project/3/schedules/validate", map[string]any{"cron_format": "* * * * *"}},
		{"views", func() error { _, e := c.ListViews(ctx, p); return e }, "GET /api/project/3/views", nil},
		{"view templates", func() error { _, e := c.ListViewTemplates(ctx, p, 2); return e }, "GET /api/project/3/views/2/templates", nil},
		{"global runners", func() error { _, e := c.ListRunners(ctx, GlobalScope); return e }, "GET /api/runners", nil},
		{"project runners", func() error { _, e := c.ListRunners(ctx, p); return e }, "GET /api/project/3/runners", nil},
		{"runner active", func() error { return c.SetRunnerActive(ctx, GlobalScope, 6, true) }, "POST /api/runners/6/active", map[string]any{"active": true}},
		{"runner token", func() error { _, e := c.RegenerateRunnerToken(ctx, p, 6); return e }, "POST /api/project/3/runners/6/registration-token", nil},
		{"runner cache", func() error { return c.ClearRunnerCache(ctx, GlobalScope, 6) }, "DELETE /api/runners/6/cache", nil},
		{"delete runner", func() error { return c.DeleteRunner(ctx, p, 6) }, "DELETE /api/project/3/runners/6", nil},
		{"info", func() error { _, e := c.Info(ctx); return e }, "GET /api/info", nil},
		{"me", func() error { _, e := c.Me(ctx); return e }, "GET /api/user", nil},
		{"login methods", func() error { _, e := c.LoginMethods(ctx); return e }, "GET /api/auth/login", nil},
		{"tokens", func() error { _, e := c.ListTokens(ctx); return e }, "GET /api/user/tokens", nil},
		{"delete token", func() error { return c.DeleteToken(ctx, "abcd1234") }, "DELETE /api/user/tokens/abcd1234", nil},
		{"logout", func() error { return c.Logout(ctx) }, "POST /api/auth/logout", nil},
		{"verify", func() error { return c.Verify(ctx, "123456") }, "POST /api/auth/verify", map[string]any{"passcode": "123456"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec.answer = nil
			require.NoError(t, tt.call())
			assert.Equal(t, tt.want, rec.method+" "+rec.path)
			if tt.body != nil {
				for k, v := range tt.body {
					assert.Equal(t, v, rec.body[k], "body field %s", k)
				}
			}
		})
	}
}

func TestCreateSetsProjectAndDecodes(t *testing.T) {
	rec := &recorder{answer: db.Template{ID: 11, Name: "deploy"}, status: http.StatusCreated}
	srv := httptest.NewServer(rec.handler(t))
	defer srv.Close()
	c := newTestClient(t, srv, nil)

	created, err := c.CreateTemplate(context.Background(), 3, db.Template{Name: "deploy"})
	require.NoError(t, err)
	assert.Equal(t, 11, created.ID)
	assert.EqualValues(t, 3, rec.body["project_id"])

	rec.answer = db.AccessKey{ID: 5}
	key, err := c.CreateKey(context.Background(), 3, db.AccessKey{Name: "k", Type: db.AccessKeySSH})
	require.NoError(t, err)
	assert.Equal(t, 5, key.ID)
	assert.EqualValues(t, 3, rec.body["project_id"])

	rec.answer = RunnerWithToken{Runner: db.Runner{ID: 2, Name: "r"}, Token: "reg-token"}
	runner, err := c.CreateRunner(context.Background(), GlobalScope, db.Runner{Name: "r"})
	require.NoError(t, err)
	assert.Equal(t, "POST /api/runners", rec.method+" "+rec.path)
	assert.Equal(t, "reg-token", runner.Token)
	assert.Equal(t, 2, runner.ID)
}

func TestCreateUser_SendsPassword(t *testing.T) {
	rec := &recorder{answer: db.User{ID: 9}}
	srv := httptest.NewServer(rec.handler(t))
	defer srv.Close()
	c := newTestClient(t, srv, nil)

	user, err := c.CreateUser(context.Background(), db.User{Username: "bob", Name: "Bob", Email: "b@x"}, "pw")
	require.NoError(t, err)
	assert.Equal(t, 9, user.ID)
	assert.Equal(t, "bob", rec.body["username"])
	assert.Equal(t, "pw", rec.body["password"])
}

func TestCreateToken(t *testing.T) {
	exp := time.Date(2027, 1, 2, 3, 4, 5, 0, time.UTC)
	rec := &recorder{answer: db.APIToken{ID: "full-token-value", Name: "cli"}, status: http.StatusCreated}
	srv := httptest.NewServer(rec.handler(t))
	defer srv.Close()
	c := newTestClient(t, srv, nil)

	tok, err := c.CreateToken(context.Background(), "cli", &exp)
	require.NoError(t, err)
	assert.Equal(t, "POST /api/user/tokens", rec.method+" "+rec.path)
	assert.Equal(t, "full-token-value", tok.ID)
	assert.Equal(t, "cli", rec.body["name"])
	assert.Equal(t, "2027-01-02T03:04:05Z", rec.body["expires_at"])

	_, err = c.CreateToken(context.Background(), "", nil)
	require.NoError(t, err)
	assert.Nil(t, rec.body["expires_at"])
}

func TestLogin_KeepsSessionCookieInMemory(t *testing.T) {
	rec := &recorder{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			rec.body = map[string]any{"auth": body["auth"], "password": body["password"]}
			http.SetCookie(w, &http.Cookie{Name: "semaphore", Value: "sess-1", Path: "/"})
			w.WriteHeader(http.StatusNoContent)
		case "/api/user":
			if c, err := r.Cookie("semaphore"); err != nil || c.Value != "sess-1" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			_ = json.NewEncoder(w).Encode(Me{User: db.User{ID: 1, Username: "admin"}})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv, func(o *Options) { o.Token = "" })
	require.NoError(t, c.Login(context.Background(), "admin", "pw"))
	assert.Equal(t, map[string]any{"auth": "admin", "password": "pw"}, rec.body)

	me, err := c.Me(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "admin", me.Username)

	// A second client has its own empty jar: the cookie did not leak anywhere.
	other := newTestClient(t, srv, func(o *Options) { o.Token = "" })
	_, err = other.Me(context.Background())
	assert.ErrorIs(t, err, ErrUnauthorized)
}

func TestProjectBackupRestore(t *testing.T) {
	rec := &recorder{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method + " " + r.URL.Path {
		case "GET /api/project/3/backup":
			_, _ = w.Write([]byte(`{"meta":{"name":"p"}}`))
		case "POST /api/projects/restore":
			buf := make([]byte, 1024)
			n, _ := r.Body.Read(buf)
			rec.rawBody = buf[:n]
			_ = json.NewEncoder(w).Encode(db.Project{ID: 12, Name: "p"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	c := newTestClient(t, srv, nil)

	var backup bytes.Buffer
	require.NoError(t, c.BackupProject(context.Background(), 3, &backup))
	assert.Equal(t, `{"meta":{"name":"p"}}`, backup.String())

	restored, err := c.RestoreProject(context.Background(), &backup)
	require.NoError(t, err)
	assert.Equal(t, 12, restored.ID)
	assert.JSONEq(t, `{"meta":{"name":"p"}}`, string(rec.rawBody))
}
