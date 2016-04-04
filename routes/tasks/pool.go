package tasks

import (
	"fmt"
	"strconv"

	"github.com/ansible-semaphore/semaphore/models"
	"github.com/ansible-semaphore/semaphore/routes/sockets"
	"github.com/gin-gonic/gin"
)

type taskPool struct {
	queue    []*models.Task
	register chan *models.Task
}

var pool = taskPool{
	queue:    make([]*models.Task, 0),
	register: make(chan *models.Task),
}

func (p *taskPool) run() {
	for {
		select {
		case task := <-p.register:
			fmt.Println(task, len(p.queue))
			if len(p.queue) == 0 {
				go runTask(task)
				continue
			}

			p.queue = append(p.queue, task)
		}
	}
}

func runTask(task *models.Task) {
	sockets.Broadcast([]byte("Running:" + strconv.Itoa(task.ID)))
}

func StartRunner() {
	pool.run()
}

func AddTask(c *gin.Context) {
	var task models.Task
	if err := c.Bind(&task); err != nil {
		return
	}

	pool.register <- &task
}
