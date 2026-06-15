package tasks

import (
	"fmt"
	"strconv"
	"time"

	"github.com/semaphoreui/semaphore/pkg/random"
	"github.com/semaphoreui/semaphore/pkg/tz"
	"github.com/semaphoreui/semaphore/pro/pkg/stage_parsers"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	"github.com/semaphoreui/semaphore/services/server"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db_lib"
	"github.com/semaphoreui/semaphore/pkg/task_logger"

	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
)

type logRecord struct {
	task         *TaskRunner
	output       string
	time         time.Time
	currentStage *db.TaskStage
}

type EventType uint

const (
	EventTypeNew      EventType = 0 // EventTypeNew represents an event when a new task is created, typically sent during a periodic check or timer.
	EventTypeFinished EventType = 1 // EventTypeFinished represents an event when a task finishes, typically sent during a periodic check or timer.
	EventTypeFailed   EventType = 2 // EventTypeFailed represents an event when a task fails, typically sent during a periodic check or timer.
	EventTypeEmpty    EventType = 3 // EventTypeEmpty represents an event when the queue is empty, typically sent during a periodic check or timer.
	EventTypeRequeued EventType = 4 // EventTypeRequeued represents an event when a task is moved back to the waiting state for reprocessing.
)

const (
	TaskOutputBatchSize        = 500
	TaskOutputInsertIntervalMs = 500
)

type PoolEvent struct {
	eventType EventType
	task      *TaskRunner
}

type TaskPool struct {
	// register channel used to put tasks to queue.
	register chan *TaskRunner

	// logger channel used to putting log records to database.
	logger chan logRecord

	store                  db.Store
	ansibleTaskRepo        db.AnsibleTaskRepository
	logWriteService        pro_interfaces.LogWriteService
	inventoryService       server.InventoryService
	encryptionService      server.AccessKeyEncryptionService
	keyInstallationService server.AccessKeyInstallationService

	queueEvents chan PoolEvent

	// state provides pluggable storage for Queue, active projects, running tasks and aliases
	state TaskStateStore

	// workflowService orchestrates workflow runs (a Pro feature). It is injected
	// after construction via SetWorkflowService; the pool only calls back into it
	// when a workflow task finishes. nil in tests / before wiring.
	workflowService pro_interfaces.WorkflowService
	// stop signals the background loops started by Run to exit. Closing it (via
	// Stop) terminates the runner-task reconcile loop and Run's own select.
	// Channels are used rather than sync.WaitGroup/sync.Once because TaskPool is
	// returned by value from the constructors, and copying a struct that embeds
	// a lock is flagged by go vet (copylocks).
	stop chan struct{}

	// reconcileDone is closed by runnerTasksReconcileLoop when it exits, so Stop
	// can block until the loop has actually finished reading shared state (e.g.
	// util.Config). It is never closed if Run was not started, so Stop must only
	// be called after Run.
	reconcileDone chan struct{}
}

func CreateTaskPool(
	store db.Store,
	state TaskStateStore,
	ansibleTaskRepo db.AnsibleTaskRepository,
	inventoryService server.InventoryService,
	encryptionService server.AccessKeyEncryptionService,
	keyInstallationService server.AccessKeyInstallationService,
	logWriteService pro_interfaces.LogWriteService,
) TaskPool {
	p := TaskPool{
		register:               make(chan *TaskRunner),      // add TaskRunner to queue
		logger:                 make(chan logRecord, 10000), // store log records to database
		store:                  store,
		state:                  state,
		queueEvents:            make(chan PoolEvent),
		inventoryService:       inventoryService,
		ansibleTaskRepo:        ansibleTaskRepo,
		encryptionService:      encryptionService,
		logWriteService:        logWriteService,
		keyInstallationService: keyInstallationService,
		stop:                   make(chan struct{}),
		reconcileDone:          make(chan struct{}),
	}
	// attempt to start HA state store (no-op for memory)
	_ = p.state.Start(p.hydrateTaskRunner)
	return p
}

// StateStore returns the pluggable task state backend. Used by the Cluster
// Dashboard to reach an optional TaskStateInspector implementation.
func (p *TaskPool) StateStore() TaskStateStore {
	return p.state
}

// SetWorkflowService injects the workflow orchestration service. It is wired
// after the pool is created (the service needs the pool as its task enqueuer,
// and the pool needs the service to progress runs as tasks finish).
func (p *TaskPool) SetWorkflowService(svc pro_interfaces.WorkflowService) {
	p.workflowService = svc
}

// HandleWorkflowTaskCompletion notifies the workflow service that a task that
// belongs to a workflow run has finished, so it can progress the run. It is a
// thin delegator so the open task lifecycle (TaskRunner) need not know about the
// Pro workflow service; a no-op when no service is wired.
func (p *TaskPool) HandleWorkflowTaskCompletion(task db.Task) error {
	if p.workflowService == nil {
		return nil
	}
	return p.workflowService.HandleWorkflowTaskCompletion(task)
}

// GetWorkflowRunArtifacts returns the merged upstream artifacts for a workflow
// run, delegating to the workflow service. Returns an empty map when no service
// is wired.
func (p *TaskPool) GetWorkflowRunArtifacts(projectID int, runID int, currentTaskID *int) (map[string]any, error) {
	if p.workflowService == nil {
		return nil, nil
	}
	return p.workflowService.GetWorkflowRunArtifacts(projectID, runID, currentTaskID)
}

func (p *TaskPool) GetNumberOfRunningTasksOfRunner(runnerID int) (res int) {
	for _, task := range p.state.RunningRange() {
		if task.Task.RunnerID != nil && *task.Task.RunnerID == runnerID {
			res++
		}
	}
	return
}

func (p *TaskPool) GetRunningTasks() (res []*TaskRunner) {
	return p.state.RunningRange()
}

func (p *TaskPool) GetTask(id int) (task *TaskRunner, err error) {
	for _, t := range p.state.QueueRange() {
		if t.Task.ID == id {
			task = t
			break
		}
	}

	if task == nil {
		for _, t := range p.state.RunningRange() {
			if t.Task.ID == id {
				task = t
				break
			}
		}
	}

	if util.HAEnabled() {
		if task == nil {
			task, err = p.HydrateTaskRunnerFromDB(id)
		}
	}

	return
}

func (p *TaskPool) GetTaskByAlias(alias string) (task *TaskRunner) {
	return p.state.GetByAlias(alias)
}

// nolint: gocyclo
func (p *TaskPool) Run() {
	ticker := time.NewTicker(5 * time.Second)

	defer ticker.Stop()

	// In HA mode the state store relays cross-node stop requests: when another
	// node stops a workflow run, tasks owned by this node must be killed here.
	if broadcaster, ok := p.state.(TaskStopBroadcaster); ok {
		broadcaster.SetTaskStopHandler(p.stopLocalTask)
	}

	go p.handleQueue()
	go p.handleLogs()
	go func() {
		// reconcileDone lets Stop block until the reconcile loop has actually
		// finished reading shared state (e.g. util.Config). Closing it here,
		// rather than inside the loop, keeps runnerTasksReconcileLoop reusable
		// by tests that call it directly with their own lifecycle channels.
		defer close(p.reconcileDone)
		p.runnerTasksReconcileLoop()
	}()

	for {
		select {
		case task := <-p.register: // new task created by API or schedule

			task.Log("Task " + task.Template.Name + " added to queue")
			log.WithFields(log.Fields{
				"task_id":   task.Task.ID,
				"task_name": task.Template.Name,
			}).Info("Task added to queue")
			task.saveStatus()

			p.queueEvents <- PoolEvent{EventTypeNew, task}

		case <-ticker.C: // timer 5 seconds
			p.queueEvents <- PoolEvent{EventTypeEmpty, nil}

		case <-p.stop:
			return
		}
	}
}

// Stop signals Run's background loops to exit and blocks until the runner-task
// reconcile loop has finished. It must be called at most once, and only after
// Run has been started (it waits on reconcileDone, which only the reconcile
// loop closes). Production runs the pool for the whole process lifetime and
// never calls Stop; it exists so tests can terminate the reconcile goroutine
// before they mutate shared globals such as util.Config.
func (p *TaskPool) Stop() {
	close(p.stop)
	<-p.reconcileDone
}

func getTaskName(t *TaskRunner) string {
	return t.Template.Name + " (" + strconv.Itoa(t.Task.ID) + ")"
}

func (p *TaskPool) handleQueue() {
	for t := range p.queueEvents {
		// When a task is re-queued (e.g., no remote runner available), we should
		// clean up its "running" bookkeeping but avoid immediately retrying it in
		// the same queue pass to prevent hot retry loops.
		skipTaskID := 0

		switch t.eventType {
		case EventTypeRequeued:
			// Task was started but moved back to waiting. It must not remain in
			// running/active sets and must release its claim so it can be picked
			// up again later.
			p.onTaskStop(t.task)
			// Avoid immediate retry in this same event handling iteration; it
			// will be retried on the next periodic tick or when another event
			// triggers queue processing.
			skipTaskID = t.task.Task.ID
		case EventTypeNew:
			p.state.Enqueue(t.task)
		case EventTypeFinished:
			p.onTaskStop(t.task)
		}

		// Snapshot the queue once per pass and address every task by ID. In HA
		// mode multiple nodes mutate the shared Redis queue concurrently, so a
		// position-based walk (QueueGet(i) + DequeueAt(i)) races: the list can
		// shift between the read and the dequeue, removing a different task than
		// the one that was claimed. Iterating a snapshot and claiming by ID
		// (ClaimAndDequeue) removes that hazard.
		for _, curr := range p.state.QueueRange() {
			if curr == nil { // item may no longer be available, move ahead
				continue
			}

			// When handling a requeue event, don't immediately start the same task again.
			if skipTaskID != 0 && curr.Task.ID == skipTaskID {
				continue
			}

			if curr.Task.Status == task_logger.TaskFailStatus {
				//delete failed TaskRunner from queue
				p.state.DequeueByID(curr.Task.ID)
				log.Info("Task " + getTaskName(curr) + " removed from queue")
				continue
			}

			if p.blocks(curr) {
				continue
			}

			// Atomically claim and remove the task so exactly one node runs it.
			// On failure another node owns it (or it is already gone); leave it.
			if !p.state.ClaimAndDequeue(curr.Task.ID) {
				continue
			}

			runTask(curr, p)
		}
	}
}

func (p *TaskPool) handleLogs() {
	logTicker := time.NewTicker(TaskOutputInsertIntervalMs * time.Millisecond)

	defer logTicker.Stop()

	logs := make([]logRecord, 0)

	for {

		select {
		case record := <-p.logger:
			logs = append(logs, record)

			if len(logs) >= TaskOutputBatchSize {
				p.flushLogs(&logs)
			}
		case <-logTicker.C:
			p.flushLogs(&logs)
		}
	}
}

func (p *TaskPool) flushLogs(logs *[]logRecord) {
	if len(*logs) > 0 {
		p.writeLogs(*logs)
		*logs = (*logs)[:0]
	}
}

func (p *TaskPool) writeLogs(logs []logRecord) {

	taskOutput := make([]db.TaskOutput, 0)

	for _, record := range logs {
		newOutput := db.TaskOutput{
			TaskID: record.task.Task.ID,
			Output: record.output,
			Time:   record.time,
		}

		currentOutput := record.task.currentOutput
		record.task.currentOutput = &newOutput

		newStage, newState, err := stage_parsers.MoveToNextStage(
			p.store,
			p.ansibleTaskRepo,
			p.logWriteService,
			record.task.Template.App,
			record.task.Task.ProjectID,
			record.task.currentState,
			record.task.currentStage,
			currentOutput,
			newOutput)

		if err != nil {
			log.Error(err)
			return
		}

		record.task.currentState = newState

		if newStage != nil {
			record.task.currentStage = newStage
		}

		if record.task.currentStage != nil {
			newOutput.StageID = &record.task.currentStage.ID
		}

		taskOutput = append(taskOutput, newOutput)
	}

	err := p.store.InsertTaskOutputBatch(taskOutput)
	if err != nil {
		log.Error(err)
		return
	}
}

func runTask(task *TaskRunner, p *TaskPool) {
	// Mark the task as actively dispatched by this process before it becomes
	// visible in the running set (onTaskRun -> SetRunning). The reconciler relies
	// on this to distinguish a live dispatch from a stale "starting" stub left in
	// the running set by a previous process that died mid-dispatch.
	task.dispatching.Store(true)

	log.WithFields(log.Fields{
		"context":   "task_pool",
		"task_id":   task.Task.ID,
		"task_name": task.Template.Name,
	}).Info("Set resource locker")
	p.onTaskRun(task)

	log.WithFields(log.Fields{
		"context":   "task_pool",
		"task_id":   task.Task.ID,
		"task_name": task.Template.Name,
	}).Info("Task started")
	go func() {
		time.Sleep(1 * time.Second)
		task.run()
	}()
}

func (p *TaskPool) onTaskRun(t *TaskRunner) {
	p.state.AddActive(t.Task.ProjectID, t)
	p.state.SetRunning(t)
	if t.Alias != "" {
		p.state.SetAlias(t.Alias, t)
	}
}

func (p *TaskPool) onTaskStop(t *TaskRunner) {
	p.state.RemoveActive(t.Task.ProjectID, t.Task.ID)
	p.state.DeleteRunning(t.Task.ID)
	p.state.DeleteClaim(t.Task.ID)
	if t.Alias != "" {
		p.state.DeleteAlias(t.Alias)
	}
}

// FinalizeRemoteTask completes a remote (runner) task once it has reached a
// terminal status. It runs the finish webhook (when a runner is provided),
// queues any autorun child templates, and releases the task's pool/Redis state
// (End time, EventTypeFinished -> onTaskStop).
//
// Because remote completion is reported by the runner to an arbitrary node,
// this is what decouples a task's lifecycle from the node that dispatched it:
// whichever node receives the terminal report finalizes the task. It is safe to
// call from several racing signals (runner report, timeout, force stop) — the
// state store's TryFinalize guard ensures it runs at most once per task across
// the cluster (Redis SETNX in HA) and within a process (in-memory sync.Map).
func (p *TaskPool) FinalizeRemoteTask(tsk *TaskRunner, runner *db.Runner) {
	if tsk == nil {
		return
	}

	if !p.state.TryFinalize(tsk.Task.ID) {
		return
	}
	defer p.state.DeleteFinalize(tsk.Task.ID)

	p.finalizeRemoteTaskLocked(tsk, runner)
}

// finalizeRemoteTaskLocked completes a remote task after the caller has won
// the state store's finalize lock for tsk.Task.ID.
func (p *TaskPool) finalizeRemoteTaskLocked(tsk *TaskRunner, runner *db.Runner) {
	if util.HAEnabled() {
		p.refreshTaskStatusFromDB(tsk)
	}

	if tsk.Task.End != nil {
		// Another node may have persisted End before onTaskStop ran (e.g.
		// crash between saveStatus and the queue drain). Release any stale
		// shared pool state without re-running finish or autorun.
		p.onTaskStop(tsk)
		return
	}

	if runner != nil {
		if err := callRunnerWebhook(runner, tsk, "finish"); err != nil {
			log.WithError(err).WithField("task_id", tsk.Task.ID).Warn("remote task finish webhook failed")
		}
	}

	// Persist End before enqueueing autorun children so the HA DB backstop
	// above (tsk.Task.End != nil) becomes a real second guard: a late
	// duplicate finalize on another node observes End set and skips autorun,
	// even if the cluster-wide finalize lock has already been released.
	tsk.finishRun()
	tsk.startAutorunTasks()
}

func applyDBPersistedTaskSnapshot(dst *db.Task, src db.Task) {
	dst.Status = src.Status
	dst.Start = src.Start
	dst.End = src.End
	dst.RunnerID = src.RunnerID
	dst.CommitHash = src.CommitHash
	dst.CommitMessage = src.CommitMessage
}

// refreshTaskStatusFromDB updates tsk with the persisted task row. In HA mode
// the in-memory pool can be stale after another node finalizes the task.
func (p *TaskPool) refreshTaskStatusFromDB(tsk *TaskRunner) {
	row, err := p.store.GetTaskByID(tsk.Task.ID)
	if err != nil {
		log.WithError(err).WithField("task_id", tsk.Task.ID).Warn("failed to refresh task status from DB")
		return
	}
	applyDBPersistedTaskSnapshot(&tsk.Task, row)
}

// hydrateTaskRunner builds a TaskRunner for an existing task from DB without starting it
func (p *TaskPool) hydrateTaskRunner(taskID int, projectID int) (*TaskRunner, error) {
	task, err := p.store.GetTask(projectID, taskID)
	if err != nil {
		return nil, err
	}

	tr := NewTaskRunner(task, p, "", p.keyInstallationService)
	if err = tr.populateDetails(); err != nil {
		return nil, err
	}

	// load runtime fields from the HA store (e.g., Redis)
	if p.state != nil {
		p.state.LoadRuntimeFields(tr)
	}

	// Persisted row from DB must win over runtime-store fields: Redis may still hold a
	// snapshot from enqueue time (e.g. status "starting") after the runner updated the DB.
	applyDBPersistedTaskSnapshot(&tr.Task, task)

	// set the appropriate job handler for consistency (not run)
	var job Job
	if util.Config.UseRemoteRunner || tr.Template.RunnerTag != nil || tr.Inventory.RunnerTag != nil {
		tag := tr.Template.RunnerTag
		if tag == nil {
			tag = tr.Inventory.RunnerTag
		}
		job = &RemoteJob{RunnerTag: tag, Task: tr.Task, taskPool: p}
	} else {
		app := db_lib.CreateApp(tr.Template, tr.Repository, tr.Inventory, tr)
		job = &LocalExecutor{
			Task:         tr.Task,
			Template:     tr.Template,
			Inventory:    tr.Inventory,
			Repository:   tr.Repository,
			Environment:  tr.Environment,
			Secret:       "{}",
			Logger:       app.SetLogger(tr),
			App:          app,
			KeyInstaller: p.keyInstallationService,
		}
	}
	tr.job = job
	return tr, nil
}

// HydrateTaskRunnerFromDB loads a task row by ID and builds a TaskRunner for API-side updates
// (e.g. runner progress on an HA node that did not enqueue the task).
func (p *TaskPool) HydrateTaskRunnerFromDB(taskID int) (*TaskRunner, error) {
	row, err := p.store.GetTaskByID(taskID)
	if err != nil {
		return nil, err
	}
	tr, err := p.hydrateTaskRunner(taskID, row.ProjectID)
	if err != nil {
		return nil, err
	}
	if row.RunnerID != nil {
		tr.Task.RunnerID = row.RunnerID
	}
	return tr, nil
}

func (p *TaskPool) blocks(t *TaskRunner) bool {

	if util.Config.MaxParallelTasks > 0 && p.state.RunningCount() >= util.Config.MaxParallelTasks {
		return true
	}

	if p.state.ActiveCount(t.Task.ProjectID) == 0 {
		return false
	}

	for _, r := range p.state.GetActive(t.Task.ProjectID) {
		if r.Task.Status.IsFinished() {
			continue
		}
		if r.Template.ID == t.Task.TemplateID && !r.Template.AllowParallelTasks {
			return true
		}
	}

	proj, err := p.store.GetProject(t.Task.ProjectID)

	if err != nil {
		log.Error(err)
		return false
	}

	res := proj.MaxParallelTasks > 0 && p.state.ActiveCount(t.Task.ProjectID) >= proj.MaxParallelTasks

	if res {
		return true
	}

	return res
}

func (p *TaskPool) ConfirmTask(targetTask db.Task) error {
	tsk, err := p.GetTask(targetTask.ID)

	if err != nil {
		return err
	}

	if tsk == nil { // task not active, but exists in database
		return fmt.Errorf("task is not active")
	}

	tsk.SetStatus(task_logger.TaskConfirmed)

	return nil
}

func (p *TaskPool) RejectTask(targetTask db.Task) error {
	tsk, err := p.GetTask(targetTask.ID)

	if err != nil {
		return err
	}

	if tsk == nil { // task not active, but exists in database
		return fmt.Errorf("task is not active")
	}

	tsk.SetStatus(task_logger.TaskRejected)

	return nil
}

func (p *TaskPool) stopTaskRunner(t *TaskRunner, forceStop bool) {
	prevStatus := t.Task.Status
	if forceStop {
		t.SetStatus(task_logger.TaskStoppedStatus)
	} else {
		t.SetStatus(task_logger.TaskStoppingStatus)
	}
	if prevStatus == task_logger.TaskRunningStatus {
		t.kill()
	}

	// A force-stopped remote task reaches "stopped" immediately (SetStatus
	// above always transitions to it) and will not get a runner completion
	// report, so finalize (cleanup) it here — otherwise it leaks in the
	// running/active sets. A graceful stop stays "stopping" and is finalized
	// when the runner reports it stopped via the runner API.
	if forceStop && t.job != nil && t.job.Async() && t.Task.Status.IsFinished() {
		go p.FinalizeRemoteTask(t, nil)
	}
}

// stopLocalTask force-stops a task in response to a cross-node stop broadcast
// (TaskStopBroadcaster). Only tasks this node actually holds are affected:
// queued waiting tasks are dequeued, running ones go through stopTaskRunner.
// Tasks not found locally are ignored — the broadcast reaches their owner too.
func (p *TaskPool) stopLocalTask(taskID int) {
	for _, t := range p.state.QueueRange() {
		if t != nil && t.Task.ID == taskID && t.Task.Status == task_logger.TaskWaitingStatus {
			t.SetStatus(task_logger.TaskStoppedStatus)
			p.state.DequeueByID(taskID)
			return
		}
	}
	for _, t := range p.state.RunningRange() {
		if t != nil && t.Task.ID == taskID && !t.Task.Status.IsFinished() {
			p.stopTaskRunner(t, true)
			return
		}
	}
}

func (p *TaskPool) StopTask(targetTask db.Task, forceStop bool) error {
	tsk, err := p.GetTask(targetTask.ID)
	if err != nil {
		return err
	}

	// task not active, but exists in database. For non-HA mode
	if tsk == nil {
		tsk = NewTaskRunner(targetTask, p, "", p.keyInstallationService)

		err := tsk.populateDetails()
		if err != nil {
			return err
		}
		tsk.SetStatus(task_logger.TaskStoppedStatus)
		tsk.createTaskEvent()
		return nil
	}

	p.stopTaskRunner(tsk, forceStop)

	return nil
}

// StopTasksByTemplate stops all active (queued or running) tasks that belong to
// the specified project and template. If forceStop is true, tasks are marked as
// stopped immediately and running tasks are killed; otherwise tasks are marked
// as stopping and will gracefully transition to stopped.
//
// Waiting tasks (which have no running process) are dequeued and bulk-updated in
// the database in a single query, avoiding expensive per-task hydration.
// Non-waiting tasks go through the regular per-task SetStatus path.
func (p *TaskPool) StopTasksByTemplate(projectID int, templateID int, forceStop bool) {

	stoppedTasks := map[int]struct{}{}

	// Bulk-update all waiting tasks in DB in a single query.
	// This is the fast path -- waiting tasks have no running process.
	if err := p.store.SetWaitingTasksToStopped(projectID, templateID); err != nil {
		log.Error(err)
	}

	// Snapshot the queue and dequeue by task ID. In HA mode the shared queue can
	// shift between QueueGet(i) and DequeueAt(i); DequeueByID matches handleQueue.
	for _, t := range p.state.QueueRange() {
		if t == nil {
			continue
		}
		if t.Task.ProjectID != projectID || t.Task.TemplateID != templateID {
			continue
		}
		if t.Task.Status.IsFinished() {
			continue
		}

		if t.Task.Status == task_logger.TaskWaitingStatus {
			stoppedTasks[t.Task.ID] = struct{}{}
			p.state.DequeueByID(t.Task.ID)
			continue
		}

		if forceStop {
			t.SetStatus(task_logger.TaskStoppedStatus)
		} else {
			t.SetStatus(task_logger.TaskStoppingStatus)
		}
		stoppedTasks[t.Task.ID] = struct{}{}
	}

	// Handle running tasks -- these need per-task SetStatus and kill.
	for _, t := range p.state.RunningRange() {
		if t == nil {
			continue
		}
		if t.Task.ProjectID != projectID || t.Task.TemplateID != templateID {
			continue
		}
		if t.Task.Status.IsFinished() {
			continue
		}

		p.stopTaskRunner(t, forceStop)

		stoppedTasks[t.Task.ID] = struct{}{}
	}

	// Handle non-waiting tasks in DB that are neither queued nor running locally
	// (e.g., HA mode or tasks created but not present in this instance's memory).
	tasks, err := p.store.GetTemplateTasks(projectID, templateID, db.RetrieveQueryParams{
		TaskFilter: &db.TaskFilter{
			Status: task_logger.UnfinishedTaskStatuses(),
		},
	})

	if err != nil {
		log.Error(err)
		return
	}

	for _, twt := range tasks {

		if _, ok := stoppedTasks[twt.ID]; ok {
			continue
		}

		tsk, taskErr := p.GetTask(twt.ID)
		if taskErr != nil {
			log.WithError(taskErr).WithFields(log.Fields{
				"task_id": twt.ID,
				"context": "task_pool",
			}).Warn("can't get task")

			continue
		}

		if tsk == nil {
			tsk = NewTaskRunner(twt.Task, p, "", p.keyInstallationService)
			if trErr := tsk.populateDetails(); trErr != nil {
				log.Error(trErr)
				continue
			}
		}

		tsk.SetStatus(task_logger.TaskStoppedStatus)

		// In HA a remote task dispatched on another node lives in the shared
		// running/active/claim sets but has no goroutine on any node that will
		// run finishRun for it. Once we mark it finished in the DB the runner's
		// terminal report is ignored (UpdateRunner skips finished tasks) and the
		// timeout backstop bails on IsFinished(), so without finalizing here the
		// shared pool state (parallel-task capacity, runner slots) would leak
		// until restart. FinalizeRemoteTask releases it (finishRun -> onTaskStop)
		// and also emits the finished task event; TryFinalize dedups across nodes
		// and against the running-tasks loop above, so it runs at most once.
		if tsk.job != nil && tsk.job.Async() {
			go p.FinalizeRemoteTask(tsk, nil)
		} else {
			tsk.createTaskEvent()
		}
	}
}

// StopTasksByWorkflowRun stops every active (queued or running) task that
// belongs to the given workflow run. It mirrors StopTasksByTemplate but scopes
// the selection by workflow_run_id instead of template_id, and is used when a
// user stops a whole workflow run.
//
// Waiting tasks are marked stopped and dequeued from the in-memory queue (so the
// queue loop never starts them); running tasks go through stopTaskRunner (kill +
// status transition); tasks that exist in the DB but are not in this instance's
// memory (HA, or a remote task dispatched elsewhere) are marked stopped and
// finalized so their pool bookkeeping is released.
func (p *TaskPool) StopTasksByWorkflowRun(projectID int, runID int, forceStop bool) {
	stoppedTasks := map[int]struct{}{}

	belongsToRun := func(t *TaskRunner) bool {
		return t != nil &&
			t.Task.ProjectID == projectID &&
			t.Task.WorkflowRunID != nil && *t.Task.WorkflowRunID == runID &&
			!t.Task.Status.IsFinished()
	}

	// Waiting tasks have no running process: mark them stopped and remove them
	// from the queue so the queue loop does not later pick them up and run them
	// (run() only converts the "stopping" status to "stopped", not a queued task
	// already set to "stopped").
	for _, t := range p.state.QueueRange() {
		if !belongsToRun(t) || t.Task.Status != task_logger.TaskWaitingStatus {
			continue
		}
		t.SetStatus(task_logger.TaskStoppedStatus)
		p.state.DequeueByID(t.Task.ID)
		stoppedTasks[t.Task.ID] = struct{}{}
	}

	// Running tasks need a per-task stop (kill + status transition).
	for _, t := range p.state.RunningRange() {
		if !belongsToRun(t) {
			continue
		}
		p.stopTaskRunner(t, forceStop)
		stoppedTasks[t.Task.ID] = struct{}{}
	}

	// Unfinished tasks in the DB that are neither queued nor running locally
	// (e.g. HA mode, or a remote task dispatched on another node).
	tasks, err := p.store.GetProjectTasks(projectID, db.RetrieveQueryParams{
		TaskFilter: &db.TaskFilter{
			Status: task_logger.UnfinishedTaskStatuses(),
		},
	})
	if err != nil {
		log.Error(err)
		return
	}

	for _, twt := range tasks {
		if twt.WorkflowRunID == nil || *twt.WorkflowRunID != runID {
			continue
		}
		if _, ok := stoppedTasks[twt.ID]; ok {
			continue
		}

		// A task not handled locally may be queued or running on another HA
		// node: ask its owner to kill it. Fire-and-forget — the stopped status
		// persisted below and the orphan cleaner remain the backstops.
		if broadcaster, ok := p.state.(TaskStopBroadcaster); ok {
			broadcaster.BroadcastTaskStop(twt.ID)
		}

		tsk, taskErr := p.GetTask(twt.ID)
		if taskErr != nil {
			log.WithError(taskErr).WithFields(log.Fields{
				"task_id": twt.ID,
				"context": "task_pool",
			}).Warn("can't get task")
			continue
		}

		if tsk == nil {
			tsk = NewTaskRunner(twt.Task, p, "", p.keyInstallationService)
			if trErr := tsk.populateDetails(); trErr != nil {
				log.Error(trErr)
				continue
			}
		}

		tsk.SetStatus(task_logger.TaskStoppedStatus)

		// A remote task has no goroutine on this node that would run finishRun,
		// so finalize it here to release the shared pool state; a local task is
		// already done from the pool's perspective and only needs its event.
		if tsk.job != nil && tsk.job.Async() {
			go p.FinalizeRemoteTask(tsk, nil)
		} else {
			tsk.createTaskEvent()
		}
	}
}

// GetQueuedTasks returns a snapshot of tasks currently queued
func (p *TaskPool) GetQueuedTasks() []*TaskRunner {
	return p.state.QueueRange()
}

// AddTask creates and queues a new task for execution in the task pool.
//
// Parameters:
//   - taskObj: The task object with initial configuration
//   - userID: Optional ID of the user initiating the task
//   - username: Username of the user initiating the task
//   - projectID: ID of the project this task belongs to
//   - needAlias: Whether to generate a unique alias for the task
//
// The method:
//   - Sets initial task properties (created time, waiting status, etc.)
//   - Validates the task against its template
//   - For build templates, calculates the next version number
//   - Creates the task record in the database
//   - Sets up appropriate job handler (local or remote)
//   - Queues the task for execution
//
// Returns:
//   - The newly created task with all properties set
//   - An error if task creation or validation fails
func (p *TaskPool) AddTask(
	taskObj db.Task,
	userID *int,
	username string,
	projectID int,
	needAlias bool,
) (newTask db.Task, err error) {
	taskObj.Created = tz.Now()
	taskObj.Status = task_logger.TaskWaitingStatus
	taskObj.UserID = userID
	taskObj.ProjectID = projectID
	extraSecretVars := taskObj.Secret
	taskObj.Secret = "{}"

	tpl, err := p.store.GetTemplate(projectID, taskObj.TemplateID)
	if err != nil {
		return
	}

	err = taskObj.ValidateNewTask(tpl)
	if err != nil {
		return
	}

	if tpl.Type == db.TemplateBuild { // get next version for TaskRunner if it is a Build
		var builds []db.TaskWithTpl
		builds, err = p.store.GetTemplateTasks(tpl.ProjectID, tpl.ID, db.RetrieveQueryParams{Count: 1})
		if err != nil {
			return
		}
		if len(builds) == 0 || builds[0].Version == nil {
			taskObj.Version = tpl.StartVersion
		} else {
			v := db.GetNextBuildVersion(*tpl.StartVersion, *builds[0].Version)
			taskObj.Version = &v
		}
	}

	newTask, err = p.store.CreateTask(taskObj, util.Config.MaxTasksPerTemplate)
	if err != nil {
		return
	}

	taskRunner := NewTaskRunner(newTask, p, username, p.keyInstallationService)

	if needAlias {
		// A unique, randomly-generated identifier that persists throughout the task's lifecycle.
		taskRunner.Alias = random.String(32)
	}

	err = taskRunner.populateDetails()
	if err != nil {
		taskRunner.Log("Error: " + err.Error())
		taskRunner.SetStatus(task_logger.TaskFailStatus)
		return
	}

	var job Job

	if util.Config.UseRemoteRunner ||
		taskRunner.Template.RunnerTag != nil ||
		taskRunner.Inventory.RunnerTag != nil {

		tag := taskRunner.Template.RunnerTag
		if tag == nil {
			tag = taskRunner.Inventory.RunnerTag
		}

		job = &RemoteJob{
			RunnerTag: tag,
			Task:      taskRunner.Task,
			taskPool:  p,
		}
	} else {
		app := db_lib.CreateApp(
			taskRunner.Template,
			taskRunner.Repository,
			taskRunner.Inventory,
			taskRunner)

		job = &LocalExecutor{
			Task:         taskRunner.Task,
			Template:     taskRunner.Template,
			Inventory:    taskRunner.Inventory,
			Repository:   taskRunner.Repository,
			Environment:  taskRunner.Environment,
			Secret:       extraSecretVars,
			Logger:       app.SetLogger(taskRunner),
			App:          app,
			KeyInstaller: p.keyInstallationService,
		}
	}

	taskRunner.job = job

	p.register <- taskRunner

	taskRunner.createTaskEvent()

	return
}
