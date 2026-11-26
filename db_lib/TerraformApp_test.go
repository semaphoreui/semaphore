//go:build !windows

package db_lib

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type terraformApprovalLogger struct {
	task_logger.NopLogger
	commands [][]string
	statuses []task_logger.TaskStatus
	messages []string
	listener task_logger.StatusListener
}

func (l *terraformApprovalLogger) LogCmd(cmd *exec.Cmd) func() {
	l.commands = append(l.commands, append([]string(nil), cmd.Args[1:]...))
	return func() {}
}

func (l *terraformApprovalLogger) Log(message string) {
	l.messages = append(l.messages, message)
}

func (l *terraformApprovalLogger) AddStatusListener(listener task_logger.StatusListener) {
	l.listener = listener
}

func (l *terraformApprovalLogger) SetStatus(status task_logger.TaskStatus) {
	l.statuses = append(l.statuses, status)
	// Reject a manual prompt immediately so a broken auto-approval path cannot
	// hang the test or accidentally execute apply.
	if status == task_logger.TaskWaitingConfirmation {
		status = task_logger.TaskRejected
	}
	l.listener(status)
}

func readTerraformReaderWithTimeout(t *testing.T, reader *terraformReader, timeout time.Duration) (n int, buf []byte, err error) {
	t.Helper()
	buf = make([]byte, 8)
	done := make(chan struct{})
	go func() {
		n, err = reader.Read(buf)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(timeout):
		t.Fatal("terraformReader.Read hung")
	}
	return n, buf, err
}

// TestTerraformReader_ReadReleasesOnTerminalStatus guards against hanging when a
// force-stop or finish moves the task out of waiting_confirmation during or before
// terraform init's backend migration prompt.
func TestTerraformReader_ReadReleasesOnTerminalStatus(t *testing.T) {
	tests := []struct {
		name          string
		initialStatus task_logger.TaskStatus
		afterDelay    task_logger.TaskStatus
		delay         time.Duration
		checkSecondEOF bool
	}{
		{
			name:          "stopping flips after read entered wait",
			initialStatus: task_logger.TaskWaitingConfirmation,
			afterDelay:    task_logger.TaskStoppingStatus,
			delay:         50 * time.Millisecond,
		},
		{
			name:          "stopped flips after read entered wait",
			initialStatus: task_logger.TaskWaitingConfirmation,
			afterDelay:    task_logger.TaskStoppedStatus,
			delay:         20 * time.Millisecond,
			checkSecondEOF: true,
		},
		{
			name:          "stopping before first read",
			initialStatus: task_logger.TaskStoppingStatus,
		},
		{
			name:          "stopped before first read",
			initialStatus: task_logger.TaskStoppedStatus,
			checkSecondEOF: true,
		},
	}

	timeout := 12 * time.Second
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &terraformReader{
				logger: &task_logger.NopLogger{},
				status: tt.initialStatus,
			}
			if tt.delay > 0 {
				go func() {
					time.Sleep(tt.delay)
					reader.setStatus(tt.afterDelay)
				}()
			}

			n, buf, err := readTerraformReaderWithTimeout(t, reader, timeout)
			require.NoError(t, err)
			assert.True(t, reader.EOF)
			assert.Equal(t, 1, n)
			assert.Equal(t, byte('\n'), buf[0])

			if tt.checkSecondEOF {
				_, err = reader.Read(buf)
				require.ErrorIs(t, err, io.EOF)
			}
		})
	}
}

func TestTerraformApp_RunAutoApprove(t *testing.T) {
	tests := []struct {
		name       string
		taskParams db.MapStringAnyField
		tplParams  db.TerraformTemplateParams
		wantApply  bool
		wantPrompt bool
		wantNotice bool
	}{
		{
			name:       "destroy with task auto approval allowed",
			taskParams: db.MapStringAnyField{"destroy": true, "auto_approve": true},
			tplParams:  db.TerraformTemplateParams{AllowAutoApprove: true},
			wantApply:  true,
		},
		{
			name:       "destroy with task auto approval forbidden",
			taskParams: db.MapStringAnyField{"destroy": true, "auto_approve": true},
			wantPrompt: true,
			wantNotice: true,
		},
		{
			name:       "destroy with template auto approval",
			taskParams: db.MapStringAnyField{"destroy": true},
			tplParams:  db.TerraformTemplateParams{AutoApprove: true},
			wantApply:  true,
		},
		{
			name:       "destroy still requires task opt in",
			taskParams: db.MapStringAnyField{"destroy": true},
			tplParams:  db.TerraformTemplateParams{AllowAutoApprove: true},
			wantPrompt: true,
		},
		{
			name:       "plan only never applies destroy",
			taskParams: db.MapStringAnyField{"destroy": true, "auto_approve": true, "plan": true},
			tplParams:  db.TerraformTemplateParams{AllowAutoApprove: true},
		},
		{
			name:       "regular apply with task auto approval",
			taskParams: db.MapStringAnyField{"auto_approve": true},
			tplParams:  db.TerraformTemplateParams{AllowAutoApprove: true},
			wantApply:  true,
		},
	}

	for _, app := range []db.TemplateApp{db.AppTerraform, db.AppTofu, db.AppTerragrunt} {
		for _, tt := range tests {
			t.Run(string(app)+"/"+tt.name, func(t *testing.T) {
				original := util.Config
				t.Cleanup(func() { util.Config = original })
				dir := t.TempDir()
				binary := filepath.Join(dir, "terraform")
				require.NoError(t, os.WriteFile(binary, []byte("#!/bin/sh\nexit 0\n"), 0700))
				util.Config = &util.ConfigType{Apps: map[string]util.App{
					string(app): {AppPath: binary},
				}}

				// Workflow nodes create tasks through TaskParams.CreateTask; the
				// executor extracts the same params for standalone tasks.
				task := (db.TaskParams{Params: tt.taskParams}).CreateTask(1)
				var params db.TerraformTaskParams
				require.NoError(t, task.ExtractParams(&params))
				args := []string{}
				if params.Destroy {
					args = append(args, "-destroy")
				}

				logger := &terraformApprovalLogger{}
				runner := TerraformApp{Name: string(app), Repository: db.Repository{GitURL: dir}}
				runner.SetLogger(logger)
				err := runner.Run(LocalAppRunningArgs{
					CliArgs:        map[string][]string{"default": args},
					TaskParams:     &params,
					TemplateParams: &tt.tplParams,
				})
				require.NoError(t, err)

				wantCommands := [][]string{append([]string{"plan", "-lock=false"}, args...)}
				if tt.wantApply {
					wantCommands = append(wantCommands, append([]string{"apply", "-auto-approve", "-lock=false"}, args...))
				}
				if app == db.AppTerragrunt {
					for i := range wantCommands {
						wantCommands[i] = append(wantCommands[i], "--tf-path=terraform")
					}
				}
				assert.Equal(t, wantCommands, logger.commands)
				if tt.wantPrompt {
					assert.Contains(t, logger.statuses, task_logger.TaskWaitingConfirmation)
				} else {
					assert.NotContains(t, logger.statuses, task_logger.TaskWaitingConfirmation)
				}
				if tt.wantNotice {
					assert.Contains(t, logger.messages, "Auto approval was requested, but the template does not allow it. Enable Auto approve in the template's task prompts to allow tasks and workflows to skip confirmation.")
				} else {
					assert.Empty(t, logger.messages)
				}
			})
		}
	}
}