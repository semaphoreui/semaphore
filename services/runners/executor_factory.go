package runners

import (
	"fmt"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db_lib"
	"github.com/semaphoreui/semaphore/pro/services/tasks/docker"
	"github.com/semaphoreui/semaphore/pro/services/tasks/k8s"
	"github.com/semaphoreui/semaphore/services/tasks"
	"github.com/semaphoreui/semaphore/util"
)

// newExecutorProvider picks the ExecutorProvider implementation that matches the
// runner's configured executor type. Called once at runner startup; the resulting
// Provider lives on the JobPool for the process lifetime.
//
// Adding a new executor strategy means: write a Provider, add a case here. Nothing
// else (JobPool, executor lifecycle, access-key hydration) needs to change.
func newExecutorProvider(executorCfg *util.ExecutorConfig, keyInstaller db_lib.AccessKeyInstaller) (tasks.ExecutorProvider, error) {
	switch resolveExecutorType(executorCfg) {
	case util.ExecutorTypeLocal:
		return tasks.NewLocalExecutorProvider(keyInstaller), nil
	case util.ExecutorTypeKubernetes:
		k8sCfg := util.RunnerK8sConfig{}
		if executorCfg != nil {
			k8sCfg = executorCfg.K8s
		}
		return k8s.NewProvider(k8sCfg)
	case util.ExecutorTypeDocker:
		dockerCfg := util.RunnerDockerConfig{}
		if executorCfg != nil {
			dockerCfg = executorCfg.Docker
		}
		return docker.NewProvider(dockerCfg)
	default:
		return nil, fmt.Errorf("unknown runner executor type %q", resolveExecutorType(executorCfg))
	}
}

// resolveExecutorType returns the executor type from config, defaulting to "local"
// when the field is missing or empty. Defaulting in one place keeps the rest of the
// runner code free of nil-checks against the config tree.
func resolveExecutorType(executorCfg *util.ExecutorConfig) util.ExecutorType {
	if executorCfg == nil || executorCfg.Type == "" {
		return util.ExecutorTypeLocal
	}
	return executorCfg.Type
}

// newExecutor wires per-task data through the Provider. Access keys are hydrated
// here (not inside each Provider) so the behaviour is identical regardless of
// strategy: ansible vault passwords, SSH keys, inventory keys, and the inventory
// repo SSH key all land on the JobData before the Executor is built.
func newExecutor(
	jobData JobData,
	accessKeys map[int]db.AccessKey,
	provider tasks.ExecutorProvider,
) (tasks.Executor, error) {
	if provider == nil {
		return nil, fmt.Errorf("executor provider is not initialised (check runner executor config)")
	}

	hydrateJobAccessKeys(&jobData, accessKeys)

	return provider.NewExecutor(
		jobData.Task,
		jobData.Template,
		jobData.Inventory,
		jobData.Repository,
		jobData.SubmoduleCredentials,
		jobData.Environment,
		jobData.JWT,
	)
}

// hydrateJobAccessKeys decrypts/wires the access keys the server sent us into the
// per-task data. Lives in the factory (not the Provider) so the behaviour is
// identical across strategies — every executor sees the same shape of JobData.
func hydrateJobAccessKeys(jobData *JobData, accessKeys map[int]db.AccessKey) {
	jobData.Repository.SSHKey = accessKeys[jobData.Repository.SSHKeyID]

	for i := range jobData.SubmoduleCredentials {
		jobData.SubmoduleCredentials[i].AccessKey = accessKeys[jobData.SubmoduleCredentials[i].AccessKeyID]
	}

	if jobData.Inventory.SSHKeyID != nil {
		jobData.Inventory.SSHKey = accessKeys[*jobData.Inventory.SSHKeyID]
	}
	if jobData.Inventory.BecomeKeyID != nil {
		jobData.Inventory.BecomeKey = accessKeys[*jobData.Inventory.BecomeKeyID]
	}

	if jobData.Template.Vaults != nil {
		vaults := make([]db.TemplateVault, 0, len(jobData.Template.Vaults))
		for _, vault := range jobData.Template.Vaults {
			v := vault
			if v.VaultKeyID != nil {
				key := accessKeys[*v.VaultKeyID]
				v.Vault = &key
			}
			vaults = append(vaults, v)
		}
		jobData.Template.Vaults = vaults
	}

	if jobData.Inventory.RepositoryID != nil && jobData.Inventory.Repository != nil {
		jobData.Inventory.Repository.SSHKey = accessKeys[jobData.Inventory.Repository.SSHKeyID]
	}

	for i := range jobData.Inventory.SubmoduleCredentials {
		jobData.Inventory.SubmoduleCredentials[i].AccessKey = accessKeys[jobData.Inventory.SubmoduleCredentials[i].AccessKeyID]
	}
}
