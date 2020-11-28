package tasks

import (
	"net/http"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/db"

	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gorilla/context"
	"github.com/masterminds/squirrel"
)

// AddTask inserts a task into the database and returns a header or returns error
func AddTask(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	user := context.Get(r, "user").(*db.User)

	var taskObj db.Task
	if err := util.Bind(w, r, &taskObj); err != nil {
		return
	}

	taskObj.Created = time.Now()
	taskObj.Status = "waiting"
	taskObj.UserID = &user.ID

	if err := db.Sql.Insert(&taskObj); err != nil {
		util.LogErrorWithFields(err, log.Fields{"error": "Bad request. Cannot create new task"})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pool.register <- &task{
		task:      taskObj,
		projectID: project.ID,
	}

	objType := taskTypeID
	desc := "Task ID " + strconv.Itoa(taskObj.ID) + " queued for running"
	if err := (db.Event{
		ProjectID:   &project.ID,
		ObjectType:  &objType,
		ObjectID:    &taskObj.ID,
		Description: &desc,
	}.Insert()); err != nil {
		util.LogErrorWithFields(err, log.Fields{"error": "Cannot write new event to database"})
	}

	util.WriteJSON(w, http.StatusCreated, taskObj)
}

// GetTasksList returns a list of tasks for the current project in desc order to limit or error
func GetTasksList(w http.ResponseWriter, r *http.Request, limit uint64) {
	project := context.Get(r, "project").(db.Project)

	q := squirrel.Select("task.*, tpl.playbook as tpl_playbook, user.name as user_name, tpl.alias as tpl_alias").
		From("task").
		Join("project__template as tpl on task.template_id=tpl.id").
		LeftJoin("user on task.user_id=user.id");

	if tpl := context.Get(r, "template"); tpl != nil {
		q = q.Where("tpl.project_id=? AND task.template_id=?", project.ID, tpl.(db.Template).ID)
	} else {
		q = q.Where("tpl.project_id=?", project.ID)
	}

	q = q.OrderBy("task.created desc, id desc")

	if limit > 0 {
		q = q.Limit(limit)
	}

	query, args, _ := q.ToSql()

	var tasks []struct {
		db.Task

		TemplatePlaybook string  `db:"tpl_playbook" json:"tpl_playbook"`
		TemplateAlias    string  `db:"tpl_alias" json:"tpl_alias"`
		UserName         *string `db:"user_name" json:"user_name"`
	}
	if _, err := db.Sql.Select(&tasks, query, args...); err != nil {
		util.LogErrorWithFields(err, log.Fields{"error": "Bad request. Cannot get tasks list from database"})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	util.WriteJSON(w, http.StatusOK, tasks)
}

// GetAllTasks returns all tasks for the current project
func GetAllTasks(w http.ResponseWriter, r *http.Request) {
	GetTasksList(w, r, 0)
}

// GetLastTasks returns the hundred most recent tasks
func GetLastTasks(w http.ResponseWriter, r *http.Request) {
	GetTasksList(w, r, 200)
}

// GetTask returns a task based on its id
func GetTask(w http.ResponseWriter, r *http.Request) {
	task := context.Get(r, taskTypeID).(db.Task)
	util.WriteJSON(w, http.StatusOK, task)
}

// GetTaskMiddleware is middleware that gets a task by id and sets the context to it or panics
func GetTaskMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		taskID, err := util.GetIntParam("task_id", w, r)
		if err != nil {
			panic(err)
		}

		var task db.Task
		if err := db.Sql.SelectOne(&task, "select * from task where id=?", taskID); err != nil {
			panic(err)
		}

		context.Set(r, taskTypeID, task)
		next.ServeHTTP(w, r)
	})
}

// GetTaskOutput returns the logged task output by id and writes it as json or returns error
func GetTaskOutput(w http.ResponseWriter, r *http.Request) {
	task := context.Get(r, taskTypeID).(db.Task)

	var output []db.TaskOutput
	if _, err := db.Sql.Select(&output, "select task_id, task, time, output from task__output where task_id=? order by time asc", task.ID); err != nil {
		util.LogErrorWithFields(err, log.Fields{"error": "Bad request. Cannot get task output from database"})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	util.WriteJSON(w, http.StatusOK, output)
}

// RemoveTask removes a task from the database
func RemoveTask(w http.ResponseWriter, r *http.Request) {
	task := context.Get(r, taskTypeID).(db.Task)
	editor := context.Get(r, "user").(*db.User)

	if !editor.Admin {
		log.Warn(editor.Username + " is not permitted to delete task logs")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	statements := []string{
		"delete from task__output where task_id=?",
		"delete from task where id=?",
	}

	for _, statement := range statements {
		_, err := db.Sql.Exec(statement, task.ID)
		if err != nil {
			util.LogErrorWithFields(err, log.Fields{"error": "Bad request. Cannot delete task from database"})
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
