package debug

import (
	"net/http"
	"runtime"
)

func GC(w http.ResponseWriter, r *http.Request) {
	runtime.GC()
	w.WriteHeader(http.StatusNoContent)
}
