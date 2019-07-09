package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ansible-semaphore/semaphore/api/projects"
	"github.com/ansible-semaphore/semaphore/api/sockets"
	"github.com/ansible-semaphore/semaphore/api/tasks"

	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gobuffalo/packr"
	"github.com/gorilla/mux"
	"github.com/russross/blackfriday"
)

var publicAssets = packr.NewBox("../web/public")

//JSONMiddleware ensures that all the routes respond with Json, this is added by default to all routes
func JSONMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		next.ServeHTTP(w, r)
	})
}

//plainTextMiddleware resets headers to Plain Text if needed
func plainTextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "text/plain; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}

func pongHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("404 not found"))
	fmt.Println(r.Method, ":", r.URL.String(), "--> 404 Not Found")
}

// Route declares all routes
func Route() *mux.Router {
	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(servePublic)

	webPath := "/"
	if util.WebHostURL != nil {
		webPath = util.WebHostURL.RequestURI()
	}

	r.HandleFunc(webPath, http.HandlerFunc(servePublic))
	r.Use(mux.CORSMethodMiddleware(r), JSONMiddleware)

	r.HandleFunc("/api/auth/login", login).Methods("POST")
	r.HandleFunc("/api/auth/logout", logout).Methods("POST")
	r.HandleFunc("/api/ping", pongHandler).Methods("GET", "HEAD").Subrouter().Use(plainTextMiddleware)

	// set up the namespace
	api := r.PathPrefix(webPath + "api").Subrouter()

	api.Use(authentication)

	api.HandleFunc("/ws", sockets.Handler).Methods("GET", "HEAD")
	api.HandleFunc("/info", getSystemInfo).Methods("GET", "HEAD")
	api.HandleFunc("/upgrade", checkUpgrade).Methods("GET", "HEAD")
	api.HandleFunc("/upgrade", doUpgrade).Methods("POST")

	user := api.PathPrefix("/user").Subrouter()

	user.HandleFunc("", getUser).Methods("GET", "HEAD")
	user.HandleFunc("/tokens", getAPITokens).Methods("GET", "HEAD")
	user.HandleFunc("/tokens", createAPIToken).Methods("POST")
	user.HandleFunc("/tokens/{token_id}", expireAPIToken).Methods("DELETE")

	api.HandleFunc("/projects", projects.GetProjects).Methods("GET", "HEAD")
	api.HandleFunc("/projects", projects.AddProject).Methods("POST")
	api.HandleFunc("/events", getAllEvents).Methods("GET", "HEAD")
	api.HandleFunc("/events/last", getLastEvents).Methods("GET", "HEAD")

	api.HandleFunc("/users", getUsers).Methods("GET", "HEAD")
	api.HandleFunc("/users", addUser).Methods("POST")
	api.HandleFunc("/users/{user_id}", getUser).Methods("GET", "HEAD").Subrouter().Use(getUserMiddleware)
	api.HandleFunc("/users/{user_id}", updateUser).Methods("PUT").Subrouter().Use(getUserMiddleware)
	api.HandleFunc("/users/{user_id}/password", updateUserPassword).Methods("POST").Subrouter().Use(getUserMiddleware)
	api.HandleFunc("/users/{user_id}", deleteUser).Methods("DELETE").Subrouter().Use(getUserMiddleware)

	project := api.PathPrefix("/project/{project_id}").Subrouter()

	project.Use(projects.ProjectMiddleware)

	project.HandleFunc("", projects.GetProject).Methods("GET", "HEAD")
	project.HandleFunc("", projects.UpdateProject).Methods("PUT").Subrouter().Use(projects.MustBeAdmin)
	project.HandleFunc("", projects.DeleteProject).Methods("DELETE").Subrouter().Use(projects.MustBeAdmin)

	project.HandleFunc("/events", getAllEvents).Methods("GET", "HEAD")
	project.HandleFunc("/events/last", getLastEvents).Methods("GET", "HEAD")

	project.HandleFunc("/users", projects.GetUsers).Methods("GET", "HEAD")
	project.HandleFunc("/users", projects.AddUser).Methods("POST").Subrouter().Use(projects.MustBeAdmin)
	project.HandleFunc("/users/{user_id}/admin", projects.MakeUserAdmin).Methods("POST").Subrouter().Use(projects.UserMiddleware, projects.MustBeAdmin)
	project.HandleFunc("/users/{user_id}/admin", projects.MakeUserAdmin).Methods("DELETE").Subrouter().Use(projects.UserMiddleware, projects.MustBeAdmin)
	project.HandleFunc("/users/{user_id}", projects.RemoveUser).Methods("DELETE").Subrouter().Use(projects.UserMiddleware, projects.MustBeAdmin)

	project.HandleFunc("/keys", projects.GetKeys).Methods("GET", "HEAD")
	project.HandleFunc("/keys", projects.AddKey).Methods("POST")
	project.HandleFunc("/keys/{key_id}", projects.UpdateKey).Methods("PUT").Subrouter().Use(projects.KeyMiddleware)
	project.HandleFunc("/keys/{key_id}", projects.RemoveKey).Methods("DELETE").Subrouter().Use(projects.KeyMiddleware)

	project.HandleFunc("/repositories", projects.GetRepositories).Methods("GET", "HEAD")
	project.HandleFunc("/repositories", projects.AddRepository).Methods("POST")
	project.HandleFunc("/repositories/{repository_id}", projects.UpdateRepository).Methods("PUT").Subrouter().Use(projects.RepositoryMiddleware)
	project.HandleFunc("/repositories/{repository_id}", projects.RemoveRepository).Methods("DELETE").Subrouter().Use(projects.RepositoryMiddleware)

	project.HandleFunc("/inventory", projects.GetInventory).Methods("GET", "HEAD")
	project.HandleFunc("/inventory", projects.AddInventory).Methods("POST")
	project.HandleFunc("/inventory/{inventory_id}", projects.UpdateInventory).Methods("PUT").Subrouter().Use(projects.InventoryMiddleware)
	project.HandleFunc("/inventory/{inventory_id}", projects.RemoveInventory).Methods("DELETE").Subrouter().Use(projects.InventoryMiddleware)

	project.HandleFunc("/environment", projects.GetEnvironment).Methods("GET", "HEAD")
	project.HandleFunc("/environment", projects.AddEnvironment).Methods("POST")
	project.HandleFunc("/environment/{environment_id}", projects.UpdateEnvironment).Methods("PUT").Subrouter().Use(projects.EnvironmentMiddleware)
	project.HandleFunc("/environment/{environment_id}", projects.RemoveEnvironment).Methods("DELETE").Subrouter().Use(projects.EnvironmentMiddleware)

	project.HandleFunc("/templates", projects.GetTemplates).Methods("GET", "HEAD")
	project.HandleFunc("/templates", projects.AddTemplate).Methods("POST")
	project.HandleFunc("/templates/{template_id}", projects.UpdateTemplate).Methods("PUT").Subrouter().Use(projects.TemplatesMiddleware)
	project.HandleFunc("/templates/{template_id}", projects.RemoveTemplate).Methods("DELETE").Subrouter().Use(projects.TemplatesMiddleware)

	project.HandleFunc("/tasks", tasks.GetAllTasks).Methods("GET", "HEAD")
	project.HandleFunc("/tasks/last", tasks.GetLastTasks).Methods("GET", "HEAD")
	project.HandleFunc("/tasks", tasks.AddTask).Methods("POST")
	project.HandleFunc("/tasks/{task_id}/output", tasks.GetTaskOutput).Methods("GET", "HEAD").Subrouter().Use(tasks.GetTaskMiddleware)
	project.HandleFunc("/tasks/{task_id}", tasks.GetTask).Methods("GET", "HEAD").Subrouter().Use(tasks.GetTaskMiddleware)
	project.HandleFunc("/tasks/{task_id}", tasks.RemoveTask).Methods("DELETE").Subrouter().Use(tasks.GetTaskMiddleware)

	return r
}

//nolint: gocyclo
func servePublic(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if strings.HasPrefix(path, "/api") {
		notFoundHandler(w, r)
		return
	}

	webPath := "/"
	if util.WebHostURL != nil {
		webPath = util.WebHostURL.RequestURI()
	}

	if !strings.HasPrefix(path, webPath+"public") {
		if len(strings.Split(path, ".")) > 1 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		path = "/html/index.html"
	}

	path = strings.Replace(path, webPath+"public/", "", 1)
	split := strings.Split(path, ".")
	suffix := split[len(split)-1]

	res, err := publicAssets.MustBytes(path)
	if err != nil {
		notFoundHandler(w, r)
		return
	}

	// replace base path
	if util.WebHostURL != nil && path == "/html/index.html" {
		res = []byte(strings.Replace(string(res),
			"<base href=\"/\">",
			"<base href=\""+util.WebHostURL.String()+"\">",
			1))
	}

	contentType := "text/plain"
	switch suffix {
	case "png":
		contentType = "image/png"
	case "jpg", "jpeg":
		contentType = "image/jpeg"
	case "gif":
		contentType = "image/gif"
	case "js":
		contentType = "application/javascript"
	case "css":
		contentType = "text/css"
	case "woff":
		contentType = "application/x-font-woff"
	case "ttf":
		contentType = "application/x-font-ttf"
	case "otf":
		contentType = "application/x-font-otf"
	case "html":
		contentType = "text/html"
	}

	w.Header().Set("content-type", contentType)
	_, err = w.Write(res)
	util.LogWarning(err)
}

func getSystemInfo(w http.ResponseWriter, r *http.Request) {
	body := map[string]interface{}{
		"version": util.Version,
		"update":  util.UpdateAvailable,
		"config": map[string]string{
			"dbHost":  util.Config.MySQL.Hostname,
			"dbName":  util.Config.MySQL.DbName,
			"dbUser":  util.Config.MySQL.Username,
			"path":    util.Config.TmpPath,
			"cmdPath": util.FindSemaphore(),
		},
	}

	if util.UpdateAvailable != nil {
		body["updateBody"] = string(blackfriday.MarkdownCommon([]byte(*util.UpdateAvailable.Body)))
	}

	util.WriteJSON(w, http.StatusOK, body)
}

func checkUpgrade(w http.ResponseWriter, r *http.Request) {
	if err := util.CheckUpdate(util.Version); err != nil {
		util.WriteJSON(w, 500, err)
		return
	}

	if util.UpdateAvailable != nil {
		getSystemInfo(w, r)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func doUpgrade(w http.ResponseWriter, r *http.Request) {
	util.LogError(util.DoUpgrade(util.Version))
}
