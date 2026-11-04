package apiclient

import (
	"context"
	"net/http"
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

// GlobalScope selects the instance-wide runners instead of a project's.
const GlobalScope = 0

// RunnerWithToken is a freshly created runner together with the registration
// token the server shows exactly once.
type RunnerWithToken struct {
	db.Runner
	Token string `json:"token"`
}

// RegistrationToken is the answer to a token re-issue.
type RegistrationToken struct {
	RegistrationToken string `json:"registration_token"`
	RunnerID          int    `json:"runner_id"`
}

// runnersPath is "/runners..." for GlobalScope and "/project/{id}/runners..."
// otherwise; both trees expose the same operations.
func runnersPath(projectID int, suffix string) string {
	if projectID == GlobalScope {
		return "/runners" + suffix
	}
	return projectPath(projectID, "/runners%s", suffix)
}

func (c *Client) ListRunners(ctx context.Context, projectID int) ([]db.Runner, error) {
	var items []db.Runner
	err := c.Do(ctx, http.MethodGet, runnersPath(projectID, ""), nil, &items)
	return items, err
}

func (c *Client) GetRunner(ctx context.Context, projectID, runnerID int) (db.Runner, error) {
	var item db.Runner
	err := c.Do(ctx, http.MethodGet, runnersPath(projectID, "/"+strconv.Itoa(runnerID)), nil, &item)
	return item, err
}

// CreateRunner registers a runner record; the returned Token is shown once.
func (c *Client) CreateRunner(ctx context.Context, projectID int, runner db.Runner) (RunnerWithToken, error) {
	var created RunnerWithToken
	err := c.Do(ctx, http.MethodPost, runnersPath(projectID, ""), runner, &created)
	return created, err
}

func (c *Client) UpdateRunner(ctx context.Context, projectID int, runner db.Runner) error {
	return c.Do(ctx, http.MethodPut, runnersPath(projectID, "/"+strconv.Itoa(runner.ID)), runner, nil)
}

func (c *Client) SetRunnerActive(ctx context.Context, projectID, runnerID int, active bool) error {
	body := struct {
		Active bool `json:"active"`
	}{Active: active}
	return c.Do(ctx, http.MethodPost, runnersPath(projectID, "/"+strconv.Itoa(runnerID)+"/active"), body, nil)
}

// RegenerateRunnerToken re-issues the registration token of a runner.
func (c *Client) RegenerateRunnerToken(ctx context.Context, projectID, runnerID int) (RegistrationToken, error) {
	var token RegistrationToken
	err := c.Do(ctx, http.MethodPost, runnersPath(projectID, "/"+strconv.Itoa(runnerID)+"/registration-token"), nil, &token)
	return token, err
}

// ClearRunnerCache asks the runner to drop its repository cache.
func (c *Client) ClearRunnerCache(ctx context.Context, projectID, runnerID int) error {
	return c.Do(ctx, http.MethodDelete, runnersPath(projectID, "/"+strconv.Itoa(runnerID)+"/cache"), nil, nil)
}

func (c *Client) DeleteRunner(ctx context.Context, projectID, runnerID int) error {
	return c.Do(ctx, http.MethodDelete, runnersPath(projectID, "/"+strconv.Itoa(runnerID)), nil, nil)
}
