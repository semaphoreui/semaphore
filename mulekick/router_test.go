package mulekick

import "net/http"

func ExampleRouter_Use() {
	r := &Router{}

	r.Get("/hello", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	r.Use(func(next http.Handler) http.Handler {
		// sample middleware
		return next
	})
	r.Post("/world", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	// /hello call will not be affected by middleware
	// /world will have the middleware in its stack
}
