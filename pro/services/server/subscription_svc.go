package server

import (
	"github.com/Digital-Data-Co/forge/db"
	"github.com/Digital-Data-Co/forge/pro_interfaces"
)

func NewSubscriptionService(userRepo db.UserManager, optionsRepo db.OptionsManager) pro_interfaces.SubscriptionService {
	return &SubscriptionServiceImpl{}
}

type SubscriptionServiceImpl struct {
}

func (s *SubscriptionServiceImpl) HasActiveSubscription() bool {
	return false
}

func (s *SubscriptionServiceImpl) CanAddProUser() (ok bool, err error) {
	return false, nil
}

func (s *SubscriptionServiceImpl) StartValidationCron() {

}
