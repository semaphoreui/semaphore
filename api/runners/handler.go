package runners

import (
	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gorilla/mux"
	"net/http"
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

	pingRouter := r.Path(webPath + "api/runners/register").Subrouter()

	pingRouter.Methods("POST", "HEAD").HandlerFunc(registerRunner)

	return r
}

func registerRunner(w http.ResponseWriter, r *http.Request) {
	var register struct {
		RegistrationToken string `json:"registration_token" binding:"required"`
	}

	if !helpers.Bind(w, r, &register) {
		return
	}

	if register.RegistrationToken != util.Config.RegistrationToken {
		return
	}

	// TODO: add runner to database
}
