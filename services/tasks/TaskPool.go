package tasks

import (
	"errors"
	"fmt"
	"github.com/semaphoreui/semaphore/pkg/random"
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
	done      chan bool
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

	store db.Store

	queueEvents chan PoolEvent

	aliases map[string]*TaskRunner
}

var ErrInvalidSubscription = errors.New("has no active subscription")

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

// nolint: gocyclo
func (p *TaskPool) Run() {
	ticker := time.NewTicker(5 * time.Second)

	defer func() {
		ticker.Stop()
	}()

	go p.handleQueue()

	for {
		select {
		case record := <-p.logger: // new log message which should be put to database
			db.StoreSession(p.store, "logger", func() {
				_, err := p.store.CreateTaskOutput(db.TaskOutput{
					TaskID: record.task.Task.ID,
					Output: record.output,
					Time:   record.time,
				})
				if err != nil {
					log.Error(err)
				}
			})

		case task := <-p.register: // new task created by API or schedule

			db.StoreSession(p.store, "new task", func() {
				//p.Queue = append(p.Queue, task)
				log.Debug(task)
				msg := "Task " + strconv.Itoa(task.Task.ID) + " added to queue"
				task.Log(msg)
				log.Info(msg)
				task.saveStatus()
			})
			p.queueEvents <- PoolEvent{EventTypeNew, task, make(chan bool)}

		case <-ticker.C: // timer 5 seconds
			p.queueEvents <- PoolEvent{EventTypeEmpty, nil, make(chan bool)}

		}
	}
}

func (p *TaskPool) handleQueue() {
	for t := range p.queueEvents {
		func() {
			defer close(t.done)

			switch t.eventType {
			case EventTypeNew:
				p.Queue = append(p.Queue, t.task)
			case EventTypeFinished:
				p.onTaskStop(t.task)
			}

			if len(p.Queue) == 0 {
				return
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
		}()
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

	return proj.MaxParallelTasks > 0 && len(p.activeProj[t.Task.ProjectID]) >= proj.MaxParallelTasks
}

func CreateTaskPool(store db.Store) TaskPool {
	return TaskPool{
		Queue:        make([]*TaskRunner, 0), // queue of waiting tasks
		register:     make(chan *TaskRunner), // add TaskRunner to queue
		activeProj:   make(map[int]map[int]*TaskRunner),
		RunningTasks: make(map[int]*TaskRunner),   // working tasks
		logger:       make(chan logRecord, 10000), // store log records to database
		store:        store,
		queueEvents:  make(chan PoolEvent),
		aliases:      make(map[string]*TaskRunner),
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
		tsk = &TaskRunner{
			Task: targetTask,
			pool: p,
		}
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

func (p *TaskPool) AddTask(taskObj db.Task, userID *int, projectID int, needAlias bool) (newTask db.Task, err error) {
	taskObj.Created = time.Now()
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

	taskRunner := TaskRunner{
		Task: newTask,
		pool: p,
	}

	if needAlias {
		taskRunner.Alias = random.String(32)
	}

	err = taskRunner.populateDetails()
	if err != nil {
		taskRunner.Log("Error: " + err.Error())
		taskRunner.SetStatus(task_logger.TaskFailStatus)
		return
	}

	var job Job

	if util.Config.UseRemoteRunner || taskRunner.Template.RunnerTag != nil {
		job = &RemoteJob{
			RunnerTag: taskRunner.Template.RunnerTag,
			Task:      taskRunner.Task,
			taskPool:  p,
		}
	} else {
		app := db_lib.CreateApp(
			taskRunner.Template,
			taskRunner.Repository,
			taskRunner.Inventory,
			&taskRunner)

		job = &LocalJob{
			Task:        taskRunner.Task,
			Template:    taskRunner.Template,
			Inventory:   taskRunner.Inventory,
			Repository:  taskRunner.Repository,
			Environment: taskRunner.Environment,
			Secret:      extraSecretVars,
			Logger:      app.SetLogger(&taskRunner),
			App:         app,
		}
	}

	taskRunner.job = job

	p.register <- &taskRunner

	objType := db.EventTask
	desc := "Task ID " + strconv.Itoa(newTask.ID) + " queued for running"
	_, err = p.store.CreateEvent(db.Event{
		UserID:      userID,
		ProjectID:   &projectID,
		ObjectType:  &objType,
		ObjectID:    &newTask.ID,
		Description: &desc,
	})

	return
}
