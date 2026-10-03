//go:build !windows

package tasks

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db_lib"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test_LocalExecutor_KillIsIdempotent verifies that repeated kill requests
// deliver only one SIGTERM to the command's process group.
//
// Test flow:
//  1. Start LocalExecutor.Run with a Bash command.
//  2. Wait for READY_FILE.
//  3. Call executor.Kill.
//  4. Wait until SIGNAL_LOG contains one TERM.
//  5. Call executor.Kill again.
//  6. Create PROBE_FILE and wait for PROBE_ACK_FILE.
//     This proves the shell ran again after the second call
//     and processed any pending signal.
//  7. Assert SIGNAL_LOG still contains exactly one TERM.
//  8. Create EXIT_FILE so the shell exits without waiting for the 15-second
//     grace period.
//  9. Wait for LocalExecutor.Run to return.
func Test_LocalExecutor_KillIsIdempotent(t *testing.T) {
	previousConfig := util.Config
	util.Config = &util.ConfigType{Process: &util.ConfigProcess{}}
	t.Cleanup(func() {
		util.Config = previousConfig
	})

	dir := t.TempDir()
	pidFile := filepath.Join(dir, "pid")
	readyFile := filepath.Join(dir, "ready")
	signalLog := filepath.Join(dir, "signals")
	probeFile := filepath.Join(dir, "probe")
	probeAckFile := filepath.Join(dir, "probe_ack")
	exitFile := filepath.Join(dir, "exit")

	const script = `
trap 'echo TERM >> "$SIGNAL_LOG"' TERM
echo "$$" > "$PID_FILE"
touch "$READY_FILE"

while [ ! -e "$EXIT_FILE" ]; do
    if [ -e "$PROBE_FILE" ]; then
        touch "$PROBE_ACK_FILE"
    fi
    sleep 0.05
done
`

	logger := task_logger.NopLogger{}
	app := &db_lib.ShellApp{
		Logger:     logger,
		Template:   db.Template{ID: 1},
		Repository: db.Repository{GitURL: dir},
		App:        db.AppBash,
	}
	executor := &LocalExecutor{
		App:      app,
		Logger:   logger,
		prepared: true,
		preparedArgsMap: map[string][]string{
			"default": {"-c", script},
		},
		preparedEnv: []string{
			"PID_FILE=" + pidFile,
			"READY_FILE=" + readyFile,
			"SIGNAL_LOG=" + signalLog,
			"PROBE_FILE=" + probeFile,
			"PROBE_ACK_FILE=" + probeAckFile,
			"EXIT_FILE=" + exitFile,
		},
	}

	runCh := make(chan error, 1)
	runReturned := false
	go func() {
		runCh <- executor.Run("", nil, "")
	}()

	t.Cleanup(func() {
		if runReturned {
			return
		}
		_ = os.WriteFile(exitFile, nil, 0o600)
		if value, err := os.ReadFile(pidFile); err == nil {
			if pid, err := strconv.Atoi(strings.TrimSpace(string(value))); err == nil {
				_ = syscall.Kill(-pid, syscall.SIGKILL)
			}
		}
		select {
		case <-runCh:
		case <-time.After(5 * time.Second):
		}
	})

	require.Eventually(t, func() bool {
		_, err := os.Stat(readyFile)
		return err == nil
	}, 5*time.Second, 10*time.Millisecond, "shell command did not become ready")

	executor.Kill()
	require.Eventually(t, func() bool {
		value, err := os.ReadFile(signalLog)
		return err == nil && len(strings.Fields(string(value))) >= 1
	}, 5*time.Second, 10*time.Millisecond, "shell command did not receive SIGTERM")

	executor.Kill()
	require.NoError(t, os.WriteFile(probeFile, nil, 0o600))
	require.Eventually(t, func() bool {
		_, err := os.Stat(probeAckFile)
		return err == nil
	}, 5*time.Second, 10*time.Millisecond, "shell command did not process the probe")

	value, err := os.ReadFile(signalLog)
	require.NoError(t, err)
	assert.Equal(t, []string{"TERM"}, strings.Fields(string(value)))

	require.NoError(t, os.WriteFile(exitFile, nil, 0o600))
	select {
	case err := <-runCh:
		runReturned = true
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		require.FailNow(t, "LocalExecutor.Run did not return")
	}
}
