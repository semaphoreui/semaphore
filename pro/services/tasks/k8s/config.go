// Package k8s is an open-source stub for the Kubernetes-backed task executor.
//
// The real implementation lives in the pro_impl module — when that module is swapped
// in via the top-level go.mod replace directive (or a workspace), the same import
// path resolves to a fully functional executor that runs each task in an ephemeral
// Pod. In the open-source build this stub keeps the executor_factory and job_pool
// imports compiling while making it explicit that "kubernetes" as a runner executor
// type requires the proprietary build.
//
// Stub behaviour:
//   - ConfigFromRunnerConfig always returns an error so JobPool refuses to start the
//     Kubernetes executor and surfaces a clear message to the operator.
//   - Executor methods never panic — they error out cleanly if something does reach
//     them (which shouldn't happen because Config initialisation already failed).
package k8s

import (
	"errors"

	"github.com/semaphoreui/semaphore/util"
)

// ErrNotAvailable is returned anywhere a Kubernetes-executor codepath is hit in the
// OSS build. The proprietary build replaces this whole package with a working one.
var ErrNotAvailable = errors.New("k8s executor: kubernetes executor is not available in this build")

// Config mirrors the proprietary type's exported shape so call sites in
// services/runners can keep referring to *k8s.Config without conditional compilation.
// The single field is intentionally untyped (any) so this stub does not pull k8s.io
// client-go into the OSS module's dependency tree.
type Config struct {
	Clientset      any
	Namespace      string
	Image          string
	HelperImage    string
	ServiceAccount string
	PullSecrets    []string
}

// ConfigFromRunnerConfig always fails in the stub. JobPool logs the error and leaves
// its k8sConfig nil, which makes the executor factory reject any subsequent
// Kubernetes task with the same error.
func ConfigFromRunnerConfig(_ util.RunnerK8sConfig) (Config, error) {
	return Config{}, ErrNotAvailable
}
