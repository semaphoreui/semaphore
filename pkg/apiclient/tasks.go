package apiclient

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

// MaxTasksPage is the largest page the server returns for a task list.
const MaxTasksPage = 200

// TaskPage is one page of the task list. Tasks are ordered newest first;
// HasNext says whether older tasks exist beyond the last one.
type TaskPage struct {
	Tasks   []db.TaskWithTpl
	HasNext bool
}

// Before returns the cursor for the next page: the id of the oldest task on
// this page, or 0 when the page is empty.
func (p TaskPage) Before() int {
	if len(p.Tasks) == 0 {
		return 0
	}
	return p.Tasks[len(p.Tasks)-1].ID
}

// RunTask creates a task. task.TemplateID or task.TemplateName selects the
// template; everything else overrides the template where it allows it.
func (c *Client) RunTask(ctx context.Context, projectID int, task db.Task) (db.Task, error) {
	task.ProjectID = projectID
	var created db.Task
	err := c.Do(ctx, http.MethodPost, projectPath(projectID, "/tasks"), task, &created)
	return created, err
}

// GetTask returns one task.
func (c *Client) GetTask(ctx context.Context, projectID, taskID int) (db.TaskWithTpl, error) {
	var task db.TaskWithTpl
	err := c.Do(ctx, http.MethodGet, projectPath(projectID, "/tasks/%d", taskID), nil, &task)
	return task, err
}

// ListTasks returns a page of the project's tasks (templateID 0) or of one
// template's tasks. count is capped by the server at MaxTasksPage; before is
// a task id and 0 means "from the newest".
func (c *Client) ListTasks(ctx context.Context, projectID, templateID, count, before int) (TaskPage, error) {
	q := url.Values{}
	if count > 0 {
		q.Set("count", strconv.Itoa(count))
	}
	if before > 0 {
		q.Set("before", strconv.Itoa(before))
	}

	apiPath := projectPath(projectID, "/tasks/last")
	if templateID > 0 {
		apiPath = projectPath(projectID, "/templates/%d/tasks/last", templateID)
	}

	var page TaskPage
	header, err := c.Call(ctx, http.MethodGet, withQuery(apiPath, q), nil, &page.Tasks)
	if err != nil {
		return page, err
	}
	page.HasNext = header.Get("X-Has-Next") == "true"
	return page, nil
}

// ListAllTasks walks the pages until limit tasks are collected or the history
// ends. limit 0 means everything.
func (c *Client) ListAllTasks(ctx context.Context, projectID, templateID, limit int) ([]db.TaskWithTpl, error) {
	var all []db.TaskWithTpl
	before := 0
	for {
		count := MaxTasksPage
		if limit > 0 && limit-len(all) < count {
			count = limit - len(all)
		}
		page, err := c.ListTasks(ctx, projectID, templateID, count, before)
		if err != nil {
			return all, err
		}
		all = append(all, page.Tasks...)
		if !page.HasNext || len(page.Tasks) == 0 || (limit > 0 && len(all) >= limit) {
			return all, nil
		}
		before = page.Before()
	}
}

// GetTaskOutput returns the stored log lines of a task.
func (c *Client) GetTaskOutput(ctx context.Context, projectID, taskID int) ([]db.TaskOutput, error) {
	var lines []db.TaskOutput
	err := c.Do(ctx, http.MethodGet, projectPath(projectID, "/tasks/%d/output", taskID), nil, &lines)
	return lines, err
}

// GetTaskRawOutput streams the plain-text log (ANSI codes stripped) into w.
func (c *Client) GetTaskRawOutput(ctx context.Context, projectID, taskID int, w io.Writer) error {
	resp, err := c.DoRaw(ctx, http.MethodGet, projectPath(projectID, "/tasks/%d/raw_output", taskID), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(w, resp.Body)
	return err
}

// StopTask asks the server to stop a task; force kills the process at once.
func (c *Client) StopTask(ctx context.Context, projectID, taskID int, force bool) error {
	body := struct {
		Force bool `json:"force"`
	}{Force: force}
	return c.Do(ctx, http.MethodPost, projectPath(projectID, "/tasks/%d/stop", taskID), body, nil)
}

// ConfirmTask lets a task in waiting_confirmation proceed.
func (c *Client) ConfirmTask(ctx context.Context, projectID, taskID int) error {
	return c.Do(ctx, http.MethodPost, projectPath(projectID, "/tasks/%d/confirm", taskID), nil, nil)
}

// RejectTask refuses a task in waiting_confirmation.
func (c *Client) RejectTask(ctx context.Context, projectID, taskID int) error {
	return c.Do(ctx, http.MethodPost, projectPath(projectID, "/tasks/%d/reject", taskID), nil, nil)
}

// DeleteTask removes a finished task and its log.
func (c *Client) DeleteTask(ctx context.Context, projectID, taskID int) error {
	return c.Do(ctx, http.MethodDelete, projectPath(projectID, "/tasks/%d", taskID), nil, nil)
}
