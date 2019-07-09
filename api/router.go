package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ansible-semaphore/semaphore/api/projects"
	"github.com/ansible-semaphore/semaphore/api/sockets"
	"github.com/ansible-semaphore/semaphore/api/tasks"
	"github.com/ansible-semaphore/semaphore/mulekick"
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
		if next != nil {
			next.ServeHTTP(w, r)
		}
	})
}

//PlainTextMiddleware resets headers to Plain Text if needed
func PlainTextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "text/plain; charset=utf-8")
		if next != nil {
			next.ServeHTTP(w, r)
		}
	})
}

func printRegisteredRoutes(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
	pathTemplate, err := route.GetPathTemplate()
	if err == nil && len(pathTemplate) > 0 {
		fmt.Println("ROUTE:", pathTemplate)
	}
	pathRegexp, err := route.GetPathRegexp()
	if err == nil && len(pathRegexp) > 0 {
		fmt.Println("Path regexp:", pathRegexp)
	}
	queriesTemplates, err := route.GetQueriesTemplates()
	if err == nil && len(queriesTemplates) > 0 {
		fmt.Println("Queries templates:", strings.Join(queriesTemplates, ","))
	}
	queriesRegexps, err := route.GetQueriesRegexp()
	if err == nil && len(queriesRegexps) > 0 {
		fmt.Println("Queries regexps:", strings.Join(queriesRegexps, ","))
	}
	methods, err := route.GetMethods()
	if err == nil && len(methods) > 0 {
		fmt.Println("Methods:", strings.Join(methods, ","))
	}
	fmt.Println()
	return nil
}

// Route declares all routes
func Route() mulekick.Router {
	r := mulekick.New(mux.NewRouter())

	r.Use(mux.CORSMethodMiddleware(r.Router))

	webPath := "/"
	if util.WebHostURL != nil {
		webPath = util.WebHostURL.RequestURI()
	}

	r.NotFoundHandler = servePublic(nil)
	r.Handle(webPath, servePublic(nil))

	r.Use(JSONMiddleware)

	r.Get(webPath+"api/ping", PlainTextMiddleware, mulekick.PongHandler)

	// set up the namespace
	api := mulekick.New(r.Path(webPath + "api").Subrouter())
	api.Post("/login", login)
	api.Post("/logout", logout)

	api.Use(authentication)

	api.Get("/ws", sockets.Handler)

	api.Get("/info", getSystemInfo)
	api.Get("/upgrade", checkUpgrade)
	api.Post("/upgrade", doUpgrade)

	api.Get("", getUser)
	// api.PUT("/user", misc.UpdateUser)

	api.Get("/tokens", getAPITokens)
	api.Post("/tokens", createAPIToken)
	api.Delete("/tokens/{token_id}", expireAPIToken)

	api.Get("/projects", projects.GetProjects)
	api.Post("/projects", projects.AddProject)
	api.Get("/events", getAllEvents)
	api.Get("/events/last", getLastEvents)

	api.Get("/users", getUsers)
	api.Post("/users", addUser)
	api.Get("/users/{user_id}", getUserMiddleware, getUser)
	api.Put("/users/{user_id}", getUserMiddleware, updateUser)
	api.Post("/users/{user_id}/password", getUserMiddleware, updateUserPassword)
	api.Delete("/users/{user_id}", getUserMiddleware, deleteUser)

	project := mulekick.New(api.Path("/project/{project_id}").Subrouter())

	project.Use(projects.ProjectMiddleware)

	project.Get("", projects.GetProject)
	project.Put("", projects.MustBeAdmin, projects.UpdateProject)
	project.Delete("", projects.MustBeAdmin, projects.DeleteProject)

	project.Get("/events", getAllEvents)
	project.Get("/events/last", getLastEvents)

	project.Get("/users", projects.GetUsers)
	project.Post("/users", projects.MustBeAdmin, projects.AddUser)
	project.Post("/users/{user_id}/admin", projects.MustBeAdmin, projects.UserMiddleware, projects.MakeUserAdmin)
	project.Delete("/users/{user_id}/admin", projects.MustBeAdmin, projects.UserMiddleware, projects.MakeUserAdmin)
	project.Delete("/users/{user_id}", projects.MustBeAdmin, projects.UserMiddleware, projects.RemoveUser)

	project.Get("/keys", projects.GetKeys)
	project.Post("/keys", projects.AddKey)
	project.Put("/keys/{key_id}", projects.KeyMiddleware, projects.UpdateKey)
	project.Delete("/keys/{key_id}", projects.KeyMiddleware, projects.RemoveKey)

	project.Get("/repositories", projects.GetRepositories)
	project.Post("/repositories", projects.AddRepository)
	project.Put("/repositories/{repository_id}", projects.RepositoryMiddleware, projects.UpdateRepository)
	project.Delete("/repositories/{repository_id}", projects.RepositoryMiddleware, projects.RemoveRepository)

	project.Get("/inventory", projects.GetInventory)
	project.Post("/inventory", projects.AddInventory)
	project.Put("/inventory/{inventory_id}", projects.InventoryMiddleware, projects.UpdateInventory)
	project.Delete("/inventory/{inventory_id}", projects.InventoryMiddleware, projects.RemoveInventory)

	project.Get("/environment", projects.GetEnvironment)
	project.Post("/environment", projects.AddEnvironment)
	project.Put("/environment/{environment_id}", projects.EnvironmentMiddleware, projects.UpdateEnvironment)
	project.Delete("/environment/{environment_id}", projects.EnvironmentMiddleware, projects.RemoveEnvironment)

	project.Get("/templates", projects.GetTemplates)
	project.Post("/templates", projects.AddTemplate)
	project.Put("/templates/{template_id}", projects.TemplatesMiddleware, projects.UpdateTemplate)
	project.Delete("/templates/{template_id}", projects.TemplatesMiddleware, projects.RemoveTemplate)

	project.Get("/tasks", tasks.GetAllTasks)
	project.Get("/tasks/last", tasks.GetLastTasks)
	project.Post("/tasks", tasks.AddTask)
	project.Get("/tasks/{task_id}/output", tasks.GetTaskMiddleware, tasks.GetTaskOutput)
	project.Get("/tasks/{task_id}", tasks.GetTaskMiddleware, tasks.GetTask)
	project.Delete("/tasks/{task_id}", tasks.GetTaskMiddleware, tasks.RemoveTask)
	return r
}

//nolint: gocyclo
func servePublic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if strings.HasPrefix(path, "/api") {
			mulekick.NotFoundHandler(next).ServeHTTP(w, r)
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
			mulekick.NotFoundHandler(next).ServeHTTP(w, r)
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
	})
}

func getSystemInfo(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		if next != nil {
			next.ServeHTTP(w, r)
		}
	})
}

func checkUpgrade(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := util.CheckUpdate(util.Version); err != nil {
			mulekick.WriteJSON(w, 500, err)
			return
		}

		if util.UpdateAvailable != nil {
			getSystemInfo(next).ServeHTTP(w, r)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		if next != nil {
			next.ServeHTTP(w, r)
		}
	})
}

func doUpgrade(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		util.LogError(util.DoUpgrade(util.Version))

		if next != nil {
			next.ServeHTTP(w, r)
		}
	})
}
