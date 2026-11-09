package projects

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type nopLogWriter struct{}

func (nopLogWriter) WriteEventLog(pro_interfaces.EventLogRecord) error { return nil }
func (nopLogWriter) WriteTaskLog(pro_interfaces.TaskLogRecord) error   { return nil }
func (nopLogWriter) WriteResult(any) error                             { return nil }

// postRepository sends POST /repositories through the middlewares api/router.go puts
// in front of it, so the request passes the real membership and permission checks.
func postRepository(store db.Store, user db.User, projectID int, body string) *httptest.ResponseRecorder {
	router := mux.NewRouter()
	projectUserAPI := router.PathPrefix("/api/project/{project_id}").Subrouter()
	projectUserAPI.Use(ProjectMiddleware, GetMustCanMiddleware(db.CanManageProjectResources))
	projectUserAPI.Path("/repositories").HandlerFunc(AddRepository).Methods(http.MethodPost)

	r := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/project/%d/repositories", projectID), strings.NewReader(body))
	r = helpers.SetContextValue(r, "store", store)
	r = helpers.SetContextValue(r, "user", &user)
	r = helpers.SetContextValue(r, "log_writer", nopLogWriter{})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	return w
}

// TestAddRepository_WritesOnlyToTheProjectOfTheURL reproduces GHSA-q9cq-8j9q-r93m.
//
// The route checks membership and permissions against the project of the URL.
// AddRepository answered 400 to a body naming another project but did not return:
// the key was looked up in the project of the body and the repository was inserted
// there. A manager of one project could so create a repository in a project they are
// not a member of, knowing only the id of one of its keys — and every project has a
// None key.
//
// The test asserts the secure outcome: the case "foreign project and its key in the
// body" fails where the return is missing (2-19-stable up to v2.19.21). The other
// cases describe the boundary around it and pass either way.
func TestAddRepository_WritesOnlyToTheProjectOfTheURL(t *testing.T) {
	const own, foreign = "own", "foreign"
	tests := []struct {
		name        string
		url         string // project of the URL
		body        string // project_id of the body
		key         string // project of ssh_key_id
		wantCode    int
		wantCreated bool
	}{
		{"own project", own, own, own, http.StatusCreated, true},
		{"foreign project and its key in the body", own, foreign, foreign, http.StatusBadRequest, false},
		{"foreign key in the body", own, own, foreign, http.StatusNotFound, false},
		{"foreign project in the URL", foreign, foreign, foreign, http.StatusNotFound, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := sql.InitConfigCreateTestStore()
			project, err := store.CreateProject(db.Project{Name: "p"})
			require.NoError(t, err)
			other, err := store.CreateProject(db.Project{Name: "other"})
			require.NoError(t, err)
			manager, err := store.CreateUser(db.UserWithPwd{Pwd: "verystrongpassword1", User: db.User{Username: "member", Name: "member", Email: "member@example.com"}})
			require.NoError(t, err)
			_, err = store.CreateProjectUser(db.ProjectUser{ProjectID: project.ID, UserID: manager.ID, Role: db.ProjectManager})
			require.NoError(t, err)
			_, err = store.GetProjectUser(other.ID, manager.ID)
			require.ErrorIs(t, err, db.ErrNotFound, "the manager must not be a member of the foreign project")
			projects := map[string]db.Project{own: project, foreign: other}
			keys := map[string]db.AccessKey{}
			for _, side := range []string{own, foreign} {
				// AddProject gives every project such a key.
				keys[side], err = store.CreateAccessKey(db.AccessKey{Name: "None", Type: db.AccessKeyNone, ProjectID: new(projects[side].ID)})
				require.NoError(t, err)
			}
			body := fmt.Sprintf(`{"name":"ops","project_id":%d,"git_url":"https://git.example.com/ops.git","git_branch":"main","ssh_key_id":%d}`,
				projects[tt.body].ID, keys[tt.key].ID)

			w := postRepository(store, manager, projects[tt.url].ID, body)

			assert.Equal(t, tt.wantCode, w.Code, w.Body.String())
			inForeign, err := store.GetRepositories(other.ID, db.RetrieveQueryParams{})
			require.NoError(t, err)
			assert.Empty(t, inForeign, "a repository was created in a project the manager is not a member of")
			inOwn, err := store.GetRepositories(project.ID, db.RetrieveQueryParams{})
			require.NoError(t, err)
			if tt.wantCreated {
				require.Len(t, inOwn, 1)
				assert.Equal(t, "ops", inOwn[0].Name)
				return
			}
			assert.Empty(t, inOwn)
			assert.NotContains(t, w.Body.String(), `"ops"`, "a rejected request must not answer with a repository")
		})
	}
}
