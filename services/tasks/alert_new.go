package tasks

import (
	"context"
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



// sendNewNotifications sends notifications using the new notification system
func (t *TaskRunner) sendNewNotifications() {
	if !t.alert {
		return
	}

	if t.Template.SuppressSuccessAlerts && t.Task.Status == task_logger.TaskSuccessStatus {
		return
	}

	// Get project information for notification context
	project, err := t.pool.store.GetProject(t.Task.ProjectID)
	if err != nil {
		t.Log("Failed to get project for notifications: " + err.Error())
		return
	}

	// Create notification service with global configuration
	notificationService := NewNotificationService()

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