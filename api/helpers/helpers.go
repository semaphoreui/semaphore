package helpers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/semaphoreui/semaphore/db"
)

func Store(r *http.Request) db.Store {
	return GetFromContext(r, "store").(db.Store)
}

func isXHR(w http.ResponseWriter, r *http.Request) bool {
	accept := r.Header.Get("Accept")
	return !strings.Contains(accept, "text/html")
}

// H just a string-to-anything map
type H map[string]any

// Bind decodes json into object
func Bind(w http.ResponseWriter, r *http.Request, out any) bool {
	err := json.NewDecoder(r.Body).Decode(out)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	return err == nil
}
