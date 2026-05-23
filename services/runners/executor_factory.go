package runners

import (
	"fmt"

	"github.com/semaphoreui/semaphore/db_lib"
	"github.com/semaphoreui/semaphore/services/tasks"
	"github.com/semaphoreui/semaphore/util"
)

// ExecutorType identifies the strategy the runner uses to execute each task. The default
// "local" executor runs tasks as subprocesses on the runner host. Future values (e.g.
// "kubernetes") will dispatch each task into an ephemeral pod, GitLab-runner-style.
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

// newExecutor builds the per-task executor the runner will dispatch into. The result is
// a *tasks.LocalExecutor today (the only implementation); the factory exists so adding a
// KubernetesExecutor in a future phase is a single switch arm here, not a series of
// edits in checkNewJobs.
//
// Returning the concrete type (rather than the tasks.Executor interface) is intentional
// for Phase 1: downstream code in this package still touches LocalExecutor's fields
// (Task, Template, Logger, App, Repository, …) directly. Lifting that to the interface
// is a follow-up step, scheduled for the phase that adds the second executor.
func newExecutor(jobData JobData, keyInstaller db_lib.AccessKeyInstaller) (*tasks.LocalExecutor, error) {
	executorType := resolveExecutorType()

	switch executorType {
	case ExecutorTypeLocal:
		return newLocalExecutor(jobData, keyInstaller), nil
	case ExecutorTypeKubernetes:
		return nil, fmt.Errorf("kubernetes executor is not implemented yet")
	default:
		return nil, fmt.Errorf("unknown runner executor type %q", executorType)
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

