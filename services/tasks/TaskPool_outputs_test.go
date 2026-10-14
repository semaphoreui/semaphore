package tasks

import (
	"errors"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskPool_StoreRemoteTaskOutputs(t *testing.T) {
	document := `{"values":{"image_tag":"1.4.2"}}`
	runID := 3

	t.Run("task outside a workflow run is ignored", func(t *testing.T) {
		fixture := newTaskRunnerRunFixture(t)
		collector := &outputsCollectorFake{}
		fixture.pool.SetTaskOutputsCollector(collector)
		tsk := &TaskRunner{Task: fixture.task, pool: &fixture.pool}

		require.NoError(t, fixture.pool.StoreRemoteTaskOutputs(tsk, document))

		assert.Empty(t, collector.validated)
		assert.Nil(t, tsk.Task.Artifacts)
		persisted, err := fixture.store.GetTaskByID(fixture.task.ID)
		require.NoError(t, err)
		assert.Nil(t, persisted.Artifacts)
	})

	t.Run("without a collector a workflow task cannot have outputs", func(t *testing.T) {
		fixture := newTaskRunnerRunFixture(t)
		tsk := &TaskRunner{Task: fixture.task, pool: &fixture.pool}
		tsk.Task.WorkflowRunID = &runID

		assert.Error(t, fixture.pool.StoreRemoteTaskOutputs(tsk, document))
		assert.Nil(t, tsk.Task.Artifacts)
	})

	t.Run("valid document is stored in canonical form", func(t *testing.T) {
		fixture := newTaskRunnerRunFixture(t)
		canonical := `{"values":{"image_tag":"1.4.2"},"skipped":{"pw":"sensitive"}}`
		collector := &outputsCollectorFake{validate: func(string) (*string, error) {
			return &canonical, nil
		}}
		fixture.pool.SetTaskOutputsCollector(collector)
		tsk := &TaskRunner{Task: fixture.task, pool: &fixture.pool}
		tsk.Task.WorkflowRunID = &runID

		require.NoError(t, fixture.pool.StoreRemoteTaskOutputs(tsk, document))

		assert.Equal(t, []string{document}, collector.validated)
		require.NotNil(t, tsk.Task.Artifacts)
		assert.Equal(t, canonical, *tsk.Task.Artifacts)
		persisted, err := fixture.store.GetTaskByID(fixture.task.ID)
		require.NoError(t, err)
		require.NotNil(t, persisted.Artifacts)
		assert.Equal(t, canonical, *persisted.Artifacts)
	})

	t.Run("invalid document is an error and nothing is stored", func(t *testing.T) {
		fixture := newTaskRunnerRunFixture(t)
		collector := &outputsCollectorFake{err: errors.New("invalid outputs")}
		fixture.pool.SetTaskOutputsCollector(collector)
		tsk := &TaskRunner{Task: fixture.task, pool: &fixture.pool}
		tsk.Task.WorkflowRunID = &runID

		assert.ErrorContains(t, fixture.pool.StoreRemoteTaskOutputs(tsk, "not json"), "invalid outputs")

		assert.Nil(t, tsk.Task.Artifacts)
		persisted, err := fixture.store.GetTaskByID(fixture.task.ID)
		require.NoError(t, err)
		assert.Nil(t, persisted.Artifacts)
	})

	t.Run("empty document stores nothing", func(t *testing.T) {
		fixture := newTaskRunnerRunFixture(t)
		collector := &outputsCollectorFake{validate: func(string) (*string, error) {
			return nil, nil
		}}
		fixture.pool.SetTaskOutputsCollector(collector)
		tsk := &TaskRunner{Task: fixture.task, pool: &fixture.pool}
		tsk.Task.WorkflowRunID = &runID

		require.NoError(t, fixture.pool.StoreRemoteTaskOutputs(tsk, `{"values":{}}`))

		assert.Nil(t, tsk.Task.Artifacts)
	})
}

func TestApplyDBPersistedTaskSnapshot_CarriesArtifacts(t *testing.T) {
	document := `{"values":{"image_tag":"1.4.2"}}`
	var dst db.Task
	applyDBPersistedTaskSnapshot(&dst, db.Task{Artifacts: &document})
	require.NotNil(t, dst.Artifacts)
	assert.Equal(t, document, *dst.Artifacts)
}
