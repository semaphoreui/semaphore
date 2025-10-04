package helpers

import (
	"encoding/json"
	"errors"
	"net/http"
	"runtime/debug"

	"github.com/Digital-Data-Co/forge/db"
	"github.com/Digital-Data-Co/forge/pkg/common_errors"
	log "github.com/sirupsen/logrus"
)

// WriteJSON writes object as JSON
func WriteJSON(w http.ResponseWriter, code int, out any) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(out); err != nil {
		log.Error(err)
		debug.PrintStack()
	}
}

func WriteErrorStatus(w http.ResponseWriter, err string, code int) {
	WriteJSON(w, code, map[string]string{
		"error": err,
	})
}

func WriteError(w http.ResponseWriter, err error) {
	if errors.Is(err, common_errors.ErrInvalidSubscription) {
		WriteErrorStatus(w, "You have no subscription.", http.StatusForbidden)
		return
	}

	if errors.Is(err, db.ErrNotFound) {
		WriteErrorStatus(w, "Resource not found", http.StatusNotFound)
		return
	}

	if errors.Is(err, db.ErrInvalidOperation) {
		WriteErrorStatus(w, "Invalid operation", http.StatusConflict)
		return
	}

	var validationError *db.ValidationError
	switch {
	case errors.As(err, &validationError):
		WriteErrorStatus(w, validationError.Error(), http.StatusBadRequest)
	default:
		log.Error(err)
		debug.PrintStack()
		WriteErrorStatus(w, "Internal server error", http.StatusInternalServerError)
	}
}
