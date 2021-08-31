package tasks

import (
	"github.com/ansible-semaphore/semaphore/db"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/util"
)

type logRecord struct {
	task *task
	output string
	time time.Time
}

type taskPool struct {
	queue        []*task
	register     chan *task
	activeProj   map[int]*task
	activeNodes  map[string]*task
	running      int
	runningTasks map[int]*task
	logger	     chan logRecord
}

type resourceLock struct {
	lock   bool
	holder *task
}

var pool = taskPool{
	queue:        make([]*task, 0), // queue of waiting tasks
	register:     make(chan *task), // add task to queue
	activeProj:   make(map[int]*task),
	activeNodes:  make(map[string]*task),
	running:      0, // number of running tasks
	runningTasks: make(map[int]*task), // working tasks
	logger:       make(chan logRecord, 10000), // store log records to database
}

var resourceLocker = make(chan *resourceLock)

func (p *taskPool) getTask(id int) (task *task){

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
func (p *taskPool) run() {
	ticker := time.NewTicker(5 * time.Second)

	defer func() {
		close(resourceLocker)
		ticker.Stop()
	}()

	// Lock or unlock resources when running a task
	go func(locker <-chan *resourceLock) {
		for l := range locker {
			t := l.holder

			if l.lock {
				if p.blocks(t) {
					panic("Trying to lock an already locked resource!")
				}

				p.activeProj[t.projectID] = t

				for _, node := range t.hosts {
					p.activeNodes[node] = t
				}

				p.running++
				p.runningTasks[t.task.ID] = t
				continue
			}

			if p.activeProj[t.projectID] == t {
				delete(p.activeProj, t.projectID)
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
			_, err := record.task.store.CreateTaskOutput(db.TaskOutput{
				TaskID: record.task.task.ID,
				Output: record.output,
				Time: record.time,
			})

			if err != nil {
				log.Error(err)
			}
		case task := <-p.register:
			p.queue = append(p.queue, task)
			log.Debug(task)
			msg := "Task " + strconv.Itoa(task.task.ID) + " added to queue"
			task.log(msg)
			log.Info(msg)

		case <-ticker.C:
			if len(p.queue) == 0 {
				continue
			}

			//get task from top of queue
			t := p.queue[0]
			if t.task.Status == taskFailStatus {
				//delete failed task from queue
				p.queue = p.queue[1:]
				log.Info("Task " + strconv.Itoa(t.task.ID) + " removed from queue")
				continue
			}
			if p.blocks(t) {
				//move blocked task to end of queue
				p.queue = append(p.queue[1:], t)
				continue
			}
			log.Info("Set resource locker with task " + strconv.Itoa(t.task.ID))
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

func (p *taskPool) blocks(t *task) bool {
	if p.running >= util.Config.MaxParallelTasks {
		return true
	}

	switch util.Config.ConcurrencyMode {
	case "project":
		return p.activeProj[t.projectID] != nil
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

// StartRunner begins the task pool, used as a goroutine
func StartRunner() {
	pool.run()
}
