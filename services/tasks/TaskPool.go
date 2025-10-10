package tasks

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
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
	EventTypeNew      EventType = 0
	EventTypeFinished EventType = 1
	EventTypeFailed   EventType = 2
	EventTypeEmpty    EventType = 3
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
	}
	// attempt to start HA state store (no-op for memory)
	_ = p.state.Start(p.hydrateTaskRunner)
	return p
}

// CreateTaskPoolWithState allows passing a custom TaskStateStore (e.g., Redis-backed)
func CreateTaskPoolWithState(
	stateStore TaskStateStore,
	store db.Store,
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
		queueEvents:            make(chan PoolEvent),
		state:                  stateStore,
		inventoryService:       inventoryService,
		ansibleTaskRepo:        ansibleTaskRepo,
		encryptionService:      encryptionService,
		logWriteService:        logWriteService,
		keyInstallationService: keyInstallationService,
	}
	_ = p.state.Start(p.hydrateTaskRunner)
	return p
}
func (p *TaskPool) GetNumberOfRunningTasksOfRunner(runnerID int) (res int) {
	for _, task := range p.state.RunningRange() {
		if task.RunnerID == runnerID {
			res++
		}
	}
	return
}

func (p *TaskPool) GetRunningTasks() (res []*TaskRunner) {
	return p.state.RunningRange()
}

func (p *TaskPool) GetTask(id int) (task *TaskRunner) {
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

	return
}

func (p *TaskPool) GetTaskByAlias(alias string) (task *TaskRunner) {
	return p.state.GetByAlias(alias)
}

// nolint: gocyclo
func (p *TaskPool) Run() {
	ticker := time.NewTicker(5 * time.Second)

	defer ticker.Stop()

	go p.handleQueue()
	go p.handleLogs()

	for {
		select {
		case task := <-p.register: // new task created by API or schedule

			db.StoreSession(p.store, "new task", func() {
				//p.Queue = append(p.Queue, task)
				msg := "Task " + getTaskName(task) + " added to queue"
				task.Log(msg)
				log.Info(msg)
				task.saveStatus()
			})
			p.queueEvents <- PoolEvent{EventTypeNew, task}

		case <-ticker.C: // timer 5 seconds
			p.queueEvents <- PoolEvent{EventTypeEmpty, nil}

		}
	}
}

func getTaskName(t *TaskRunner) string {
	return t.Template.Name + " " + strconv.Itoa(t.Task.ID)
}

func (p *TaskPool) handleQueue() {
	for t := range p.queueEvents {
		switch t.eventType {
		case EventTypeNew:
			p.state.Enqueue(t.task)
		case EventTypeFinished:
			p.onTaskStop(t.task)
		}

		if p.state.QueueLen() == 0 {
			continue
		}

		var i = 0
		for i < p.state.QueueLen() {
			curr := p.state.QueueGet(i)
			if curr == nil { // item may no longer be local, move ahead
				i = i + 1
				continue
			}

			if curr.Task.Status == task_logger.TaskFailStatus {
				//delete failed TaskRunner from queue
				_ = p.state.DequeueAt(i)
				log.Info("Task " + getTaskName(curr) + " removed from queue")
				continue
			}

			if p.blocks(curr) {
				i = i + 1
				continue
			}

			// ensure only one instance claims the task before dequeue
			if !p.state.TryClaim(curr.Task.ID) {
				i = i + 1
				continue
			}

			_ = p.state.DequeueAt(i)
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

		db.StoreSession(p.store, "logger", func() {

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
		})
		taskOutput = append(taskOutput, newOutput)
	}

	db.StoreSession(p.store, "logger", func() {
		err := p.store.InsertTaskOutputBatch(taskOutput)
		if err != nil {
			log.Error(err)
			return
		}
	})
}

func runTask(task *TaskRunner, p *TaskPool) {
	log.Info("Set resource locker with TaskRunner " + getTaskName(task))
	p.onTaskRun(task)

	log.Info("Task " + getTaskName(task) + " started")
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

// hydrateTaskRunner builds a TaskRunner for an existing task from DB without starting it
func (p *TaskPool) hydrateTaskRunner(taskID int, projectID int) (*TaskRunner, error) {
	task, err := p.store.GetTask(projectID, taskID)
	if err != nil {
		return nil, err
	}
	tr := NewTaskRunner(task, p, "", p.keyInstallationService)
	if err := tr.populateDetails(); err != nil {
		return nil, err
	}
	// load runtime fields from HA store (e.g., Redis)
	if p.state != nil {
		p.state.LoadRuntimeFields(tr)
	}
	// set appropriate job handler for consistency (not run)
	var job Job
	if util.Config.UseRemoteRunner || tr.Template.RunnerTag != nil || tr.Inventory.RunnerTag != nil {
		tag := tr.Template.RunnerTag
		if tag == nil {
			tag = tr.Inventory.RunnerTag
		}
		job = &RemoteJob{RunnerTag: tag, Task: tr.Task, taskPool: p}
	} else {
		app := db_lib.CreateApp(tr.Template, tr.Repository, tr.Inventory, tr)
		job = &LocalJob{
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
	tsk := p.GetTask(targetTask.ID)

	if tsk == nil { // task not active, but exists in database
		return fmt.Errorf("task is not active")
	}

	tsk.SetStatus(task_logger.TaskConfirmed)

	return nil
}

func (p *TaskPool) RejectTask(targetTask db.Task) error {
	tsk := p.GetTask(targetTask.ID)

	if tsk == nil { // task not active, but exists in database
		return fmt.Errorf("task is not active")
	}

	tsk.SetStatus(task_logger.TaskRejected)

	return nil
}

func (p *TaskPool) StopTask(targetTask db.Task, forceStop bool) error {
	tsk := p.GetTask(targetTask.ID)
	if tsk == nil { // task not active, but exists in database

		tsk = NewTaskRunner(targetTask, p, "", p.keyInstallationService)

		err := tsk.populateDetails()
		if err != nil {
			return err
		}
		tsk.SetStatus(task_logger.TaskStoppedStatus)
		tsk.createTaskEvent()
	} else {
		status := tsk.Task.Status

		if forceStop {
			tsk.SetStatus(task_logger.TaskStoppedStatus)
		} else {
			tsk.SetStatus(task_logger.TaskStoppingStatus)
		}

		if status == task_logger.TaskRunningStatus {
			tsk.kill()
		}
	}

	return nil
}

// StopTasksByTemplate stops all active (queued or running) tasks that belong to
// the specified project and template. If forceStop is true, tasks are marked as
// stopped immediately and running tasks are killed; otherwise tasks are marked
// as stopping and will gracefully transition to stopped.
func (p *TaskPool) StopTasksByTemplate(projectID int, templateID int, forceStop bool) {
	// Handle queued tasks
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
		if forceStop {
			t.SetStatus(task_logger.TaskStoppedStatus)
		} else {
			t.SetStatus(task_logger.TaskStoppingStatus)
		}
		// Queued tasks will be dequeued and immediately finalize to Stopped in run()
	}

	// Handle running tasks
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
		prevStatus := t.Task.Status
		if forceStop {
			t.SetStatus(task_logger.TaskStoppedStatus)
		} else {
			t.SetStatus(task_logger.TaskStoppingStatus)
		}
		if prevStatus == task_logger.TaskRunningStatus {
			t.kill()
		}
	}

	// Update tasks in DB that are neither queued nor running but still active
	// (e.g., created but not present in this instance's memory state).
	if tasks, err := p.store.GetTemplateTasks(projectID, templateID, db.RetrieveQueryParams{
		TaskFilter: &db.TaskFilter{
			Status: task_logger.UnfinishedTaskStatuses(),
		},
	}); err == nil {
		for _, twt := range tasks {

			// if task is managed locally (queued/running), it was handled above
			if p.GetTask(twt.Task.ID) != nil {
				continue
			}

			// mark non-local task as stopped and write event for history
			tr := NewTaskRunner(twt.Task, p, "", p.keyInstallationService)
			if err := tr.populateDetails(); err != nil {
				log.Error(err)
				continue
			}

			tr.SetStatus(task_logger.TaskStoppedStatus)
			tr.createTaskEvent()
		}
	} else {
		log.Error(err)
	}
}

// GetQueuedTasks returns a snapshot of tasks currently queued
func (p *TaskPool) GetQueuedTasks() []*TaskRunner {
	return p.state.QueueRange()
}

func getNextBuildVersion(startVersion string, currentVersion string) string {
	re := regexp.MustCompile(`^(.*[^\d])?(\d+)([^\d].*)?$`)
	m := re.FindStringSubmatch(startVersion)

	if m == nil {
		return startVersion
	}

	var prefix, suffix, body string

	switch len(m) - 1 {
	case 3:
		prefix = m[1]
		body = m[2]
		suffix = m[3]
	case 2:
		if _, err := strconv.Atoi(m[1]); err == nil {
			body = m[1]
			suffix = m[2]
		} else {
			prefix = m[1]
			body = m[2]
		}
	case 1:
		body = m[1]
	default:
		return startVersion
	}

	if !strings.HasPrefix(currentVersion, prefix) ||
		!strings.HasSuffix(currentVersion, suffix) {
		return startVersion
	}

	curr, err := strconv.Atoi(currentVersion[len(prefix) : len(currentVersion)-len(suffix)])
	if err != nil {
		return startVersion
	}

	start, err := strconv.Atoi(body)
	if err != nil {
		panic(err)
	}

	var newVer int
	if start > curr {
		newVer = start
	} else {
		newVer = curr + 1
	}

	return prefix + strconv.Itoa(newVer) + suffix
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
			v := getNextBuildVersion(*tpl.StartVersion, *builds[0].Version)
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

		job = &LocalJob{
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
