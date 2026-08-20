package db_lib

import (
	"io"
	"os"
	"os/exec"
	"sync"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type outputCaptureLogger struct {
	task_logger.NopLogger
	wg        sync.WaitGroup
	stdout    []byte
	stderr    []byte
	stdoutErr error
	stderrErr error
}

func (l *outputCaptureLogger) LogCmd(cmd *exec.Cmd) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		l.stdoutErr = err
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		l.stderrErr = err
		return
	}

	l.wg.Add(2)
	go func() {
		defer l.wg.Done()
		l.stdout, l.stdoutErr = io.ReadAll(stdout)
	}()
	go func() {
		defer l.wg.Done()
		l.stderr, l.stderrErr = io.ReadAll(stderr)
	}()
}

func (l *outputCaptureLogger) WaitLog() {
	l.wg.Wait()
}

func TestShellApp_RunDrainsOutputBeforeWait(t *testing.T) {
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash is not available")
	}

	// ShellApp reads the global config; do not run this test in parallel.
	previousConfig := util.Config
	util.Config = &util.ConfigType{
		Process: &util.ConfigProcess{},
	}
	t.Cleanup(func() { util.Config = previousConfig })

	logger := &outputCaptureLogger{}
	app := &ShellApp{
		Logger:     logger,
		Template:   db.Template{ID: 1},
		Repository: db.Repository{GitURL: t.TempDir()},
		App:        db.AppBash,
	}

	err := app.Run(LocalAppRunningArgs{
		CliArgs:  map[string][]string{"default": {"-c", "printf 'stdout message'; printf 'stderr message' >&2"}},
		Callback: func(*os.Process) {},
	})

	require.NoError(t, err)
	require.NoError(t, logger.stdoutErr)
	require.NoError(t, logger.stderrErr)
	assert.Equal(t, "stdout message", string(logger.stdout))
	assert.Equal(t, "stderr message", string(logger.stderr))
}
