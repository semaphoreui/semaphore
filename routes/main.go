package routes

import (
	"github.com/castawaylabs/semaphore"
	"github.com/gin-gonic/gin"
)

// Declare all routes
func Route(r *gin.Engine) {
	r.GET("/api/ping", func(c *gin.Context) {
		c.String(200, "PONG")
	})

	// serve public/ folder
	r.Group("/public", servePublic, serve404)

	// set up the namespace
	api := r.Group("/api")

	api.Use(authentication)

	// serve /api/auth

	api.Use(MustAuthenticate)

	// api.GET("/user", misc.GetUser)
	// api.PUT("/user", misc.UpdateUser)
}

func servePublic(c *gin.Context) {
	// util.asset url
	// util.Asset()
}

func serve404(c *gin.Context) {
	c.AbortWithStatus(404)
}
