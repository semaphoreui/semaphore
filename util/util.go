package util

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

func isXHR(w http.ResponseWriter, r *http.Request) bool {
	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "text/html") {
		return false
	}

	return true
}

func AuthFailed(w http.ResponseWriter, r *http.Request) {
	if isXHR(c) == false {
		c.Redirect(302, "/?hai")
	} else {
		c.Writer.WriteHeader(401)
	}

	c.Abort()

	return
}

func GetIntParam(name string, w http.ResponseWriter, r *http.Request) (int, error) {
	intParam, err := strconv.Atoi(c.Params.ByName(name))
	if err != nil {
		if isXHR(c) == false {
			c.Redirect(302, "/404")
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}

		return 0, err
	}

	return intParam, nil
}
