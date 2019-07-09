package mulekick

import (
	"net/http"
)

func PongHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
		next.ServeHTTP(w, r)
	})
}

func NotFoundHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			CorsMiddleware(next).ServeHTTP(w, r)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 not found"))
		next.ServeHTTP(w, r)
	})
}

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method != "OPTIONS" {
			return
		}

		w.Header().Set("Access-Control-Max-Age", "1728000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		w.WriteHeader(http.StatusNoContent)
		next.ServeHTTP(w, r)
	})
}
