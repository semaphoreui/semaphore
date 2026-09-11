package tasks

import (
	"errors"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db_lib"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type delayedApp struct {
	runCh     chan struct{}
	releaseCh chan struct{}
}

func (a *delayedApp) SetLogger(logger task_logger.Logger) task_logger.Logger {
	return logger
}

func (a *delayedApp) InstallRequirements(db_lib.LocalAppInstallingArgs) error {
	return nil
}

func (a *delayedApp) Run(args db_lib.LocalAppRunningArgs) error {
	close(a.runCh)
	<-a.releaseCh

	select {
	case <-args.StopCh:
		return nil
	case <-time.After(time.Second):
		return errors.New("stop request was not delivered")
	}
}

func (a *delayedApp) Clear() {}

// Test_KillBeforeRun_PreventsAppStart asserts that an existing stop request
// prevents the app from starting.
func Test_KillBeforeRun_PreventsAppStart(t *testing.T) {
	setupExecutorConfig(t)

	app := &delayedApp{
		runCh:     make(chan struct{}),
		releaseCh: make(chan struct{}),
	}
	executor := &LocalExecutor{
		App:      app,
		Logger:   task_logger.NopLogger{},
		prepared: true,
	}

	executor.Kill()
	close(app.releaseCh)
	err := executor.Run("", nil, "")

	assert.NoError(t, err)
	select {
	case <-app.runCh:
		assert.Fail(t, "app started after termination was requested")
	default:
	}
}

func Test_RunCanOnlyBeCalledOnce(t *testing.T) {
	setupExecutorConfig(t)

	app := &delayedApp{
		runCh:     make(chan struct{}),
		releaseCh: make(chan struct{}),
	}
	executor := &LocalExecutor{
		App:      app,
		Logger:   task_logger.NopLogger{},
		prepared: true,
	}
	firstRunCh := make(chan error, 1)
	go func() {
		firstRunCh <- executor.Run("", nil, "")
	}()

	<-app.runCh

	err := executor.Run("", nil, "")
	assert.ErrorContains(t, err, "local executor has already been run")

	executor.Kill()
	close(app.releaseCh)
	select {
	case err := <-firstRunCh:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		require.FailNow(t, "first LocalExecutor.Run did not return")
	}

	err = executor.Run("", nil, "")
	assert.ErrorContains(t, err, "local executor has already been run")
}
