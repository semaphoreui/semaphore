package api

import (
	"errors"
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	proFeatures "github.com/semaphoreui/semaphore/pro/pkg/features"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
)

type SystemInfoController struct {
	subscriptionService pro_interfaces.SubscriptionService
}

func NewSystemInfoController(subscriptionService pro_interfaces.SubscriptionService) *SystemInfoController {
	return &SystemInfoController{
		subscriptionService,
	}
}

func (c *SystemInfoController) GetSystemInfo(w http.ResponseWriter, r *http.Request) {
	user := helpers.GetFromContext(r, "user").(*db.User)

	var authMethods LoginAuthMethods

	if util.Config.Auth.Totp.Enabled {
		authMethods.Totp = &LoginTotpAuthMethod{
			AllowRecovery: util.Config.Auth.Totp.AllowRecovery,
		}
	}

	if util.Config.Auth.Email.Enabled {
		authMethods.Email = &LoginEmailAuthMethod{}
	}

	timezone := util.Config.Schedule.Timezone

	if timezone == "" {
		timezone = "UTC"
	}

	roles, err := helpers.Store(r).GetGlobalRoles()
	if err != nil {
		log.WithFields(log.Fields{
			"context": "system_info",
			"user_id": user.ID,
		}).WithError(err).Error("Failed to get roles")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var plan string

	token, err := c.subscriptionService.GetToken()

	switch {
	case errors.Is(err, db.ErrNotFound):
		err = nil
		plan = ""
	case err != nil:
		log.WithFields(log.Fields{
			"context": "system_info",
			"user_id": user.ID,
		}).WithError(err).Error("Failed to get subscription plan")
		err = nil
		plan = ""
	default:
		if token.State == "expired" {
			plan = ""
		} else {
			plan = token.Plan
		}
	}

	body := map[string]any{
		"version":             util.Version(),
		"ansible":             util.AnsibleVersion(),
		"web_host":            util.Config.WebHost,
		"use_remote_runner":   util.Config.UseRemoteRunner,
		"auth_methods":        authMethods,
		"login_with_password": !util.Config.PasswordLoginDisable,
		"premium_features":    proFeatures.GetFeatures(user, plan),
		"subscription_state":  token.State,
		"git_client":          util.Config.GitClientId,
		"schedule_timezone":   timezone,
		"teams":               util.Config.Teams,
		"roles":               roles,
		"boltdb_used":         util.Config.Dialect == "bolt",
	}

	helpers.WriteJSON(w, http.StatusOK, body)
}
