package projects

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Digital-Data-Co/forge/api/helpers"
	"github.com/Digital-Data-Co/forge/db"
	"github.com/Digital-Data-Co/forge/pkg/common_errors"
	"github.com/Digital-Data-Co/forge/services/files"
	"github.com/Digital-Data-Co/forge/services/tasks"
	"github.com/Digital-Data-Co/forge/util"
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

	var taskObj db.Task

	if !helpers.Bind(w, r, &taskObj) {
		return
	}

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

//type ansibleTaskResult struct {
//	App        string              `json:"app"`
//	TemplateID int                 `json:"template_id"`
//	Hosts      db.AnsibleTaskHost  `json:"hosts"`
//	Errors     db.AnsibleTaskError `json:"errors"`
//}

//func GetAnsibleTaskResult() (res ansibleTaskResult, err error) {
//	return
//}

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

// GetTaskFiles returns all files associated with a task
func GetTaskFiles(w http.ResponseWriter, r *http.Request) {
	task := helpers.GetFromContext(r, "task").(db.Task)
	project := helpers.GetFromContext(r, "project").(db.Project)

	taskFiles, err := helpers.Store(r).GetTaskFiles(project.ID, task.ID)
	if err != nil {
		util.LogErrorF(err, log.Fields{"error": "Cannot get task files from database"})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, taskFiles)
}

// GetTaskFile returns a specific task file for download
func GetTaskFile(w http.ResponseWriter, r *http.Request) {
	task := helpers.GetFromContext(r, "task").(db.Task)
	project := helpers.GetFromContext(r, "project").(db.Project)

	fileID, err := helpers.GetIntParam("fileId", w, r)
	if err != nil {
		return
	}

	taskFile, err := helpers.Store(r).GetTaskFile(project.ID, task.ID, fileID)
	if err != nil {
		util.LogErrorF(err, log.Fields{"error": "Cannot get task file from database"})
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Get file from storage
	fileStorage := files.NewTaskFileStorage()
	fileData, err := fileStorage.ReadFile(&taskFile)
	if err != nil {
		util.LogErrorF(err, log.Fields{"error": "Cannot read task file from storage"})
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Set appropriate headers for file download
	w.Header().Set("Content-Type", taskFile.MimeType)
	w.Header().Set("Content-Disposition", "attachment; filename="+taskFile.Filename)
	w.Header().Set("Content-Length", strconv.FormatInt(taskFile.FileSize, 10))
	w.WriteHeader(http.StatusOK)
	w.Write(fileData)
}

// DeleteTaskFile deletes a specific task file
func DeleteTaskFile(w http.ResponseWriter, r *http.Request) {
	task := helpers.GetFromContext(r, "task").(db.Task)
	project := helpers.GetFromContext(r, "project").(db.Project)

	fileID, err := helpers.GetIntParam("fileId", w, r)
	if err != nil {
		return
	}

	// Get task file first to delete from storage
	taskFile, err := helpers.Store(r).GetTaskFile(project.ID, task.ID, fileID)
	if err != nil {
		util.LogErrorF(err, log.Fields{"error": "Cannot get task file from database"})
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Delete from storage
	fileStorage := files.NewTaskFileStorage()
	if err := fileStorage.DeleteFile(&taskFile); err != nil {
		util.LogErrorF(err, log.Fields{"error": "Cannot delete task file from storage"})
		// Continue with database deletion even if storage deletion fails
	}

	// Delete from database
	err = helpers.Store(r).DeleteTaskFile(project.ID, task.ID, fileID)
	if err != nil {
		util.LogErrorF(err, log.Fields{"error": "Cannot delete task file from database"})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
