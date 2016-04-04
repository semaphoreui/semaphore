package projects

import "github.com/gin-gonic/gin"

func EnvironmentMiddleware(c *gin.Context) {
	c.AbortWithStatus(501)
}

func GetEnvironment(c *gin.Context) {
	c.AbortWithStatus(501)
}

func AddEnvironment(c *gin.Context) {
	c.AbortWithStatus(501)
}

func RemoveEnvironment(c *gin.Context) {
	c.AbortWithStatus(501)
}
