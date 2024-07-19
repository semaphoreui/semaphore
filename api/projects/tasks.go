package projects

import (
	"errors"
	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/services/tasks"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gorilla/context"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

// AddTask inserts a task into the database and returns a header or returns error
func AddTask(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	user := context.Get(r, "user").(*db.User)

	var taskObj db.Task

	if !helpers.Bind(w, r, &taskObj) {
		return
	}

	newTask, err := helpers.TaskPool(r).AddTask(taskObj, &user.ID, project.ID)

	if errors.Is(err, tasks.ErrInvalidSubscription) {
		helpers.WriteErrorStatus(w, "No active subscription available.", http.StatusForbidden)
		return
	} else if err != nil {

		util.LogErrorWithFields(err, log.Fields{"error": "Cannot write new event to database"})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	helpers.WriteJSON(w, http.StatusCreated, newTask)
}

// GetTasksList returns a list of tasks for the current project in desc order to limit or error
func GetTasksList(w http.ResponseWriter, r *http.Request, limit int) {
	project := context.Get(r, "project").(db.Project)
	tpl := context.Get(r, "template")

	var err error
	var tasks []db.TaskWithTpl

	if tpl != nil {
		tasks, err = helpers.Store(r).GetTemplateTasks(tpl.(db.Template).ProjectID, tpl.(db.Template).ID, db.RetrieveQueryParams{
			Count: limit,
		})
	} else {
		tasks, err = helpers.Store(r).GetProjectTasks(project.ID, db.RetrieveQueryParams{
			Count: limit,
		})
	}

	if err != nil {
		util.LogErrorWithFields(err, log.Fields{"error": "Bad request. Cannot get tasks list from database"})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, tasks)
}

// GetAllTasks returns all tasks for the current project
func GetAllTasks(w http.ResponseWriter, r *http.Request) {
	GetTasksList(w, r, 1000)
}

// GetLastTasks returns the hundred most recent tasks
func GetLastTasks(w http.ResponseWriter, r *http.Request) {
	str := r.URL.Query().Get("limit")
	limit, err := strconv.Atoi(str)
	if err != nil || limit <= 0 || limit > 200 {
		limit = 200
	}
	GetTasksList(w, r, limit)
}

// GetTask returns a task based on its id
func GetTask(w http.ResponseWriter, r *http.Request) {
	task := context.Get(r, "task").(db.Task)
	helpers.WriteJSON(w, http.StatusOK, task)
}

// GetTaskMiddleware is middleware that gets a task by id and sets the context to it or panics
func GetTaskMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := context.Get(r, "project").(db.Project)
		taskID, err := helpers.GetIntParam("task_id", w, r)

		if err != nil {
			util.LogErrorWithFields(err, log.Fields{"error": "Bad request. Cannot get task_id from request"})
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		task, err := helpers.Store(r).GetTask(project.ID, taskID)
		if err != nil {
			util.LogErrorWithFields(err, log.Fields{"error": "Bad request. Cannot get task from database"})
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		context.Set(r, "task", task)
		next.ServeHTTP(w, r)
	})
}

// GetTaskOutput returns the logged task output by id and writes it as json or returns error
func GetTaskStages(w http.ResponseWriter, r *http.Request) {
	task := context.Get(r, "task").(db.Task)
	project := context.Get(r, "project").(db.Project)

	var output []db.TaskOutput
	output, err := helpers.Store(r).GetTaskOutputs(project.ID, task.ID)

	if err != nil {
		util.LogErrorWithFields(err, log.Fields{"error": "Bad request. Cannot get task output from database"})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, output)
}

// GetTaskOutput returns the logged task output by id and writes it as json or returns error
func GetTaskOutput(w http.ResponseWriter, r *http.Request) {
	task := context.Get(r, "task").(db.Task)
	project := context.Get(r, "project").(db.Project)

	var output []db.TaskOutput
	output, err := helpers.Store(r).GetTaskOutputs(project.ID, task.ID)

	if err != nil {
		util.LogErrorWithFields(err, log.Fields{"error": "Bad request. Cannot get task output from database"})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, output)
}

func ConfirmTask(w http.ResponseWriter, r *http.Request) {
	targetTask := context.Get(r, "task").(db.Task)
	project := context.Get(r, "project").(db.Project)

	if targetTask.ProjectID != project.ID {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := helpers.TaskPool(r).ConfirmTask(targetTask)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func StopTask(w http.ResponseWriter, r *http.Request) {
	targetTask := context.Get(r, "task").(db.Task)
	project := context.Get(r, "project").(db.Project)

	if targetTask.ProjectID != project.ID {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var stopObj struct {
		Force bool `json:"force"`
	}

	if !helpers.Bind(w, r, &stopObj) {
		return
	}

	err := helpers.TaskPool(r).StopTask(targetTask, stopObj.Force)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveTask removes a task from the database
func RemoveTask(w http.ResponseWriter, r *http.Request) {
	targetTask := context.Get(r, "task").(db.Task)
	editor := context.Get(r, "user").(*db.User)
	project := context.Get(r, "project").(db.Project)

	activeTask := helpers.TaskPool(r).GetTask(targetTask.ID)

	if activeTask != nil {
		// can't delete task in queue or running
		// task must be stopped firstly
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !editor.Admin {
		log.Warn(editor.Username + " is not permitted to delete task logs")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err := helpers.Store(r).DeleteTaskWithOutputs(project.ID, targetTask.ID)
	if err != nil {
		util.LogErrorWithFields(err, log.Fields{"error": "Bad request. Cannot delete task from database"})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
