package tasks

import (
	"strconv"
	"time"

	"github.com/ansible-semaphore/semaphore/util"

	"github.com/ansible-semaphore/semaphore/database"
	"github.com/ansible-semaphore/semaphore/models"
	"github.com/gin-gonic/gin"
	"github.com/masterminds/squirrel"
)

func AddTask(c *gin.Context) {
	project := c.MustGet("project").(models.Project)

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
		task:      taskObj,
		projectID: project.ID,
	}

	objType := "task"
	desc := "Task ID " + strconv.Itoa(taskObj.ID) + " queued for running"
	if err := (models.Event{
		ProjectID:   &project.ID,
		ObjectType:  &objType,
		ObjectID:    &taskObj.ID,
		Description: &desc,
	}.Insert()); err != nil {
		panic(err)
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

func GetTaskMiddleware(c *gin.Context) {
	taskID, err := util.GetIntParam("task_id", c)
	if err != nil {
		panic(err)
	}

	var task models.Task
	if err := database.Mysql.SelectOne(&task, "select * from task where id=?", taskID); err != nil {
		panic(err)
	}

	c.Set("task", task)
	c.Next()
}

func GetTaskOutput(c *gin.Context) {
	task := c.MustGet("task").(models.Task)

	var output []models.TaskOutput
	if _, err := database.Mysql.Select(&output, "select * from task__output where task_id=? order by time desc", task.ID); err != nil {
		panic(err)
	}

	c.JSON(200, output)
}
