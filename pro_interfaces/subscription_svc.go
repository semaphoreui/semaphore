package pro_interfaces

type SubscriptionService interface {
	HasActiveSubscription() bool
	CanAddProUser() (ok bool, err error)
	CanAddRunner() (ok bool, err error)
	CanAddTerraformHTTPBackend() (ok bool, err error)
	StartValidationCron()
}
