package api

import (
	"net/http"
	"strings"

	"github.com/ansible-semaphore/semaphore/api/projects"
	"github.com/ansible-semaphore/semaphore/api/sockets"
	"github.com/ansible-semaphore/semaphore/api/tasks"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/castawaylabs/mulekick"
	"github.com/gorilla/mux"
	"github.com/russross/blackfriday"
)

// Declare all routes
func Route() mulekick.Router {
	r := mulekick.New(mux.NewRouter(), mulekick.CorsMiddleware)
	r.NotFoundHandler = http.HandlerFunc(servePublic)

	r.Get("/api/ping", mulekick.PongHandler)

	// set up the namespace
	api := r.Group("/api")

	func(api mulekick.Router) {
		api.Post("/login", login)
		api.Post("/logout", logout)
	}(api.Group("/auth"))

	api.Use(authentication)

	api.Get("/ws", sockets.Handler)

	api.Get("/info", getSystemInfo)
	api.Get("/upgrade", checkUpgrade)
	api.Post("/upgrade", doUpgrade)

	func(api mulekick.Router) {
		api.Get("", getUser)
		// api.PUT("/user", misc.UpdateUser)

		api.Get("/tokens", getAPITokens)
		api.Post("/tokens", createAPIToken)
		api.Delete("/tokens/:token_id", expireAPIToken)
	}(api.Group("/user"))

	api.Get("/projects", projects.GetProjects)
	api.Post("/projects", projects.AddProject)
	api.Get("/events", getEvents)

	api.Get("/users", getUsers)
	api.Post("/users", addUser)
	api.Get("/users/{user_id}", getUserMiddleware, getUser)
	api.Put("/users/{user_id}", getUserMiddleware, updateUser)
	api.Post("/users/{user_id}/password", getUserMiddleware, updateUserPassword)
	api.Delete("/users/{user_id}", getUserMiddleware, deleteUser)

	func(api mulekick.Router) {
		api.Use(projects.ProjectMiddleware)

		api.Get("", projects.GetProject)
		api.Put("", projects.MustBeAdmin, projects.UpdateProject)
		api.Delete("", projects.MustBeAdmin, projects.DeleteProject)

		api.Get("/events", getEvents)

		api.Get("/users", projects.GetUsers)
		api.Post("/users", projects.MustBeAdmin, projects.AddUser)
		api.Post("/users/{user_id}/admin", projects.MustBeAdmin, projects.UserMiddleware, projects.MakeUserAdmin)
		api.Delete("/users/{user_id}/admin", projects.MustBeAdmin, projects.UserMiddleware, projects.MakeUserAdmin)
		api.Delete("/users/{user_id}", projects.MustBeAdmin, projects.UserMiddleware, projects.RemoveUser)

		api.Get("/keys", projects.GetKeys)
		api.Post("/keys", projects.AddKey)
		api.Put("/keys/{key_id}", projects.KeyMiddleware, projects.UpdateKey)
		api.Delete("/keys/{key_id}", projects.KeyMiddleware, projects.RemoveKey)

		api.Get("/repositories", projects.GetRepositories)
		api.Post("/repositories", projects.AddRepository)
		api.Put("/repositories/{repository_id}", projects.RepositoryMiddleware, projects.UpdateRepository)
		api.Delete("/repositories/{repository_id}", projects.RepositoryMiddleware, projects.RemoveRepository)

		api.Get("/inventory", projects.GetInventory)
		api.Post("/inventory", projects.AddInventory)
		api.Put("/inventory/{inventory_id}", projects.InventoryMiddleware, projects.UpdateInventory)
		api.Delete("/inventory/{inventory_id}", projects.InventoryMiddleware, projects.RemoveInventory)

		api.Get("/environment", projects.GetEnvironment)
		api.Post("/environment", projects.AddEnvironment)
		api.Put("/environment/{environment_id}", projects.EnvironmentMiddleware, projects.UpdateEnvironment)
		api.Delete("/environment/{environment_id}", projects.EnvironmentMiddleware, projects.RemoveEnvironment)

		api.Get("/templates", projects.GetTemplates)
		api.Post("/templates", projects.AddTemplate)
		api.Put("/templates/{template_id}", projects.TemplatesMiddleware, projects.UpdateTemplate)
		api.Delete("/templates/{template_id}", projects.TemplatesMiddleware, projects.RemoveTemplate)

		api.Get("/tasks", tasks.GetAll)
		api.Post("/tasks", tasks.AddTask)
		api.Get("/tasks/{task_id}/output", tasks.GetTaskMiddleware, tasks.GetTaskOutput)
		api.Get("/tasks/{task_id}", tasks.GetTaskMiddleware, tasks.GetTask)
		api.Delete("/tasks/{task_id}", tasks.GetTaskMiddleware, tasks.RemoveTask)
	}(api.Group("/project/{project_id}"))

	return r
}

func servePublic(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if strings.HasPrefix(path, "/api") {
		mulekick.NotFoundHandler(w, r)
		return
	}

	if !strings.HasPrefix(path, "/public") {
		if len(strings.Split(path, ".")) > 1 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		path = "/public/html/index.html"
	}

	path = strings.Replace(path, "/", "", 1)
	split := strings.Split(path, ".")
	suffix := split[len(split)-1]

	res, err := util.Asset(path)
	if err != nil {
		mulekick.NotFoundHandler(w, r)
		return
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
	w.Write(res)
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

	mulekick.WriteJSON(w, http.StatusOK, body)
}

func checkUpgrade(w http.ResponseWriter, r *http.Request) {
	if err := util.CheckUpdate(util.Version); err != nil {
		mulekick.WriteJSON(w, 500, err)
		return
	}

	if util.UpdateAvailable != nil {
		getSystemInfo(w, r)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func doUpgrade(w http.ResponseWriter, r *http.Request) {
	util.DoUpgrade(util.Version)
}
