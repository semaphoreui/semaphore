package tasks

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
)

func testRunner(project db.Project, store db.Store) (*TaskRunner, error) {
	projectUsers, err := store.GetProjectUsers(project.ID, db.RetrieveQueryParams{})
	if err != nil {
		return nil, err
	}

	var userIDs []int
	for _, u := range projectUsers {
		userIDs = append(userIDs, u.ID)
	}

	pool := &TaskPool{
		logger: make(chan logRecord, 100),
		store:  store,
	}
	go func() {
		for range pool.logger {
		}
	}()

	return &TaskRunner{
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
		pool:      pool,
	}, nil
}

func closeTestRunner(tr *TaskRunner) {
	if tr == nil || tr.pool == nil || tr.pool.logger == nil {
		return
	}
	close(tr.pool.logger)
}

// SendProjectTestAlerts sends test alerts to all enabled notifiers for the given project.
func SendProjectTestAlerts(project db.Project, store db.Store) error {
	tr, err := testRunner(project, store)
	if err != nil {
		return err
	}
	defer closeTestRunner(tr)

	alerts, err := store.GetAlerts(project.ID, db.RetrieveQueryParams{})
	if err != nil {
		return err
	}

	var firstErr error
	for _, alert := range alerts {
		if !alert.Enabled {
			continue
		}
		if err := tr.sendProjectAlert(alert); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// SendAlertTest sends a single project alert as a test message.
func SendAlertTest(project db.Project, alert db.Alert, store db.Store) error {
	tr, err := testRunner(project, store)
	if err != nil {
		return err
	}
	defer closeTestRunner(tr)
	return tr.sendProjectAlert(alert)
}
