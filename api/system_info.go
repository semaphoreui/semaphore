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

type SystemInfo struct {
	Version           string                  `json:"version"`
	Ansible           string                  `json:"ansible"`
	WebHost           string                  `json:"web_host"`
	UseRemoteRunner   bool                    `json:"use_remote_runner"`
	AuthMethods       LoginAuthMethods        `json:"auth_methods"`
	LoginWithPassword bool                    `json:"login_with_password"`
	Features          pro_interfaces.Features `json:"features"`
	SubscriptionState string                  `json:"subscription_state"`
	GitClient         string                  `json:"git_client"`
	ScheduleTimezone  string                  `json:"schedule_timezone"`
	Teams             *util.TeamsConfig       `json:"teams"`
	Roles             []db.Role               `json:"roles"`
	BoltdbUsed        bool                    `json:"boltdb_used"`
}

func NewSystemInfoController(subscriptionService pro_interfaces.SubscriptionService) *SystemInfoController {
	return &SystemInfoController{
		subscriptionService,
	}
}

func (c *SystemInfoController) GetSystemInfo(w http.ResponseWriter, r *http.Request) {
	user := helpers.GetFromContext(r, "user").(*db.User)

	var authMethods LoginAuthMethods

	if util.Config.Mfa.Totp.Enabled {
		authMethods.Totp = &LoginTotpAuthMethod{
			AllowRecovery: util.Config.Mfa.Totp.AllowRecovery,
		}
	}

	if util.Config.Mfa.Email.Enabled {
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

	body := SystemInfo{
		Version:           util.Version(),
		Ansible:           util.AnsibleVersion(),
		WebHost:           util.Config.WebHost,
		UseRemoteRunner:   util.Config.UseRemoteRunner,
		AuthMethods:       authMethods,
		LoginWithPassword: !util.Config.PasswordLoginDisable,
		Features:          proFeatures.GetFeatures(user, plan),
		SubscriptionState: token.State,
		GitClient:         util.Config.GitClientId,
		ScheduleTimezone:  timezone,
		Teams:             util.Config.Teams,
		Roles:             roles,
		BoltdbUsed:        util.Config.Dialect == "bolt",
	}

	helpers.WriteJSON(w, http.StatusOK, body)
}
