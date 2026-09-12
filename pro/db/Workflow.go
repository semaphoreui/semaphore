package db

import (
	coreDB "github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
)

func WorkflowConditionMatches(status task_logger.TaskStatus, condition coreDB.WorkflowEdgeCondition) bool {
	return false
}

func ValidateWorkflowTemplate(d coreDB.WorkflowTemplateValidationStore, workflow coreDB.WorkflowTemplate) error {
	return nil
}

func WorkflowRootNode(workflow coreDB.WorkflowTemplate) (coreDB.WorkflowNode, error) {
	return coreDB.WorkflowNode{}, nil
}
