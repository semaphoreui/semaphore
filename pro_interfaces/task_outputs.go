package pro_interfaces

// TaskOutputsCollector captures the outputs a workflow task hands to downstream
// nodes: a JSON object the task process writes to the file named by the
// SEMAPHORE_OUTPUTS_FILE environment variable. Workflows are a Pro feature —
// the open-source build wires nil (pro/services/tasks), so callers must
// nil-check; the licensed build provides the implementation
// (pro_impl/services/tasks/artifacts).
type TaskOutputsCollector interface {
	// Begin creates the outputs file of one task in its own directory under
	// dir. The returned capture must be closed once the task has finished.
	Begin(dir string, taskID int) (TaskOutputsCapture, error)
}

// TaskOutputsCapture is the outputs file of one running task.
type TaskOutputsCapture interface {
	// Env returns the environment variables the task process must receive.
	Env() []string

	// AnsibleCallbackDir extracts the callback plugin that turns set_stats data
	// into the outputs file and returns the directory to add to Ansible's
	// callback plugin path.
	AnsibleCallbackDir() (string, error)

	// Collect reads and validates what the task wrote. terraformOutputs is the
	// raw result of `terraform output -json`, or nil for other apps; it is
	// captured leniently, while a file written by the process is validated
	// strictly and makes Collect fail. The returned document is what is stored
	// on the task (nil when there is nothing to store); notes are lines for the
	// task log about outputs that were not captured.
	Collect(terraformOutputs []byte) (document *string, notes []string, err error)

	// Close removes the files created by Begin and AnsibleCallbackDir.
	Close()
}
