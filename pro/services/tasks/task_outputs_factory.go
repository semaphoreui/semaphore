package tasks

import (
	"github.com/semaphoreui/semaphore/pro_interfaces"
)

// NewTaskOutputsCollector is the open-source stub: workflows are a Pro feature,
// so no task has outputs to capture. Callers must nil-check.
func NewTaskOutputsCollector() pro_interfaces.TaskOutputsCollector {
	return nil
}
