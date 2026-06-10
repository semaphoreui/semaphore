package server

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pro_interfaces"
)

// workflowService is the open-source no-op stub for the Pro workflow
// orchestration service. The task pool still calls HandleWorkflowTaskCompletion
// / GetWorkflowRunArtifacts on every finished task, so the methods must be safe
// no-ops; the workflow API itself is disabled via the stub controller and the
// Workflows feature flag.
type workflowService struct{}

func NewWorkflowService(workflowRepo db.WorkflowManager, templateReceiver db.WorkflowTemplateValidationStore, enqueuer pro_interfaces.WorkflowTaskEnqueuer, locker pro_interfaces.WorkflowRunLocker) pro_interfaces.WorkflowService {
	return &workflowService{}
}

// NewWorkflowReconciler is the open-source no-op stub: workflows are a Pro
// feature, so there is nothing to reconcile. Callers must nil-check.
func NewWorkflowReconciler(_ db.WorkflowManager, _ pro_interfaces.WorkflowService) pro_interfaces.WorkflowReconciler {
	return nil
}

func (s *workflowService) StartWorkflow(workflow db.WorkflowTemplate, user *db.User) (db.WorkflowRun, error) {
	return db.WorkflowRun{}, nil
}

func (s *workflowService) ProgressWorkflowRun(projectID int, runID int, user *db.User) error {
	return nil
}

func (s *workflowService) StopWorkflowRun(projectID int, runID int, user *db.User) (db.WorkflowRun, error) {
	return db.WorkflowRun{}, nil
}

func (s *workflowService) ResolveWorkflowApproval(projectID int, workflowID int, runID int, nodeID int, status db.WorkflowApprovalStatus, user *db.User) (db.WorkflowApproval, error) {
	return db.WorkflowApproval{}, nil
}

func (s *workflowService) HandleWorkflowTaskCompletion(task db.Task) error {
	return nil
}

func (s *workflowService) GetWorkflowRunArtifacts(projectID int, runID int, currentTaskID *int) (map[string]any, error) {
	return map[string]any{}, nil
}
