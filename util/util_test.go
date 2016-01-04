package util

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
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
	_, err := GetIntParam("test_id", c)
	if err != nil {
		return
	}

	c.AbortWithStatus(200)
}
