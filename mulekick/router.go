package mulekick

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Router struct {
	*mux.Router
}

func (r Router) Get(endpoint string, hf http.HandlerFunc) *mux.Route {
	return r.HandleFunc(endpoint, hf).Methods("GET", "HEAD")
}

func (r Router) GetHandle(endpoint string, hf http.Handler) *mux.Route {
	return r.Handle(endpoint, hf).Methods("GET", "HEAD")
}

func (r Router) Post(endpoint string, hf http.HandlerFunc) *mux.Route {
	return r.HandleFunc(endpoint, hf).Methods("POST")
}

func (r Router) PostHandle(endpoint string, hf http.Handler) *mux.Route {
	return r.Handle(endpoint, hf).Methods("POST")
}

func (r Router) Put(endpoint string, hf http.HandlerFunc) *mux.Route {
	return r.HandleFunc(endpoint, hf).Methods("PUT")
}

func (r Router) PutHandle(endpoint string, hf http.Handler) *mux.Route {
	return r.Handle(endpoint, hf).Methods("PUT")
}

func (r Router) Delete(endpoint string, hf http.HandlerFunc) *mux.Route {
	return r.HandleFunc(endpoint, hf).Methods("DELETE")
}

func (r Router) DeleteHandle(endpoint string, hf http.Handler) *mux.Route {
	return r.Handle(endpoint, hf).Methods("DELETE")
}

func (r Router) Patch(endpoint string, hf http.HandlerFunc) *mux.Route {
	return r.HandleFunc(endpoint, hf).Methods("PATCH")
}

func (r Router) PatchHandle(endpoint string, hf http.Handler) *mux.Route {
	return r.Handle(endpoint, hf).Methods("PATCH")
}

func (r Router) Options(endpoint string, hf http.HandlerFunc) *mux.Route {
	return r.HandleFunc(endpoint, hf).Methods("OPTIONS")
}

func (r Router) OptionsHandle(endpoint string, hf http.Handler) *mux.Route {
	return r.Handle(endpoint, hf).Methods("OPTIONS")
}

// Group creates a new sub-router, enabling you to group handlers
func (r Router) Group(str string, middleware ...mux.MiddlewareFunc) Router {
	subR := r.PathPrefix(str).Subrouter()
	subR.Use(middleware...)

	return Router{subR}
}
