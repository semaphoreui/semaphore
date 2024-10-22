package tasks

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/db_lib"
	"github.com/ansible-semaphore/semaphore/pkg/task_logger"

	"github.com/ansible-semaphore/semaphore/util"
	log "github.com/sirupsen/logrus"
)

type logRecord struct {
	task   *TaskRunner
	output string
	time   time.Time
}

type resourceLock struct {
	lock   bool
	holder *TaskRunner
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

	resourceLocker chan *resourceLock
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

// nolint: gocyclo
func (p *TaskPool) Run() {
	tickerDuration := time.Duration(int64(util.Config.TaskPoolTickerSec)) * time.Second
	if tickerDuration.Seconds() == 0 {
		tickerDuration = 5 * time.Second
	}
	ticker := time.NewTicker(tickerDuration)

	defer func() {
		close(p.resourceLocker)
		ticker.Stop()
	}()

	// Lock or unlock resources when running a TaskRunner
	go func(locker <-chan *resourceLock) {
		for l := range locker {
			t := l.holder

			if l.lock {
				if p.blocks(t) {
					panic("Trying to lock an already locked resource!")
				}

				projTasks, ok := p.activeProj[t.Task.ProjectID]
				if !ok {
					projTasks = make(map[int]*TaskRunner)
					p.activeProj[t.Task.ProjectID] = projTasks
				}
				projTasks[t.Task.ID] = t
				p.RunningTasks[t.Task.ID] = t
				continue
			}

			if p.activeProj[t.Task.ProjectID] != nil && p.activeProj[t.Task.ProjectID][t.Task.ID] != nil {
				delete(p.activeProj[t.Task.ProjectID], t.Task.ID)
				if len(p.activeProj[t.Task.ProjectID]) == 0 {
					delete(p.activeProj, t.Task.ProjectID)
				}
			}

			delete(p.RunningTasks, t.Task.ID)
		}
	}(p.resourceLocker)

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
				p.Queue = append(p.Queue, task)
				log.Debug(task)
				msg := "Task " + strconv.Itoa(task.Task.ID) + " added to queue"
				task.Log(msg)
				log.Info(msg)
				task.saveStatus()
			})

		case <-ticker.C: // timer 5 seconds
			if len(p.Queue) == 0 {
				break
			}

			//get TaskRunner from top of queue
			t := p.Queue[0]
			if t.Task.Status == task_logger.TaskFailStatus {
				//delete failed TaskRunner from queue
				p.Queue = p.Queue[1:]
				log.Info("Task " + strconv.Itoa(t.Task.ID) + " removed from queue")
				break
			}

			if p.blocks(t) {
				//move blocked TaskRunner to end of queue
				p.Queue = append(p.Queue[1:], t)
				break
			}

			log.Info("Set resource locker with TaskRunner " + strconv.Itoa(t.Task.ID))
			p.resourceLocker <- &resourceLock{lock: true, holder: t}

			go t.run()

			p.Queue = p.Queue[1:]
			log.Info("Task " + strconv.Itoa(t.Task.ID) + " removed from queue")
		}
	}
}

func (p *TaskPool) blocks(t *TaskRunner) bool {

	if util.Config.MaxParallelTasks > 0 && len(p.RunningTasks) >= util.Config.MaxParallelTasks {
		return true
	}

	if p.activeProj[t.Task.ProjectID] == nil || len(p.activeProj[t.Task.ProjectID]) == 0 {
		return false
	}

	if !util.Config.TaskFreeConcurrencyMode {
		for _, r := range p.activeProj[t.Task.ProjectID] {
			if r.Task.Status.IsFinished() {
				continue
			}
			if r.Template.ID == t.Task.TemplateID {
				return true
			}
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
		Queue:          make([]*TaskRunner, 0), // queue of waiting tasks
		register:       make(chan *TaskRunner), // add TaskRunner to queue
		activeProj:     make(map[int]map[int]*TaskRunner),
		RunningTasks:   make(map[int]*TaskRunner),   // working tasks
		logger:         make(chan logRecord, 10000), // store log records to database
		store:          store,
		resourceLocker: make(chan *resourceLock),
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

func (p *TaskPool) AddTask(taskObj db.Task, userID *int, projectID int) (newTask db.Task, err error) {
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

	err = taskRunner.populateDetails()
	if err != nil {
		taskRunner.Log("Error: " + err.Error())
		taskRunner.SetStatus(task_logger.TaskFailStatus)
		return
	}

	var job Job

	if util.Config.UseRemoteRunner {
		job = &RemoteJob{
			Task:     taskRunner.Task,
			taskPool: p,
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
