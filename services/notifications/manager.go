package notifications

import (
	"context"
	"fmt"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
)

// CreateManager creates a new notification manager with all services registered
func CreateManager() *NotificationManager {
	manager := NewNotificationManager()
	
	// Register all notification services
	// New configuration approach
	if util.Config.Notifications != nil {
		for serviceName, config := range util.Config.Notifications {
			switch serviceName {
			case "telegram":
				manager.RegisterService(NewTelegramService(&config))
			case "slack":
				manager.RegisterService(NewSlackService(&config))
			case "gotify":
				manager.RegisterService(NewGotifyService(&config))
			case "dingtalk":
				manager.RegisterService(NewDingTalkService(&config))
			}
		}
	} else {
		// Legacy configuration fallback - register services with nil config to use legacy settings
		manager.RegisterService(NewTelegramService(nil))
		manager.RegisterService(NewSlackService(nil))
		manager.RegisterService(NewGotifyService(nil))
		manager.RegisterService(NewDingTalkService(nil))
	}
	
	return manager
}

// SendTaskNotification sends a notification for a task using the provided manager
func SendTaskNotification(ctx context.Context, manager *NotificationManager, task *db.Task, template *db.Template, project *db.Project, author string, taskURL string) error {
	// Get task version
	version := ""
	if task.Version != nil {
		version = *task.Version
	} else if template.Type != db.TemplateTask {
		// This would need access to store to get incoming version
		// For now, we'll leave it empty in this context
		version = ""
	}
	
	// Prepare recipients for legacy project chat override
	var recipients []string
	if project.AlertChat != nil && *project.AlertChat != "" {
		recipients = append(recipients, *project.AlertChat)
	}
	
	// Create notification
	notification := &Notification{
		TemplateName:          template.Name,
		TemplateID:            template.ID,
		ProjectID:             project.ID,
		TaskID:                task.ID,
		TaskStatus:            task.Status,
		TaskMessage:           task.Message,
		TaskVersion:           version,
		TaskURL:               taskURL,
		AuthorName:            author,
		Recipients:            recipients,
		ProjectConfig:         make(map[string]interface{}),
		SuppressSuccessAlerts: template.SuppressSuccessAlerts,
	}
	
	// Add project-specific alert chat to project config for new services
	if project.AlertChat != nil && *project.AlertChat != "" {
		notification.ProjectConfig["telegram_chat"] = *project.AlertChat
	}
	
	return manager.SendNotification(ctx, notification)
}

// SendTestNotifications sends test notifications using the provided manager
func SendTestNotifications(ctx context.Context, manager *NotificationManager, project *db.Project) error {
	return manager.SendTestNotification(ctx, *project)
}



// Helper function to create task link (matches original implementation)
func CreateTaskLink(webHost string, projectID, templateID, taskID int) string {
	return fmt.Sprintf(
		"%s/project/%d/templates/%d?t=%d",
		webHost,
		projectID,
		templateID,
		taskID,
	)
}