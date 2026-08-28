package db_lib

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type outputCaptureLogger struct {
	task_logger.NopLogger
	stdout bytes.Buffer
	stderr bytes.Buffer
}

func (l *outputCaptureLogger) LogCmd(cmd *exec.Cmd) func() {
	cmd.Stdout = &l.stdout
	cmd.Stderr = &l.stderr
	return func() {}
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
	assert.Equal(t, "stdout message", logger.stdout.String())
	assert.Equal(t, "stderr message", logger.stderr.String())
}

func TestShellApp_RunBoundsOutputDrainWhenBackgroundChildKeepsPipesOpen(t *testing.T) {
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash is not available")
	}

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

	startedAt := time.Now()
	err := app.Run(LocalAppRunningArgs{
		CliArgs: map[string][]string{
			"default": {"-c", "printf 'stdout message'; printf 'stderr message' >&2; sleep 2 &"},
		},
		Callback: func(*os.Process) {},
	})

	require.NoError(t, err)
	assert.Equal(t, "stdout message", logger.stdout.String())
	assert.Equal(t, "stderr message", logger.stderr.String())
	assert.Less(t, time.Since(startedAt), 1500*time.Millisecond)
}
