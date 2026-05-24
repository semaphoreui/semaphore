package runners

import (
	"fmt"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db_lib"
	"github.com/semaphoreui/semaphore/services/tasks"
	"github.com/semaphoreui/semaphore/services/tasks/k8s"
	"github.com/semaphoreui/semaphore/util"
)

// ExecutorType identifies the strategy the runner uses to execute each task. The default
// "local" executor runs tasks as subprocesses on the runner host. "kubernetes" dispatches
// each task into an ephemeral pod, GitLab-runner-style.
type ExecutorType string

const (
	ExecutorTypeLocal      ExecutorType = "local"
	ExecutorTypeKubernetes ExecutorType = "kubernetes"
)

// resolveExecutorType returns the executor type configured for this runner process. Empty
// (or unset) config falls back to local so existing deployments keep their behaviour.
func resolveExecutorType() ExecutorType {
	if util.Config == nil || util.Config.Runner == nil {
		return ExecutorTypeLocal
	}
	t := ExecutorType(util.Config.Runner.Executor)
	if t == "" {
		return ExecutorTypeLocal
	}
	return t
}

// newExecutor builds the per-task executor the runner will dispatch into. Access keys
// retrieved from the server are wired into the JobData up front so each concrete
// constructor sees a fully-hydrated record — this keeps job_pool.go agnostic of which
// executor it ends up dispatching to.
func newExecutor(
	jobData JobData,
	accessKeys map[int]db.AccessKey,
	keyInstaller db_lib.AccessKeyInstaller,
	k8sCfg *k8s.Config,
) (tasks.Executor, error) {
	hydrateJobAccessKeys(&jobData, accessKeys)

	switch resolveExecutorType() {
	case ExecutorTypeLocal:
		return newLocalExecutor(jobData, keyInstaller), nil
	case ExecutorTypeKubernetes:
		if k8sCfg == nil {
			return nil, fmt.Errorf("kubernetes executor requested but k8s config is not initialized")
		}
		return newK8sExecutor(jobData, *k8sCfg), nil
	default:
		return nil, fmt.Errorf("unknown runner executor type %q", resolveExecutorType())
	}
}

func newLocalExecutor(jobData JobData, keyInstaller db_lib.AccessKeyInstaller) *tasks.LocalExecutor {
	return &tasks.LocalExecutor{
		Task:         jobData.Task,
		Template:     jobData.Template,
		Inventory:    jobData.Inventory,
		Repository:   jobData.Repository,
		Environment:  jobData.Environment,
		KeyInstaller: keyInstaller,
		App: db_lib.CreateApp(
			jobData.Template,
			jobData.Repository,
			jobData.Inventory,
			nil),
	}
}

func newK8sExecutor(jobData JobData, cfg k8s.Config) *k8s.Executor {
	return k8s.New(cfg, jobData.Task, jobData.Template, jobData.Inventory, jobData.Repository, jobData.Environment)
}

// hydrateJobAccessKeys decrypts/wires the access keys the server sent us into the
// per-task data. Lives in the factory (not the per-executor constructor) so the
// behaviour is identical across executor types: ansible vault passwords, SSH keys,
// inventory keys, and the inventory repo SSH key all get attached to the JobData
// regardless of whether the task ends up running locally or in a Pod.
func hydrateJobAccessKeys(jobData *JobData, accessKeys map[int]db.AccessKey) {
	jobData.Repository.SSHKey = accessKeys[jobData.Repository.SSHKeyID]

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
}
