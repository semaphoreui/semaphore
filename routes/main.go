package routes

import (
	"strings"

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

	// serve /api/auth

	api.Use(MustAuthenticate)

	// api.GET("/user", misc.GetUser)
	// api.PUT("/user", misc.UpdateUser)
}

func servePublic(c *gin.Context) {
	path := c.Request.URL.Path

	if !strings.HasPrefix(path, "/public") {
		c.Next()
		return
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
