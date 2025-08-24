package tasks

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/services/notifications"
	"github.com/semaphoreui/semaphore/util"
)

// NewNotificationService creates a new notification service with the configured providers
func NewNotificationService() *notifications.NotificationManager {
	manager := notifications.NewNotificationManager()

	// Add providers based on global configuration
	if util.Config.Notifications != nil {
		if util.Config.Notifications.Telegram != nil {
			telegramConfig := &notifications.TelegramConfig{
				NotificationConfig: notifications.NotificationConfig{
					Enabled: util.Config.Notifications.Telegram.Enabled,
					Token:   util.Config.Notifications.Telegram.Token,
					Channel: util.Config.Notifications.Telegram.Channel,
				},
				ChatID: util.Config.Notifications.Telegram.ChatID,
			}
			manager.AddProvider(notifications.NewTelegramProvider(telegramConfig))
		}

		if util.Config.Notifications.Slack != nil {
			slackConfig := &notifications.SlackConfig{
				NotificationConfig: notifications.NotificationConfig{
					Enabled: util.Config.Notifications.Slack.Enabled,
					Token:   util.Config.Notifications.Slack.Token,
					Channel: util.Config.Notifications.Slack.Channel,
				},
				WebhookURL: util.Config.Notifications.Slack.WebhookURL,
			}
			manager.AddProvider(notifications.NewSlackProvider(slackConfig))
		}

		if util.Config.Notifications.Gotify != nil {
			gotifyConfig := &notifications.GotifyConfig{
				NotificationConfig: notifications.NotificationConfig{
					Enabled: util.Config.Notifications.Gotify.Enabled,
					Token:   util.Config.Notifications.Gotify.Token,
					Channel: util.Config.Notifications.Gotify.Channel,
				},
				URL:      util.Config.Notifications.Gotify.URL,
				Priority: util.Config.Notifications.Gotify.Priority,
			}
			manager.AddProvider(notifications.NewGotifyProvider(gotifyConfig))
		}

		if util.Config.Notifications.Dingtalk != nil {
			dingtalkConfig := &notifications.DingtalkConfig{
				NotificationConfig: notifications.NotificationConfig{
					Enabled: util.Config.Notifications.Dingtalk.Enabled,
					Token:   util.Config.Notifications.Dingtalk.Token,
					Channel: util.Config.Notifications.Dingtalk.Channel,
				},
				WebhookURL: util.Config.Notifications.Dingtalk.WebhookURL,
				Secret:     util.Config.Notifications.Dingtalk.Secret,
			}
			manager.AddProvider(notifications.NewDingtalkProvider(dingtalkConfig))
		}
	}

	return manager
}

// CreateProjectNotificationService creates a notification service with project-specific configuration
func CreateProjectNotificationService(project db.Project) *notifications.NotificationManager {
	manager := notifications.NewNotificationManager()

	// If project has specific notification configuration, use that
	if project.NotificationsConfig != nil && *project.NotificationsConfig != "" {
		var projectNotifications util.NotificationsConfig
		if err := json.Unmarshal([]byte(*project.NotificationsConfig), &projectNotifications); err == nil {
			if projectNotifications.Telegram != nil {
				telegramConfig := &notifications.TelegramConfig{
					NotificationConfig: notifications.NotificationConfig{
						Enabled: projectNotifications.Telegram.Enabled,
						Token:   projectNotifications.Telegram.Token,
						Channel: projectNotifications.Telegram.Channel,
					},
					ChatID: projectNotifications.Telegram.ChatID,
				}
				manager.AddProvider(notifications.NewTelegramProvider(telegramConfig))
			}

			if projectNotifications.Slack != nil {
				slackConfig := &notifications.SlackConfig{
					NotificationConfig: notifications.NotificationConfig{
						Enabled: projectNotifications.Slack.Enabled,
						Token:   projectNotifications.Slack.Token,
						Channel: projectNotifications.Slack.Channel,
					},
					WebhookURL: projectNotifications.Slack.WebhookURL,
				}
				manager.AddProvider(notifications.NewSlackProvider(slackConfig))
			}

			if projectNotifications.Gotify != nil {
				gotifyConfig := &notifications.GotifyConfig{
					NotificationConfig: notifications.NotificationConfig{
						Enabled: projectNotifications.Gotify.Enabled,
						Token:   projectNotifications.Gotify.Token,
						Channel: projectNotifications.Gotify.Channel,
					},
					URL:      projectNotifications.Gotify.URL,
					Priority: projectNotifications.Gotify.Priority,
				}
				manager.AddProvider(notifications.NewGotifyProvider(gotifyConfig))
			}

			if projectNotifications.Dingtalk != nil {
				dingtalkConfig := &notifications.DingtalkConfig{
					NotificationConfig: notifications.NotificationConfig{
						Enabled: projectNotifications.Dingtalk.Enabled,
						Token:   projectNotifications.Dingtalk.Token,
						Channel: projectNotifications.Dingtalk.Channel,
					},
					WebhookURL: projectNotifications.Dingtalk.WebhookURL,
					Secret:     projectNotifications.Dingtalk.Secret,
				}
				manager.AddProvider(notifications.NewDingtalkProvider(dingtalkConfig))
			}

			return manager
		}
	}

	// Fallback to global configuration
	return NewNotificationService()
}

// sendNewNotifications sends notifications using the new notification system
func (t *TaskRunner) sendNewNotifications() {
	if !t.alert {
		return
	}

	if t.Template.SuppressSuccessAlerts && t.Task.Status == task_logger.TaskSuccessStatus {
		return
	}

	// Get project information for project-specific notifications
	project, err := t.pool.store.GetProject(t.Task.ProjectID)
	if err != nil {
		t.Log("Failed to get project for notifications: " + err.Error())
		return
	}

	// Create notification service with project-specific or global configuration
	notificationService := CreateProjectNotificationService(project)

	// Prepare notification message
	author, version := t.alertInfos()
	message := &notifications.NotificationMessage{
		Title:       t.Template.Name,
		Content:     t.Task.Message,
		Color:       t.alertColor(""),
		ProjectName: project.Name,
		TaskID:      strconv.Itoa(t.Task.ID),
		TaskURL:     t.taskLink(),
		Status:      t.Task.Status.Format(),
		Author:      author,
		Version:     version,
		Description: t.Task.Message,
		Timestamp:   time.Now(),
	}

	// Send notification
	ctx := context.Background()
	notificationService.SendNotification(ctx, message)
}