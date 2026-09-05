package helpers

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// GetStrParamOrAbort fetches a parameter from the route variables as a string.
// On failure it writes the response itself (redirect to /404 or 400 status)
// and returns false — the caller must only return.
func GetStrParamOrAbort(name string, w http.ResponseWriter, r *http.Request) (string, bool) {
	strParam, ok := mux.Vars(r)[name]

	if !ok {
		log.WithFields(log.Fields{
			"context": "route_params",
			"param":   name,
			"method":  r.Method,
			"url":     r.URL.String(),
		}).Debug("Route param is missing")

		if !isXHR(w, r) {
			http.Redirect(w, r, "/404", http.StatusFound)
		} else {
			WriteErrorStatus(w, "Bad request", http.StatusBadRequest)
		}

		return "", false
	}

	return strParam, true
}

func HasParam(name string, r *http.Request) bool {
	_, ok := mux.Vars(r)[name]
	return ok
}

func GetIntParamR(name string, r *http.Request) (int, error) {
	intParam, err := strconv.Atoi(mux.Vars(r)[name])

	if err != nil {
		return 0, err
	}

	return intParam, nil
}

// GetIntParamOrAbort fetches a parameter from the route variables as an integer.
// On failure it writes the response itself (redirect to /404 or 400 status)
// and returns false — the caller must only return.
func GetIntParamOrAbort(name string, w http.ResponseWriter, r *http.Request) (int, bool) {
	intParam, err := GetIntParamR(name, r)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"context":   "route_params",
			"param":     name,
			"raw_value": mux.Vars(r)[name],
			"method":    r.Method,
			"url":       r.URL.String(),
		}).Debug("Route param is not a valid integer")

		if !isXHR(w, r) {
			http.Redirect(w, r, "/404", http.StatusFound)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}

		return 0, false
	}

	return intParam, true
}
