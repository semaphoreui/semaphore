package tasks

import (
	"fmt"
	"time"

	"github.com/ansible-semaphore/semaphore/database"
	"github.com/ansible-semaphore/semaphore/models"
	"github.com/gin-gonic/gin"
	"github.com/masterminds/squirrel"
)

type taskPool struct {
	queue    []*task
	register chan *task
	running  *task
}

var pool = taskPool{
	queue:    make([]*task, 0),
	register: make(chan *task),
	running:  nil,
}

func (p *taskPool) run() {
	ticker := time.NewTicker(10 * time.Second)

	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case task := <-p.register:
			fmt.Println(task)
			if p.running == nil {
				go task.run()
				continue
			}

			p.queue = append(p.queue, task)
		case <-ticker.C:
			if len(p.queue) == 0 || p.running != nil {
				continue
			}

			fmt.Println("Running a task.")
			go pool.queue[0].run()
			pool.queue = pool.queue[1:]
		}
	}
}

func StartRunner() {
	pool.run()
}

func AddTask(c *gin.Context) {
	var taskObj models.Task
	if err := c.Bind(&taskObj); err != nil {
		return
	}

	taskObj.Created = time.Now()
	taskObj.Status = "waiting"

	if err := database.Mysql.Insert(&taskObj); err != nil {
		panic(err)
	}

	pool.register <- &task{
		task: taskObj,
	}

	c.JSON(201, taskObj)
}

func GetAll(c *gin.Context) {
	project := c.MustGet("project").(models.Project)

	query, args, _ := squirrel.Select("task.*, tpl.playbook as tpl_playbook").
		From("task").
		Join("project__template as tpl on task.template_id=tpl.id").
		Where("tpl.project_id=?", project.ID).
		OrderBy("task.created desc").
		ToSql()

	var tasks []struct {
		models.Task

		TemplatePlaybook string `db:"tpl_playbook" json:"tpl_playbook"`
	}
	if _, err := database.Mysql.Select(&tasks, query, args...); err != nil {
		panic(err)
	}

	c.JSON(200, tasks)
}
