package health

import (
	"context"
	"net/http"
	"time"

	"github.com/semaphoreui/semaphore/pro_interfaces"
)

const probeTimeout = 2 * time.Second

// DatabasePinger is the readiness check against the primary store.
// db.Store satisfies this interface.
type DatabasePinger interface {
	Ping(ctx context.Context) error
}

// Server serves Kubernetes /livez and /readyz on a dedicated listener.
type Server struct {
	db    DatabasePinger
	redis pro_interfaces.HARedis
}

// New returns a health Server. redis may be nil outside HA / Pro builds.
func New(db DatabasePinger, redis pro_interfaces.HARedis) *Server {
	return &Server{db: db, redis: redis}
}

// Handler returns the mux that serves /livez and /readyz.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/livez", s.livez)
	mux.HandleFunc("/readyz", s.readyz)
	return mux
}

func (s *Server) livez(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) readyz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	ctx, cancel := context.WithTimeout(r.Context(), probeTimeout)
	defer cancel()

	if err := s.db.Ping(ctx); err != nil {
		http.Error(w, "database unavailable", http.StatusServiceUnavailable)
		return
	}

	if s.redis != nil {
		if err := s.redis.Ping(ctx); err != nil {
			http.Error(w, "redis unavailable", http.StatusServiceUnavailable)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
