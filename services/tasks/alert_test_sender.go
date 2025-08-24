package tasks

import (
	"context"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/services/notifications"
)

// SendProjectTestAlerts sends test alerts to all enabled notifiers for the given project.
func SendProjectTestAlerts(project db.Project, store db.Store, notificationManager *notifications.NotificationManager) (err error) {
	// Use new notification system first if available
	if notificationManager != nil {
		ctx := context.Background()
		err = notifications.SendTestNotifications(ctx, notificationManager, &project)
		if err != nil {
			// Log error but continue with legacy system as fallback
			// In production, you might want proper logging here
		}
	}

	// Legacy fallback for services not yet covered by new system or if new system fails
	projectUsers, err := store.GetProjectUsers(project.ID, db.RetrieveQueryParams{})
	if err != nil {
		return
	}

	var userIDs []int
	for _, u := range projectUsers {
		userIDs = append(userIDs, u.ID)
	}

	tr := &TaskRunner{
		Task: db.Task{
			ProjectID:  project.ID,
			TemplateID: 0,
			Status:     task_logger.TaskSuccessStatus,
			Message:    "This is a test notification",
		},
		Template: db.Template{
			ID:        0,
			ProjectID: project.ID,
			Name:      "Test Notification",
			Type:      db.TemplateTask,
		},
		users:                userIDs,
		alert:                project.Alert,
		alertChat:            project.AlertChat,
		notificationManager:  notificationManager,
		pool: &TaskPool{
			logger: make(chan logRecord, 100),
			store:  store,
		},
	}

	// Send legacy notifications for services not covered by new system
	tr.sendRocketChatAlert()
	tr.sendMicrosoftTeamsAlert()
	tr.sendMailAlert()

	return
}
