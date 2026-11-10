package tasks

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ansibleTaskRepoFake struct {
	db.AnsibleTaskRepository
	calls     int
	projectID int
	taskID    int
	hosts     []db.AnsibleTaskHost
}

func (f *ansibleTaskRepoFake) ReplaceAnsibleTaskHosts(projectID int, taskID int, hosts []db.AnsibleTaskHost) error {
	f.calls++
	f.projectID = projectID
	f.taskID = taskID
	f.hosts = hosts
	return nil
}

func newHostsPool(repo *ansibleTaskRepoFake) (*TaskPool, *TaskRunner) {
	tsk := &TaskRunner{Task: db.Task{ID: 42, ProjectID: 7}}
	return &TaskPool{ansibleTaskRepo: repo}, tsk
}

func TestTaskPool_StoreTaskHosts(t *testing.T) {
	repo := &ansibleTaskRepoFake{}
	pool, tsk := newHostsPool(repo)
	document := `{"version":1,"hosts":{"web-02":{"ok":1},"web-01":{"ok":2,"failed":1}}}`

	require.NoError(t, pool.StoreTaskHosts(tsk, document))

	assert.Equal(t, 1, repo.calls)
	assert.Equal(t, 7, repo.projectID)
	assert.Equal(t, 42, repo.taskID)
	require.Len(t, repo.hosts, 2)
	assert.Equal(t, "web-01", repo.hosts[0].Host)
	assert.Equal(t, "web-02", repo.hosts[1].Host)
}
