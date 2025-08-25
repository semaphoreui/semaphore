package pro_interfaces

import "net/http"

type SubscriptionController interface {
	GetSubscription(w http.ResponseWriter, r *http.Request)
	Activate(w http.ResponseWriter, r *http.Request)
}
