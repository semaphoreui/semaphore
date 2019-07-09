package mulekick

import (
	"github.com/gorilla/mux"
)

type Router struct {
	*mux.Router
}

func (r Router) makeRouteWithMiddleware(endpoint string, middleware []mux.MiddlewareFunc) *mux.Route {
	route := r.NewRoute()
	route.Path(endpoint)
	route.Subrouter().Use(middleware...)
	return route
}

func (r Router) Get(endpoint string, middleware ...mux.MiddlewareFunc) *mux.Route {
	return r.makeRouteWithMiddleware(endpoint, middleware).Methods("GET", "HEAD")
}

func (r Router) Post(endpoint string, middleware ...mux.MiddlewareFunc) *mux.Route {
	return r.makeRouteWithMiddleware(endpoint, middleware).Methods("POST")
}

func (r Router) Put(endpoint string, middleware ...mux.MiddlewareFunc) *mux.Route {
	return r.makeRouteWithMiddleware(endpoint, middleware).Methods("PUT")
}

func (r Router) Delete(endpoint string, middleware ...mux.MiddlewareFunc) *mux.Route {
	return r.makeRouteWithMiddleware(endpoint, middleware).Methods("DELETE")
}

func (r Router) Patch(endpoint string, middleware ...mux.MiddlewareFunc) *mux.Route {
	return r.makeRouteWithMiddleware(endpoint, middleware).Methods("PATCH")
}

func (r Router) Options(endpoint string, middleware ...mux.MiddlewareFunc) *mux.Route {
	return r.makeRouteWithMiddleware(endpoint, middleware).Methods("OPTIONS")
}

// Group creates a new sub-router, enabling you to group handlers
func (r Router) Group(str string, middleware ...mux.MiddlewareFunc) Router {
	subR := r.PathPrefix(str).Subrouter()
	for _, md := range middleware {
		subR.Use(md)
	}

	return Router{subR}
}
