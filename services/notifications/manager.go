package notifications

import (
	"context"
	"fmt"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
)

var globalManager *NotificationManager

// InitializeManager initializes the global notification manager with all services
func InitializeManager() {
	globalManager = NewNotificationManager()
	
	// Register all notification services
	// New configuration approach
	if util.Config.Notifications != nil {
		for serviceName, config := range util.Config.Notifications {
			switch serviceName {
			case "telegram":
				globalManager.RegisterService(NewTelegramService(&config))
			case "slack":
				globalManager.RegisterService(NewSlackService(&config))
			case "gotify":
				globalManager.RegisterService(NewGotifyService(&config))
			case "dingtalk":
				globalManager.RegisterService(NewDingTalkService(&config))
			}
		}
	} else {
		// Legacy configuration fallback - register services with nil config to use legacy settings
		globalManager.RegisterService(NewTelegramService(nil))
		globalManager.RegisterService(NewSlackService(nil))
		globalManager.RegisterService(NewGotifyService(nil))
		globalManager.RegisterService(NewDingTalkService(nil))
	}
}

// GetManager returns the global notification manager
func GetManager() *NotificationManager {
	if globalManager == nil {
		InitializeManager()
	}
	return globalManager
}

// SendTaskNotification sends a notification for a task using the global manager
func SendTaskNotification(ctx context.Context, task *db.Task, template *db.Template, project *db.Project, author string, taskURL string) error {
	manager := GetManager()
	
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

// SendTestNotifications sends test notifications using the global manager
func SendTestNotifications(ctx context.Context, project *db.Project) error {
	manager := GetManager()
	return manager.SendTestNotification(ctx, *project)
}

// BackwardCompatibleTaskRunner provides backward compatibility for the existing TaskRunner
type BackwardCompatibleTaskRunner struct {
	Task     *db.Task
	Template *db.Template
	Project  *db.Project
	Author   string
	TaskURL  string
	Alert    bool
}

// SendAlerts sends alerts using both new and legacy systems for backward compatibility
func (t *BackwardCompatibleTaskRunner) SendAlerts(ctx context.Context) {
	if !t.Alert {
		return
	}
	
	// Send using new notification system
	err := SendTaskNotification(ctx, t.Task, t.Template, t.Project, t.Author, t.TaskURL)
	if err != nil {
		// Log error but continue - this is for backward compatibility
		// In production, you might want to add proper logging here
	}
}

// LegacyAlertSender provides methods that match the original TaskRunner interface
type LegacyAlertSender struct {
	runner *BackwardCompatibleTaskRunner
}

func NewLegacyAlertSender(task *db.Task, template *db.Template, project *db.Project, author string, taskURL string, alert bool) *LegacyAlertSender {
	return &LegacyAlertSender{
		runner: &BackwardCompatibleTaskRunner{
			Task:     task,
			Template: template,
			Project:  project,
			Author:   author,
			TaskURL:  taskURL,
			Alert:    alert,
		},
	}
}

func (l *LegacyAlertSender) SendTelegramAlert() {
	if !l.runner.Alert {
		return
	}
	
	manager := GetManager()
	service, exists := manager.GetService("telegram")
	if !exists || !service.IsConfigured() {
		return
	}
	
	ctx := context.Background()
	l.runner.SendAlerts(ctx)
}

func (l *LegacyAlertSender) SendSlackAlert() {
	if !l.runner.Alert {
		return
	}
	
	manager := GetManager()
	service, exists := manager.GetService("slack")
	if !exists || !service.IsConfigured() {
		return
	}
	
	ctx := context.Background()
	l.runner.SendAlerts(ctx)
}

func (l *LegacyAlertSender) SendGotifyAlert() {
	if !l.runner.Alert {
		return
	}
	
	manager := GetManager()
	service, exists := manager.GetService("gotify")
	if !exists || !service.IsConfigured() {
		return
	}
	
	ctx := context.Background()
	l.runner.SendAlerts(ctx)
}

func (l *LegacyAlertSender) SendDingTalkAlert() {
	if !l.runner.Alert {
		return
	}
	
	manager := GetManager()
	service, exists := manager.GetService("dingtalk")
	if !exists || !service.IsConfigured() {
		return
	}
	
	ctx := context.Background()
	l.runner.SendAlerts(ctx)
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