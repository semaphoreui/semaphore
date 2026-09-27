package tasks

import (
	"context"

	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/services/alerting"
)

// notifyStatus hands the status change to the alerting service. Delivery
// runs in the background: a slow webhook must not stall the task runner.
func (t *TaskRunner) notifyStatus(status task_logger.TaskStatus) {
	if t.pool == nil || t.pool.alertService == nil {
		return
	}
	if _, ok := alerting.EventForStatus(status); !ok {
		return
	}

	task := t.Task
	template := t.Template
	service := t.pool.alertService

	go func() {
		// Errors are already written to the task log by the service.
		_ = service.Notify(context.Background(), task, template, status, t)
	}()
}
