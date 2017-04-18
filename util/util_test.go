package util

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetIntParam(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/test/:test_id", mockParam)
	req, _ := http.NewRequest("GET", "/test/123", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Response code should be 200 %d", w.Code)
	}
}

func mockParam(c *gin.Context) {
	_, err := GetIntParam("test_id", c.Writer, c.Request)
	if err != nil {
		return
	}

	c.AbortWithStatus(200)
}
