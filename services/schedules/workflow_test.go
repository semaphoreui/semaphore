package schedules

import (
	"context"
	"fmt"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type workflowScheduleStore struct {
	db.Store
	schedule db.Schedule
}

func (s workflowScheduleStore) GetSchedule(projectID, scheduleID int) (db.Schedule, error) {
	return s.schedule, nil
}

type scheduleWorkflowRepo struct {
	db.WorkflowManager
	workflow db.WorkflowTemplate
}

func (r scheduleWorkflowRepo) GetWorkflowTemplate(projectID, workflowID int) (db.WorkflowTemplate, error) {
	return r.workflow, nil
}

type scheduleWorkflowService struct {
	pro_interfaces.WorkflowService
	calls  int
	params db.WorkflowRunParams
	actor  audit.Actor
}

func (s *scheduleWorkflowService) StartWorkflow(ctx context.Context, workflow db.WorkflowTemplate, params db.WorkflowRunParams, user *db.User) (db.WorkflowRun, error) {
	s.calls++
	s.params = params
	s.actor = audit.ActorFrom(ctx)
	return db.WorkflowRun{}, nil
}
func TestScheduleRunsWorkflowWithHADeduplication(t *testing.T) {
	for _, allowed := range []bool{true, false} {
		t.Run(map[bool]string{true: "winning node", false: "other node"}[allowed], func(t *testing.T) {
			service := &scheduleWorkflowService{}
			schedule := db.Schedule{ID: 8, ProjectID: 2, WorkflowID: new(3), Active: true, TaskParams: &db.TaskParams{Environment: `{"region":"eu"}`}}
			dedup := newMockDeduplicator()
			dedup.setAllowExecution(8, allowed)
			pool := SchedulePool{store: workflowScheduleStore{schedule: schedule}, dedup: dedup}
			pool.SetWorkflowService(scheduleWorkflowRepo{workflow: db.WorkflowTemplate{ID: 3, ProjectID: 2}}, service)
			runner := CreateScheduleRunner(2, 8, &pool, nil, nil)
			runner.Run()
			assert.Equal(t, 1, dedup.getLockAttempts(8))
			if allowed {
				assert.Equal(t, 1, service.calls)
				assert.Equal(t, db.MapStringAnyField{"region": "eu"}, service.params.Environment)
				assert.Equal(t, audit.SystemActor(audit.ComponentScheduler), service.actor)
			} else {
				assert.Zero(t, service.calls)
			}
		})
	}
}

func TestWorkflowOneTimeScheduleCleanup(t *testing.T) {
	for _, deleteAfter := range []bool{false, true} {
		t.Run(fmt.Sprint(deleteAfter), func(t *testing.T) {
			pool, store := setupTestSchedulePool(t)
			project, err := store.CreateProject(db.Project{Name: "scheduled workflows"})
			require.NoError(t, err)
			sqlStore := store.(*sql.SqlDb)
			_, err = sqlStore.Sql().Exec(sqlStore.PrepareQuery("insert into project__workflow_template (project_id,name) values (?,?)"), project.ID, "deploy")
			require.NoError(t, err)
			id, err := sqlStore.Sql().SelectInt("select id from project__workflow_template where name='deploy'")
			require.NoError(t, err)
			workflowID := int(id)
			runAt := time.Now().Add(-time.Minute)
			schedule, err := store.CreateSchedule(db.Schedule{ProjectID: project.ID, WorkflowID: &workflowID, Type: db.ScheduleTypeRunAt, RunAt: &runAt, Active: true, DeleteAfterRun: deleteAfter})
			require.NoError(t, err)
			service := &scheduleWorkflowService{}
			pool.SetWorkflowService(scheduleWorkflowRepo{workflow: db.WorkflowTemplate{ID: workflowID, ProjectID: project.ID}}, service)
			CreateScheduleRunner(project.ID, schedule.ID, pool, nil, nil).Run()
			assert.Equal(t, 1, service.calls)
			loaded, err := store.GetSchedule(project.ID, schedule.ID)
			if deleteAfter {
				assert.ErrorIs(t, err, db.ErrNotFound)
			} else {
				require.NoError(t, err)
				assert.False(t, loaded.Active)
			}
		})
	}
}

func TestWorkflowScheduleEnvironment(t *testing.T) {
	for _, tc := range []struct {
		name   string
		params *db.TaskParams
		calls  int
	}{
		{"defaults", nil, 1},
		{"empty environment", &db.TaskParams{}, 1},
		{"invalid JSON", &db.TaskParams{Environment: "invalid"}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service := &scheduleWorkflowService{}
			pool := SchedulePool{}
			pool.SetWorkflowService(scheduleWorkflowRepo{workflow: db.WorkflowTemplate{ID: 3, ProjectID: 2}}, service)
			runner := CreateScheduleRunner(2, 8, &pool, nil, nil)
			runner.runWorkflow(db.Schedule{ID: 8, ProjectID: 2, WorkflowID: new(3), TaskParams: tc.params})
			assert.Equal(t, tc.calls, service.calls)
			assert.Empty(t, service.params.Environment)
		})
	}
}
