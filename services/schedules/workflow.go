package schedules

import (
	"context"
	"encoding/json"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	"github.com/semaphoreui/semaphore/services/audit"
	log "github.com/sirupsen/logrus"
)

// SetWorkflowService wires the workflow engine before the scheduler starts.
func (p *SchedulePool) SetWorkflowService(repo db.WorkflowManager, service pro_interfaces.WorkflowService) {
	p.workflowRepo, p.workflowService = repo, service
}

func (r ScheduleRunner) runWorkflow(schedule db.Schedule) {
	logger := log.WithFields(log.Fields{"project_id": schedule.ProjectID, "schedule_id": schedule.ID, "workflow_id": *schedule.WorkflowID})
	if r.pool.workflowRepo == nil || r.pool.workflowService == nil {
		logger.Error("Workflows are not available")
		return
	}
	workflow, err := r.pool.workflowRepo.GetWorkflowTemplate(schedule.ProjectID, *schedule.WorkflowID)
	if err != nil {
		logger.WithError(err).Error("failed to load schedule workflow")
		return
	}
	if workflow.ID == 0 {
		logger.Error("Workflows are not available in this build")
		return
	}
	params := db.WorkflowRunParams{}
	if schedule.TaskParams != nil && schedule.TaskParams.Environment != "" {
		if err = json.Unmarshal([]byte(schedule.TaskParams.Environment), &params.Environment); err != nil {
			logger.WithError(err).Error("failed to read scheduled workflow survey answers")
			return
		}
	}
	_, err = r.pool.workflowService.StartWorkflow(
		audit.WithActor(context.Background(), audit.SystemActor(audit.ComponentScheduler)),
		workflow, params, nil)
	if err != nil {
		logger.WithError(err).Error("failed to start scheduled workflow")
	}
}
