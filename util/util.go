package util

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

func isXHR(c *gin.Context) bool {
	accept := c.Request.Header.Get("Accept")
	if strings.Contains(accept, "text/html") {
		return false
	}

	return true
}

func AuthFailed(c *gin.Context) {
	if isXHR(c) == false {
		c.Redirect(302, "/?hai")
	} else {
		c.Writer.WriteHeader(401)
	}

	c.Abort()

	return
}

func GetIntParam(name string, c *gin.Context) (int, error) {
	intParam, err := strconv.Atoi(c.Params.ByName(name))
	if err != nil {
		if isXHR(c) == false {
			c.Redirect(302, "/404")
		} else {
			c.AbortWithStatus(400)
		}

		return 0, err
	}

	return intParam, nil
}
