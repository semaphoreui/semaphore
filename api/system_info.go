package api

import (
	"errors"
	"fmt"
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
		log.WithError(err).Error("Failed to get roles")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var plan string

	token, err := c.subscriptionService.GetToken()

	if errors.Is(err, db.ErrNotFound) {
		err = nil
	}

	if err != nil {
		log.WithError(err).Error("Failed to get subscription plan")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	switch {
	case errors.Is(err, db.ErrNotFound):
		err = nil
		plan = ""
	case err != nil:
		log.WithError(err).Error("Failed to get subscription plan")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	default:
		plan = token.Plan
	}

	// Check if the user has seen the intro
	seenIntroKey := fmt.Sprintf("seen_intro_%d", user.ID)
	seenIntroVersion, err := helpers.Store(r).GetOption(seenIntroKey)
	seenIntro := seenIntroVersion == util.Ver

	if err != nil {
		log.WithError(err).Error("Failed to get seen_intro option")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	body := map[string]any{
		"version":           util.Version(),
		"ansible":           util.AnsibleVersion(),
		"web_host":          util.Config.WebHost,
		"use_remote_runner": util.Config.UseRemoteRunner,
		"auth_methods":      authMethods,
		"premium_features":  proFeatures.GetFeatures(user, plan),
		"git_client":        util.Config.GitClientId,
		"schedule_timezone": timezone,
		"teams":             util.Config.Teams,
		"roles":             roles,
		"seen_intro":        seenIntro,
	}

	helpers.WriteJSON(w, http.StatusOK, body)
}

func (c *SystemInfoController) SetSeenIntro(w http.ResponseWriter, r *http.Request) {
	user := helpers.GetFromContext(r, "user").(*db.User)

	var reqBody struct {
		Version string `json:"version"`
	}

	if !helpers.Bind(w, r, &reqBody) {
		return
	}

	seenIntroKey := fmt.Sprintf("seen_intro_%d", user.ID)

	err := helpers.Store(r).SetOption(seenIntroKey, reqBody.Version)
	if err != nil {
		log.WithError(err).Error("Failed to set seen_intro option")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}
