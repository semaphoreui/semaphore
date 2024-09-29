package api

import (
	"bytes"
	"embed"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/ansible-semaphore/semaphore/api/runners"

	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/api/projects"
	"github.com/ansible-semaphore/semaphore/api/sockets"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gorilla/mux"
)

var startTime = time.Now().UTC()

//go:embed public/*
var publicAssets embed.FS

// StoreMiddleware WTF?
func StoreMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		store := helpers.Store(r)
		//var url = r.URL.String()

		db.StoreSession(store, util.RandString(12), func() {
			next.ServeHTTP(w, r)
		})
	})
}

// JSONMiddleware ensures that all the routes respond with Json, this is added by default to all routes
func JSONMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// plainTextMiddleware resets headers to Plain Text if needed
func plainTextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "text/plain; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}

func pongHandler(w http.ResponseWriter, r *http.Request) {
	//nolint: errcheck
	w.Write([]byte("pong"))
}

// Route declares all routes
func Route() *mux.Router {
	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(servePublic)

	webPath := "/"
	if util.WebHostURL != nil {
		webPath = util.WebHostURL.Path
		if !strings.HasSuffix(webPath, "/") {
			webPath += "/"
		}
	}

	r.Use(mux.CORSMethodMiddleware(r))

	pingRouter := r.Path(webPath + "api/ping").Subrouter()
	pingRouter.Use(plainTextMiddleware)
	pingRouter.Methods("GET", "HEAD").HandlerFunc(pongHandler)

	publicAPIRouter := r.PathPrefix(webPath + "api").Subrouter()
	publicAPIRouter.Use(StoreMiddleware, JSONMiddleware)

	publicAPIRouter.HandleFunc("/auth/login", login).Methods("GET", "POST")
	publicAPIRouter.HandleFunc("/auth/logout", logout).Methods("POST")
	publicAPIRouter.HandleFunc("/auth/oidc/{provider}/login", oidcLogin).Methods("GET")
	publicAPIRouter.HandleFunc("/auth/oidc/{provider}/redirect", oidcRedirect).Methods("GET")
	publicAPIRouter.HandleFunc("/auth/oidc/{provider}/redirect/{redirect_path:.*}", oidcRedirect).Methods("GET")

	internalAPI := publicAPIRouter.PathPrefix("/internal").Subrouter()
	internalAPI.HandleFunc("/runners", runners.RegisterRunner).Methods("POST")

	runnersAPI := internalAPI.PathPrefix("/runners").Subrouter()
	runnersAPI.Use(runners.RunnerMiddleware)
	runnersAPI.Path("").HandlerFunc(runners.GetRunner).Methods("GET", "HEAD")
	runnersAPI.Path("").HandlerFunc(runners.UpdateRunner).Methods("PUT")
	runnersAPI.Path("").HandlerFunc(runners.UnregisterRunner).Methods("DELETE")

	publicWebHookRouter := r.PathPrefix(webPath + "api").Subrouter()
	publicWebHookRouter.Use(StoreMiddleware, JSONMiddleware)
	publicWebHookRouter.Path("/integrations/{integration_alias}").HandlerFunc(ReceiveIntegration).Methods("POST", "GET", "OPTIONS")

	authenticatedWS := r.PathPrefix(webPath + "api").Subrouter()
	authenticatedWS.Use(JSONMiddleware, authenticationWithStore)
	authenticatedWS.Path("/ws").HandlerFunc(sockets.Handler).Methods("GET", "HEAD")

	authenticatedAPI := r.PathPrefix(webPath + "api").Subrouter()
	authenticatedAPI.Use(StoreMiddleware, JSONMiddleware, authentication)

	authenticatedAPI.Path("/info").HandlerFunc(getSystemInfo).Methods("GET", "HEAD")

	authenticatedAPI.Path("/projects").HandlerFunc(projects.GetProjects).Methods("GET", "HEAD")
	authenticatedAPI.Path("/projects").HandlerFunc(projects.AddProject).Methods("POST")
	authenticatedAPI.Path("/projects/restore").HandlerFunc(projects.Restore).Methods("POST")
	authenticatedAPI.Path("/events").HandlerFunc(getAllEvents).Methods("GET", "HEAD")
	authenticatedAPI.HandleFunc("/events/last", getLastEvents).Methods("GET", "HEAD")

	authenticatedAPI.Path("/users").HandlerFunc(getUsers).Methods("GET", "HEAD")
	authenticatedAPI.Path("/users").HandlerFunc(addUser).Methods("POST")
	authenticatedAPI.Path("/user").HandlerFunc(getUser).Methods("GET", "HEAD")

	authenticatedAPI.Path("/apps").HandlerFunc(getApps).Methods("GET", "HEAD")

	tokenAPI := authenticatedAPI.PathPrefix("/user").Subrouter()
	tokenAPI.Path("/tokens").HandlerFunc(getAPITokens).Methods("GET", "HEAD")
	tokenAPI.Path("/tokens").HandlerFunc(createAPIToken).Methods("POST")
	tokenAPI.HandleFunc("/tokens/{token_id}", expireAPIToken).Methods("DELETE")

	adminAPI := authenticatedAPI.NewRoute().Subrouter()
	adminAPI.Use(adminMiddleware)
	adminAPI.Path("/options").HandlerFunc(getOptions).Methods("GET", "HEAD")
	adminAPI.Path("/options").HandlerFunc(setOption).Methods("POST")

	adminAPI.Path("/runners").HandlerFunc(getGlobalRunners).Methods("GET", "HEAD")
	adminAPI.Path("/runners").HandlerFunc(addGlobalRunner).Methods("POST", "HEAD")

	globalRunnersAPI := adminAPI.PathPrefix("/runners").Subrouter()
	globalRunnersAPI.Use(globalRunnerMiddleware)
	globalRunnersAPI.Path("/{runner_id}").HandlerFunc(getGlobalRunner).Methods("GET", "HEAD")
	globalRunnersAPI.Path("/{runner_id}").HandlerFunc(updateGlobalRunner).Methods("PUT", "POST")
	globalRunnersAPI.Path("/{runner_id}/active").HandlerFunc(setGlobalRunnerActive).Methods("POST")
	globalRunnersAPI.Path("/{runner_id}").HandlerFunc(deleteGlobalRunner).Methods("DELETE")

	appsAPI := adminAPI.PathPrefix("/apps").Subrouter()
	appsAPI.Use(appMiddleware)
	appsAPI.Path("/{app_id}").HandlerFunc(getApp).Methods("GET", "HEAD")
	appsAPI.Path("/{app_id}").HandlerFunc(setApp).Methods("PUT", "POST")
	appsAPI.Path("/{app_id}/active").HandlerFunc(setAppActive).Methods("POST")
	appsAPI.Path("/{app_id}").HandlerFunc(deleteApp).Methods("DELETE")

	userAPI := authenticatedAPI.Path("/users/{user_id}").Subrouter()
	userAPI.Use(getUserMiddleware)

	userAPI.Methods("GET", "HEAD").HandlerFunc(getUser)
	userAPI.Methods("PUT").HandlerFunc(updateUser)
	userAPI.Methods("DELETE").HandlerFunc(deleteUser)

	userPasswordAPI := authenticatedAPI.PathPrefix("/users/{user_id}").Subrouter()
	userPasswordAPI.Use(getUserMiddleware)
	userPasswordAPI.Path("/password").HandlerFunc(updateUserPassword).Methods("POST")

	projectGet := authenticatedAPI.Path("/project/{project_id}").Subrouter()
	projectGet.Use(projects.ProjectMiddleware)
	projectGet.Methods("GET", "HEAD").HandlerFunc(projects.GetProject)

	//
	// Start and Stop tasks
	projectTaskStart := authenticatedAPI.PathPrefix("/project/{project_id}").Subrouter()
	projectTaskStart.Use(projects.ProjectMiddleware, projects.GetMustCanMiddleware(db.CanRunProjectTasks))
	projectTaskStart.Path("/tasks").HandlerFunc(projects.AddTask).Methods("POST")

	projectTaskStop := authenticatedAPI.PathPrefix("/project/{project_id}").Subrouter()
	projectTaskStop.Use(projects.ProjectMiddleware, projects.GetTaskMiddleware, projects.GetMustCanMiddleware(db.CanRunProjectTasks))
	projectTaskStop.HandleFunc("/tasks/{task_id}/stop", projects.StopTask).Methods("POST")
	projectTaskStop.HandleFunc("/tasks/{task_id}/confirm", projects.ConfirmTask).Methods("POST")

	//
	// Project resources CRUD
	projectUserAPI := authenticatedAPI.PathPrefix("/project/{project_id}").Subrouter()
	projectUserAPI.Use(projects.ProjectMiddleware, projects.GetMustCanMiddleware(db.CanManageProjectResources))

	projectUserAPI.Path("/role").HandlerFunc(projects.GetUserRole).Methods("GET", "HEAD")

	projectUserAPI.Path("/events").HandlerFunc(getAllEvents).Methods("GET", "HEAD")
	projectUserAPI.HandleFunc("/events/last", getLastEvents).Methods("GET", "HEAD")

	projectUserAPI.Path("/users").HandlerFunc(projects.GetUsers).Methods("GET", "HEAD")

	projectUserAPI.Path("/keys").HandlerFunc(projects.GetKeys).Methods("GET", "HEAD")
	projectUserAPI.Path("/keys").HandlerFunc(projects.AddKey).Methods("POST")

	projectUserAPI.Path("/repositories").HandlerFunc(projects.GetRepositories).Methods("GET", "HEAD")
	projectUserAPI.Path("/repositories").HandlerFunc(projects.AddRepository).Methods("POST")

	projectUserAPI.Path("/inventory").HandlerFunc(projects.GetInventory).Methods("GET", "HEAD")
	projectUserAPI.Path("/inventory").HandlerFunc(projects.AddInventory).Methods("POST")

	projectUserAPI.Path("/environment").HandlerFunc(projects.GetEnvironment).Methods("GET", "HEAD")
	projectUserAPI.Path("/environment").HandlerFunc(projects.AddEnvironment).Methods("POST")

	projectUserAPI.Path("/tasks").HandlerFunc(projects.GetAllTasks).Methods("GET", "HEAD")
	projectUserAPI.HandleFunc("/tasks/last", projects.GetLastTasks).Methods("GET", "HEAD")

	projectUserAPI.Path("/templates").HandlerFunc(projects.GetTemplates).Methods("GET", "HEAD")
	projectUserAPI.Path("/templates").HandlerFunc(projects.AddTemplate).Methods("POST")

	projectUserAPI.Path("/schedules").HandlerFunc(projects.GetProjectSchedules).Methods("GET", "HEAD")
	projectUserAPI.Path("/schedules").HandlerFunc(projects.AddSchedule).Methods("POST")
	projectUserAPI.Path("/schedules/validate").HandlerFunc(projects.ValidateScheduleCronFormat).Methods("POST")

	projectUserAPI.Path("/views").HandlerFunc(projects.GetViews).Methods("GET", "HEAD")
	projectUserAPI.Path("/views").HandlerFunc(projects.AddView).Methods("POST")
	projectUserAPI.Path("/views/positions").HandlerFunc(projects.SetViewPositions).Methods("POST")

	projectUserAPI.Path("/integrations").HandlerFunc(projects.GetIntegrations).Methods("GET", "HEAD")
	projectUserAPI.Path("/integrations").HandlerFunc(projects.AddIntegration).Methods("POST")
	projectUserAPI.Path("/backup").HandlerFunc(projects.GetBackup).Methods("GET", "HEAD")

	//
	// Updating and deleting project
	projectAdminAPI := authenticatedAPI.Path("/project/{project_id}").Subrouter()
	projectAdminAPI.Use(projects.ProjectMiddleware, projects.GetMustCanMiddleware(db.CanUpdateProject))
	projectAdminAPI.Methods("PUT").HandlerFunc(projects.UpdateProject)
	projectAdminAPI.Methods("DELETE").HandlerFunc(projects.DeleteProject)

	meAPI := authenticatedAPI.Path("/project/{project_id}/me").Subrouter()
	meAPI.Use(projects.ProjectMiddleware)
	meAPI.HandleFunc("", projects.LeftProject).Methods("DELETE")

	//
	// Manage project users
	projectAdminUsersAPI := authenticatedAPI.PathPrefix("/project/{project_id}").Subrouter()

	projectAdminUsersAPI.Use(projects.ProjectMiddleware, projects.GetMustCanMiddleware(db.CanManageProjectUsers))
	projectAdminUsersAPI.Path("/users").HandlerFunc(projects.AddUser).Methods("POST")

	projectUserManagement := projectAdminUsersAPI.PathPrefix("/users").Subrouter()
	projectUserManagement.Use(projects.UserMiddleware)

	projectUserManagement.HandleFunc("/{user_id}", projects.GetUsers).Methods("GET", "HEAD")
	projectUserManagement.HandleFunc("/{user_id}", projects.UpdateUser).Methods("PUT")
	projectUserManagement.HandleFunc("/{user_id}", projects.RemoveUser).Methods("DELETE")

	//
	// Project resources CRUD (continue)
	projectKeyManagement := projectUserAPI.PathPrefix("/keys").Subrouter()
	projectKeyManagement.Use(projects.KeyMiddleware)

	projectKeyManagement.HandleFunc("/{key_id}", projects.GetKeys).Methods("GET", "HEAD")
	projectKeyManagement.HandleFunc("/{key_id}/refs", projects.GetKeyRefs).Methods("GET", "HEAD")
	projectKeyManagement.HandleFunc("/{key_id}", projects.UpdateKey).Methods("PUT")
	projectKeyManagement.HandleFunc("/{key_id}", projects.RemoveKey).Methods("DELETE")

	projectRepoManagement := projectUserAPI.PathPrefix("/repositories").Subrouter()
	projectRepoManagement.Use(projects.RepositoryMiddleware)

	projectRepoManagement.HandleFunc("/{repository_id}", projects.GetRepositories).Methods("GET", "HEAD")
	projectRepoManagement.HandleFunc("/{repository_id}/refs", projects.GetRepositoryRefs).Methods("GET", "HEAD")
	projectRepoManagement.HandleFunc("/{repository_id}", projects.UpdateRepository).Methods("PUT")
	projectRepoManagement.HandleFunc("/{repository_id}", projects.RemoveRepository).Methods("DELETE")

	projectInventoryManagement := projectUserAPI.PathPrefix("/inventory").Subrouter()
	projectInventoryManagement.Use(projects.InventoryMiddleware)

	projectInventoryManagement.HandleFunc("/{inventory_id}", projects.GetInventory).Methods("GET", "HEAD")
	projectInventoryManagement.HandleFunc("/{inventory_id}/refs", projects.GetInventoryRefs).Methods("GET", "HEAD")
	projectInventoryManagement.HandleFunc("/{inventory_id}", projects.UpdateInventory).Methods("PUT")
	projectInventoryManagement.HandleFunc("/{inventory_id}", projects.RemoveInventory).Methods("DELETE")

	projectEnvManagement := projectUserAPI.PathPrefix("/environment").Subrouter()
	projectEnvManagement.Use(projects.EnvironmentMiddleware)

	projectEnvManagement.HandleFunc("/{environment_id}", projects.GetEnvironment).Methods("GET", "HEAD")
	projectEnvManagement.HandleFunc("/{environment_id}/refs", projects.GetEnvironmentRefs).Methods("GET", "HEAD")
	projectEnvManagement.HandleFunc("/{environment_id}", projects.UpdateEnvironment).Methods("PUT")
	projectEnvManagement.HandleFunc("/{environment_id}", projects.RemoveEnvironment).Methods("DELETE")

	projectTmplManagement := projectUserAPI.PathPrefix("/templates").Subrouter()
	projectTmplManagement.Use(projects.TemplatesMiddleware)

	projectTmplManagement.HandleFunc("/{template_id}", projects.UpdateTemplate).Methods("PUT")
	projectTmplManagement.HandleFunc("/{template_id}", projects.RemoveTemplate).Methods("DELETE")
	projectTmplManagement.HandleFunc("/{template_id}", projects.GetTemplate).Methods("GET")
	projectTmplManagement.HandleFunc("/{template_id}/refs", projects.GetTemplateRefs).Methods("GET", "HEAD")
	projectTmplManagement.HandleFunc("/{template_id}/tasks", projects.GetAllTasks).Methods("GET")
	projectTmplManagement.HandleFunc("/{template_id}/tasks/last", projects.GetLastTasks).Methods("GET")
	projectTmplManagement.HandleFunc("/{template_id}/schedules", projects.GetTemplateSchedules).Methods("GET")

	projectTaskManagement := projectUserAPI.PathPrefix("/tasks").Subrouter()
	projectTaskManagement.Use(projects.GetTaskMiddleware)

	projectTaskManagement.HandleFunc("/{task_id}/output", projects.GetTaskOutput).Methods("GET", "HEAD")
	projectTaskManagement.HandleFunc("/{task_id}", projects.GetTask).Methods("GET", "HEAD")
	projectTaskManagement.HandleFunc("/{task_id}", projects.RemoveTask).Methods("DELETE")

	projectScheduleManagement := projectUserAPI.PathPrefix("/schedules").Subrouter()
	projectScheduleManagement.Use(projects.SchedulesMiddleware)
	projectScheduleManagement.HandleFunc("/{schedule_id}", projects.GetSchedule).Methods("GET", "HEAD")
	projectScheduleManagement.HandleFunc("/{schedule_id}", projects.UpdateSchedule).Methods("PUT")
	projectScheduleManagement.HandleFunc("/{schedule_id}/active", projects.SetScheduleActive).Methods("PUT")
	projectScheduleManagement.HandleFunc("/{schedule_id}", projects.RemoveSchedule).Methods("DELETE")

	projectViewManagement := projectUserAPI.PathPrefix("/views").Subrouter()
	projectViewManagement.Use(projects.ViewMiddleware)
	projectViewManagement.HandleFunc("/{view_id}", projects.GetViews).Methods("GET", "HEAD")
	projectViewManagement.HandleFunc("/{view_id}", projects.UpdateView).Methods("PUT")
	projectViewManagement.HandleFunc("/{view_id}", projects.RemoveView).Methods("DELETE")
	projectViewManagement.HandleFunc("/{view_id}/templates", projects.GetViewTemplates).Methods("GET", "HEAD")

	projectIntegrationsAliasAPI := projectUserAPI.PathPrefix("/integrations").Subrouter()
	projectIntegrationsAliasAPI.Use(projects.ProjectMiddleware)
	projectIntegrationsAliasAPI.HandleFunc("/aliases", projects.GetIntegrationAlias).Methods("GET", "HEAD")
	projectIntegrationsAliasAPI.HandleFunc("/aliases", projects.AddIntegrationAlias).Methods("POST")
	projectIntegrationsAliasAPI.HandleFunc("/aliases/{alias_id}", projects.RemoveIntegrationAlias).Methods("DELETE")

	projectIntegrationsAPI := projectUserAPI.PathPrefix("/integrations").Subrouter()
	projectIntegrationsAPI.Use(projects.ProjectMiddleware, projects.IntegrationMiddleware)
	projectIntegrationsAPI.HandleFunc("/{integration_id}", projects.UpdateIntegration).Methods("PUT")
	projectIntegrationsAPI.HandleFunc("/{integration_id}", projects.DeleteIntegration).Methods("DELETE")
	projectIntegrationsAPI.HandleFunc("/{integration_id}", projects.GetIntegration).Methods("GET")
	projectIntegrationsAPI.HandleFunc("/{integration_id}/refs", projects.GetIntegrationRefs).Methods("GET", "HEAD")
	projectIntegrationsAPI.HandleFunc("/{integration_id}/matchers", projects.GetIntegrationMatchers).Methods("GET", "HEAD")
	projectIntegrationsAPI.HandleFunc("/{integration_id}/matchers", projects.AddIntegrationMatcher).Methods("POST")
	projectIntegrationsAPI.HandleFunc("/{integration_id}/values", projects.GetIntegrationExtractValues).Methods("GET", "HEAD")
	projectIntegrationsAPI.HandleFunc("/{integration_id}/values", projects.AddIntegrationExtractValue).Methods("POST")
	projectIntegrationsAPI.HandleFunc("/{integration_id}/aliases", projects.GetIntegrationAlias).Methods("GET", "HEAD")
	projectIntegrationsAPI.HandleFunc("/{integration_id}/aliases", projects.AddIntegrationAlias).Methods("POST")
	projectIntegrationsAPI.HandleFunc("/{integration_id}/aliases/{alias_id}", projects.RemoveIntegrationAlias).Methods("DELETE")

	projectIntegrationsAPI.HandleFunc("/{integration_id}/matchers/{matcher_id}", projects.GetIntegrationMatcher).Methods("GET", "HEAD")
	projectIntegrationsAPI.HandleFunc("/{integration_id}/matchers/{matcher_id}", projects.UpdateIntegrationMatcher).Methods("PUT")
	projectIntegrationsAPI.HandleFunc("/{integration_id}/matchers/{matcher_id}", projects.DeleteIntegrationMatcher).Methods("DELETE")
	projectIntegrationsAPI.HandleFunc("/{integration_id}/matchers/{matcher_id}/refs", projects.GetIntegrationMatcherRefs).Methods("GET", "HEAD")

	projectIntegrationsAPI.HandleFunc("/{integration_id}/values/{value_id}", projects.GetIntegrationExtractValue).Methods("GET", "HEAD")
	projectIntegrationsAPI.HandleFunc("/{integration_id}/values/{value_id}", projects.UpdateIntegrationExtractValue).Methods("PUT")
	projectIntegrationsAPI.HandleFunc("/{integration_id}/values/{value_id}", projects.DeleteIntegrationExtractValue).Methods("DELETE")
	projectIntegrationsAPI.HandleFunc("/{integration_id}/values/{value_id}/refs", projects.GetIntegrationExtractValueRefs).Methods("GET")

	if os.Getenv("DEBUG") == "1" {
		defer debugPrintRoutes(r)
	}

	return r
}

func debugPrintRoutes(r *mux.Router) {
	err := r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err == nil {
			fmt.Println("ROUTE:", pathTemplate)
		}
		pathRegexp, err := route.GetPathRegexp()
		if err == nil {
			fmt.Println("Path regexp:", pathRegexp)
		}
		queriesTemplates, err := route.GetQueriesTemplates()
		if err == nil {
			fmt.Println("Queries templates:", strings.Join(queriesTemplates, ","))
		}
		queriesRegexps, err := route.GetQueriesRegexp()
		if err == nil {
			fmt.Println("Queries regexps:", strings.Join(queriesRegexps, ","))
		}
		methods, err := route.GetMethods()
		if err == nil {
			fmt.Println("Methods:", strings.Join(methods, ","))
		}
		fmt.Println()
		return nil
	})

	if err != nil {
		fmt.Println(err)
	}
}

func servePublic(w http.ResponseWriter, r *http.Request) {
	webPath := "/"
	if util.WebHostURL != nil {
		webPath = util.WebHostURL.Path
		if !strings.HasSuffix(webPath, "/") {
			webPath += "/"
		}
	}

	reqPath := r.URL.Path
	apiPath := path.Join(webPath, "api")

	if reqPath == apiPath || strings.HasPrefix(reqPath, apiPath) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	if !strings.Contains(reqPath, ".") {
		serveFile(w, r, "index.html")
		return
	}

	newPath := strings.Replace(
		reqPath,
		webPath,
		"",
		1,
	)

	serveFile(w, r, newPath)
}

func serveFile(w http.ResponseWriter, r *http.Request, name string) {
	res, err := publicAssets.ReadFile(
		fmt.Sprintf("public/%s", name),
	)

	if err != nil {
		http.Error(
			w,
			http.StatusText(http.StatusNotFound),
			http.StatusNotFound,
		)

		return
	}

	if util.WebHostURL != nil && name == "index.html" {
		baseURL := util.WebHostURL.String()

		if !strings.HasSuffix(baseURL, "/") {
			baseURL += "/"
		}

		res = []byte(
			strings.Replace(
				string(res),
				`<base href="/">`,
				fmt.Sprintf(`<base href="%s">`, baseURL),
				1,
			),
		)
	}

	if !strings.HasSuffix(name, ".html") {
		w.Header().Add(
			"Cache-Control",
			fmt.Sprintf("max-age=%d, public, must-revalidate, proxy-revalidate", 24*time.Hour),
		)
	}

	http.ServeContent(
		w,
		r,
		name,
		startTime,
		bytes.NewReader(
			res,
		),
	)
}

func getSystemInfo(w http.ResponseWriter, r *http.Request) {
	host := ""

	if util.WebHostURL != nil {
		host = util.WebHostURL.String()
	}

	body := map[string]interface{}{
		"version":           util.Version(),
		"ansible":           util.AnsibleVersion(),
		"web_host":          host,
		"use_remote_runner": util.Config.UseRemoteRunner,
	}

	helpers.WriteJSON(w, http.StatusOK, body)
}
