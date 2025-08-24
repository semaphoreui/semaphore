package projects

import (
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/util"
)

// GetGlobalNotifications returns the global notification configuration
func GetGlobalNotifications(w http.ResponseWriter, r *http.Request) {
	helpers.WriteJSON(w, http.StatusOK, util.Config.Notifications)
}

// UpdateGlobalNotifications updates the global notification configuration
func UpdateGlobalNotifications(w http.ResponseWriter, r *http.Request) {
	var body util.NotificationsConfig

	if !helpers.Bind(w, r, &body) {
		return
	}

	// Validate notification configurations
	if body.Telegram != nil && body.Telegram.Enabled {
		if body.Telegram.Token == "" {
			helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Telegram token is required when Telegram notifications are enabled",
			})
			return
		}
		if body.Telegram.Channel == "" && body.Telegram.ChatID == "" {
			helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Telegram channel or chat_id is required when Telegram notifications are enabled",
			})
			return
		}
	}

	if body.Slack != nil && body.Slack.Enabled {
		if body.Slack.WebhookURL == "" {
			helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Slack webhook URL is required when Slack notifications are enabled",
			})
			return
		}
	}

	if body.Gotify != nil && body.Gotify.Enabled {
		if body.Gotify.URL == "" || body.Gotify.Token == "" {
			helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Gotify URL and token are required when Gotify notifications are enabled",
			})
			return
		}
	}

	if body.Dingtalk != nil && body.Dingtalk.Enabled {
		if body.Dingtalk.WebhookURL == "" {
			helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Dingtalk webhook URL is required when Dingtalk notifications are enabled",
			})
			return
		}
	}

	// Update global configuration
	util.Config.Notifications = &body

	helpers.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Global notification configuration updated successfully",
	})
}