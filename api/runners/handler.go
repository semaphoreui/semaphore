package runners

import (
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gorilla/mux"
	"strings"
)

func RunnerRoute() *mux.Router {
	r := mux.NewRouter()

	webPath := "/"
	if util.WebHostURL != nil {
		webPath = util.WebHostURL.Path
		if !strings.HasSuffix(webPath, "/") {
			webPath += "/"
		}
	}

	pingRouter := r.Path(webPath + "api/runner/register").Subrouter()

	pingRouter.Methods("GET", "HEAD").HandlerFunc(nil)

	return r
}
