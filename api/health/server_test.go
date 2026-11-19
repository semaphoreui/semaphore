package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeDB struct {
	err error
}

func (f *fakeDB) Ping(ctx context.Context) error {
	return f.err
}

type fakeRedis struct {
	err error
}

func (f *fakeRedis) Ping(ctx context.Context) error { return f.err }
func (f *fakeRedis) Close() error                   { return nil }
func (f *fakeRedis) Describe() string               { return "fake-redis" }

func TestLivez_AlwaysOK(t *testing.T) {
	srv := New(&fakeDB{}, nil)
	req := httptest.NewRequest(http.MethodGet, "/livez", nil)
	w := httptest.NewRecorder()

	srv.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ok", w.Body.String())
	assert.Contains(t, w.Header().Get("Content-Type"), "text/plain")
}

func TestReadyz_OK(t *testing.T) {
	srv := New(&fakeDB{}, nil)
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()

	srv.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ok", w.Body.String())
}

func TestReadyz_DatabaseUnavailable(t *testing.T) {
	srv := New(&fakeDB{err: errors.New("down")}, nil)
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()

	srv.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "database unavailable")
}

func TestReadyz_RedisUnavailable(t *testing.T) {
	srv := New(&fakeDB{}, &fakeRedis{err: errors.New("redis down")})
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()

	srv.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "redis unavailable")
}

func TestReadyz_NilRedisSkipsRedisCheck(t *testing.T) {
	srv := New(&fakeDB{}, nil)
	require.Nil(t, srv.redis)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()

	srv.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReadyz_HEAD(t *testing.T) {
	srv := New(&fakeDB{}, nil)
	req := httptest.NewRequest(http.MethodHead, "/readyz", nil)
	w := httptest.NewRecorder()

	srv.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
