package api

import (
	"net/http"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pro_interfaces"
)

func NewSubscriptionController(
	optionsRepo db.OptionsManager,
	userRepo db.UserManager,
	runnerRepo db.RunnerManager,
	tfRepo db.TerraformStore,
) pro_interfaces.SubscriptionController {
	return &subscriptionControllerImpl{}
}

type subscriptionControllerImpl struct {
}

func (ctrl *subscriptionControllerImpl) Delete(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (ctrl *subscriptionControllerImpl) Activate(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (ctrl *subscriptionControllerImpl) GetSubscription(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (ctrl *subscriptionControllerImpl) Refresh(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}
