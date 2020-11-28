package util

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

func isXHR(w http.ResponseWriter, r *http.Request) bool {
	accept := r.Header.Get("Accept")
	return !strings.Contains(accept, "text/html")
}

// AuthFailed write a status unauthorized header unless it is an XHR request
// TODO - never called!
func AuthFailed(w http.ResponseWriter, r *http.Request) {
	if !isXHR(w, r) {
		http.Redirect(w, r, "/?hai", http.StatusFound)
		return
	}

	w.WriteHeader(http.StatusUnauthorized)
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

// ScanErrorChecker deals with errors encountered while scanning lines
// since we do not fail on these errors currently we can simply note them
// and move on
func ScanErrorChecker(n int, err error) {
	if err != nil {
		log.Warn("An input error occurred:" + err.Error())
	}
}

//H just a string-to-anything map
type H map[string]interface{}

//Bind decodes json into object
func Bind(w http.ResponseWriter, r *http.Request, out interface{}) error {
	err := json.NewDecoder(r.Body).Decode(out)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	return err
}

//WriteJSON writes object as JSON
func WriteJSON(w http.ResponseWriter, code int, out interface{}) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(out); err != nil {
		panic(err)
	}
}
