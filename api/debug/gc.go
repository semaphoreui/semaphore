package debug

import (
	"net/http"
	"runtime"

	"github.com/gorilla/context"
)

func GC(w http.ResponseWriter, r *http.Request) {
	context.Purge(600)
	runtime.GC()
	w.WriteHeader(http.StatusNoContent)
}
