package metrics

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetrics_ServeHTTP(t *testing.T) {
	m := NewMetrics()

	req := httptest.NewRequest("GET", "/api/metrics", nil)
	w := httptest.NewRecorder()

	m.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "go_goroutines")
}
