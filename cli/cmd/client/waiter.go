package client

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/apiclient"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
	"github.com/spf13/cobra"
)

// defaultPollInterval is how often the waiter asks the server about a task.
const defaultPollInterval = 2 * time.Second

// maxPollFailures is how many consecutive transport or 5xx errors the waiter
// tolerates before giving up; a stand restarting behind a proxy needs a few.
const maxPollFailures = 5

// waitOptions tune waitTask.
type waitOptions struct {
	// follow prints new log lines to stdout as they appear.
	follow bool
	// timeout bounds the whole wait; zero means forever.
	timeout time.Duration
	// stopOnInterrupt sends a stop request when the wait is interrupted.
	stopOnInterrupt bool
	// onStatus is called on every status change (watch uses it to print).
	onStatus func(task db.TaskWithTpl)
}

// waitTask polls a task until it reaches a final status (success, error,
// stopped) and returns it. It never stops the task itself: Ctrl+C only
// disconnects the client unless stopOnInterrupt is set, and a timeout returns
// a WaitTimeoutError while the task keeps running on the server.
func (r *runtime) waitTask(cmd *cobra.Command, api *apiclient.Client, projectID, taskID int, opts waitOptions) (db.TaskWithTpl, error) {
	parent := cmd.Context()
	ctx, cancel := context.WithCancel(parent)
	if opts.timeout > 0 {
		ctx, cancel = context.WithTimeout(parent, opts.timeout)
	}
	defer cancel()

	interval := r.pollInterval
	if interval <= 0 {
		interval = defaultPollInterval
	}

	var (
		last       db.TaskWithTpl
		lastStatus task_logger.TaskStatus
		printed    int
		failures   int
		hinted     bool
	)

	for {
		task, err := r.pollOnce(ctx, api, projectID, taskID)
		switch {
		case err == nil:
			failures = 0
			last = task
			if task.Status != lastStatus {
				lastStatus = task.Status
				if opts.onStatus != nil {
					opts.onStatus(task)
				}
				if task.Status == task_logger.TaskWaitingConfirmation && !hinted {
					hinted = true
					r.warnf("Task %d waits for confirmation: run 'semaphore client task confirm %d' (or reject).", taskID, taskID)
				}
			}
			if opts.follow {
				n, err := r.printNewOutput(ctx, api, projectID, taskID, printed)
				if err == nil {
					printed = n
				} else if !isTransient(err) {
					return last, err
				}
			}
			if task.Status.IsFinished() {
				return last, nil
			}
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			return last, r.waitEnded(ctx, parent, api, projectID, taskID, opts)
		case isTransient(err):
			failures++
			if failures >= maxPollFailures {
				return last, fmt.Errorf("lost the server while waiting for task %d: %w", taskID, err)
			}
			r.warnf("warning: %v (retry %d/%d)", err, failures, maxPollFailures)
		default:
			return last, err
		}

		select {
		case <-ctx.Done():
			return last, r.waitEnded(ctx, parent, api, projectID, taskID, opts)
		case <-time.After(interval):
		}
	}
}

// pollOnce reads the task with the per-request timeout on top of ctx.
func (r *runtime) pollOnce(ctx context.Context, api *apiclient.Client, projectID, taskID int) (db.TaskWithTpl, error) {
	reqCtx, cancel := context.WithTimeout(ctx, r.cfg.Timeout)
	defer cancel()
	return api.GetTask(reqCtx, projectID, taskID)
}

// printNewOutput prints the log lines beyond printed and returns the new count.
func (r *runtime) printNewOutput(ctx context.Context, api *apiclient.Client, projectID, taskID, printed int) (int, error) {
	reqCtx, cancel := context.WithTimeout(ctx, r.cfg.Timeout)
	defer cancel()
	lines, err := api.GetTaskOutput(reqCtx, projectID, taskID)
	if err != nil {
		return printed, err
	}
	if len(lines) < printed {
		// The server trimmed or replaced the log; start over rather than skip.
		printed = 0
	}
	for _, line := range lines[printed:] {
		r.printLogLine(line.Output)
	}
	return len(lines), nil
}

// printLogLine writes one task log line to stdout, with ANSI colours kept only
// for a colour-capable terminal.
func (r *runtime) printLogLine(text string) {
	opts := r.printer.Options()
	if !(opts.TTY && opts.Color) {
		text = util.ClearFromAnsiCodes(text)
	}
	r.printer.Printf("%s\n", strings.TrimRight(text, "\r\n"))
}

// waitEnded turns a cancelled or expired wait into the right error, sending a
// stop request first when the user asked for it.
func (r *runtime) waitEnded(ctx, parent context.Context, api *apiclient.Client, projectID, taskID int, opts waitOptions) error {
	if parent.Err() == nil && errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return &WaitTimeoutError{TaskID: taskID}
	}
	// Interrupted. The task goes on unless --stop-on-interrupt was given.
	if opts.stopOnInterrupt {
		stopCtx, cancel := context.WithTimeout(context.Background(), r.cfg.Timeout)
		defer cancel()
		if err := api.StopTask(stopCtx, projectID, taskID, false); err != nil {
			r.warnf("warning: could not stop task %d: %v", taskID, err)
		} else {
			r.warnf("Stop requested for task %d.", taskID)
		}
	} else {
		r.warnf("Interrupted; task %d keeps running on the server. Follow it again with 'semaphore client task logs %d -f'.", taskID, taskID)
	}
	return context.Canceled
}

// isTransient says whether an error is worth retrying: network trouble and
// server-side 5xx. A 401, 403 or 404 is final.
func isTransient(err error) bool {
	var tErr *apiclient.TransportError
	if errors.As(err, &tErr) {
		return !errors.Is(err, context.Canceled)
	}
	var apiErr *apiclient.APIError
	if errors.As(err, &apiErr) {
		return apiErr.Status >= 500
	}
	return false
}

// finishedError maps a final task status to the error that sets the exit
// code: nil for success, TaskFinishedError for error and stopped.
func finishedError(task db.TaskWithTpl) error {
	switch task.Status {
	case task_logger.TaskSuccessStatus:
		return nil
	case task_logger.TaskFailStatus, task_logger.TaskStoppedStatus:
		return &TaskFinishedError{TaskID: task.ID, Status: string(task.Status)}
	}
	return &TaskFinishedError{TaskID: task.ID, Status: string(task.Status)}
}
