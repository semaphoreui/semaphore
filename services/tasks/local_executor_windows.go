//go:build windows

package tasks

import "os"

// Windows does not support POSIX SIGTERM, so terminate the process immediately.
func (t *LocalExecutor) stopProcess(process *os.Process) {
	t.killProcess(process)
}
