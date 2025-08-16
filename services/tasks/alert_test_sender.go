package tasks

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
)

// SendProjectTestAlerts sends test alerts to all enabled notifiers for the given project.
func SendProjectTestAlerts(project db.Project, store db.Store) (err error) {

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
		users:     userIDs,
		alert:     project.Alert,
		alertChat: project.AlertChat,
		pool: &TaskPool{
			logger: make(chan logRecord, 100),
			store:  store,
		},
	}

	tr.sendTelegramAlert()
	tr.sendSlackAlert()
	tr.sendRocketChatAlert()
	tr.sendMicrosoftTeamsAlert()
	tr.sendDingTalkAlert()
	tr.sendGotifyAlert()
	tr.sendMailAlert()

	return
}
