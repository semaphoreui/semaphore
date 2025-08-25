package pro_interfaces

type SubscriptionService interface {
	HasActiveSubscription() bool
	CanAddProUser() (ok bool, err error)
	StartValidationCron()
}
