//go:build !pro

package api

import (
	"net/http"
)

func verifySessionByEmail(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	return
}

func startEmailVerification(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	return
}
