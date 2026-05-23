package runners

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// restoreConfig captures the global util.Config so tests that mutate it can restore
// the original between runs. The runner-side code reads util.Config eagerly, so leaking
// state between tests is a real risk.
func restoreConfig(t *testing.T) {
	original := util.Config
	t.Cleanup(func() {
		util.Config = original
	})
}

func TestResolveExecutorType_Default(t *testing.T) {
	restoreConfig(t)

	tests := []struct {
		name   string
		config *util.ConfigType
		want   ExecutorType
	}{
		{
			name:   "nil config falls back to local",
			config: nil,
			want:   ExecutorTypeLocal,
		},
		{
			name:   "nil Runner falls back to local",
			config: &util.ConfigType{},
			want:   ExecutorTypeLocal,
		},
		{
			name:   "empty Executor field falls back to local",
			config: &util.ConfigType{Runner: &util.RunnerConfig{Executor: ""}},
			want:   ExecutorTypeLocal,
		},
		{
			name:   "explicit local is honored",
			config: &util.ConfigType{Runner: &util.RunnerConfig{Executor: "local"}},
			want:   ExecutorTypeLocal,
		},
		{
			name:   "kubernetes is honored",
			config: &util.ConfigType{Runner: &util.RunnerConfig{Executor: "kubernetes"}},
			want:   ExecutorTypeKubernetes,
		},
		{
			name:   "unknown values pass through (factory rejects them)",
			config: &util.ConfigType{Runner: &util.RunnerConfig{Executor: "docker"}},
			want:   ExecutorType("docker"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			util.Config = tt.config
			assert.Equal(t, tt.want, resolveExecutorType())
		})
	}
}

func TestNewExecutor_Local(t *testing.T) {
	restoreConfig(t)
	util.Config = &util.ConfigType{Runner: &util.RunnerConfig{Executor: "local"}}

	jobData := JobData{
		Task:       db.Task{ID: 42},
		Template:   db.Template{ID: 7, App: db.AppAnsible},
		Inventory:  db.Inventory{ID: 3},
		Repository: db.Repository{ID: 5},
	}

	exec, err := newExecutor(jobData, nil)
	require.NoError(t, err)
	require.NotNil(t, exec)

	assert.Equal(t, 42, exec.Task.ID)
	assert.Equal(t, 7, exec.Template.ID)
	assert.Equal(t, 3, exec.Inventory.ID)
	assert.Equal(t, 5, exec.Repository.ID)
	assert.NotNil(t, exec.App, "factory must populate App so prepareRun has somewhere to install requirements")
}

func TestNewExecutor_Kubernetes_NotImplemented(t *testing.T) {
	restoreConfig(t)
	util.Config = &util.ConfigType{Runner: &util.RunnerConfig{Executor: "kubernetes"}}

	exec, err := newExecutor(JobData{}, nil)

	assert.Nil(t, exec)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "kubernetes")
}

func TestNewExecutor_UnknownType(t *testing.T) {
	restoreConfig(t)
	util.Config = &util.ConfigType{Runner: &util.RunnerConfig{Executor: "docker"}}

	exec, err := newExecutor(JobData{}, nil)

	assert.Nil(t, exec)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "docker", "error should name the offending type")
}
