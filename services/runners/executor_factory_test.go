package runners

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/tasks"
	"github.com/semaphoreui/semaphore/services/tasks/k8s"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes/fake"
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

	exec, err := newExecutor(jobData, nil, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, exec)

	local, ok := exec.(*tasks.LocalExecutor)
	require.True(t, ok, "local executor type expected when config is local")
	assert.Equal(t, 42, local.Task.ID)
	assert.Equal(t, 7, local.Template.ID)
	assert.Equal(t, 3, local.Inventory.ID)
	assert.Equal(t, 5, local.Repository.ID)
	assert.NotNil(t, local.App, "factory must populate App so prepareRun has somewhere to install requirements")
}

func TestNewExecutor_Kubernetes(t *testing.T) {
	restoreConfig(t)
	util.Config = &util.ConfigType{Runner: &util.RunnerConfig{Executor: "kubernetes"}}

	k8sCfg := &k8s.Config{
		Clientset: fake.NewSimpleClientset(),
		Namespace: "semaphore-test",
		Image:     "alpine:latest",
	}

	jobData := JobData{
		Task:     db.Task{ID: 99},
		Template: db.Template{ID: 1, App: db.AppAnsible},
	}

	exec, err := newExecutor(jobData, nil, nil, k8sCfg)
	require.NoError(t, err)
	require.NotNil(t, exec)

	_, ok := exec.(*k8s.Executor)
	assert.True(t, ok, "kubernetes executor type expected when config is kubernetes")
}

func TestNewExecutor_KubernetesMissingConfig(t *testing.T) {
	restoreConfig(t)
	util.Config = &util.ConfigType{Runner: &util.RunnerConfig{Executor: "kubernetes"}}

	exec, err := newExecutor(JobData{}, nil, nil, nil)

	assert.Nil(t, exec)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "k8s config")
}

func TestNewExecutor_UnknownType(t *testing.T) {
	restoreConfig(t)
	util.Config = &util.ConfigType{Runner: &util.RunnerConfig{Executor: "docker"}}

	exec, err := newExecutor(JobData{}, nil, nil, nil)

	assert.Nil(t, exec)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "docker", "error should name the offending type")
}

func TestHydrateJobAccessKeys_WiresKeysIntoJobData(t *testing.T) {
	sshKeyID := 11
	becomeKeyID := 22
	vaultKeyID := 33
	inventoryRepoID := 44
	inventoryRepoSSHKeyID := 55

	accessKeys := map[int]db.AccessKey{
		7:                     {ID: 7, Type: db.AccessKeySSH}, // template repo
		sshKeyID:              {ID: sshKeyID, Type: db.AccessKeySSH},
		becomeKeyID:           {ID: becomeKeyID, Type: db.AccessKeyLoginPassword},
		vaultKeyID:            {ID: vaultKeyID, Type: db.AccessKeyLoginPassword},
		inventoryRepoSSHKeyID: {ID: inventoryRepoSSHKeyID, Type: db.AccessKeySSH},
	}

	jobData := JobData{
		Repository: db.Repository{SSHKeyID: 7},
		Inventory: db.Inventory{
			SSHKeyID:     &sshKeyID,
			BecomeKeyID:  &becomeKeyID,
			RepositoryID: &inventoryRepoID,
			Repository: &db.Repository{
				ID:       inventoryRepoID,
				SSHKeyID: inventoryRepoSSHKeyID,
			},
		},
		Template: db.Template{
			Vaults: []db.TemplateVault{{VaultKeyID: &vaultKeyID}},
		},
	}

	hydrateJobAccessKeys(&jobData, accessKeys)

	assert.Equal(t, 7, jobData.Repository.SSHKey.ID, "template repo SSH key wired")
	assert.Equal(t, sshKeyID, jobData.Inventory.SSHKey.ID, "inventory SSH key wired")
	assert.Equal(t, becomeKeyID, jobData.Inventory.BecomeKey.ID, "inventory become key wired")
	require.Len(t, jobData.Template.Vaults, 1)
	require.NotNil(t, jobData.Template.Vaults[0].Vault)
	assert.Equal(t, vaultKeyID, jobData.Template.Vaults[0].Vault.ID, "vault key wired")
	assert.Equal(t, inventoryRepoSSHKeyID, jobData.Inventory.Repository.SSHKey.ID, "inventory repo SSH key wired")
}
