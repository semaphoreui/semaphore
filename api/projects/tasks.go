package projects

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/semaphoreui/semaphore/services/tasks"
	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
)

// maxTasksPageSize limits how many tasks can be requested in a single page.
const maxTasksPageSize = 200

type TaskController struct {
	store           db.Store
	ansibleTaskRepo db.AnsibleTaskRepository
}

func NewTaskController(store db.Store, ansibleTaskRepo db.AnsibleTaskRepository) *TaskController {
	return &TaskController{
		store:           store,
		ansibleTaskRepo: ansibleTaskRepo,
	}
}

func taskPool(r *http.Request) *tasks.TaskPool {
	return helpers.GetFromContext(r, "task_pool").(*tasks.TaskPool)
}

// AddTask inserts a task into the database and returns a header or returns error
// resolveTaskTemplate returns the template of the task, which may be referenced
// either by id or by name. The resolved id is written back to the task so the
// rest of the pipeline only deals with ids.
func (c *TaskController) resolveTaskTemplate(projectID int, task *db.Task) (tpl db.Template, err error) {
	// The name is resolved to an id here, so it is cleared to keep it out of the
	// stored task and the response.
	name := task.TemplateName
	task.TemplateName = ""

	if task.TemplateID == 0 && name == "" {
		err = common_errors.NewValidationError("template_id or template_name is required")
		return
	}

	if task.TemplateID != 0 {
		tpl, err = c.store.GetTemplate(projectID, task.TemplateID)
		return
	}

	tpl, err = c.store.GetTemplateByName(projectID, name)
	if err != nil {
		return
	}

	task.TemplateID = tpl.ID
	return
}

func (c *TaskController) AddTask(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	user := helpers.GetFromContext(r, "user").(*db.User)
	taskObj := helpers.GetFromContext(r, "task").(db.Task)

	tpl, err := c.resolveTaskTemplate(project.ID, &taskObj)
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
	}

	if err != nil {
		log.WithFields(log.Fields{
			"context":     "AddTask",
			"project_id":  project.ID,
			"template_id": taskObj.TemplateID,
			"user_id":     user.ID,
		}).WithError(err).Error("Cannot add task")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	helpers.WriteJSON(w, http.StatusCreated, newTask)
}

// writeTasksList retrieves the tasks list (project-wide or per-template
// depending on the request context) applying the given query params and writes
// the result as JSON.
//
// When pageSize > 0 it implements keyset pagination: params.Count is expected to
// already request one extra row (pageSize + 1) so that the presence of a next
// page can be detected without an expensive COUNT(*). The extra row is trimmed
// off and whether it existed is reported via the X-Has-Next header. No total
// count is computed — that would not scale to projects with millions of tasks.
func (c *TaskController) writeTasksList(w http.ResponseWriter, r *http.Request, params db.RetrieveQueryParams, pageSize int) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	tpl := helpers.GetFromContext(r, "template")

	var err error
	var taskList []db.TaskWithTpl

	if tpl != nil {
		template := tpl.(db.Template)
		taskList, err = c.store.GetTemplateTasks(template.ProjectID, template.ID, params)
	} else {
		taskList, err = c.store.GetProjectTasks(project.ID, params)
	}

	if err != nil {
		util.LogErrorF(err, log.Fields{"error": "Bad request. Cannot get tasks list from database"})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if pageSize > 0 {
		hasNext := len(taskList) > pageSize
		if hasNext {
			taskList = taskList[:pageSize]
		}
		w.Header().Set("X-Has-Next", strconv.FormatBool(hasNext))
	}

	helpers.WriteJSON(w, http.StatusOK, taskList)
}

// GetAllTasks returns all tasks for the current project
func (c *TaskController) GetAllTasks(w http.ResponseWriter, r *http.Request) {
	params := helpers.QueryParams(r.URL)
	params.Count = 1000
	c.writeTasksList(w, r, params, 0)
}

// parseTasksPageParams fills keyset pagination fields into the given base params
// from the request query and returns the effective page size. It supports the
// `count` parameter (with backward compatibility for the legacy `limit`) and the
// `before` cursor (a task id; only older tasks are returned). The page size is
// capped at maxTasksPageSize. params.Count is set to pageSize + 1 so the caller
// can detect whether a next page exists.
func parseTasksPageParams(query url.Values, base db.RetrieveQueryParams) (db.RetrieveQueryParams, int) {
	pageSize := maxTasksPageSize

	if str := query.Get("count"); str != "" {
		if v, err := strconv.Atoi(str); err == nil && v > 0 {
			pageSize = v
		}
	} else if str := query.Get("limit"); str != "" {
		if v, err := strconv.Atoi(str); err == nil && v > 0 {
			pageSize = v
		}
	}

	if pageSize > maxTasksPageSize {
		pageSize = maxTasksPageSize
	}

	// Fetch one extra row to detect the presence of a next page.
	base.Count = pageSize + 1

	if str := query.Get("before"); str != "" {
		if v, err := strconv.Atoi(str); err == nil && v > 0 {
			base.BeforeID = v
		}
	}

	return base, pageSize
}

// GetLastTasks returns a page of the most recent tasks using keyset pagination.
// The page size is controlled by the `count` query parameter (legacy `limit` is
// still accepted) and the `before` cursor selects the next, older page. The
// X-Has-Next response header reports whether more (older) tasks are available.
func (c *TaskController) GetLastTasks(w http.ResponseWriter, r *http.Request) {
	params, pageSize := parseTasksPageParams(r.URL.Query(), helpers.QueryParams(r.URL))
	c.writeTasksList(w, r, params, pageSize)
}

// GetTask returns a task based on its id
func (c *TaskController) GetTask(w http.ResponseWriter, r *http.Request) {
	task := helpers.GetFromContext(r, "task").(db.Task)
	helpers.WriteJSON(w, http.StatusOK, task)
}

func (c *TaskController) GetTaskPermissionsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := helpers.GetFromContext(r, "project").(db.Project)
		user := helpers.GetFromContext(r, "user").(*db.User)
		task := helpers.GetFromContext(r, "task").(db.Task)

		permissions := helpers.GetFromContext(r, "permissions").(db.ProjectUserPermission)

		perm, err := c.store.GetTemplatePermission(project.ID, task.TemplateID, user.ID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}

		permissions |= perm

		r = helpers.SetContextValue(r, "permissions", permissions)
		next.ServeHTTP(w, r)
	})
}

// GetTaskMiddleware is middleware that gets a task by id and sets the context to it or panics
func (c *TaskController) GetTaskMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := helpers.GetFromContext(r, "project").(db.Project)
		taskID, err := helpers.GetIntParam("task_id", w, r)

		if err != nil {
			util.LogErrorF(err, log.Fields{"error": "Bad request. Cannot get task_id from request"})
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		task, err := c.store.GetTask(project.ID, taskID)
		if err != nil {
			util.LogErrorF(err, log.Fields{"error": "Bad request. Cannot get task from database"})
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		r = helpers.SetContextValue(r, "task", task)
		next.ServeHTTP(w, r)
	})
}

// NewTaskMiddleware is middleware that binds a task from the request body and sets the context to it
func (c *TaskController) NewTaskMiddleware(next http.Handler) http.Handler {
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
func (c *TaskController) GetTaskStages(w http.ResponseWriter, r *http.Request) {
	task := helpers.GetFromContext(r, "task").(db.Task)
	project := helpers.GetFromContext(r, "project").(db.Project)

	stages, err := c.store.GetTaskStages(project.ID, task.ID)

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
func (c *TaskController) GetTaskOutput(w http.ResponseWriter, r *http.Request) {
	task := helpers.GetFromContext(r, "task").(db.Task)
	project := helpers.GetFromContext(r, "project").(db.Project)

	var output []db.TaskOutput
	output, err := c.store.GetTaskOutputs(project.ID, task.ID, db.RetrieveQueryParams{})

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

func (c *TaskController) GetTaskRawOutput(w http.ResponseWriter, r *http.Request) {
	task := helpers.GetFromContext(r, "task").(db.Task)
	project := helpers.GetFromContext(r, "project").(db.Project)

	const chunkSize = 10000
	offset := 0

	eof := false
	for !eof {
		var output []db.TaskOutput
		output, err := c.store.GetTaskOutputs(project.ID, task.ID, db.RetrieveQueryParams{Offset: offset, Count: chunkSize})

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

func (c *TaskController) ConfirmTask(w http.ResponseWriter, r *http.Request) {
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

func (c *TaskController) RejectTask(w http.ResponseWriter, r *http.Request) {
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

func (c *TaskController) StopTask(w http.ResponseWriter, r *http.Request) {
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
func (c *TaskController) RemoveTask(w http.ResponseWriter, r *http.Request) {
	targetTask := helpers.GetFromContext(r, "task").(db.Task)
	editor := helpers.GetFromContext(r, "user").(*db.User)
	project := helpers.GetFromContext(r, "project").(db.Project)

	activeTask, err := taskPool(r).GetTask(targetTask.ID)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

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

	err = c.store.DeleteTaskWithOutputs(project.ID, targetTask.ID)
	if err != nil {
		util.LogErrorF(err, log.Fields{"error": "Bad request. Cannot delete task from database"})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *TaskController) GetTaskStats(w http.ResponseWriter, r *http.Request) {
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

	stats, err := c.store.GetTaskStats(project.ID, tplID, db.TaskStatUnitDay, filter)
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
