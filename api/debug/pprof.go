package debug

import (
	"net/http"
	"os"
	"path"
	"runtime/pprof"
	"strconv"
	"time"

	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
)

func Dump(w http.ResponseWriter, r *http.Request) {
	if util.Config.Debugging.PprofDumpDir == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	f, err := os.Create(path.Join(util.Config.Debugging.PprofDumpDir, "mem-"+strconv.Itoa(int(time.Now().Unix()))+".prof"))

	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"context": "pprof",
		}).Error("error creating mem.prof")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer f.Close()

	err = pprof.WriteHeapProfile(f)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"context": "pprof",
		}).Error("Failed to write memory profile")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
