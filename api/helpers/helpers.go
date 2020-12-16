package helpers

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/gorilla/context"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

func Store(r *http.Request) db.Store {
	return context.Get(r, "store").(db.Store)
}

func isXHR(w http.ResponseWriter, r *http.Request) bool {
	accept := r.Header.Get("Accept")
	return !strings.Contains(accept, "text/html")
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

//H just a string-to-anything map
type H map[string]interface{}

//Bind decodes json into object
func Bind(w http.ResponseWriter, r *http.Request, out interface{}) bool {
	err := json.NewDecoder(r.Body).Decode(out)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	return err == nil
}

//WriteJSON writes object as JSON
func WriteJSON(w http.ResponseWriter, code int, out interface{}) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(out); err != nil {
		panic(err)
	}
}


func WriteError(w http.ResponseWriter, err error) {
	if err == db.ErrNotFound {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	log.Error(err)
	debug.PrintStack()
	w.WriteHeader(http.StatusInternalServerError)
}
