package apiclient

import (
	"context"
	"net/http"

	"github.com/semaphoreui/semaphore/db"
)

// Templates. The list omits vaults and survey details; read a template by id
// before updating it, because PUT without vaults deletes them.

func (c *Client) ListTemplates(ctx context.Context, projectID int) ([]db.Template, error) {
	var items []db.Template
	err := c.Do(ctx, http.MethodGet, projectPath(projectID, "/templates"), nil, &items)
	return items, err
}

func (c *Client) GetTemplate(ctx context.Context, projectID, templateID int) (db.Template, error) {
	var item db.Template
	err := c.Do(ctx, http.MethodGet, projectPath(projectID, "/templates/%d", templateID), nil, &item)
	return item, err
}

func (c *Client) CreateTemplate(ctx context.Context, projectID int, tpl db.Template) (db.Template, error) {
	tpl.ProjectID = projectID
	var created db.Template
	err := c.Do(ctx, http.MethodPost, projectPath(projectID, "/templates"), tpl, &created)
	return created, err
}

func (c *Client) UpdateTemplate(ctx context.Context, projectID int, tpl db.Template) error {
	tpl.ProjectID = projectID
	return c.Do(ctx, http.MethodPut, projectPath(projectID, "/templates/%d", tpl.ID), tpl, nil)
}

func (c *Client) DeleteTemplate(ctx context.Context, projectID, templateID int) error {
	return c.Do(ctx, http.MethodDelete, projectPath(projectID, "/templates/%d", templateID), nil, nil)
}

// StopAllTemplateTasks stops every active task of a template.
func (c *Client) StopAllTemplateTasks(ctx context.Context, projectID, templateID int) error {
	return c.Do(ctx, http.MethodPost, projectPath(projectID, "/templates/%d/stop_all_tasks", templateID), nil, nil)
}

// Inventories.

func (c *Client) ListInventories(ctx context.Context, projectID int) ([]db.Inventory, error) {
	var items []db.Inventory
	err := c.Do(ctx, http.MethodGet, projectPath(projectID, "/inventory"), nil, &items)
	return items, err
}

func (c *Client) GetInventory(ctx context.Context, projectID, inventoryID int) (db.Inventory, error) {
	var item db.Inventory
	err := c.Do(ctx, http.MethodGet, projectPath(projectID, "/inventory/%d", inventoryID), nil, &item)
	return item, err
}

func (c *Client) CreateInventory(ctx context.Context, projectID int, inv db.Inventory) (db.Inventory, error) {
	inv.ProjectID = projectID
	var created db.Inventory
	err := c.Do(ctx, http.MethodPost, projectPath(projectID, "/inventory"), inv, &created)
	return created, err
}

func (c *Client) UpdateInventory(ctx context.Context, projectID int, inv db.Inventory) error {
	inv.ProjectID = projectID
	return c.Do(ctx, http.MethodPut, projectPath(projectID, "/inventory/%d", inv.ID), inv, nil)
}

func (c *Client) DeleteInventory(ctx context.Context, projectID, inventoryID int) error {
	return c.Do(ctx, http.MethodDelete, projectPath(projectID, "/inventory/%d", inventoryID), nil, nil)
}

// Environments (variable groups).

func (c *Client) ListEnvironments(ctx context.Context, projectID int) ([]db.Environment, error) {
	var items []db.Environment
	err := c.Do(ctx, http.MethodGet, projectPath(projectID, "/environment"), nil, &items)
	return items, err
}

func (c *Client) GetEnvironment(ctx context.Context, projectID, environmentID int) (db.Environment, error) {
	var item db.Environment
	err := c.Do(ctx, http.MethodGet, projectPath(projectID, "/environment/%d", environmentID), nil, &item)
	return item, err
}

func (c *Client) CreateEnvironment(ctx context.Context, projectID int, env db.Environment) (db.Environment, error) {
	env.ProjectID = projectID
	var created db.Environment
	err := c.Do(ctx, http.MethodPost, projectPath(projectID, "/environment"), env, &created)
	return created, err
}

func (c *Client) UpdateEnvironment(ctx context.Context, projectID int, env db.Environment) error {
	env.ProjectID = projectID
	return c.Do(ctx, http.MethodPut, projectPath(projectID, "/environment/%d", env.ID), env, nil)
}

func (c *Client) DeleteEnvironment(ctx context.Context, projectID, environmentID int) error {
	return c.Do(ctx, http.MethodDelete, projectPath(projectID, "/environment/%d", environmentID), nil, nil)
}

// Access keys. The server never returns the secret.

func (c *Client) ListKeys(ctx context.Context, projectID int) ([]db.AccessKey, error) {
	var items []db.AccessKey
	err := c.Do(ctx, http.MethodGet, projectPath(projectID, "/keys"), nil, &items)
	return items, err
}

func (c *Client) GetKey(ctx context.Context, projectID, keyID int) (db.AccessKey, error) {
	var item db.AccessKey
	err := c.Do(ctx, http.MethodGet, projectPath(projectID, "/keys/%d", keyID), nil, &item)
	return item, err
}

func (c *Client) CreateKey(ctx context.Context, projectID int, key db.AccessKey) (db.AccessKey, error) {
	key.ProjectID = &projectID
	var created db.AccessKey
	err := c.Do(ctx, http.MethodPost, projectPath(projectID, "/keys"), key, &created)
	return created, err
}

func (c *Client) UpdateKey(ctx context.Context, projectID int, key db.AccessKey) error {
	key.ProjectID = &projectID
	return c.Do(ctx, http.MethodPut, projectPath(projectID, "/keys/%d", key.ID), key, nil)
}

func (c *Client) DeleteKey(ctx context.Context, projectID, keyID int) error {
	return c.Do(ctx, http.MethodDelete, projectPath(projectID, "/keys/%d", keyID), nil, nil)
}

// Repositories.

func (c *Client) ListRepositories(ctx context.Context, projectID int) ([]db.Repository, error) {
	var items []db.Repository
	err := c.Do(ctx, http.MethodGet, projectPath(projectID, "/repositories"), nil, &items)
	return items, err
}

func (c *Client) GetRepository(ctx context.Context, projectID, repositoryID int) (db.Repository, error) {
	var item db.Repository
	err := c.Do(ctx, http.MethodGet, projectPath(projectID, "/repositories/%d", repositoryID), nil, &item)
	return item, err
}

func (c *Client) CreateRepository(ctx context.Context, projectID int, repo db.Repository) (db.Repository, error) {
	repo.ProjectID = projectID
	var created db.Repository
	err := c.Do(ctx, http.MethodPost, projectPath(projectID, "/repositories"), repo, &created)
	return created, err
}

func (c *Client) UpdateRepository(ctx context.Context, projectID int, repo db.Repository) error {
	repo.ProjectID = projectID
	return c.Do(ctx, http.MethodPut, projectPath(projectID, "/repositories/%d", repo.ID), repo, nil)
}

func (c *Client) DeleteRepository(ctx context.Context, projectID, repositoryID int) error {
	return c.Do(ctx, http.MethodDelete, projectPath(projectID, "/repositories/%d", repositoryID), nil, nil)
}

// Schedules.

// ListSchedules returns the project's schedules with the template name
// (tpl_name) the UI shows next to each.
func (c *Client) ListSchedules(ctx context.Context, projectID int) ([]db.ScheduleWithTpl, error) {
	var items []db.ScheduleWithTpl
	err := c.Do(ctx, http.MethodGet, projectPath(projectID, "/schedules"), nil, &items)
	return items, err
}

// ListTemplateSchedules returns the schedules of one template.
func (c *Client) ListTemplateSchedules(ctx context.Context, projectID, templateID int) ([]db.Schedule, error) {
	var items []db.Schedule
	err := c.Do(ctx, http.MethodGet, projectPath(projectID, "/templates/%d/schedules", templateID), nil, &items)
	return items, err
}

func (c *Client) GetSchedule(ctx context.Context, projectID, scheduleID int) (db.Schedule, error) {
	var item db.Schedule
	err := c.Do(ctx, http.MethodGet, projectPath(projectID, "/schedules/%d", scheduleID), nil, &item)
	return item, err
}

func (c *Client) CreateSchedule(ctx context.Context, projectID int, schedule db.Schedule) (db.Schedule, error) {
	schedule.ProjectID = projectID
	var created db.Schedule
	err := c.Do(ctx, http.MethodPost, projectPath(projectID, "/schedules"), schedule, &created)
	return created, err
}

func (c *Client) UpdateSchedule(ctx context.Context, projectID int, schedule db.Schedule) error {
	schedule.ProjectID = projectID
	return c.Do(ctx, http.MethodPut, projectPath(projectID, "/schedules/%d", schedule.ID), schedule, nil)
}

// SetScheduleActive enables or disables a schedule without touching its fields.
func (c *Client) SetScheduleActive(ctx context.Context, projectID, scheduleID int, active bool) error {
	body := struct {
		Active bool `json:"active"`
	}{Active: active}
	return c.Do(ctx, http.MethodPut, projectPath(projectID, "/schedules/%d/active", scheduleID), body, nil)
}

func (c *Client) DeleteSchedule(ctx context.Context, projectID, scheduleID int) error {
	return c.Do(ctx, http.MethodDelete, projectPath(projectID, "/schedules/%d", scheduleID), nil, nil)
}

// ValidateCron asks the server whether a cron expression is acceptable; a bad
// one comes back as an *APIError with the server's explanation.
func (c *Client) ValidateCron(ctx context.Context, projectID int, cronFormat string) error {
	body := struct {
		CronFormat string `json:"cron_format"`
	}{CronFormat: cronFormat}
	return c.Do(ctx, http.MethodPost, projectPath(projectID, "/schedules/validate"), body, nil)
}

// Views (template groups on the project page).

func (c *Client) ListViews(ctx context.Context, projectID int) ([]db.View, error) {
	var items []db.View
	err := c.Do(ctx, http.MethodGet, projectPath(projectID, "/views"), nil, &items)
	return items, err
}

func (c *Client) GetView(ctx context.Context, projectID, viewID int) (db.View, error) {
	var item db.View
	err := c.Do(ctx, http.MethodGet, projectPath(projectID, "/views/%d", viewID), nil, &item)
	return item, err
}

// ListViewTemplates returns the templates shown in a view.
func (c *Client) ListViewTemplates(ctx context.Context, projectID, viewID int) ([]db.Template, error) {
	var items []db.Template
	err := c.Do(ctx, http.MethodGet, projectPath(projectID, "/views/%d/templates", viewID), nil, &items)
	return items, err
}

func (c *Client) CreateView(ctx context.Context, projectID int, view db.View) (db.View, error) {
	view.ProjectID = projectID
	var created db.View
	err := c.Do(ctx, http.MethodPost, projectPath(projectID, "/views"), view, &created)
	return created, err
}

func (c *Client) UpdateView(ctx context.Context, projectID int, view db.View) error {
	view.ProjectID = projectID
	return c.Do(ctx, http.MethodPut, projectPath(projectID, "/views/%d", view.ID), view, nil)
}

func (c *Client) DeleteView(ctx context.Context, projectID, viewID int) error {
	return c.Do(ctx, http.MethodDelete, projectPath(projectID, "/views/%d", viewID), nil, nil)
}
