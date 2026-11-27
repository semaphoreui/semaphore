package projects

import (
	"github.com/semaphoreui/semaphore/db"
	proFeatures "github.com/semaphoreui/semaphore/pro/pkg/features"
	"github.com/semaphoreui/semaphore/pro_interfaces"
)

// proFeatureGate answers which Pro features a user has under the current license. The plan comes
// from the subscription service on every call, so a license activated or expired at runtime takes
// effect without a restart and on every node alike. getFeatures is the build's feature table
// (pro/pkg/features); tests substitute their own.
type proFeatureGate struct {
	subscription pro_interfaces.SubscriptionService
	getFeatures  func(user *db.User, plan string) pro_interfaces.Features
}

func newProFeatureGate(subscription pro_interfaces.SubscriptionService) proFeatureGate {
	return proFeatureGate{subscription: subscription, getFeatures: proFeatures.GetFeatures}
}

// features returns the user's features; the error is a subscription store failure, not a
// missing license — no license is a plan of "" with every feature off.
func (g proFeatureGate) features(user *db.User) (pro_interfaces.Features, error) {
	plan, err := proFeatures.CurrentPlan(g.subscription)
	if err != nil {
		return pro_interfaces.Features{}, err
	}
	return g.getFeatures(user, plan), nil
}
