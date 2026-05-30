// Package k8s is an open-source stub for the Kubernetes-backed ExecutorProvider.
//
// The real implementation lives in the pro_impl module. When that module is swapped
// in via the top-level go.mod replace directive (or a workspace), the same import
// path resolves to a fully functional provider that runs each task in an ephemeral
// Pod. In the open-source build this stub keeps the executor_factory and job_pool
// imports compiling while making it explicit that the "k8s" runner executor type
// requires the proprietary build.
//
// The only entry point is NewProvider; it always fails so JobPool refuses to start
// and the operator sees a clear message in the logs.
package k8s

import (
	"errors"

	"github.com/semaphoreui/semaphore/services/tasks"
	"github.com/semaphoreui/semaphore/util"
)

// ErrNotAvailable is returned by NewProvider in the OSS build. The proprietary build
// replaces this whole package with a working one.
var ErrNotAvailable = errors.New("k8s executor: kubernetes executor is not available in this build")

// NewProvider always returns ErrNotAvailable. Callers (job_pool, executor_factory)
// log the error and treat the provider as unavailable — no panics, no half-built
// state, just a clean refusal to dispatch Kubernetes tasks.
func NewProvider(_ util.RunnerK8sConfig) (tasks.ExecutorProvider, error) {
	return nil, ErrNotAvailable
}
