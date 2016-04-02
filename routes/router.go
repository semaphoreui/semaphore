package routes

import (
	"strings"

	"github.com/ansible-semaphore/semaphore/routes/auth"
	"github.com/ansible-semaphore/semaphore/routes/projects"
	"github.com/ansible-semaphore/semaphore/routes/sockets"
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
		api.POST("/login", auth.Login)
		api.POST("/logout", auth.Logout)
	}(api.Group("/auth"))

	api.Use(MustAuthenticate)

	api.GET("/ws", sockets.Handler)

	api.GET("/user", getUser)
	// api.PUT("/user", misc.UpdateUser)

	api.GET("/projects", projects.GetProjects)
	api.POST("/projects", projects.AddProject)

	func(api *gin.RouterGroup) {
		api.Use(projects.ProjectMiddleware)

		api.GET("", projects.GetProject)

		api.GET("/users", projects.GetProjectUsers)
		api.POST("/users", projects.AddProjectUser)
		api.DELETE("/users/:user_id", projects.RemoveProjectUser)

		api.GET("/keys", projects.GetProjectKeys)
		api.POST("/keys", projects.AddProjectKey)
		api.DELETE("/keys", projects.RemoveProjectKey)

		api.GET("/repositories", projects.GetProjectRepositories)
		api.POST("/repositories", projects.AddProjectRepository)
		api.DELETE("/repositories/:user_id", projects.RemoveProjectRepository)

		api.GET("/inventory", projects.GetProjectInventories)
		api.POST("/inventory", projects.AddProjectInventory)
		api.DELETE("/inventory/:user_id", projects.RemoveProjectInventory)

		api.GET("/environment", projects.GetProjectEnvironment)
		api.POST("/environment", projects.AddProjectEnvironment)
		api.DELETE("/environment/:user_id", projects.RemoveProjectEnvironment)

		api.GET("/templates", projects.GetProjectUsers)
		api.POST("/templates", projects.AddProjectUser)
		api.DELETE("/templates/:user_id", projects.RemoveProjectUser)
	}(api.Group("/project/:project_id"))
}

func servePublic(c *gin.Context) {
	path := c.Request.URL.Path

	if strings.HasPrefix(path, "/api") {
		c.Next()
		return
	}

	if !strings.HasPrefix(path, "/public") {
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

func getUser(c *gin.Context) {
	c.JSON(200, c.MustGet("user"))
}
