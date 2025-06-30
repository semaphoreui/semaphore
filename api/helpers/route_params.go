package helpers

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

// GetStrParam fetches a parameter from the route variables as an integer
// redirects to a 404 or writes bad request state depending on error state
func GetStrParam(name string, w http.ResponseWriter, r *http.Request) (string, error) {
	strParam, ok := mux.Vars(r)[name]

	if !ok {
		if !isXHR(w, r) {
			http.Redirect(w, r, "/404", http.StatusFound)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}

		return "", fmt.Errorf("parameter missed")
	}

	return strParam, nil
}

// GetIntParam fetches a parameter from the route variables as an integer
// redirects to a 404 or writes bad request state depending on error state
func GetIntParam(name string, w http.ResponseWriter, r *http.Request) (int, error) {
	intParam, err := strconv.Atoi(mux.Vars(r)[name])

	if err != nil {
		if !isXHR(w, r) {
			http.Redirect(w, r, "/404", http.StatusFound)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}

		return 0, err
	}

	return intParam, nil
}
