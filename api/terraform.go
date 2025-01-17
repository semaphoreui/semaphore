//go:build !pro

package api

import "net/http"

func TerraformInventoryAliasMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func getTerraformState(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func addTerraformState(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func lockTerraformState(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func unlockTerraformState(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}
