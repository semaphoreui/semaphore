package pro_interfaces

import "time"

type SubscriptionToken struct {
	Company   string    `json:"company,omitempty"`
	State     string    `json:"state"`
	Key       string    `json:"key"`
	Plan      string    `json:"plan"`
	Users     int       `json:"users"`
	ExpiresAt time.Time `json:"expiresAt"`
	Nodes     int       `json:"nodes,omitempty"`
	UIs       int       `json:"uis,omitempty"`
}

func (t *SubscriptionToken) Validate() error {
	return nil
}

type SubscriptionService interface {
	HasActiveSubscription() bool
	CanAddProUser() (ok bool, err error)
	CanAddRunner() (ok bool, err error)
	CanAddTerraformHTTPBackend() (ok bool, err error)
	StartValidationCron()
	GetToken() (res SubscriptionToken, err error)
}
