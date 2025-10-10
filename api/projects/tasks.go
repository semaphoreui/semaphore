package projects

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/semaphoreui/semaphore/services/tasks"
	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
)

type TaskController struct {
	ansibleTaskRepo db.AnsibleTaskRepository
}

func NewTaskController(ansibleTaskRepo db.AnsibleTaskRepository) *TaskController {
	return &TaskController{
		ansibleTaskRepo: ansibleTaskRepo,
	}
}

func taskPool(r *http.Request) *tasks.TaskPool {
	return helpers.GetFromContext(r, "task_pool").(*tasks.TaskPool)
}

// AddTask inserts a task into the database and returns a header or returns error
func AddTask(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	user := helpers.GetFromContext(r, "user").(*db.User)
	taskObj := helpers.GetFromContext(r, "task").(db.Task)

	tpl, err := helpers.Store(r).GetTemplate(project.ID, taskObj.TemplateID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	newTask, err := taskPool(r).AddTask(
		taskObj,
		&user.ID,
		user.Username,
		project.ID,
		tpl.App.NeedTaskAlias(),
	)

	if errors.Is(err, common_errors.ErrInvalidSubscription) {
		helpers.WriteErrorStatus(w, "No active subscription available.", http.StatusForbidden)
		return
	} else if err != nil {
		util.LogErrorF(err, log.Fields{"error": "Cannot write new event to database"})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	helpers.WriteJSON(w, http.StatusCreated, newTask)
}

// GetTasksList returns a list of tasks for the current project in desc order to limit or error
func GetTasksList(w http.ResponseWriter, r *http.Request, limit int) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	tpl := helpers.GetFromContext(r, "template")

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
		util.LogErrorF(err, log.Fields{"error": "Bad request. Cannot get tasks list from database"})
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
	task := helpers.GetFromContext(r, "task").(db.Task)
	helpers.WriteJSON(w, http.StatusOK, task)
}

func GetTaskPermissionsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := helpers.GetFromContext(r, "project").(db.Project)
		user := helpers.GetFromContext(r, "user").(*db.User)
		task := helpers.GetFromContext(r, "task").(db.Task)

		permissions := helpers.GetFromContext(r, "permissions").(db.ProjectUserPermission)

		perm, err := helpers.Store(r).GetTemplatePermission(project.ID, task.TemplateID, user.ID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}

		permissions |= perm

		r = helpers.SetContextValue(r, "permissions", permissions)
		next.ServeHTTP(w, r)
	})
}

// GetTaskMiddleware is middleware that gets a task by id and sets the context to it or panics
func GetTaskMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := helpers.GetFromContext(r, "project").(db.Project)
		taskID, err := helpers.GetIntParam("task_id", w, r)

		if err != nil {
			util.LogErrorF(err, log.Fields{"error": "Bad request. Cannot get task_id from request"})
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		task, err := helpers.Store(r).GetTask(project.ID, taskID)
		if err != nil {
			util.LogErrorF(err, log.Fields{"error": "Bad request. Cannot get task from database"})
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		r = helpers.SetContextValue(r, "task", task)
		next.ServeHTTP(w, r)
	})
}

// GetTaskMiddleware is middleware that gets a task by id and sets the context to it or panics
func NewTaskMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var taskObj db.Task

		if !helpers.Bind(w, r, &taskObj) {
			return
		}

		r = helpers.SetContextValue(r, "task", taskObj)
		next.ServeHTTP(w, r)
	})
}

func (c *TaskController) GetAnsibleTaskHosts(w http.ResponseWriter, r *http.Request) {
	task := helpers.GetFromContext(r, "task").(db.Task)
	project := helpers.GetFromContext(r, "project").(db.Project)
	hosts, err := c.ansibleTaskRepo.GetAnsibleTaskHosts(project.ID, task.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, hosts)
}

func (c *TaskController) GetAnsibleTaskErrors(w http.ResponseWriter, r *http.Request) {
	task := helpers.GetFromContext(r, "task").(db.Task)
	project := helpers.GetFromContext(r, "project").(db.Project)
	hosts, err := c.ansibleTaskRepo.GetAnsibleTaskErrors(project.ID, task.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, hosts)
}

// GetTaskStages returns the logged task stages by id and writes it as json or returns error
func GetTaskStages(w http.ResponseWriter, r *http.Request) {
	task := helpers.GetFromContext(r, "task").(db.Task)
	project := helpers.GetFromContext(r, "project").(db.Project)

	stages, err := helpers.Store(r).GetTaskStages(project.ID, task.ID)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	for i := range stages {
		if stages[i].JSON == "" {
			continue
		}
		var res any
		err = json.Unmarshal([]byte(stages[i].JSON), &res)
		if err != nil {
			helpers.WriteError(w, err)
			return
		}
		stages[i].Result = res
	}

	helpers.WriteJSON(w, http.StatusOK, stages)
}

// GetTaskOutput returns the logged task output by id and writes it as json or returns error
func GetTaskOutput(w http.ResponseWriter, r *http.Request) {
	task := helpers.GetFromContext(r, "task").(db.Task)
	project := helpers.GetFromContext(r, "project").(db.Project)

	var output []db.TaskOutput
	output, err := helpers.Store(r).GetTaskOutputs(project.ID, task.ID, db.RetrieveQueryParams{})

	if err != nil {
		util.LogErrorF(err, log.Fields{"error": "Bad request. Cannot get task output from database"})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, output)
}

func outputToBytes(lines []db.TaskOutput) []byte {
	var buffer bytes.Buffer
	for _, line := range lines {
		output := util.ClearFromAnsiCodes(line.Output)
		buffer.WriteString(output)
		buffer.WriteByte('\n')
	}
	return buffer.Bytes()
}

func GetTaskRawOutput(w http.ResponseWriter, r *http.Request) {
	task := helpers.GetFromContext(r, "task").(db.Task)
	project := helpers.GetFromContext(r, "project").(db.Project)

	const chunkSize = 10000
	offset := 0

	eof := false
	for !eof {
		var output []db.TaskOutput
		output, err := helpers.Store(r).GetTaskOutputs(project.ID, task.ID, db.RetrieveQueryParams{Offset: offset, Count: chunkSize})

		if err != nil {
			if offset == 0 {
				util.LogErrorF(err, log.Fields{"error": "Bad request. Cannot get task output from database"})
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			util.LogErrorF(err, log.Fields{"error": "Cannot get task output from database"})
			return
		}

		if offset == 0 {
			w.Header().Set("content-type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
		}

		readSize := len(output)

		if readSize > 0 {
			offset += readSize
			data := outputToBytes(output)
			if _, err := w.Write(data); err != nil {
				return
			}
		}

		eof = readSize < chunkSize
	}
}

func ConfirmTask(w http.ResponseWriter, r *http.Request) {
	targetTask := helpers.GetFromContext(r, "task").(db.Task)
	project := helpers.GetFromContext(r, "project").(db.Project)

	if targetTask.ProjectID != project.ID {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := taskPool(r).ConfirmTask(targetTask)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func RejectTask(w http.ResponseWriter, r *http.Request) {
	targetTask := helpers.GetFromContext(r, "task").(db.Task)
	project := helpers.GetFromContext(r, "project").(db.Project)

	if targetTask.ProjectID != project.ID {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := taskPool(r).RejectTask(targetTask)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func StopTask(w http.ResponseWriter, r *http.Request) {
	targetTask := helpers.GetFromContext(r, "task").(db.Task)
	project := helpers.GetFromContext(r, "project").(db.Project)

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

	err := taskPool(r).StopTask(targetTask, stopObj.Force)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveTask removes a task from the database
func RemoveTask(w http.ResponseWriter, r *http.Request) {
	targetTask := helpers.GetFromContext(r, "task").(db.Task)
	editor := helpers.GetFromContext(r, "user").(*db.User)
	project := helpers.GetFromContext(r, "project").(db.Project)

	activeTask := taskPool(r).GetTask(targetTask.ID)

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
		util.LogErrorF(err, log.Fields{"error": "Bad request. Cannot delete task from database"})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func GetTaskStats(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)

	var tplID *int
	if tpl := helpers.GetFromContext(r, "template"); tpl != nil {
		id := tpl.(db.Template).ID
		tplID = &id
	}

	filter := db.TaskFilter{}

	if start := r.URL.Query().Get("start"); start != "" {
		d, err := time.Parse("2006-01-02", start)
		if err != nil {
			helpers.WriteErrorStatus(w, "Invalid start date", http.StatusBadRequest)
			return
		}
		filter.Start = &d
	}

	if end := r.URL.Query().Get("end"); end != "" {
		d, err := time.Parse("2006-01-02", end)
		if err != nil {
			helpers.WriteErrorStatus(w, "Invalid end date", http.StatusBadRequest)
			return
		}
		filter.End = &d
	}

	if userId := r.URL.Query().Get("user_id"); userId != "" {
		u, err := strconv.Atoi(userId)
		if err != nil {
			helpers.WriteErrorStatus(w, "Invalid user_id", http.StatusBadRequest)
			return
		}
		filter.UserID = &u
	}

	stats, err := helpers.Store(r).GetTaskStats(project.ID, tplID, db.TaskStatUnitDay, filter)
	if err != nil {
		util.LogErrorF(err, log.Fields{"error": "Bad request. Cannot get task stats from database"})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, stats)
}

func (c *TaskController) StopAllTasks(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	tpl := helpers.GetFromContext(r, "template").(db.Template)

	var stopObj struct {
		Force bool `json:"force"`
	}

	// optional body; ignore bind error and default Force=false
	if ok := helpers.Bind(w, r, &stopObj); !ok {
		helpers.WriteErrorStatus(w, "Not allowed", http.StatusBadRequest)
		return
	}

	taskPool(r).StopTasksByTemplate(project.ID, tpl.ID, stopObj.Force)
	w.WriteHeader(http.StatusNoContent)
}
