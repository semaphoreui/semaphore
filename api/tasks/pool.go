package tasks

import (
	"fmt"
	"time"

	"github.com/ansible-semaphore/semaphore/util"
)

type taskPool struct {
	queue       []*task
	register    chan *task
	activeProj  map[int]*task
	activeNodes map[string]*task
	running int
}

var pool = taskPool{
	queue:       make([]*task, 0),
	register:    make(chan *task),
	activeProj:  make(map[int]*task),
	activeNodes: make(map[string]*task),
	running: 0,
}

type resourceLock struct {
	lock     bool
	holder   *task
}

var resourceLocker = make(chan *resourceLock)

func (p *taskPool) run() {

	defer func() {
		close(resourceLocker)
	}()

	ticker := time.NewTicker(5 * time.Second)

	defer func() {
		ticker.Stop()
	}()

	// Lock or unlock resources when running a task
	go func (locker <-chan *resourceLock) {
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
				p.running += 1
			} else {
				if p.activeProj[t.projectID] == t {
					delete(p.activeProj, t.projectID)
				}
				for _, node := range t.hosts {
					delete(p.activeNodes, node)
				}
				p.running -= 1
			}
		}
	}(resourceLocker)

	for {
		select {
		case task := <-p.register:
			fmt.Println(task)
			go task.prepareRun()
			p.queue = append(p.queue, task)
		case <-ticker.C:
			if len(p.queue) == 0 {
				continue
			} else if t := p.queue[0]; t.task.Status != "error" && (!t.prepared || p.blocks(t)) {
				p.queue = append(p.queue[1:], t)
				continue
			}

			if t := pool.queue[0]; t.task.Status != "error" {
				fmt.Println("Running a task.")
				resourceLocker <- &resourceLock{lock: true, holder: t,}
				go t.run()
			}
			pool.queue = pool.queue[1:]
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
		collision := false
		for _, node := range t.hosts {
			if p.activeNodes[node] != nil {
				collision = true
				break
			}
		}
		return collision
	default:
		return p.running > 0
	}
}

func StartRunner() {
	pool.run()
}
