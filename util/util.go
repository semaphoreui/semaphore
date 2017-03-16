package util

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

func isXHR(w http.ResponseWriter, r *http.Request) bool {
	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "text/html") {
		return false
	}

	return true
}

func AuthFailed(w http.ResponseWriter, r *http.Request) {
	if isXHR(w, r) == false {
		http.Redirect(w, r, "/?hai", http.StatusFound)
		return
	}

	w.WriteHeader(http.StatusUnauthorized)
	return
}

func GetIntParam(name string, w http.ResponseWriter, r *http.Request) (int, error) {
	intParam, err := strconv.Atoi(mux.Vars(r)[name])

	if err != nil {
		if isXHR(w, r) == false {
			http.Redirect(w, r, "/404", http.StatusFound)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}

		return 0, err
	}

	return intParam, nil
}
