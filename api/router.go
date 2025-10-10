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

	"github.com/semaphoreui/semaphore/pro_interfaces"

	proApi "github.com/semaphoreui/semaphore/pro/api"
	proProjects "github.com/semaphoreui/semaphore/pro/api/projects"
	proFeatures "github.com/semaphoreui/semaphore/pro/pkg/features"
	"github.com/semaphoreui/semaphore/services/server"
	taskServices "github.com/semaphoreui/semaphore/services/tasks"

	"github.com/semaphoreui/semaphore/api/debug"
	"github.com/semaphoreui/semaphore/api/tasks"
	"github.com/semaphoreui/semaphore/pkg/tz"
	log "github.com/sirupsen/logrus"

	"github.com/semaphoreui/semaphore/api/runners"

	"github.com/gorilla/mux"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/api/projects"
	"github.com/semaphoreui/semaphore/api/sockets"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
)

var startTime = tz.Now()

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

// DelayMiddleware adds artificial delay to simulate slow network conditions
func DelayMiddleware(delay time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(delay)
			next.ServeHTTP(w, r)
		})
	}
}

// Route declares all routes
func Route(
	store db.Store,
	terraformStore db.TerraformStore,
	ansibleTaskRepo db.AnsibleTaskRepository,
	taskPool *taskServices.TaskPool,
	projectService server.ProjectService,
	integrationService server.IntegrationService,
	encryptionService server.AccessKeyEncryptionService,
	accessKeyInstallationService server.AccessKeyInstallationService,
	secretStorageService server.SecretStorageService,
	accessKeyService server.AccessKeyService,
	environmentService server.EnvironmentService,
	subscriptionService pro_interfaces.SubscriptionService,
) *mux.Router {

	projectController := &projects.ProjectController{ProjectService: projectService}
	runnerController := runners.NewRunnerController(store, taskPool, encryptionService)
	integrationController := NewIntegrationController(integrationService)
	environmentController := projects.NewEnvironmentController(store, encryptionService, accessKeyService, environmentService)
	secretStorageController := projects.NewSecretStorageController(store, secretStorageService)
	repositoryController := projects.NewRepositoryController(accessKeyInstallationService)
	keyController := projects.NewKeyController(accessKeyService)
	projectsController := projects.NewProjectsController(accessKeyService)
	terraformController := proApi.NewTerraformController(encryptionService, terraformStore, store)
	terraformInventoryController := proProjects.NewTerraformInventoryController(terraformStore)
	userController := NewUserController(subscriptionService)
	usersController := NewUsersController(subscriptionService)
	subscriptionController := proApi.NewSubscriptionController(store, store)
	projectRunnerController := proProjects.NewProjectRunnerController()
	taskController := projects.NewTaskController(ansibleTaskRepo)
	rolesController := proApi.NewRolesController(store)
	templateController := projects.NewTemplateController(store, store)

	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(servePublic)

	if util.Config.Debugging.ApiDelay != "" {
		delay, err := time.ParseDuration(util.Config.Debugging.ApiDelay)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"context": "debugging",
			}).Panic("Invalid API delay format")
		}
		r.Use(DelayMiddleware(delay))
	}

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
	publicAPIRouter.HandleFunc("/auth/verify/email", startEmailVerification).Methods("POST")
	publicAPIRouter.HandleFunc("/auth/login/email", loginEmail).Methods("GET", "POST")
	publicAPIRouter.HandleFunc("/auth/login/email/resend", resendEmailOtp).Methods("GET", "POST")
	publicAPIRouter.HandleFunc("/auth/verify", verifySession).Methods("POST")
	publicAPIRouter.HandleFunc("/auth/recovery", recoverySession).Methods("POST")

	publicAPIRouter.HandleFunc("/auth/logout", logout).Methods("POST")
	publicAPIRouter.HandleFunc("/auth/oidc/{provider}/login", oidcLogin).Methods("GET")
	publicAPIRouter.HandleFunc("/auth/oidc/{provider}/redirect", oidcRedirect).Methods("GET")
	publicAPIRouter.HandleFunc("/auth/oidc/{provider}/redirect/{redirect_path:.*}", oidcRedirect).Methods("GET")

	internalAPI := publicAPIRouter.PathPrefix("/internal").Subrouter()
	internalAPI.HandleFunc("/runners", runners.RegisterRunner).Methods("POST")

	runnersAPI := internalAPI.PathPrefix("/runners").Subrouter()
	runnersAPI.Use(runners.RunnerMiddleware)
	runnersAPI.Path("").HandlerFunc(runnerController.GetRunner).Methods("GET", "HEAD")
	runnersAPI.Path("").HandlerFunc(runnerController.UpdateRunner).Methods("PUT")
	runnersAPI.Path("").HandlerFunc(runners.UnregisterRunner).Methods("DELETE")

	publicWebHookRouter := r.PathPrefix(webPath + "api").Subrouter()
	publicWebHookRouter.Use(StoreMiddleware, JSONMiddleware)
	publicWebHookRouter.Path("/integrations/{integration_alias}").HandlerFunc(
		integrationController.ReceiveIntegration).Methods("POST", "GET", "OPTIONS")

	terraformWebhookRouter := publicWebHookRouter.PathPrefix("/terraform").Subrouter()
	terraformWebhookRouter.Use(terraformController.TerraformInventoryAliasMiddleware)
	terraformWebhookRouter.Path("/{alias}").HandlerFunc(terraformController.GetTerraformState).Methods("GET")
	terraformWebhookRouter.Path("/{alias}").HandlerFunc(terraformController.AddTerraformState).Methods("POST")
	terraformWebhookRouter.Path("/{alias}").HandlerFunc(terraformController.LockTerraformState).Methods("LOCK")
	terraformWebhookRouter.Path("/{alias}").HandlerFunc(terraformController.UnlockTerraformState).Methods("UNLOCK")

	authenticatedWS := r.PathPrefix(webPath + "api").Subrouter()
	authenticatedWS.Use(JSONMiddleware, authenticationWithStore)
	authenticatedWS.Path("/ws").HandlerFunc(sockets.Handler).Methods("GET", "HEAD")

	authenticatedAPI := r.PathPrefix(webPath + "api").Subrouter()
	authenticatedAPI.Use(StoreMiddleware, JSONMiddleware, authentication)

	authenticatedAPI.Path("/info").HandlerFunc(getSystemInfo).Methods("GET", "HEAD")

	authenticatedAPI.Path("/subscription").HandlerFunc(subscriptionController.Activate).Methods("POST")
	authenticatedAPI.Path("/subscription").HandlerFunc(subscriptionController.GetSubscription).Methods("GET")

	authenticatedAPI.Path("/projects").HandlerFunc(projects.GetProjects).Methods("GET", "HEAD")
	authenticatedAPI.Path("/projects").HandlerFunc(projectsController.AddProject).Methods("POST")
	authenticatedAPI.Path("/projects/restore").HandlerFunc(projects.Restore).Methods("POST")
	authenticatedAPI.Path("/events").HandlerFunc(getAllEvents).Methods("GET", "HEAD")
	authenticatedAPI.HandleFunc("/events/last", getLastEvents).Methods("GET", "HEAD")

	authenticatedAPI.Path("/users").HandlerFunc(usersController.GetUsers).Methods("GET", "HEAD")
	authenticatedAPI.Path("/users").HandlerFunc(usersController.AddUser).Methods("POST")
	authenticatedAPI.Path("/user").HandlerFunc(userController.GetUser).Methods("GET", "HEAD")

	authenticatedAPI.Path("/apps").HandlerFunc(getApps).Methods("GET", "HEAD")

	tokenAPI := authenticatedAPI.PathPrefix("/user").Subrouter()
	tokenAPI.Path("/tokens").HandlerFunc(getAPITokens).Methods("GET", "HEAD")
	tokenAPI.Path("/tokens").HandlerFunc(createAPIToken).Methods("POST")
	tokenAPI.HandleFunc("/tokens/{token_id}", deleteAPIToken).Methods("DELETE")

	adminAPI := authenticatedAPI.NewRoute().Subrouter()
	adminAPI.Use(adminMiddleware)
	adminAPI.Path("/options").HandlerFunc(getOptions).Methods("GET", "HEAD")
	adminAPI.Path("/options").HandlerFunc(setOption).Methods("POST")

	adminAPI.Path("/runners").HandlerFunc(getAllRunners).Methods("GET", "HEAD")
	adminAPI.Path("/runners").HandlerFunc(addGlobalRunner).Methods("POST", "HEAD")

	adminAPI.Path("/roles").HandlerFunc(rolesController.GetRoles).Methods("GET", "HEAD")
	adminAPI.Path("/roles").HandlerFunc(rolesController.AddRole).Methods("POST", "HEAD")

	adminAPI.Path("/cache").HandlerFunc(clearCache).Methods("DELETE", "HEAD")

	debugAPI := adminAPI.PathPrefix("/debug").Subrouter()
	debugAPI.Path("/gc").HandlerFunc(debug.GC).Methods("POST")
	debugAPI.Path("/pprof/dump").HandlerFunc(debug.Dump).Methods("POST")

	globalRunnersAPI := adminAPI.PathPrefix("/runners").Subrouter()
	globalRunnersAPI.Use(globalRunnerMiddleware)
	globalRunnersAPI.Path("/{runner_id}").HandlerFunc(getGlobalRunner).Methods("GET", "HEAD")
	globalRunnersAPI.Path("/{runner_id}").HandlerFunc(updateGlobalRunner).Methods("PUT", "POST")
	globalRunnersAPI.Path("/{runner_id}/active").HandlerFunc(setGlobalRunnerActive).Methods("POST")
	globalRunnersAPI.Path("/{runner_id}").HandlerFunc(deleteGlobalRunner).Methods("DELETE")
	globalRunnersAPI.Path("/{runner_id}/cache").HandlerFunc(clearGlobalRunnerCache).Methods("DELETE")

	rolesAPI := adminAPI.PathPrefix("/roles").Subrouter()
	rolesAPI.Path("/{role_slug}").HandlerFunc(rolesController.GetGlobalRole).Methods("GET", "HEAD")
	rolesAPI.Path("/{role_slug}").HandlerFunc(rolesController.UpdateRole).Methods("PUT", "POST")
	rolesAPI.Path("/{role_slug}").HandlerFunc(rolesController.DeleteRole).Methods("DELETE")

	appsAPI := adminAPI.PathPrefix("/apps").Subrouter()
	appsAPI.Use(appMiddleware)
	appsAPI.Path("/{app_id}").HandlerFunc(getApp).Methods("GET", "HEAD")
	appsAPI.Path("/{app_id}").HandlerFunc(setApp).Methods("PUT", "POST")
	appsAPI.Path("/{app_id}/active").HandlerFunc(setAppActive).Methods("POST")
	appsAPI.Path("/{app_id}").HandlerFunc(deleteApp).Methods("DELETE")

	adminAPI.Path("/tasks").HandlerFunc(tasks.GetTasks).Methods("GET", "HEAD")
	tasksAPI := adminAPI.PathPrefix("/tasks").Subrouter()
	tasksAPI.Use(tasks.TaskMiddleware)
	tasksAPI.Path("/{task_id}").HandlerFunc(tasks.GetTasks).Methods("GET", "HEAD")
	tasksAPI.Path("/{task_id}").HandlerFunc(tasks.DeleteTask).Methods("DELETE")

	userUserAPI := authenticatedAPI.Path("/users/{user_id}").Subrouter()
	userUserAPI.Use(readonlyUserMiddleware)
	userUserAPI.Methods("GET", "HEAD").HandlerFunc(userController.GetUser)

	userAPI := authenticatedAPI.Path("/users/{user_id}").Subrouter()
	userAPI.Use(getUserMiddleware)

	userAPI.Methods("PUT").HandlerFunc(usersController.UpdateUser)
	userAPI.Methods("DELETE").HandlerFunc(deleteUser)

	userPasswordAPI := authenticatedAPI.PathPrefix("/users/{user_id}").Subrouter()
	userPasswordAPI.Use(getUserMiddleware)
	userPasswordAPI.Path("/password").HandlerFunc(updateUserPassword).Methods("POST")
	userPasswordAPI.Path("/2fas/totp").HandlerFunc(enableTotp).Methods("POST")
	userPasswordAPI.Path("/2fas/totp/{totp_id}/qr").HandlerFunc(totpQr).Methods("GET")
	userPasswordAPI.Path("/2fas/totp/{totp_id}").HandlerFunc(disableTotp).Methods("DELETE")

	projectGet := authenticatedAPI.Path("/project/{project_id}").Subrouter()
	projectGet.Use(projects.ProjectMiddleware)
	projectGet.Methods("GET", "HEAD").HandlerFunc(projects.GetProject)

	//
	// Start and Stop tasks
	projectTaskStart := authenticatedAPI.PathPrefix("/project/{project_id}").Subrouter()
	projectTaskStart.Use(projects.ProjectMiddleware, projects.NewTaskMiddleware, projects.GetTaskPermissionsMiddleware, projects.GetMustCanMiddleware(db.CanRunProjectTasks))
	projectTaskStart.Path("/tasks").HandlerFunc(projects.AddTask).Methods("POST")

	projectTaskStop := authenticatedAPI.PathPrefix("/project/{project_id}").Subrouter()
	projectTaskStop.Use(projects.ProjectMiddleware, projects.GetTaskMiddleware, projects.GetTaskPermissionsMiddleware, projects.GetMustCanMiddleware(db.CanRunProjectTasks))
	projectTaskStop.HandleFunc("/tasks/{task_id}/stop", projects.StopTask).Methods("POST")
	projectTaskStop.HandleFunc("/tasks/{task_id}/confirm", projects.ConfirmTask).Methods("POST")
	projectTaskStop.HandleFunc("/tasks/{task_id}/reject", projects.RejectTask).Methods("POST")

	//
	// Project resources CRUD
	projectUserAPI := authenticatedAPI.PathPrefix("/project/{project_id}").Subrouter()
	projectUserAPI.Use(projects.ProjectMiddleware, projects.GetMustCanMiddleware(db.CanManageProjectResources))

	projectUserAPI.Path("/role").HandlerFunc(projects.GetUserRole).Methods("GET", "HEAD")

	projectUserAPI.Path("/events").HandlerFunc(getAllEvents).Methods("GET", "HEAD")
	projectUserAPI.HandleFunc("/events/last", getLastEvents).Methods("GET", "HEAD")

	projectUserAPI.Path("/users").HandlerFunc(projects.GetUsers).Methods("GET", "HEAD")

	projectUserAPI.Path("/keys").HandlerFunc(projects.GetKeys).Methods("GET", "HEAD")
	projectUserAPI.Path("/keys").HandlerFunc(keyController.AddKey).Methods("POST")

	projectUserAPI.Path("/secret_storages").HandlerFunc(secretStorageController.GetSecretStorages).Methods("GET", "HEAD")
	projectUserAPI.Path("/secret_storages").HandlerFunc(secretStorageController.Add).Methods("POST")

	projectUserAPI.Path("/repositories").HandlerFunc(projects.GetRepositories).Methods("GET", "HEAD")
	projectUserAPI.Path("/repositories").HandlerFunc(projects.AddRepository).Methods("POST")

	projectUserAPI.Path("/inventory").HandlerFunc(projects.GetInventory).Methods("GET", "HEAD")
	projectUserAPI.Path("/inventory").HandlerFunc(projects.AddInventory).Methods("POST")

	projectUserAPI.Path("/environment").HandlerFunc(projects.GetEnvironment).Methods("GET", "HEAD")
	projectUserAPI.Path("/environment").HandlerFunc(environmentController.AddEnvironment).Methods("POST")

	projectUserAPI.Path("/tasks").HandlerFunc(projects.GetAllTasks).Methods("GET", "HEAD")
	projectUserAPI.HandleFunc("/tasks/last", projects.GetLastTasks).Methods("GET", "HEAD")

	projectUserAPI.Path("/stats").HandlerFunc(projects.GetTaskStats).Methods("GET", "HEAD")

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
	projectUserAPI.Path("/notifications/test").HandlerFunc(projectController.SendTestNotification).Methods("POST")

	projectUserAPI.Path("/runners").HandlerFunc(projectRunnerController.GetRunners).Methods("GET", "HEAD")
	projectUserAPI.Path("/runners").HandlerFunc(projectRunnerController.AddRunner).Methods("POST")
	projectUserAPI.Path("/runner_tags").HandlerFunc(projectRunnerController.GetRunnerTags).Methods("GET", "HEAD")

	projectRunnersAPI := projectUserAPI.PathPrefix("/runners").Subrouter()
	projectRunnersAPI.Use(projectRunnerController.RunnerMiddleware)
	projectRunnersAPI.Path("/{runner_id}").HandlerFunc(projectRunnerController.GetRunner).Methods("GET", "HEAD")
	projectRunnersAPI.Path("/{runner_id}").HandlerFunc(projectRunnerController.UpdateRunner).Methods("PUT", "POST")
	projectRunnersAPI.Path("/{runner_id}/active").HandlerFunc(projectRunnerController.SetRunnerActive).Methods("POST")
	projectRunnersAPI.Path("/{runner_id}").HandlerFunc(projectRunnerController.DeleteRunner).Methods("DELETE")
	projectRunnersAPI.Path("/{runner_id}/cache").HandlerFunc(projectRunnerController.ClearRunnerCache).Methods("DELETE")

	projectUserAPI.Path("/roles").HandlerFunc(rolesController.GetProjectRoles).Methods("GET", "HEAD")
	projectUserAPI.Path("/roles/all").HandlerFunc(rolesController.GetProjectAndGlobalRoles).Methods("GET", "HEAD")
	projectUserAPI.Path("/roles").HandlerFunc(rolesController.AddProjectRole).Methods("POST")

	projectRolesAPI := projectUserAPI.PathPrefix("/roles").Subrouter()
	projectRolesAPI.Path("/{role_slug}").HandlerFunc(rolesController.GetProjectRole).Methods("GET", "HEAD")
	projectRolesAPI.Path("/{role_slug}").HandlerFunc(rolesController.UpdateProjectRole).Methods("PUT", "POST")
	projectRolesAPI.Path("/{role_slug}").HandlerFunc(rolesController.DeleteProjectRole).Methods("DELETE")

	//
	// Updating and deleting project
	projectAdminAPI := authenticatedAPI.Path("/project/{project_id}").Subrouter()
	projectAdminAPI.Use(projects.ProjectMiddleware, projects.GetMustCanMiddleware(db.CanUpdateProject))
	projectAdminAPI.Methods("PUT").HandlerFunc(projectController.UpdateProject)
	projectAdminAPI.Methods("DELETE").HandlerFunc(projectController.DeleteProject)

	meAPI := authenticatedAPI.Path("/project/{project_id}/me").Subrouter()
	meAPI.Use(projects.ProjectMiddleware)
	meAPI.HandleFunc("", projects.LeftProject).Methods("DELETE")

	cacheAPI := authenticatedAPI.Path("/project/{project_id}/cache").Subrouter()
	cacheAPI.Use(projects.ProjectMiddleware)
	cacheAPI.HandleFunc("", projects.ClearCache).Methods("DELETE")

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
	// Manage project invites
	projectInvitesAPI := authenticatedAPI.PathPrefix("/project/{project_id}").Subrouter()
	projectInvitesAPI.Use(projects.ProjectMiddleware, projects.GetMustCanMiddleware(db.CanManageProjectUsers))
	projectInvitesAPI.Path("/invites").HandlerFunc(projects.GetInvites).Methods("GET", "HEAD")
	projectInvitesAPI.Path("/invites").HandlerFunc(projects.CreateInvite).Methods("POST")

	projectInviteManagement := projectInvitesAPI.PathPrefix("/invites").Subrouter()
	projectInviteManagement.Use(projects.InviteMiddleware)
	projectInviteManagement.HandleFunc("/{invite_id}", projects.GetInvites).Methods("GET", "HEAD")
	projectInviteManagement.HandleFunc("/{invite_id}", projects.UpdateInvite).Methods("PUT")
	projectInviteManagement.HandleFunc("/{invite_id}", projects.DeleteInvite).Methods("DELETE")

	// Accept invite endpoint (doesn't require project context)
	authenticatedAPI.Path("/invites/accept").HandlerFunc(projects.AcceptInvite).Methods("POST")

	//
	// Project resources CRUD (continue)
	projectKeyManagement := projectUserAPI.PathPrefix("/keys").Subrouter()
	projectKeyManagement.Use(projects.KeyMiddleware)

	projectKeyManagement.HandleFunc("/{key_id}", projects.GetKeys).Methods("GET", "HEAD")
	projectKeyManagement.HandleFunc("/{key_id}/refs", projects.GetKeyRefs).Methods("GET", "HEAD")
	projectKeyManagement.HandleFunc("/{key_id}", keyController.UpdateKey).Methods("PUT")
	projectKeyManagement.HandleFunc("/{key_id}", keyController.RemoveKey).Methods("DELETE")

	projectSecretStorageManagement := projectUserAPI.PathPrefix("/secret_storages").Subrouter()
	projectSecretStorageManagement.Use(projects.SecretStorageMiddleware)
	projectSecretStorageManagement.HandleFunc("/{storage_id}", secretStorageController.GetSecretStorage).Methods("GET", "HEAD")
	projectSecretStorageManagement.HandleFunc("/{storage_id}/refs", secretStorageController.GetRefs).Methods("GET", "HEAD")
	projectSecretStorageManagement.HandleFunc("/{storage_id}", secretStorageController.Update).Methods("PUT")
	projectSecretStorageManagement.HandleFunc("/{storage_id}", secretStorageController.Remove).Methods("DELETE")

	projectRepoManagement := projectUserAPI.PathPrefix("/repositories").Subrouter()
	projectRepoManagement.Use(projects.RepositoryMiddleware)

	projectRepoManagement.HandleFunc("/{repository_id}", projects.GetRepositories).Methods("GET", "HEAD")
	projectRepoManagement.HandleFunc("/{repository_id}/refs", projects.GetRepositoryRefs).Methods("GET", "HEAD")
	projectRepoManagement.HandleFunc("/{repository_id}", projects.UpdateRepository).Methods("PUT")
	projectRepoManagement.HandleFunc("/{repository_id}", projects.RemoveRepository).Methods("DELETE")
	projectRepoManagement.HandleFunc("/{repository_id}/branches", repositoryController.GetRepositoryBranches).Methods("GET", "HEAD")

	projectInventoryManagement := projectUserAPI.PathPrefix("/inventory").Subrouter()
	projectInventoryManagement.Use(projects.InventoryMiddleware)

	projectInventoryManagement.HandleFunc("/{inventory_id}", projects.GetInventory).Methods("GET", "HEAD")
	projectInventoryManagement.HandleFunc("/{inventory_id}/refs", projects.GetInventoryRefs).Methods("GET", "HEAD")
	projectInventoryManagement.HandleFunc("/{inventory_id}", projects.UpdateInventory).Methods("PUT")
	projectInventoryManagement.HandleFunc("/{inventory_id}", projects.RemoveInventory).Methods("DELETE")

	projectInventoryManagement.HandleFunc("/{inventory_id}/terraform/aliases", terraformInventoryController.GetTerraformInventoryAliases).Methods("GET", "HEAD")
	projectInventoryManagement.HandleFunc("/{inventory_id}/terraform/aliases", terraformInventoryController.AddTerraformInventoryAlias).Methods("POST")
	projectInventoryManagement.HandleFunc("/{inventory_id}/terraform/aliases/{alias_id}", terraformInventoryController.GetTerraformInventoryAlias).Methods("GET")
	projectInventoryManagement.HandleFunc("/{inventory_id}/terraform/aliases/{alias_id}", terraformInventoryController.DeleteTerraformInventoryAlias).Methods("DELETE")
	projectInventoryManagement.HandleFunc("/{inventory_id}/terraform/aliases/{alias_id}", terraformInventoryController.SetTerraformInventoryAliasAccessKey).Methods("PUT")

	projectInventoryManagement.HandleFunc("/{inventory_id}/terraform/states", terraformInventoryController.GetTerraformInventoryStates).Methods("GET", "HEAD")
	projectInventoryManagement.HandleFunc("/{inventory_id}/terraform/states/latest", terraformInventoryController.GetTerraformInventoryLatestState).Methods("GET", "HEAD")
	projectInventoryManagement.HandleFunc("/{inventory_id}/terraform/states/{state_id}", terraformInventoryController.GetTerraformInventoryState).Methods("GET")
	projectInventoryManagement.HandleFunc("/{inventory_id}/terraform/states/{state_id}", terraformInventoryController.DeleteTerraformInventoryState).Methods("DELETE")

	projectEnvManagement := projectUserAPI.PathPrefix("/environment").Subrouter()
	projectEnvManagement.Use(environmentController.EnvironmentMiddleware)

	projectEnvManagement.HandleFunc("/{environment_id}", projects.GetEnvironment).Methods("GET", "HEAD")
	projectEnvManagement.HandleFunc("/{environment_id}/refs", projects.GetEnvironmentRefs).Methods("GET", "HEAD")
	projectEnvManagement.HandleFunc("/{environment_id}", environmentController.UpdateEnvironment).Methods("PUT")
	projectEnvManagement.HandleFunc("/{environment_id}", environmentController.RemoveEnvironment).Methods("DELETE")

	projectTmplManagement := projectUserAPI.PathPrefix("/templates").Subrouter()
	projectTmplManagement.Use(projects.TemplatesMiddleware)

	projectTmplManagement.HandleFunc("/{template_id}", projects.UpdateTemplate).Methods("PUT")
	projectTmplManagement.HandleFunc("/{template_id}/description", projects.UpdateTemplateDescription).Methods("PUT")
	projectTmplManagement.HandleFunc("/{template_id}", projects.RemoveTemplate).Methods("DELETE")
	projectTmplManagement.HandleFunc("/{template_id}", projects.GetTemplate).Methods("GET")
	projectTmplManagement.HandleFunc("/{template_id}/refs", projects.GetTemplateRefs).Methods("GET", "HEAD")
	projectTmplManagement.HandleFunc("/{template_id}/tasks", projects.GetAllTasks).Methods("GET")
	projectTmplManagement.HandleFunc("/{template_id}/tasks/last", projects.GetLastTasks).Methods("GET")
	projectTmplManagement.HandleFunc("/{template_id}/schedules", projects.GetTemplateSchedules).Methods("GET")
	projectTmplManagement.HandleFunc("/{template_id}/stats", projects.GetTaskStats).Methods("GET")
	projectTmplManagement.HandleFunc("/{template_id}/stop_all_tasks", taskController.StopAllTasks).Methods("POST")

	projectTmplManagement.HandleFunc("/{template_id}/perms", templateController.GetTemplatePerms).Methods("GET")
	projectTmplManagement.HandleFunc("/{template_id}/perms", templateController.AddTemplatePerm).Methods("POST")
	projectTmplManagement.HandleFunc("/{template_id}/perms/{perm_id}", templateController.GetTemplatePerm).Methods("GET")
	projectTmplManagement.HandleFunc("/{template_id}/perms/{perm_id}", templateController.UpdateTemplatePerm).Methods("PUT")
	projectTmplManagement.HandleFunc("/{template_id}/perms/{perm_id}", templateController.DeleteTemplatePerm).Methods("DELETE")

	projectTmplInvManagement := projectTmplManagement.PathPrefix("/{template_id}/inventory").Subrouter()
	projectTmplInvManagement.Use(projects.InventoryMiddleware)
	projectTmplInvManagement.HandleFunc("/{inventory_id}/set_default", projects.SetTemplateInventory).Methods("POST")
	projectTmplInvManagement.HandleFunc("/{inventory_id}/attach", projects.AttachInventory).Methods("POST")
	projectTmplInvManagement.HandleFunc("/{inventory_id}/detach", projects.DetachInventory).Methods("POST")

	projectTaskManagement := projectUserAPI.PathPrefix("/tasks").Subrouter()
	projectTaskManagement.Use(projects.GetTaskMiddleware)

	projectTaskManagement.HandleFunc("/{task_id}/output", projects.GetTaskOutput).Methods("GET", "HEAD")
	projectTaskManagement.HandleFunc("/{task_id}/raw_output", projects.GetTaskRawOutput).Methods("GET", "HEAD")
	projectTaskManagement.HandleFunc("/{task_id}", projects.GetTask).Methods("GET", "HEAD")
	projectTaskManagement.HandleFunc("/{task_id}", projects.RemoveTask).Methods("DELETE")
	projectTaskManagement.HandleFunc("/{task_id}/stages", projects.GetTaskStages).Methods("GET", "HEAD")
	projectTaskManagement.HandleFunc("/{task_id}/ansible/hosts", taskController.GetAnsibleTaskHosts).Methods("GET", "HEAD")
	projectTaskManagement.HandleFunc("/{task_id}/ansible/errors", taskController.GetAnsibleTaskErrors).Methods("GET", "HEAD")

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

	// Check if this is a request for the swagger UI
	swaggerPath := path.Join(webPath, "swagger")
	if reqPath == swaggerPath || reqPath == swaggerPath+"/" {
		serveFile(w, r, "swagger/index.html")
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
	user := helpers.GetFromContext(r, "user").(*db.User)

	var authMethods LoginAuthMethods

	if util.Config.Auth.Totp.Enabled {
		authMethods.Totp = &LoginTotpAuthMethod{
			AllowRecovery: util.Config.Auth.Totp.AllowRecovery,
		}
	}

	if util.Config.Auth.Email.Enabled {
		authMethods.Email = &LoginEmailAuthMethod{}
	}

	timezone := util.Config.Schedule.Timezone

	if timezone == "" {
		timezone = "UTC"
	}

	roles, err := helpers.Store(r).GetGlobalRoles()
	if err != nil {
		log.WithError(err).Error("Failed to get roles")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	body := map[string]any{
		"version":           util.Version(),
		"ansible":           util.AnsibleVersion(),
		"web_host":          util.Config.WebHost,
		"use_remote_runner": util.Config.UseRemoteRunner,
		"auth_methods":      authMethods,
		"premium_features":  proFeatures.GetFeatures(user),
		"git_client":        util.Config.GitClientId,
		"schedule_timezone": timezone,
		"teams":             util.Config.Teams,
		"roles":             roles,
	}

	helpers.WriteJSON(w, http.StatusOK, body)
}
