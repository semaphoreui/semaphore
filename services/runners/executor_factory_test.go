package runners

import (
	"errors"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pro/services/tasks/docker"
	"github.com/semaphoreui/semaphore/pro/services/tasks/k8s"
	"github.com/semaphoreui/semaphore/services/tasks"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveExecutorType_Default(t *testing.T) {
	tests := []struct {
		name string
		cfg  *util.ExecutorConfig
		want util.ExecutorType
	}{
		{"nil config falls back to local", nil, util.ExecutorTypeLocal},
		{"empty type falls back to local", &util.ExecutorConfig{Type: ""}, util.ExecutorTypeLocal},
		{"explicit local is honored", &util.ExecutorConfig{Type: util.ExecutorTypeLocal}, util.ExecutorTypeLocal},
		{"kubernetes is honored", &util.ExecutorConfig{Type: util.ExecutorTypeKubernetes}, util.ExecutorTypeKubernetes},
		{"docker is honored", &util.ExecutorConfig{Type: util.ExecutorTypeDocker}, util.ExecutorTypeDocker},
		{"unknown values pass through (factory rejects them)", &util.ExecutorConfig{Type: util.ExecutorType("nomad")}, util.ExecutorType("nomad")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, resolveExecutorType(tt.cfg))
		})
	}
}

func TestNewExecutorProvider_Local(t *testing.T) {
	provider, err := newExecutorProvider(&util.ExecutorConfig{Type: util.ExecutorTypeLocal}, nil)
	require.NoError(t, err)
	_, ok := provider.(*tasks.LocalExecutorProvider)
	assert.True(t, ok, "local executor type must yield a LocalExecutorProvider")
}

func TestNewExecutorProvider_NilConfigDefaultsToLocal(t *testing.T) {
	// Old runner deployments may have no Executor block at all; we must not refuse
	// to start them — that would be a silent regression.
	provider, err := newExecutorProvider(nil, nil)
	require.NoError(t, err)
	_, ok := provider.(*tasks.LocalExecutorProvider)
	assert.True(t, ok)
}

func TestNewExecutorProvider_KubernetesStubBuildErrors(t *testing.T) {
	// In the OSS stub build (what these tests run against) the K8s provider
	// constructor always errors with ErrNotAvailable. The proprietary build returns
	// a real provider — different test suite, lives in pro_impl.
	provider, err := newExecutorProvider(&util.ExecutorConfig{Type: util.ExecutorTypeKubernetes}, nil)
	assert.Nil(t, provider)
	require.Error(t, err)
	assert.True(t, errors.Is(err, k8s.ErrNotAvailable),
		"OSS stub must surface k8s.ErrNotAvailable so operators see why the runner refuses k8s jobs")
}

func TestNewExecutorProvider_DockerStubBuildErrors(t *testing.T) {
	// Like the K8s case: in the OSS stub build the Docker provider constructor always
	// errors with ErrNotAvailable. The proprietary build returns a real provider.
	provider, err := newExecutorProvider(&util.ExecutorConfig{Type: util.ExecutorTypeDocker}, nil)
	assert.Nil(t, provider)
	require.Error(t, err)
	assert.True(t, errors.Is(err, docker.ErrNotAvailable),
		"OSS stub must surface docker.ErrNotAvailable so operators see why the runner refuses docker jobs")
}

func TestNewExecutorProvider_UnknownType(t *testing.T) {
	provider, err := newExecutorProvider(&util.ExecutorConfig{Type: util.ExecutorType("nomad")}, nil)
	assert.Nil(t, provider)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nomad", "error must name the offending type")
}

func TestNewExecutor_DispatchesToProvider(t *testing.T) {
	provider, err := newExecutorProvider(&util.ExecutorConfig{Type: util.ExecutorTypeLocal}, nil)
	require.NoError(t, err)

	jobData := JobData{
		Task:       db.Task{ID: 42},
		Template:   db.Template{ID: 7, App: db.AppAnsible},
		Inventory:  db.Inventory{ID: 3},
		Repository: db.Repository{ID: 5},
	}

	exec, err := newExecutor(jobData, nil, provider)
	require.NoError(t, err)
	require.NotNil(t, exec)

	local, ok := exec.(*tasks.LocalExecutor)
	require.True(t, ok, "local provider must return a *LocalExecutor")
	assert.Equal(t, 42, local.Task.ID)
	assert.Equal(t, 7, local.Template.ID)
	assert.Equal(t, 3, local.Inventory.ID)
	assert.Equal(t, 5, local.Repository.ID)
	assert.NotNil(t, local.App, "provider must populate App so Prepare has somewhere to install requirements")
}

func TestNewExecutor_RejectsNilProvider(t *testing.T) {
	// JobPool may end up with a nil provider when the runner config is malformed at
	// startup. Dispatch must refuse cleanly with a useful message instead of
	// panicking on a nil-interface call.
	exec, err := newExecutor(JobData{}, nil, nil)
	assert.Nil(t, exec)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "provider")
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
