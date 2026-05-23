package tasks

// Executor encapsulates the strategy used to run a single task. The job pool dispatches
// every queued task to an Executor, which orchestrates the full lifecycle: preparation
// of the working environment, execution of the underlying tool (ansible-playbook,
// terraform, shell, ...), and cleanup of any resources it allocated.
//
// Today the only implementation is LocalExecutor, which runs the tool as a subprocess
// on the runner host. The interface exists so future implementations (e.g. a Kubernetes
// executor that runs each task in an ephemeral Pod, GitLab-runner-style) can plug into
// the same job pool without changes to TaskRunner / job_pool.
//
// Execution model:
//   - Run is the single entry point used by the job pool. Implementations are free to
//     organize the work internally; LocalExecutor calls Prepare → run-the-app → Cleanup.
//   - Prepare and Cleanup are exposed so that callers that want to observe the lifecycle
//     (status updates, phased progress reporting) can do so. Run remains the default
//     orchestrator and must call them itself when invoked directly.
//   - Kill is invoked from a different goroutine when a stop is requested by the user.
type Executor interface {
	Job

	// Prepare materializes everything required before the underlying tool can start:
	// project tmp dir, repository checkout, inventory file, installed access keys and
	// vault password files. The alias is needed for Terraform tasks (it shapes the
	// TF_HTTP_ADDRESS env var); pass "" for non-Terraform apps. Calling Prepare twice
	// on the same executor is a no-op.
	Prepare(username string, incomingVersion *string, alias string) error

	// Cleanup releases every resource Prepare allocated. It is invoked unconditionally
	// at the end of Run (via defer); implementations must tolerate being called even
	// when Prepare failed partway through.
	Cleanup()
}
