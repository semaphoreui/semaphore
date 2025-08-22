package api

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	"net/http"
)

func NewSubscriptionController(
	optionsRepo db.OptionsManager,
	userRepo db.UserManager,
) pro_interfaces.SubscriptionController {
	return &subscriptionControllerImpl{}
}

type subscriptionControllerImpl struct {
}

func (ctrl *subscriptionControllerImpl) Activate(w http.ResponseWriter, r *http.Request) {}

func (ctrl *subscriptionControllerImpl) GetSubscription(w http.ResponseWriter, r *http.Request) {}
