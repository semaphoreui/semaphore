package routes

import (
	"strings"

	"github.com/ansible-semaphore/semaphore/routes/projects"
	"github.com/ansible-semaphore/semaphore/routes/sockets"
	"github.com/ansible-semaphore/semaphore/routes/tasks"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gin-gonic/gin"
)

// Declare all routes
func Route(r *gin.Engine) {
	r.GET("/api/ping", func(c *gin.Context) {
		c.String(200, "PONG")
	})

	r.NoRoute(servePublic)

	// set up the namespace
	api := r.Group("/api")

	api.Use(authentication)

	func(api *gin.RouterGroup) {
		api.POST("/login", login)
		api.POST("/logout", logout)
	}(api.Group("/auth"))

	api.Use(MustAuthenticate)

	api.GET("/ws", sockets.Handler)

	func(api *gin.RouterGroup) {
		api.GET("", getUser)
		// api.PUT("/user", misc.UpdateUser)

		api.GET("/tokens", getAPITokens)
		api.POST("/tokens", createAPIToken)
		api.DELETE("/tokens/:token_id", expireAPIToken)
	}(api.Group("/user"))

	api.GET("/projects", projects.GetProjects)
	api.POST("/projects", projects.AddProject)

	func(api *gin.RouterGroup) {
		api.Use(projects.ProjectMiddleware)

		api.GET("", projects.GetProject)

		api.GET("/users", projects.GetUsers)
		api.POST("/users", projects.AddUser)
		api.POST("/users/:user_id/admin", projects.UserMiddleware, projects.MakeUserAdmin)
		api.DELETE("/users/:user_id/admin", projects.UserMiddleware, projects.MakeUserAdmin)
		api.DELETE("/users/:user_id", projects.UserMiddleware, projects.RemoveUser)

		api.GET("/keys", projects.GetKeys)
		api.POST("/keys", projects.AddKey)
		api.PUT("/keys/:key_id", projects.KeyMiddleware, projects.UpdateKey)
		api.DELETE("/keys/:key_id", projects.KeyMiddleware, projects.RemoveKey)

		api.GET("/repositories", projects.GetRepositories)
		api.POST("/repositories", projects.AddRepository)
		api.DELETE("/repositories/:repository_id", projects.RepositoryMiddleware, projects.RemoveRepository)

		api.GET("/inventory", projects.GetInventory)
		api.POST("/inventory", projects.AddInventory)
		api.PUT("/inventory/:inventory_id", projects.InventoryMiddleware, projects.UpdateInventory)
		api.DELETE("/inventory/:inventory_id", projects.InventoryMiddleware, projects.RemoveInventory)

		api.GET("/environment", projects.GetEnvironment)
		api.POST("/environment", projects.AddEnvironment)
		api.DELETE("/environment/:environment_id", projects.EnvironmentMiddleware, projects.RemoveEnvironment)

		api.GET("/templates", projects.GetTemplates)
		api.POST("/templates", projects.AddTemplate)
		api.PUT("/templates/:template_id", projects.TemplatesMiddleware, projects.UpdateTemplate)
		api.DELETE("/templates/:template_id", projects.TemplatesMiddleware, projects.RemoveTemplate)

		api.POST("/tasks", tasks.AddTask)
	}(api.Group("/project/:project_id"))
}

func servePublic(c *gin.Context) {
	path := c.Request.URL.Path

	if strings.HasPrefix(path, "/api") {
		c.Next()
		return
	}

	if !strings.HasPrefix(path, "/public") {
		if len(strings.Split(path, ".")) > 1 {
			c.AbortWithStatus(404)
			return
		}

		path = "/public/html/index.html"
	}

	path = strings.Replace(path, "/", "", 1)
	split := strings.Split(path, ".")
	suffix := split[len(split)-1]

	res, err := util.Asset(path)
	if err != nil {
		c.Next()
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

	c.Writer.Header().Set("content-type", contentType)
	c.String(200, string(res))
}
