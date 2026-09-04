package tasks

import (
	"errors"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/pkg/tz"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestTaskPool(t *testing.T) (*sql.SqlDb, *TaskPool, db.Project, db.Task) {
	prevCfg := util.Config
	t.Cleanup(func() { util.Config = prevCfg })
	util.Config = &util.ConfigType{}

	store := sql.InitConfigCreateTestStore()
	proj, err := store.CreateProject(db.Project{Name: "Test Proj"})
	require.NoError(t, err)

	key, err := store.CreateAccessKey(db.AccessKey{ProjectID: &proj.ID, Type: db.AccessKeyNone})
	require.NoError(t, err)

	repo, err := store.CreateRepository(db.Repository{
		ProjectID: proj.ID,
		SSHKeyID:  key.ID,
		Name:      "Test Repo",
		GitURL:    "git@example.com:test/test",
		GitBranch: "master",
	})
	require.NoError(t, err)

	inv, err := store.CreateInventory(db.Inventory{ProjectID: proj.ID})
	require.NoError(t, err)

	tpl, err := store.CreateTemplate(db.Template{
		Name:         "Test Template",
		Playbook:     "test.yml",
		ProjectID:    proj.ID,
		RepositoryID: repo.ID,
		InventoryID:  &inv.ID,
	})
	require.NoError(t, err)

	task, err := store.CreateTask(db.Task{
		ProjectID:  proj.ID,
		TemplateID: tpl.ID,
		Status:     task_logger.TaskWaitingStatus,
	}, 0)
	require.NoError(t, err)

	state := NewMemoryTaskStateStore()
	pool := &TaskPool{
		store:  store,
		state:  state,
		logger: make(chan logRecord, 1000),
	}

	taskRunner := &TaskRunner{
		Task:     task,
		Template: tpl,
		pool:     pool,
	}
	state.Enqueue(taskRunner)

	return store, pool, proj, task
}

func TestRemoteJob_ProjectRunnerMatchOnline(t *testing.T) {
	store, pool, proj, task := setupTestTaskPool(t)

	pTag := "P11"
	r, err := store.CreateRunner(db.Runner{
		ProjectID: &proj.ID,
		Name:      "project-runner-1",
		Active:    true,
		Token:     "token-proj-1",
		Tags:      []string{pTag},
	})
	require.NoError(t, err)
	err = store.TouchRunner(r)
	require.NoError(t, err)

	job := &RemoteJob{
		RunnerTag: &pTag,
		Task:      task,
		taskPool:  pool,
	}

	err = job.Run("user", nil, "")
	assert.NoError(t, err)

	updatedTask, err := store.GetTask(proj.ID, task.ID)
	require.NoError(t, err)
	require.NotNil(t, updatedTask.RunnerID)
	assert.Equal(t, r.ID, *updatedTask.RunnerID)
}

func TestRemoteJob_GlobalDefaultFallbackWhenProjectRunnerOffline(t *testing.T) {
	store, pool, proj, task := setupTestTaskPool(t)

	pTag := "P11"
	// Project runner exists and is active, but offline (touched is nil)
	projRunner, err := store.CreateRunner(db.Runner{
		ProjectID: &proj.ID,
		Name:      "project-runner-offline",
		Active:    true,
		Token:     "token-proj-offline",
		Tags:      []string{pTag},
	})
	require.NoError(t, err)
	assert.False(t, projRunner.IsOnline(tz.Now(), util.Config.RunnersOfflineTimeout()))

	// Global runner with is_default=true and online
	globalRunner, err := store.CreateRunner(db.Runner{
		ProjectID: nil,
		Name:      "global-runner-default",
		Active:    true,
		IsDefault: true,
		Token:     "token-global-default",
		Tags:      []string{"G1"},
	})
	require.NoError(t, err)
	err = store.TouchRunner(globalRunner)
	require.NoError(t, err)

	job := &RemoteJob{
		RunnerTag: &pTag,
		Task:      task,
		taskPool:  pool,
	}

	err = job.Run("user", nil, "")
	assert.NoError(t, err)

	updatedTask, err := store.GetTask(proj.ID, task.ID)
	require.NoError(t, err)
	require.NotNil(t, updatedTask.RunnerID)
	assert.Equal(t, globalRunner.ID, *updatedTask.RunnerID, "should delegate to global default runner when project runner is offline")
}

func TestRemoteJob_GlobalTaggedFallbackWhenProjectOffline(t *testing.T) {
	store, pool, proj, task := setupTestTaskPool(t)

	pTag := "P11"
	// Project runner exists, but offline
	_, err := store.CreateRunner(db.Runner{
		ProjectID: &proj.ID,
		Name:      "project-runner-offline",
		Active:    true,
		Token:     "token-proj-offline",
		Tags:      []string{pTag},
	})
	require.NoError(t, err)

	// Global runner with same tag "P11" is online
	globalRunner, err := store.CreateRunner(db.Runner{
		ProjectID: nil,
		Name:      "global-runner-p11",
		Active:    true,
		Token:     "token-global-p11",
		Tags:      []string{pTag},
	})
	require.NoError(t, err)
	err = store.TouchRunner(globalRunner)
	require.NoError(t, err)

	job := &RemoteJob{
		RunnerTag: &pTag,
		Task:      task,
		taskPool:  pool,
	}

	err = job.Run("user", nil, "")
	assert.NoError(t, err)

	updatedTask, err := store.GetTask(proj.ID, task.ID)
	require.NoError(t, err)
	require.NotNil(t, updatedTask.RunnerID)
	assert.Equal(t, globalRunner.ID, *updatedTask.RunnerID)
}

func TestRemoteJob_OfflineRunnersKeepTaskWaiting(t *testing.T) {
	store, pool, proj, task := setupTestTaskPool(t)

	pTag := "P11"
	// Project runner tagged P11 is offline
	_, err := store.CreateRunner(db.Runner{
		ProjectID: &proj.ID,
		Name:      "project-runner-offline",
		Active:    true,
		Token:     "token-proj-offline",
		Tags:      []string{pTag},
	})
	require.NoError(t, err)

	// Global default runner is also offline
	_, err = store.CreateRunner(db.Runner{
		ProjectID: nil,
		Name:      "global-runner-offline",
		Active:    true,
		IsDefault: true,
		Token:     "token-global-offline",
	})
	require.NoError(t, err)

	job := &RemoteJob{
		RunnerTag: &pTag,
		Task:      task,
		taskPool:  pool,
	}

	err = job.Run("user", nil, "")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrAllRunnersBusy), "expected ErrAllRunnersBusy, got %v", err)
}

func TestRemoteJob_DisabledRunnersKeepTaskWaiting(t *testing.T) {
	store, pool, proj, task := setupTestTaskPool(t)

	pTag := "P11"
	// Project runner tagged P11 exists in DB, but Active = false
	_, err := store.CreateRunner(db.Runner{
		ProjectID: &proj.ID,
		Name:      "project-runner-disabled",
		Active:    false,
		Token:     "token-proj-disabled",
		Tags:      []string{pTag},
	})
	require.NoError(t, err)

	job := &RemoteJob{
		RunnerTag: &pTag,
		Task:      task,
		taskPool:  pool,
	}

	err = job.Run("user", nil, "")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrAllRunnersBusy), "expected ErrAllRunnersBusy for disabled matching runner, got %v", err)
}

func TestRemoteJob_NoRunnersConfiguredFailsImmediately(t *testing.T) {
	_, pool, _, task := setupTestTaskPool(t)

	nonExistentTag := "non-existent-tag"
	job := &RemoteJob{
		RunnerTag: &nonExistentTag,
		Task:      task,
		taskPool:  pool,
	}

	err := job.Run("user", nil, "")
	assert.Error(t, err)
	assert.False(t, errors.Is(err, ErrAllRunnersBusy))
	assert.Contains(t, err.Error(), "no runners available")
}

func TestRemoteJob_NoTagSpecified_ProjectAndGlobalDefault(t *testing.T) {
	store, pool, proj, task := setupTestTaskPool(t)

	// Global default runner is online
	globalRunner, err := store.CreateRunner(db.Runner{
		ProjectID: nil,
		Name:      "global-default",
		Active:    true,
		IsDefault: true,
		Token:     "token-global-default",
	})
	require.NoError(t, err)
	err = store.TouchRunner(globalRunner)
	require.NoError(t, err)

	job := &RemoteJob{
		RunnerTag: nil,
		Task:      task,
		taskPool:  pool,
	}

	err = job.Run("user", nil, "")
	assert.NoError(t, err)

	updatedTask, err := store.GetTask(proj.ID, task.ID)
	require.NoError(t, err)
	require.NotNil(t, updatedTask.RunnerID)
	assert.Equal(t, globalRunner.ID, *updatedTask.RunnerID)
}
