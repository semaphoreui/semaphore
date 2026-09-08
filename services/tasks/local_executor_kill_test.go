package tasks

import (
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db_lib"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
)

const timeoutEnv = "SEMAPHORE_TEST_PROCESS_TIMEOUT"

// newProcessHelperCommand re-executes the test binary so tests can use a
// portable sleeping process without relying on platform-specific commands such
// as sleep, which is not available on Windows.
func newProcessHelperCommand(timeout time.Duration) *exec.Cmd {
	cmd := exec.Command(os.Args[0], "-test.run=^TestProcessHelper$")
	cmd.Env = append(os.Environ(), timeoutEnv+"="+timeout.String())
	return cmd
}

func TestProcessHelper(t *testing.T) {
	timeoutValue, ok := os.LookupEnv(timeoutEnv)
	if !ok {
		return
	}

	timeout, err := time.ParseDuration(timeoutValue)
	if err != nil {
		t.Fatalf("parse process helper timeout: %v", err)
	}
	if timeout <= 0 {
		t.Fatal("process helper timeout must be positive")
	}

	time.Sleep(timeout)
}

type delayedProcessApp struct {
	runCh     chan struct{}
	killCh    chan struct{}
	processCh chan *os.Process
}

func (a *delayedProcessApp) SetLogger(logger task_logger.Logger) task_logger.Logger {
	return logger
}

func (a *delayedProcessApp) InstallRequirements(db_lib.LocalAppInstallingArgs) error {
	return nil
}

func (a *delayedProcessApp) Run(args db_lib.LocalAppRunningArgs) error {
	close(a.runCh)
	<-a.killCh

	cmd := newProcessHelperCommand(30 * time.Second)
	if err := cmd.Start(); err != nil {
		return err
	}

	args.OnProcessStarted(cmd.Process)
	a.processCh <- cmd.Process
	return cmd.Wait()
}

func (a *delayedProcessApp) Clear() {}

// TestKillBeforeProcessStart asserts that a stop requested after the
// killRequested check is applied when the process is registered. It fails until
// killRequested and Process are synchronized.
func TestKillBeforeProcessStart(t *testing.T) {
	setupExecutorConfig(t)

	app := &delayedProcessApp{
		runCh:     make(chan struct{}),
		killCh:    make(chan struct{}),
		processCh: make(chan *os.Process),
	}
	executor := &LocalExecutor{
		App:      app,
		Logger:   task_logger.NopLogger{},
		prepared: true,
	}
	execCh := make(chan error, 1)
	go func() {
		execCh <- executor.Run("", nil, "")
	}()

	<-app.runCh
	executor.Kill() // Sets killRequested, sees Process == nil, and returns.
	close(app.killCh)

	var process *os.Process

	// wait on process sturtup
	select {
	case process = <-app.processCh:
	case err := <-execCh:
		t.Fatalf("start delayed process: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for delayed process to start")
	}

	// wait on process termination
	select {
	case <-execCh:
		// The pending stop was applied when the process was registered.
	case <-time.After(time.Second):
		_ = process.Kill()
		select {
		case <-execCh:
		case <-time.After(5 * time.Second):
			t.Fatal("LocalExecutor.Run did not return after test cleanup")
		}
		t.Fatal("process survived the stop request made before registration")
	}
}
