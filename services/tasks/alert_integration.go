package tasks

import (
	"context"

	"github.com/semaphoreui/semaphore/services/notifications"
	"github.com/semaphoreui/semaphore/util"
)

// sendNewSystemAlerts sends alerts using the new notification system
func (t *TaskRunner) sendNewSystemAlerts() {
	if !t.alert {
		return
	}

	// Get project information
	project, err := t.pool.store.GetProject(t.Template.ProjectID)
	if err != nil {
		t.Log("Failed to get project for notifications: " + err.Error())
		return
	}

	// Get author information
	author, _ := t.alertInfos()

	// Create task URL
	taskURL := notifications.CreateTaskLink(
		util.Config.WebHost,
		t.Template.ProjectID,
		t.Template.ID,
		t.Task.ID,
	)

	// Send notification using new system
	ctx := context.Background()
	err = notifications.SendTaskNotification(ctx, &t.Task, &t.Template, &project, author, taskURL)
	if err != nil {
		t.Log("Failed to send new system notifications: " + err.Error())
	}
}

// Enhanced alert methods that use both old and new systems for maximum compatibility
func (t *TaskRunner) sendEnhancedTelegramAlert() {
	// Send using new system first
	t.sendNewSystemAlerts()
	
	// Fallback to original system if new system fails or isn't configured
	t.sendTelegramAlert()
}

func (t *TaskRunner) sendEnhancedSlackAlert() {
	// Send using new system first  
	t.sendNewSystemAlerts()
	
	// Fallback to original system if new system fails or isn't configured
	t.sendSlackAlert()
}

func (t *TaskRunner) sendEnhancedGotifyAlert() {
	// Send using new system first
	t.sendNewSystemAlerts()
	
	// Fallback to original system if new system fails or isn't configured
	t.sendGotifyAlert()
}

func (t *TaskRunner) sendEnhancedDingTalkAlert() {
	// Send using new system first
	t.sendNewSystemAlerts()
	
	// Fallback to original system if new system fails or isn't configured
	t.sendDingTalkAlert()
}

// sendAllNotifications is a unified method to send all notifications
func (t *TaskRunner) sendAllNotifications() {
	if !t.alert {
		return
	}

	// Try new notification system first
	t.sendNewSystemAlerts()
	
	// For email, we still use the original system since it's more complex
	// and requires user-specific handling
	t.sendMailAlert()
	
	// For other services that might not be covered by the new system,
	// we maintain the original calls as fallbacks
	t.sendRocketChatAlert()
	t.sendMicrosoftTeamsAlert()
}

// Wrapper methods for backward compatibility that can be called from TaskRunner_logging.go
func (t *TaskRunner) SendNotifications() {
	t.sendAllNotifications()
}

// Individual service methods for backward compatibility
func (t *TaskRunner) SendTelegramNotification() {
	if !t.alert {
		return
	}
	
	// Use new system
	manager := notifications.GetManager()
	service, exists := manager.GetService("telegram")
	if exists && service.IsConfigured() {
		t.sendNewSystemAlerts()
		return
	}
	
	// Fallback to legacy
	t.sendTelegramAlert()
}

func (t *TaskRunner) SendSlackNotification() {
	if !t.alert {
		return
	}
	
	// Use new system
	manager := notifications.GetManager()
	service, exists := manager.GetService("slack")
	if exists && service.IsConfigured() {
		t.sendNewSystemAlerts()
		return
	}
	
	// Fallback to legacy
	t.sendSlackAlert()
}

func (t *TaskRunner) SendGotifyNotification() {
	if !t.alert {
		return
	}
	
	// Use new system
	manager := notifications.GetManager()
	service, exists := manager.GetService("gotify")
	if exists && service.IsConfigured() {
		t.sendNewSystemAlerts()
		return
	}
	
	// Fallback to legacy
	t.sendGotifyAlert()
}

func (t *TaskRunner) SendDingTalkNotification() {
	if !t.alert {
		return
	}
	
	// Use new system
	manager := notifications.GetManager()
	service, exists := manager.GetService("dingtalk")
	if exists && service.IsConfigured() {
		t.sendNewSystemAlerts()
		return
	}
	
	// Fallback to legacy
	t.sendDingTalkAlert()
}