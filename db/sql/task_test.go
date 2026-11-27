package sql

import (
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateTaskEffectiveInventory_KeepsOtherColumns(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, repositoryID := newTemplateTestProject(t, store)
	inventory, err := store.CreateInventory(db.Inventory{ProjectID: projectID, Name: "inv"})
	require.NoError(t, err)
	tpl, err := store.CreateTemplate(db.Template{ProjectID: projectID, RepositoryID: repositoryID, Name: "tpl", Playbook: "tpl.yml"})
	require.NoError(t, err)
	task, err := store.CreateTask(db.Task{ProjectID: projectID, TemplateID: tpl.ID, Created: time.Now()}, 0)
	require.NoError(t, err)
	task.Status = task_logger.TaskRunningStatus
	require.NoError(t, store.UpdateTask(task))

	require.NoError(t, store.UpdateTaskEffectiveInventory(projectID, task.ID, &inventory.ID))

	stored, err := store.GetTask(projectID, task.ID)
	require.NoError(t, err)
	require.NotNil(t, stored.EffectiveInventoryID)
	assert.Equal(t, inventory.ID, *stored.EffectiveInventoryID)
	assert.Equal(t, task_logger.TaskRunningStatus, stored.Status)
	assert.Nil(t, stored.InventoryID)
}

func TestUpdateTask_KeepsRecordedInventory(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, repositoryID := newTemplateTestProject(t, store)
	inventory, err := store.CreateInventory(db.Inventory{ProjectID: projectID, Name: "inv"})
	require.NoError(t, err)
	tpl, err := store.CreateTemplate(db.Template{ProjectID: projectID, RepositoryID: repositoryID, Name: "tpl", Playbook: "tpl.yml"})
	require.NoError(t, err)
	task, err := store.CreateTask(db.Task{ProjectID: projectID, TemplateID: tpl.ID, Created: time.Now()}, 0)
	require.NoError(t, err)
	require.NoError(t, store.UpdateTaskEffectiveInventory(projectID, task.ID, &inventory.ID))

	task.Status = task_logger.TaskSuccessStatus
	require.NoError(t, store.UpdateTask(task))

	stored, err := store.GetTask(projectID, task.ID)
	require.NoError(t, err)
	assert.Equal(t, task_logger.TaskSuccessStatus, stored.Status)
	require.NotNil(t, stored.EffectiveInventoryID)
	assert.Equal(t, inventory.ID, *stored.EffectiveInventoryID)
}
