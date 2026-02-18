package server

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pro_interfaces"
)

func NewSubscriptionService(userRepo db.UserManager, optionsRepo db.OptionsManager, runnerRepo db.RunnerManager, tfRepo db.TerraformStore) pro_interfaces.SubscriptionService {
	return &SubscriptionServiceImpl{}
}

type SubscriptionServiceImpl struct {
}

func (s *SubscriptionServiceImpl) GetToken() (res pro_interfaces.SubscriptionToken, err error) {
	err = db.ErrNotFound
	return
}

func (s *SubscriptionServiceImpl) HasActiveSubscription() bool {
	return false
}

func (s *SubscriptionServiceImpl) CanAddProUser() (ok bool, err error) {
	return false, nil
}

func (s *SubscriptionServiceImpl) StartValidationCron() {

}

func (s *SubscriptionServiceImpl) CanAddRole() (ok bool, err error) {
	return
}

func (s *SubscriptionServiceImpl) CanAddRunner() (ok bool, err error) {
	return
}

func (s *SubscriptionServiceImpl) CanAddTerraformHTTPBackend() (ok bool, err error) {
	return
}
