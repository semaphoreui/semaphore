package mulekick

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type H map[string]interface{}

func New(r *mux.Router, middleware ...http.HandlerFunc) Router {
	return Router{r}
}

func Bind(w http.ResponseWriter, r *http.Request, out interface{}) error {
	err := json.NewDecoder(r.Body).Decode(out)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	return err
}

func WriteJSON(w http.ResponseWriter, code int, out interface{}) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(out); err != nil {
		panic(err)
	}
}
