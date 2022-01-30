package tasks

import (
	"github.com/ansible-semaphore/semaphore/db"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/util"
)

type logRecord struct {
	task   *TaskRunner
	output string
	time   time.Time
}

type TaskPool struct {
	queue        []*TaskRunner
	register     chan *TaskRunner
	activeProj   map[int]*TaskRunner
	activeNodes  map[string]*TaskRunner
	running      int
	runningTasks map[int]*TaskRunner
	logger       chan logRecord
	store        db.Store
}

type resourceLock struct {
	lock   bool
	holder *TaskRunner
}

var resourceLocker = make(chan *resourceLock)

func (p *TaskPool) GetTask(id int) (task *TaskRunner) {

	for _, t := range p.queue {
		if t.task.ID == id {
			task = t
			break
		}
	}

	if task == nil {
		for _, t := range p.runningTasks {
			if t.task.ID == id {
				task = t
				break
			}
		}
	}

	return
}

//nolint: gocyclo
func (p *TaskPool) Run() {
	ticker := time.NewTicker(5 * time.Second)

	defer func() {
		close(resourceLocker)
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

				p.activeProj[t.task.ProjectID] = t

				for _, node := range t.hosts {
					p.activeNodes[node] = t
				}

				p.running++
				p.runningTasks[t.task.ID] = t
				continue
			}

			if p.activeProj[t.task.ProjectID] == t {
				delete(p.activeProj, t.task.ProjectID)
			}

			for _, node := range t.hosts {
				delete(p.activeNodes, node)
			}

			p.running--
			delete(p.runningTasks, t.task.ID)
		}
	}(resourceLocker)

	for {
		select {
		case record := <-p.logger:
			_, err := record.task.pool.store.CreateTaskOutput(db.TaskOutput{
				TaskID: record.task.task.ID,
				Output: record.output,
				Time:   record.time,
			})

			if err != nil {
				log.Error(err)
			}
		case task := <-p.register:
			p.queue = append(p.queue, task)
			log.Debug(task)
			msg := "Task " + strconv.Itoa(task.task.ID) + " added to queue"
			task.Log(msg)
			log.Info(msg)

		case <-ticker.C:
			if len(p.queue) == 0 {
				continue
			}

			//get TaskRunner from top of queue
			t := p.queue[0]
			if t.task.Status == db.TaskFailStatus {
				//delete failed TaskRunner from queue
				p.queue = p.queue[1:]
				log.Info("Task " + strconv.Itoa(t.task.ID) + " removed from queue")
				continue
			}
			if p.blocks(t) {
				//move blocked TaskRunner to end of queue
				p.queue = append(p.queue[1:], t)
				continue
			}
			log.Info("Set resource locker with TaskRunner " + strconv.Itoa(t.task.ID))
			resourceLocker <- &resourceLock{lock: true, holder: t}
			if !t.prepared {
				go t.prepareRun()
				continue
			}
			go t.run()
			p.queue = p.queue[1:]
			log.Info("Task " + strconv.Itoa(t.task.ID) + " removed from queue")
		}
	}
}

func (p *TaskPool) blocks(t *TaskRunner) bool {
	if p.running >= util.Config.MaxParallelTasks {
		return true
	}

	switch util.Config.ConcurrencyMode {
	case "project":
		return p.activeProj[t.task.ProjectID] != nil
	case "node":
		for _, node := range t.hosts {
			if p.activeNodes[node] != nil {
				return true
			}
		}

		return false
	default:
		return p.running > 0
	}
}

func CreateTaskPool(store db.Store) TaskPool {
	return TaskPool{
		queue:        make([]*TaskRunner, 0), // queue of waiting tasks
		register:     make(chan *TaskRunner), // add TaskRunner to queue
		activeProj:   make(map[int]*TaskRunner),
		activeNodes:  make(map[string]*TaskRunner),
		running:      0,                           // number of running tasks
		runningTasks: make(map[int]*TaskRunner),   // working tasks
		logger:       make(chan logRecord, 10000), // store log records to database
		store:        store,
	}
}

func (p *TaskPool) StopTask(targetTask db.Task) error {
	tsk := p.GetTask(targetTask.ID)
	if tsk == nil { // task not active, but exists in database
		tsk = &TaskRunner{
			task: targetTask,
			pool: p,
		}
		err := tsk.populateDetails()
		if err != nil {
			return err
		}
		tsk.setStatus(db.TaskStoppedStatus)
		tsk.createTaskEvent()
	} else {
		status := tsk.task.Status
		tsk.setStatus(db.TaskStoppingStatus)
		if status == db.TaskRunningStatus {
			if tsk.process == nil {
				panic("running process can not be nil")
			}
			err := tsk.process.Kill()
			if err != nil {
				return err
			}
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
	taskObj.Status = db.TaskWaitingStatus
	taskObj.UserID = userID
	taskObj.ProjectID = projectID

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

	newTask, err = p.store.CreateTask(taskObj)
	if err != nil {
		return
	}

	p.register <- &TaskRunner{
		task: newTask,
		pool: p,
	}

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
