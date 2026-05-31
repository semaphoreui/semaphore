// Package docker is an open-source stub for the Docker-backed ExecutorProvider.
//
// The real implementation lives in the pro_impl module. When that module is swapped
// in via the top-level go.work file (or the go.mod replace directive), the same import
// path resolves to a fully functional provider that runs each task in an ephemeral
// container. In the open-source build this stub keeps the executor_factory and job_pool
// imports compiling while making it explicit that the "docker" runner executor type
// requires the proprietary build.
//
// The only entry point is NewProvider; it always fails so JobPool refuses to start
// and the operator sees a clear message in the logs.
package docker

import (
	"github.com/semaphoreui/semaphore/services/tasks"
	"github.com/semaphoreui/semaphore/util"
)

func NewProvider(_ util.RunnerDockerConfig) (tasks.ExecutorProvider, error) {
	return nil, nil
}
