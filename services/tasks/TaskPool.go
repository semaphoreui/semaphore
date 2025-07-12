package tasks

import (
	"fmt"
	"github.com/semaphoreui/semaphore/pkg/random"
	"github.com/semaphoreui/semaphore/pkg/tz"
	"github.com/semaphoreui/semaphore/services/server"
	"github.com/semaphoreui/semaphore/services/tasks/stage_parsers"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db_lib"
	"github.com/semaphoreui/semaphore/pkg/task_logger"

	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
)

type logRecord struct {
	task   *TaskRunner
	output string
	time   time.Time
}

type EventType uint

const (
	EventTypeNew      EventType = 0
	EventTypeFinished EventType = 1
	EventTypeFailed   EventType = 2
	EventTypeEmpty    EventType = 3
)

type PoolEvent struct {
	eventType EventType
	task      *TaskRunner
}

type TaskPool struct {
	// Queue contains list of tasks in status TaskWaitingStatus.
	Queue []*TaskRunner

	// register channel used to put tasks to queue.
	register chan *TaskRunner

	// activeProj ???
	activeProj map[int]map[int]*TaskRunner

	// RunningTasks contains tasks with status TaskRunningStatus. Map key is a task ID.
	RunningTasks map[int]*TaskRunner

	// logger channel used to putting log records to database.
	logger chan logRecord

	store                  db.Store
	inventoryService       server.InventoryService
	encryptionService      server.AccessKeyEncryptionService
	keyInstallationService server.AccessKeyInstallationService

	queueEvents chan PoolEvent

	aliases map[string]*TaskRunner
}

func (p *TaskPool) GetNumberOfRunningTasksOfRunner(runnerID int) (res int) {
	for _, task := range p.RunningTasks {
		if task.RunnerID == runnerID {
			res++
		}
	}
	return
}

func (p *TaskPool) GetRunningTasks() (res []*TaskRunner) {
	for _, task := range p.RunningTasks {
		res = append(res, task)
	}
	return
}

func (p *TaskPool) GetTask(id int) (task *TaskRunner) {

	for _, t := range p.Queue {
		if t.Task.ID == id {
			task = t
			break
		}
	}

	if task == nil {
		for _, t := range p.RunningTasks {
			if t.Task.ID == id {
				task = t
				break
			}
		}
	}

	return
}

func (p *TaskPool) GetTaskByAlias(alias string) (task *TaskRunner) {
	return p.aliases[alias]
}

func (p *TaskPool) MoveToNextStage(
	app db.TemplateApp,
	projectID int,
	currentState any,
	currentStage *db.TaskStage,
	currentOutput *db.TaskOutput,
	newOutput db.TaskOutput,
) (newStage *db.TaskStage, newState any, err error) {

	newState = currentState
	stages := stage_parsers.GetAllTaskStages(app)

	for _, stageType := range stages {

		parser := stage_parsers.GetStageResultParser(app, stageType, newState)
		if parser == nil {
			continue
		}

		matched := false

		var oldStage *db.TaskStage

		var stage db.TaskStage

		if parser.IsEnd(currentStage, newOutput) {

			err = p.store.EndTaskStage(
				currentStage.TaskID,
				currentStage.ID,
				newOutput.Time,
				newOutput.ID,
			)

			if err != nil {
				return
			}

			stage = *currentStage
			stage.End = &newOutput.Time
			stage.EndOutputID = &newOutput.ID
			oldStage = &stage

			matched = true

		} else if parser.IsStart(currentStage, newOutput) {

			if currentStage != nil {
				err = p.store.EndTaskStage(
					currentStage.TaskID,
					currentStage.ID,
					currentOutput.Time,
					currentOutput.ID,
				)

				if err != nil {
					return
				}

				oldSt := *currentStage
				oldSt.End = &currentOutput.Time
				oldSt.EndOutputID = &currentOutput.ID
				oldStage = &oldSt
			}

			stage, err = p.store.CreateTaskStage(db.TaskStage{
				TaskID:        newOutput.TaskID,
				Start:         &newOutput.Time,
				Type:          stageType,
				StartOutputID: &newOutput.ID,
				EndOutputID:   nil,
			})

			if err != nil {
				return
			}

			matched = true
		} else {
			var ok bool
			ok, err = parser.Parse(currentStage, newOutput, p.store, projectID)
			if err != nil {
				log.Error(err.Error())
			} else if ok {
				newState = parser.State()
			}
		}

		if matched {

			newStage = &stage

			var oldParser stage_parsers.StageResultParser

			if oldStage != nil {
				oldParser = stage_parsers.GetStageResultParser(app, oldStage.Type, currentState)
			}

			if oldParser != nil && oldParser.NeedParse() {

				res := oldParser.Result()

				err = p.store.CreateTaskStageResult(oldStage.TaskID, oldStage.ID, res)
			}

			break
		}
	}

	return
}

// nolint: gocyclo
func (p *TaskPool) Run() {
	ticker := time.NewTicker(5 * time.Second)

	defer func() {
		ticker.Stop()
	}()

	go p.handleQueue()
	go p.handleLogs()

	for {
		select {
		case task := <-p.register: // new task created by API or schedule

			db.StoreSession(p.store, "new task", func() {
				//p.Queue = append(p.Queue, task)
				msg := "Task " + strconv.Itoa(task.Task.ID) + " added to queue"
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

func (p *TaskPool) handleQueue() {
	for t := range p.queueEvents {
		switch t.eventType {
		case EventTypeNew:
			p.Queue = append(p.Queue, t.task)
		case EventTypeFinished:
			p.onTaskStop(t.task)
		}

		if len(p.Queue) == 0 {
			continue
		}

		var i = 0
		for i < len(p.Queue) {
			curr := p.Queue[i]

			if curr.Task.Status == task_logger.TaskFailStatus {
				//delete failed TaskRunner from queue
				p.Queue = slices.Delete(p.Queue, i, i+1)
				log.Info("Task " + strconv.Itoa(curr.Task.ID) + " removed from queue")
				continue
			}

			if p.blocks(curr) {
				i = i + 1
				continue
			}

			p.Queue = slices.Delete(p.Queue, i, i+1)
			runTask(curr, p)
		}
	}
}

func (p *TaskPool) handleLogs() {

	for record := range p.logger {
		db.StoreSession(p.store, "logger", func() {

			newOutput, err := p.store.CreateTaskOutput(db.TaskOutput{
				TaskID: record.task.Task.ID,
				Output: record.output,
				Time:   record.time,
			})

			if err != nil {
				log.Error(err)
				return
			}

			currentOutput := record.task.currentOutput
			record.task.currentOutput = &newOutput

			newStage, newState, err := p.MoveToNextStage(
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
		})
	}
}

func runTask(task *TaskRunner, p *TaskPool) {
	log.Info("Set resource locker with TaskRunner " + strconv.Itoa(task.Task.ID))

	p.onTaskRun(task)

	log.Info("Task " + strconv.Itoa(task.Task.ID) + " started")
	go task.run()
}

func (p *TaskPool) onTaskRun(t *TaskRunner) {
	projTasks, ok := p.activeProj[t.Task.ProjectID]
	if !ok {
		projTasks = make(map[int]*TaskRunner)
		p.activeProj[t.Task.ProjectID] = projTasks
	}
	projTasks[t.Task.ID] = t
	p.RunningTasks[t.Task.ID] = t
	p.aliases[t.Alias] = t
}

func (p *TaskPool) onTaskStop(t *TaskRunner) {
	if p.activeProj[t.Task.ProjectID] != nil && p.activeProj[t.Task.ProjectID][t.Task.ID] != nil {
		delete(p.activeProj[t.Task.ProjectID], t.Task.ID)
		if len(p.activeProj[t.Task.ProjectID]) == 0 {
			delete(p.activeProj, t.Task.ProjectID)
		}
	}

	delete(p.RunningTasks, t.Task.ID)
	delete(p.aliases, t.Alias)
}

func (p *TaskPool) blocks(t *TaskRunner) bool {

	if util.Config.MaxParallelTasks > 0 && len(p.RunningTasks) >= util.Config.MaxParallelTasks {
		return true
	}

	if p.activeProj[t.Task.ProjectID] == nil || len(p.activeProj[t.Task.ProjectID]) == 0 {
		return false
	}

	for _, r := range p.activeProj[t.Task.ProjectID] {
		if r.Task.Status.IsFinished() {
			continue
		}
		if r.Template.ID == t.Task.TemplateID {
			return true
		}
	}

	proj, err := p.store.GetProject(t.Task.ProjectID)

	if err != nil {
		log.Error(err)
		return false
	}

	res := proj.MaxParallelTasks > 0 && len(p.activeProj[t.Task.ProjectID]) >= proj.MaxParallelTasks

	if res {
		return true
	}

	return res
}

func CreateTaskPool(
	store db.Store,
	inventoryService server.InventoryService,
	encryptionService server.AccessKeyEncryptionService,
	keyInstallationService server.AccessKeyInstallationService,
) TaskPool {
	return TaskPool{
		Queue:                  make([]*TaskRunner, 0), // queue of waiting tasks
		register:               make(chan *TaskRunner), // add TaskRunner to queue
		activeProj:             make(map[int]map[int]*TaskRunner),
		RunningTasks:           make(map[int]*TaskRunner),   // working tasks
		logger:                 make(chan logRecord, 10000), // store log records to database
		store:                  store,
		queueEvents:            make(chan PoolEvent),
		aliases:                make(map[string]*TaskRunner),
		inventoryService:       inventoryService,
		encryptionService:      encryptionService,
		keyInstallationService: keyInstallationService,
	}
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
