package tasks

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/castawaylabs/mulekick"
	"github.com/gorilla/context"
	"github.com/masterminds/squirrel"
)

func AddTask(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	user := context.Get(r, "user").(*db.User)

	var taskObj db.Task
	if err := mulekick.Bind(w, r, &taskObj); err != nil {
		return
	}

	taskObj.Created = time.Now()
	taskObj.Status = "waiting"
	taskObj.UserID = &user.ID

	if err := db.Mysql.Insert(&taskObj); err != nil {
		panic(err)
	}

	pool.register <- &task{
		task:      taskObj,
		projectID: project.ID,
	}

	objType := "task"
	desc := "Task ID " + strconv.Itoa(taskObj.ID) + " queued for running"
	if err := (db.Event{
		ProjectID:   &project.ID,
		ObjectType:  &objType,
		ObjectID:    &taskObj.ID,
		Description: &desc,
	}.Insert()); err != nil {
		panic(err)
	}

	mulekick.WriteJSON(w, http.StatusCreated, taskObj)
}

func GetAll(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)

	query, args, _ := squirrel.Select("task.*, tpl.playbook as tpl_playbook, user.name as user_name, tpl.alias as tpl_alias").
		From("task").
		Join("project__template as tpl on task.template_id=tpl.id").
		LeftJoin("user on task.user_id=user.id").
		Where("tpl.project_id=?", project.ID).
		OrderBy("task.created desc").
		ToSql()

	var tasks []struct {
		db.Task

		TemplatePlaybook string  `db:"tpl_playbook" json:"tpl_playbook"`
		TemplateAlias    string  `db:"tpl_alias" json:"tpl_alias"`
		UserName         *string `db:"user_name" json:"user_name"`
	}
	if _, err := db.Mysql.Select(&tasks, query, args...); err != nil {
		panic(err)
	}

	mulekick.WriteJSON(w, http.StatusOK, tasks)
}

func GetTask(w http.ResponseWriter, r *http.Request) {
	task := context.Get(r, "task").(db.Task)
	mulekick.WriteJSON(w, http.StatusOK, task)
}

func GetTaskMiddleware(w http.ResponseWriter, r *http.Request) {
	taskID, err := util.GetIntParam("task_id", w, r)
	if err != nil {
		panic(err)
	}

	var task db.Task
	if err := db.Mysql.SelectOne(&task, "select * from task where id=?", taskID); err != nil {
		panic(err)
	}

	context.Set(r, "task", task)
}

func GetTaskOutput(w http.ResponseWriter, r *http.Request) {
	task := context.Get(r, "task").(db.Task)

	var output []db.TaskOutput
	if _, err := db.Mysql.Select(&output, "select task_id, task, time, output from task__output where task_id=? order by time asc", task.ID); err != nil {
		panic(err)
	}

	mulekick.WriteJSON(w, http.StatusOK, output)
}

func RemoveTask(w http.ResponseWriter, r *http.Request) {
	task := context.Get(r, "task").(db.Task)

	statements := []string{
		"delete from task__output where task_id=?",
		"delete from task where id=?",
	}

	for _, statement := range statements {
		_, err := db.Mysql.Exec(statement, task.ID)
		if err != nil {
			panic(err)
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
