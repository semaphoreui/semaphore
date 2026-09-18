package tasks

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/stretchr/testify/assert"
)

func TestResolveAlerts_EmptyTemplateUsesDefaults(t *testing.T) {
	snap := ResolveAlerts(db.Template{}, nil, []int{4, 5})
	assert.Equal(t, []int{4, 5}, snap.AlertIDs)
	assert.True(t, snap.OnSuccess)
	assert.True(t, snap.OnError)
}

func TestResolveAlerts_TemplateIDs(t *testing.T) {
	snap := ResolveAlerts(db.Template{
		AlertMode:             db.AlertModeIDs,
		AlertIDs:              []int{3, 7},
		AlertOnSuccess:        db.BoolPtr(false),
		AlertOnError:          db.BoolPtr(true),
		SuppressSuccessAlerts: true,
	}, nil, []int{9})
	assert.Equal(t, []int{3, 7}, snap.AlertIDs)
	assert.False(t, snap.OnSuccess)
	assert.True(t, snap.OnError)
}

func TestResolveAlerts_LegacySuppress(t *testing.T) {
	snap := ResolveAlerts(db.Template{
		SuppressSuccessAlerts: true,
		SuppressErrorAlerts:   false,
	}, nil, nil)
	assert.False(t, snap.OnSuccess)
	assert.True(t, snap.OnError)
}

func TestResolveAlerts_ScheduleInheritKeepsDefaults(t *testing.T) {
	tpl := db.Template{AlertOnError: db.BoolPtr(true)}
	schedule := &db.Schedule{AlertMode: db.AlertModeInherit, AlertIDs: []int{9}}
	snap := ResolveAlerts(tpl, schedule, []int{1, 2})
	assert.Equal(t, []int{1, 2}, snap.AlertIDs)
}

func TestResolveAlerts_ScheduleOverride(t *testing.T) {
	tpl := db.Template{
		AlertIDs:       []int{1, 2},
		AlertOnSuccess: db.BoolPtr(true),
		AlertOnError:   db.BoolPtr(true),
	}
	schedule := &db.Schedule{
		AlertMode:      db.AlertModeIDs,
		AlertIDs:       []int{5},
		AlertOnSuccess: db.BoolPtr(false),
	}
	snap := ResolveAlerts(tpl, schedule, []int{1, 2})
	assert.Equal(t, []int{5}, snap.AlertIDs)
	assert.False(t, snap.OnSuccess)
	assert.True(t, snap.OnError)
}

func TestResolveAlerts_ScheduleEmptyStops(t *testing.T) {
	tpl := db.Template{AlertIDs: []int{1}}
	schedule := &db.Schedule{AlertMode: db.AlertModeIDs, AlertIDs: []int{}}
	snap := ResolveAlerts(tpl, schedule, []int{1})
	assert.Empty(t, snap.AlertIDs)
}

func TestResolveAlerts_CustomEmptyIsSilent(t *testing.T) {
	snap := ResolveAlerts(db.Template{AlertMode: db.AlertModeIDs, AlertIDs: []int{}}, nil, []int{8})
	assert.Empty(t, snap.AlertIDs)
}

func TestShouldSkipStatusAlert_Snapshot(t *testing.T) {
	runner := &TaskRunner{
		Task: db.Task{
			AlertSnapshot: &db.AlertSnapshot{
				OnSuccess: false,
				OnError:   true,
			},
		},
	}
	runner.Task.Status = task_logger.TaskSuccessStatus
	assert.True(t, runner.shouldSkipStatusAlert())
	runner.Task.Status = task_logger.TaskFailStatus
	assert.False(t, runner.shouldSkipStatusAlert())
	runner.Task.Status = task_logger.TaskWaitingConfirmation
	assert.True(t, runner.shouldSkipStatusAlert())
}
