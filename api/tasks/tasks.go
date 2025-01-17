package tasks

import (
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	task2 "github.com/semaphoreui/semaphore/services/tasks"
	"github.com/gorilla/context"
)

func TaskMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		taskID, err := helpers.GetIntParam("task_id", w, r)
		if err != nil {
			helpers.WriteErrorStatus(w, err.Error(), http.StatusBadRequest)
		}

		context.Set(r, "task_id", taskID)
		next.ServeHTTP(w, r)
	})
}

type taskLocation string

const (
	taskQueue   taskLocation = "queue"
	taskRunning taskLocation = "running"
)

type taskRes struct {
	TaskID      int                    `json:"task_id"`
	ProjectID   int                    `json:"project_id"`
	Username    string                 `json:"username,omitempty"`
	RunnerID    int                    `json:"runner_id,omitempty"`
	Status      task_logger.TaskStatus `json:"status"`
	Location    taskLocation           `json:"location"`
	RunnerName  string                 `json:"runner_name,omitempty"`
	ProjectName string                 `json:"project_name,omitempty"`
}

func GetTasks(w http.ResponseWriter, r *http.Request) {
	pool := context.Get(r, "task_pool").(*task2.TaskPool)

	res := []taskRes{}

	for _, task := range pool.Queue {
		res = append(res, taskRes{
			TaskID:    task.Task.ID,
			ProjectID: task.Task.ProjectID,
			RunnerID:  task.RunnerID,
			Username:  task.Username,
			Status:    task.Task.Status,
			Location:  taskQueue,
		})
	}

	for _, task := range pool.RunningTasks {
		res = append(res, taskRes{
			TaskID:    task.Task.ID,
			ProjectID: task.Task.ProjectID,
			RunnerID:  task.RunnerID,
			Username:  task.Username,
			Status:    task.Task.Status,
			Location:  taskRunning,
		})
	}

	helpers.WriteJSON(w, http.StatusOK, res)
}

func DeleteTask(w http.ResponseWriter, r *http.Request) {

	taskID := context.Get(r, "task_id").(int)

	pool := context.Get(r, "task_pool").(*task2.TaskPool)

	var task *db.Task

	for _, t := range pool.Queue {
		if t.Task.ID == taskID {
			task = &t.Task
			break
		}
	}

	if task == nil {
		for _, t := range pool.RunningTasks {
			if t.Task.ID == taskID {
				task = &t.Task
				break
			}
		}
	}

	if task != nil {
		err := pool.StopTask(*task, false)
		if err != nil {
			helpers.WriteErrorStatus(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	helpers.WriteJSON(w, http.StatusNoContent, nil)
}
