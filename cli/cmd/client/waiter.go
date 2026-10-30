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
	// autoReject rejects the task instead of waiting for an approval, so that
	// an unattended run ends on its own rather than hanging on a gate nobody
	// is there to open.
	autoReject bool
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
		approval   approvalState
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
			}
			if opts.follow {
				n, err := r.printNewOutput(ctx, api, projectID, taskID, printed)
				if err == nil {
					printed = n
				} else if !isTransient(err) {
					return last, err
				}
			}
			// After the log, so that a Terraform plan is on screen before the
			// question about applying it.
			if err := r.handleApproval(ctx, api, projectID, taskID, task, opts, &approval); err != nil {
				return last, err
			}
			if task.Status.IsFinished() {
				return last, nil
			}
		case ctx.Err() != nil:
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
		case answer := <-approval.answers:
			// A decision was typed: act on it and poll again at once, so the
			// new status shows up without waiting out the interval.
			if err := r.applyApproval(ctx, api, projectID, taskID, answer, &approval); err != nil {
				return last, err
			}
		case <-time.After(interval):
		}
	}
}

// approvalState tracks the conversation about a task that stopped in
// waiting_confirmation: a Terraform plan awaiting "apply", or any template
// whose app asks before it changes anything.
type approvalState struct {
	// answers delivers the line typed at the prompt. It is nil while no
	// question is outstanding, which in a select means "never ready".
	answers <-chan promptAnswer
	// told says the explanation was printed; it is printed once per wait.
	told bool
	// silent stops the client from asking again: there is no terminal, or
	// stdin ended without an answer.
	silent bool
	// sent says an answer already went to the server, so it is not repeated
	// on every poll while the task moves on.
	sent bool
}

// handleApproval reacts to the task's status: it explains the stop, rejects
// the task outright when asked to, asks the user to approve or reject, and
// notices when somebody else answered instead (in the web UI, or with
// 'task confirm' from another terminal).
func (r *runtime) handleApproval(ctx context.Context, api *apiclient.Client, projectID, taskID int, task db.TaskWithTpl, opts waitOptions, st *approvalState) error {
	if task.Status != task_logger.TaskWaitingConfirmation {
		if st.answers != nil {
			r.warnf("Task %d is no longer waiting for confirmation (status: %s).", taskID, task.Status)
			st.answers = nil
		}
		return nil
	}

	if !st.told {
		st.told = true
		r.warnf("Task %d is waiting for confirmation before it continues.", taskID)
	}
	if opts.autoReject {
		if st.sent {
			// Already rejected; the server has not moved the task on yet.
			return nil
		}
		st.sent = true
		r.warnf("Rejecting it: --auto-reject was given.")
		return r.decide(ctx, api, projectID, taskID, false)
	}
	if st.silent || !r.prompter.canAsk() {
		if !st.silent {
			st.silent = true
			r.warnf("Run 'semaphore client task confirm %d' to continue, or 'semaphore client task reject %d' to cancel it.", taskID, taskID)
		}
		return nil
	}
	if st.answers == nil {
		fmt.Fprintf(r.io.Err, "\nApply task %d? [yes/no]: ", taskID)
		st.answers = r.prompter.readAsync()
	}
	return nil
}

// applyApproval turns the typed line into a confirm or a reject request. An
// unusable answer only re-asks: the task is still waiting either way.
func (r *runtime) applyApproval(ctx context.Context, api *apiclient.Client, projectID, taskID int, answer promptAnswer, st *approvalState) error {
	st.answers = nil

	switch {
	case answer.err != nil:
		st.silent = true
		r.warnf("warning: cannot read the answer: %v", answer.err)
	case answer.eof:
		st.silent = true
	default:
		switch strings.ToLower(answer.text) {
		case "yes", "y":
			return r.decide(ctx, api, projectID, taskID, true)
		case "no", "n":
			return r.decide(ctx, api, projectID, taskID, false)
		default:
			r.warnf("Please answer yes or no.")
			return nil
		}
	}

	r.warnf("Run 'semaphore client task confirm %d' to continue, or 'semaphore client task reject %d' to cancel it.", taskID, taskID)
	return nil
}

// decide sends the answer to the server. A rejected task ends as "error" on
// the server side, so the wait finishes with exit code 3.
func (r *runtime) decide(ctx context.Context, api *apiclient.Client, projectID, taskID int, approve bool) error {
	reqCtx, cancel := context.WithTimeout(ctx, r.cfg.Timeout)
	defer cancel()

	if approve {
		if err := api.ConfirmTask(reqCtx, projectID, taskID); err != nil {
			return err
		}
		r.warnf("Task %d confirmed.", taskID)
		return nil
	}
	if err := api.RejectTask(reqCtx, projectID, taskID); err != nil {
		return err
	}
	// Said here rather than left to the final line: the server ends a rejected
	// task as "error", which exits 3 and reads like a failure otherwise.
	r.warnf("Task %d rejected; the server ends a rejected task as failed (exit 3).", taskID)
	return nil
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
