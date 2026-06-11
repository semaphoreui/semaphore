package tasks

import (
	"errors"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/pkg/tz"
	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
)

// RunnerTaskAction is the reconciler's decision for one dispatched task.
type RunnerTaskAction int

const (
	// RunnerTaskKeep leaves the task alone.
	RunnerTaskKeep RunnerTaskAction = iota
	// RunnerTaskRequeue returns a not-yet-running task to the queue so the
	// normal dispatch loop assigns it to another runner.
	RunnerTaskRequeue
	// RunnerTaskFail fails a running task whose runner is lost.
	RunnerTaskFail
)

// DecideRunnerTaskAction classifies a dispatched, unfinished task against its
// runner's liveness. runner == nil means the runner row no longer exists.
//
// Two thresholds with distinct semantics:
//
//   - offlineTimeout (heartbeat staleness): the runner is offline — it gets no
//     new tasks and its "starting" tasks are requeued. Offline does NOT mean
//     its running jobs stopped.
//   - taskFailTimeout: the runner is presumed dead — its "running" tasks are
//     failed. A runner that reconnects within this window kept its in-memory
//     job pool and simply continues; nothing is failed.
//
// A restarted runner (started_at newer than the task's start) provably lost
// its job pool, so its running task is failed immediately — there is nothing
// to wait for. Its starting tasks self-heal: the restarted runner re-pulls
// them from NewJobs.
func DecideRunnerTaskAction(
	status task_logger.TaskStatus,
	taskStart *time.Time,
	runner *db.Runner,
	now time.Time,
	offlineTimeout time.Duration,
	taskFailTimeout time.Duration,
) (RunnerTaskAction, string) {

	starting := status == task_logger.TaskStartingStatus || status == task_logger.TaskWaitingStatus
	running := status == task_logger.TaskRunningStatus

	if !starting && !running {
		return RunnerTaskKeep, ""
	}

	if runner == nil {
		if starting {
			return RunnerTaskRequeue, "runner no longer exists"
		}
		return RunnerTaskFail, "runner no longer exists"
	}

	if running && runner.StartedAt != nil && taskStart != nil &&
		runner.StartedAt.After(*taskStart) {
		return RunnerTaskFail, "runner restarted and lost the task"
	}

	// A webhook-driven runner may still be booting in response to the dispatch
	// webhook; its heartbeat history (if any) predates this task, so staleness
	// must not requeue a just-dispatched task. Once such a runner reports the
	// task running, its heartbeat is meaningful again and the fail check below
	// applies as usual.
	if starting && runner.Webhook != "" {
		return RunnerTaskKeep, ""
	}

	if runner.Touched == nil {
		// Never polled. A poll-based runner cannot have been selected without
		// a fresh heartbeat, so a starting task here is safe to give back.
		if starting {
			return RunnerTaskRequeue, "runner never polled the server"
		}
		return RunnerTaskKeep, ""
	}

	silence := now.Sub(*runner.Touched)

	if starting && silence > offlineTimeout {
		return RunnerTaskRequeue, "runner is offline"
	}

	if running && silence > taskFailTimeout {
		return RunnerTaskFail, "runner stopped responding"
	}

	return RunnerTaskKeep, ""
}

// runnerTasksReconcileLoop periodically reconciles dispatched tasks against
// runner liveness: tasks on an offline runner are requeued (starting) or
// failed (running, after the recovery window). Started from TaskPool.Run.
// A nil stop channel means run forever.
func (p *TaskPool) runnerTasksReconcileLoop(stop <-chan struct{}) {
	ticker := time.NewTicker(util.Config.RunnersReconcileInterval())
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.reconcileRunnerTasks(tz.Now())
		case <-stop:
			return
		}
	}
}

// reconcileRunnerTasks applies DecideRunnerTaskAction to every dispatched,
// unfinished task this node owns (OwnedRunningRange: in HA, the tasks whose
// claim names this node; single-node, everything). Tasks of dead nodes have
// no live claim and are reconciled by the HA orphan cleaner instead, so the
// work is partitioned across the cluster rather than repeated on every node.
// The remaining cross-actor races (cleaner vs. owner around claim expiry)
// are covered by the state store's finalize lock and the DB re-checks in the
// action helpers.
func (p *TaskPool) reconcileRunnerTasks(now time.Time) {
	offlineTimeout := util.Config.RunnersOfflineTimeout()
	taskFailTimeout := util.Config.RunnersTaskFailTimeout()

	for _, tsk := range p.state.OwnedRunningRange() {
		if tsk == nil || tsk.Task.RunnerID == nil || tsk.Task.Status.IsFinished() {
			continue
		}

		runnerID := *tsk.Task.RunnerID

		var runnerPtr *db.Runner
		runner, err := p.store.GetGlobalRunner(runnerID)
		switch {
		case err == nil:
			runnerPtr = &runner
		case errors.Is(err, db.ErrNotFound):
			runnerPtr = nil
		default:
			log.WithError(err).WithFields(log.Fields{
				"task_id":   tsk.Task.ID,
				"runner_id": runnerID,
				"context":   "runner_reconciler",
			}).Warn("failed to load runner; skipping task")
			continue
		}

		action, reason := DecideRunnerTaskAction(
			tsk.Task.Status, tsk.Task.Start, runnerPtr, now, offlineTimeout, taskFailTimeout)

		switch action {
		case RunnerTaskRequeue:
			p.requeueTaskRunnerOffline(tsk, runnerID, reason)
		case RunnerTaskFail:
			p.failTaskRunnerLost(tsk, runnerPtr, reason)
		}
	}
}

// failTaskRunnerLost fails a dispatched task whose runner is lost and runs the
// usual finalization (finish webhook, autorun children, pool/Redis cleanup).
// Idempotent: it takes the state store's finalize lock before writing the
// failure so a concurrent terminal runner report can win without being
// overwritten by the reconciler.
func (p *TaskPool) failTaskRunnerLost(tsk *TaskRunner, runner *db.Runner, reason string) {
	if !p.state.TryFinalize(tsk.Task.ID) {
		return
	}
	defer p.state.DeleteFinalize(tsk.Task.ID)

	if util.HAEnabled() {
		p.refreshTaskStatusFromDB(tsk)
	}

	if tsk.Task.Status.IsFinished() {
		return
	}

	fields := log.Fields{
		"task_id": tsk.Task.ID,
		"context": "runner_reconciler",
	}
	if tsk.Task.RunnerID != nil {
		fields["runner_id"] = *tsk.Task.RunnerID
	}
	log.WithFields(fields).Warn("Runner lost: marking task failed")

	tsk.Log("Runner lost: " + reason)
	tsk.Task.Message = reason
	tsk.SetStatus(task_logger.TaskFailStatus)

	p.finalizeRemoteTaskLocked(tsk, runner)
}

// requeueTaskRunnerOffline returns a not-yet-running task dispatched to an
// offline runner back to the queue so the dispatch loop selects another
// runner. Clearing RunnerID removes the task from the old runner's NewJobs,
// and UpdateRunner's ownership check rejects its late progress reports.
//
// In HA mode the reconciler runs on every node over the shared running set,
// so several nodes can reach this point for the same task in the same tick.
// The shared queue is a plain RPUSH (no dedup) and a duplicate queue entry
// would eventually re-run the task, so requeue must happen exactly once:
// the cluster-wide finalize lock (Redis SETNX) serializes the attempts, and
// the DB re-check under the lock makes every loser observe the cleared
// RunnerID and bail.
func (p *TaskPool) requeueTaskRunnerOffline(tsk *TaskRunner, runnerID int, reason string) {
	if !p.state.TryFinalize(tsk.Task.ID) {
		return // another node is requeueing or finalizing this task
	}
	defer p.state.DeleteFinalize(tsk.Task.ID)

	if util.HAEnabled() {
		p.refreshTaskStatusFromDB(tsk)
	}

	// Only a task that has not started executing may be reassigned: the
	// runner may have just picked it up and reported "running" concurrently.
	if tsk.Task.Status != task_logger.TaskStartingStatus &&
		tsk.Task.Status != task_logger.TaskWaitingStatus {
		return
	}

	// Already reassigned (e.g. by another node).
	if tsk.Task.RunnerID == nil || *tsk.Task.RunnerID != runnerID {
		return
	}

	log.WithFields(log.Fields{
		"task_id":   tsk.Task.ID,
		"runner_id": runnerID,
		"context":   "runner_reconciler",
	}).Warn("Runner offline: returning task to queue")

	tsk.Logf("Runner #%d lost the task: %s. Returning task to queue.", runnerID, reason)

	tsk.Task.RunnerID = nil
	tsk.SetStatus(task_logger.TaskWaitingStatus)

	// SetStatus is a no-op when the status is already "waiting"; persist the
	// cleared RunnerID explicitly so the old runner cannot pull the task again.
	if err := p.store.UpdateTask(tsk.Task); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"task_id": tsk.Task.ID,
			"context": "runner_reconciler",
		}).Error("failed to persist requeued task")
		return
	}

	// Same flow as the ErrAllRunnersBusy requeue in TaskRunner.run:
	// put the task back into the queue, then let the pool release its
	// running/active bookkeeping (EventTypeRequeued -> onTaskStop).
	p.state.Enqueue(tsk)
	p.queueEvents <- PoolEvent{EventTypeRequeued, tsk}
}
