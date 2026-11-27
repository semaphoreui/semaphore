package features

import (
	"errors"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pro_interfaces"
)

// GetFeatures is the open-source stub: no Pro feature is available in this build.
func GetFeatures(user *db.User, plan string) pro_interfaces.Features {
	return pro_interfaces.Features{}
}

// CurrentPlan is the plan of an active license, or "" when there is none or it is in any other
// state. The open-source subscription service has no license, so the answer is "" unless the
// service fails; community code calls it the same way as the Pro implementation does.
func CurrentPlan(subscription pro_interfaces.SubscriptionService) (string, error) {
	token, err := subscription.GetToken()
	switch {
	case errors.Is(err, db.ErrNotFound):
		return "", nil
	case err != nil:
		return "", err
	case token.State != "active":
		return "", nil
	}
	return token.Plan, nil
}
