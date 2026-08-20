package db_lib

import (
	"io"
	"os"
	"os/exec"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stdoutCaptureLogger struct {
	task_logger.NopLogger
	pipe   io.ReadCloser
	output []byte
	err    error
}

func (l *stdoutCaptureLogger) LogCmd(cmd *exec.Cmd) {
	l.pipe, l.err = cmd.StdoutPipe()
}

func (l *stdoutCaptureLogger) WaitLog() {
	if l.err == nil {
		l.output, l.err = io.ReadAll(l.pipe)
	}
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

	logger := &stdoutCaptureLogger{}
	app := &ShellApp{
		Logger:     logger,
		Template:   db.Template{ID: 1},
		Repository: db.Repository{GitURL: t.TempDir()},
		App:        db.AppBash,
	}

	err := app.Run(LocalAppRunningArgs{
		CliArgs:  map[string][]string{"default": {"-c", "printf 'stdout message'"}},
		Callback: func(*os.Process) {},
	})

	require.NoError(t, err)
	require.NoError(t, logger.err)
	assert.Equal(t, "stdout message", string(logger.output))
}
