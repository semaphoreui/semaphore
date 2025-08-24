package projects

import (
	"encoding/json"
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/tasks"
	"github.com/semaphoreui/semaphore/util"
)

// GetProjectNotifications returns the notification configuration for a project
func GetProjectNotifications(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)

	var notifications *util.NotificationsConfig
	if project.NotificationsConfig != nil && *project.NotificationsConfig != "" {
		notifications = &util.NotificationsConfig{}
		if err := json.Unmarshal([]byte(*project.NotificationsConfig), notifications); err != nil {
			helpers.WriteError(w, err)
			return
		}
	}

	helpers.WriteJSON(w, http.StatusOK, notifications)
}

// UpdateProjectNotifications updates the notification configuration for a project
func UpdateProjectNotifications(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
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

	// Serialize configuration to JSON
	configJSON, err := json.Marshal(body)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	configString := string(configJSON)
	project.NotificationsConfig = &configString

	// Update project in database
	err = helpers.Store(r).UpdateProject(project)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// TestProjectNotifications sends test notifications using the project's configuration
func TestProjectNotifications(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)

	// Check if notifications are enabled for the project
	if !project.Alert {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Notifications are not enabled for this project",
		})
		return
	}

	// Send test notifications using both legacy and new systems
	err := tasks.SendProjectTestAlerts(project, helpers.Store(r))
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

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