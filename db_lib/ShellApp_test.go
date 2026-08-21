package db_lib

import (
	"io"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
)

type shellAppCaptureLogger struct {
	task_logger.NopLogger
	wg sync.WaitGroup
}

func (l *shellAppCaptureLogger) LogCmd(cmd *exec.Cmd) {
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}

	l.wg.Add(2)
	go func() {
		defer l.wg.Done()
		_, _ = io.Copy(io.Discard, stderr)
	}()
	go func() {
		defer l.wg.Done()
		_, _ = io.Copy(io.Discard, stdout)
	}()
}

func (l *shellAppCaptureLogger) WaitLog() {
	l.wg.Wait()
}

func TestShellAppRunDoesNotWaitForBackgroundChildOutput(t *testing.T) {
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash is not available")
	}

	previousConfig := util.Config
	util.Config = &util.ConfigType{
		Process: &util.ConfigProcess{},
	}
	t.Cleanup(func() {
		util.Config = previousConfig
	})

	app := &ShellApp{
		Logger:     &shellAppCaptureLogger{},
		Template:   db.Template{ID: 1},
		Repository: db.Repository{GitURL: t.TempDir()},
		App:        db.AppBash,
	}

	startedAt := time.Now()
	err := app.Run(LocalAppRunningArgs{
		CliArgs: map[string][]string{
			"default": {"-c", "sleep 5 &"},
		},
		Callback: func(*os.Process) {},
	})
	elapsed := time.Since(startedAt)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if elapsed >= 3*time.Second {
		t.Fatalf("expected command to finish before background child exits, took %s", elapsed)
	}
}
