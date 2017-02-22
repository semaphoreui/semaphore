package tasks

import (
	"strconv"
	"time"

	database "github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/models"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gin-gonic/gin"
	"github.com/masterminds/squirrel"
)

func AddTask(w http.ResponseWriter, r *http.Request) {
	project := c.MustGet("project").(models.Project)
	user := c.MustGet("user").(*models.User)

	var taskObj models.Task
	if err := c.Bind(&taskObj); err != nil {
		return
	}

	taskObj.Created = time.Now()
	taskObj.Status = "waiting"
	taskObj.UserID = &user.ID

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

func GetAll(w http.ResponseWriter, r *http.Request) {
	project := c.MustGet("project").(models.Project)

	query, args, _ := squirrel.Select("task.*, tpl.playbook as tpl_playbook, user.name as user_name, tpl.alias as tpl_alias").
		From("task").
		Join("project__template as tpl on task.template_id=tpl.id").
		LeftJoin("user on task.user_id=user.id").
		Where("tpl.project_id=?", project.ID).
		OrderBy("task.created desc").
		ToSql()

	var tasks []struct {
		models.Task

		TemplatePlaybook string  `db:"tpl_playbook" json:"tpl_playbook"`
		TemplateAlias    string  `db:"tpl_alias" json:"tpl_alias"`
		UserName         *string `db:"user_name" json:"user_name"`
	}
	if _, err := database.Mysql.Select(&tasks, query, args...); err != nil {
		panic(err)
	}

	c.JSON(200, tasks)
}

func GetTaskMiddleware(w http.ResponseWriter, r *http.Request) {
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

func GetTaskOutput(w http.ResponseWriter, r *http.Request) {
	task := c.MustGet("task").(models.Task)

	var output []models.TaskOutput
	if _, err := database.Mysql.Select(&output, "select * from task__output where task_id=? order by time asc", task.ID); err != nil {
		panic(err)
	}

	c.JSON(200, output)
}

func RemoveTask(w http.ResponseWriter, r *http.Request) {
	task := c.MustGet("task").(models.Task)

	statements := []string{
		"delete from task__output where task_id=?",
		"delete from task where id=?",
	}

	for _, statement := range statements {
		_, err := database.Mysql.Exec(statement, task.ID)
		if err != nil {
			panic(err)
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
