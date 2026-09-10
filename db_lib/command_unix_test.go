//go:build !windows

package db_lib

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testFileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func testReadPID(path string) (int, error) {
	value, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(value)))
}

func TestStopCommandRespectsGracePeriod(t *testing.T) {
	cmd := exec.Command("sh", "-c", "sleep 0.1")
	require.NoError(t, cmd.Start())

	pgid, err := syscall.Getpgid(cmd.Process.Pid)
	require.NoError(t, err)
	// The command must not lead its group, so signaling -PID returns ESRCH.
	require.NotEqual(t, cmd.Process.Pid, pgid, "command must not lead its own process group")

	waitCh := make(chan error, 1)
	go func() {
		waitCh <- cmd.Wait()
	}()

	// An immediate Process.Kill would make cmd.Wait return an *exec.ExitError.
	err = stopCommand(cmd, waitCh, task_logger.NopLogger{})
	assert.NoError(t, err)
}

func Test_WaitCommand_CleansProcessGroup(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	childPIDFile := filepath.Join(dir, "child_pid")
	childReadyFile := filepath.Join(dir, "child_ready")
	childSurvivedFile := filepath.Join(dir, "child_survived")

	cmd := exec.Command("sh", "-c", `
(
    # Ignore HUP from parent-shell exit and TERM so only SIGKILL can stop the child.
    trap '' HUP TERM
    touch "$CHILD_READY_FILE"
    sleep 1
    touch "$CHILD_SURVIVED_FILE"
) &
echo "$!" > "$CHILD_PID_FILE"

# Exit as soon as the background child is ready.
while [ ! -e "$CHILD_READY_FILE" ]; do
    sleep 0.1
done
exit 42
`)
	cmd.Env = append(os.Environ(),
		"CHILD_PID_FILE="+childPIDFile,
		"CHILD_READY_FILE="+childReadyFile,
		"CHILD_SURVIVED_FILE="+childSurvivedFile,
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	require.NoError(t, cmd.Start())

	commandExited := false
	t.Cleanup(func() {
		if !commandExited {
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
	})

	exitCh := make(chan error, 1)
	go func() {
		exitCh <- waitCommand(cmd, nil, task_logger.NopLogger{})
	}()

	var commandErr error
	select {
	case commandErr = <-exitCh:
		commandExited = true
	case <-time.After(5 * time.Second):
		require.FailNow(t, "waitCommand did not return after the main command exited")
	}

	// The main command's exit status was preserved.
	var exitErr *exec.ExitError
	require.ErrorAs(t, commandErr, &exitErr)
	assert.Equal(t, 42, exitErr.ExitCode())

	// The background child was started.
	childPID, err := testReadPID(childPIDFile)
	require.NoError(t, err, "read background child PID")

	// The background child did not survive process-group cleanup.
	require.Eventually(t, func() bool {
		return errors.Is(syscall.Kill(childPID, 0), syscall.ESRCH)
	}, 1500*time.Millisecond, 15*time.Millisecond, "background child still exists after process-group cleanup")
	assert.NoFileExists(t, childSurvivedFile)
}

func Test_StopCommand_KillsTermResistantChild(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	mainReadyFile := filepath.Join(dir, "main_ready")
	childReadyFile := filepath.Join(dir, "child_ready")
	childPIDFile := filepath.Join(dir, "child_pid")
	termFile := filepath.Join(dir, "term_received")
	childSurvivedFile := filepath.Join(dir, "child_survived")

	cmd := exec.Command("sh", "-c", `
trap 'touch "$TERM_FILE"; exit 0' TERM

(
    # Ignore HUP from parent-shell exit and TERM so only SIGKILL can stop the child.
    trap '' HUP TERM
    touch "$CHILD_READY_FILE"
    sleep 1
    touch "$CHILD_SURVIVED_FILE"
) &
echo "$!" > "$CHILD_PID_FILE"

touch "$MAIN_READY_FILE"

# Keep the main script running until the test requests cancellation.
while true; do
    sleep 1
done
`)
	cmd.Env = append(os.Environ(),
		"MAIN_READY_FILE="+mainReadyFile,
		"CHILD_READY_FILE="+childReadyFile,
		"CHILD_PID_FILE="+childPIDFile,
		"TERM_FILE="+termFile,
		"CHILD_SURVIVED_FILE="+childSurvivedFile,
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	require.NoError(t, cmd.Start())

	commandExited := false
	t.Cleanup(func() {
		if !commandExited {
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
	})

	stopCh := make(chan struct{})
	exitCh := make(chan error, 1)
	go func() {
		exitCh <- waitCommand(cmd, stopCh, task_logger.NopLogger{})
	}()

	var childPID int
	require.Eventually(t, func() bool {
		pid, err := testReadPID(childPIDFile)
		if err != nil || !testFileExists(mainReadyFile) || !testFileExists(childReadyFile) {
			return false
		}
		childPID = pid
		return true
	}, 5*time.Second, 15*time.Millisecond, "command did not finish installing its signal handlers")

	// Request graceful shutdown through the stop channel.
	close(stopCh)

	select {
	case err := <-exitCh:
		commandExited = true
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		require.FailNow(t, "waitCommand did not return after cancellation")
	}

	assert.FileExists(t, termFile, "main command did not receive SIGTERM")
	require.Eventually(t, func() bool {
		return errors.Is(syscall.Kill(childPID, 0), syscall.ESRCH)
	}, 1500*time.Millisecond, 15*time.Millisecond, "SIGTERM-resistant child still exists after process-group cleanup")
	assert.NoFileExists(t, childSurvivedFile)
}
