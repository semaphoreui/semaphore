package apiclient

import (
	"context"
	"io"
	"net/http"
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

// ListProjects returns the projects the user is a member of.
func (c *Client) ListProjects(ctx context.Context) ([]db.Project, error) {
	var projects []db.Project
	err := c.Do(ctx, http.MethodGet, "/projects", nil, &projects)
	return projects, err
}

// GetProject returns one project.
func (c *Client) GetProject(ctx context.Context, projectID int) (db.Project, error) {
	var project db.Project
	err := c.Do(ctx, http.MethodGet, projectPath(projectID, ""), nil, &project)
	return project, err
}

// CreateProject creates a project; demo adds the sample resources the UI offers.
func (c *Client) CreateProject(ctx context.Context, project db.Project, demo bool) (db.Project, error) {
	body := struct {
		db.Project
		Demo bool `json:"demo"`
	}{Project: project, Demo: demo}
	var created db.Project
	err := c.Do(ctx, http.MethodPost, "/projects", body, &created)
	return created, err
}

// UpdateProject sends the whole project object; read it first.
func (c *Client) UpdateProject(ctx context.Context, project db.Project) error {
	return c.Do(ctx, http.MethodPut, projectPath(project.ID, ""), project, nil)
}

// DeleteProject removes a project with everything in it.
func (c *Client) DeleteProject(ctx context.Context, projectID int) error {
	return c.Do(ctx, http.MethodDelete, projectPath(projectID, ""), nil, nil)
}

// BackupProject streams the JSON backup of a project into w.
func (c *Client) BackupProject(ctx context.Context, projectID int, w io.Writer) error {
	resp, err := c.DoRaw(ctx, http.MethodGet, projectPath(projectID, "/backup"), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(w, resp.Body)
	return err
}

// RestoreProject creates a project from a backup produced by BackupProject.
func (c *Client) RestoreProject(ctx context.Context, backup io.Reader) (db.Project, error) {
	var project db.Project
	err := c.Do(ctx, http.MethodPost, "/projects/restore", backup, &project)
	return project, err
}

// ProjectMember is a user of a project as GET /project/{id}/users returns it:
// roles are slugs, and Role repeats the slug when there is exactly one, for
// older clients.
type ProjectMember struct {
	ID          int          `json:"id"`
	Username    string       `json:"username"`
	Name        string       `json:"name"`
	Role        string       `json:"role,omitempty"`
	Roles       []string     `json:"roles"`
	MappedRoles []MappedRole `json:"mapped_roles,omitempty"`
}

// MappedRole is a role granted through an identity-provider group mapping.
type MappedRole struct {
	Role     string `json:"role"`
	Provider string `json:"provider"`
	Group    string `json:"group"`
}

// ListProjectMembers returns the users of a project with their roles.
func (c *Client) ListProjectMembers(ctx context.Context, projectID int) ([]ProjectMember, error) {
	var members []ProjectMember
	err := c.Do(ctx, http.MethodGet, projectPath(projectID, "/users"), nil, &members)
	return members, err
}

// AddProjectMember grants a user a role in a project. role is a role slug
// (owner, manager, task_runner, guest or a custom one).
func (c *Client) AddProjectMember(ctx context.Context, projectID, userID int, role string) error {
	body := struct {
		UserID int    `json:"user_id"`
		Role   string `json:"role"`
	}{UserID: userID, Role: role}
	return c.Do(ctx, http.MethodPost, projectPath(projectID, "/users"), body, nil)
}

// UpdateProjectMember changes the role of a project member.
func (c *Client) UpdateProjectMember(ctx context.Context, projectID, userID int, role string) error {
	body := struct {
		Role string `json:"role"`
	}{Role: role}
	return c.Do(ctx, http.MethodPut, projectPath(projectID, "/users/%d", userID), body, nil)
}

// RemoveProjectMember removes a user from a project.
func (c *Client) RemoveProjectMember(ctx context.Context, projectID, userID int) error {
	return c.Do(ctx, http.MethodDelete, projectPath(projectID, "/users/%d", userID), nil, nil)
}

// ListUsers returns every user of the instance (admin only).
func (c *Client) ListUsers(ctx context.Context) ([]db.User, error) {
	var users []db.User
	err := c.Do(ctx, http.MethodGet, "/users", nil, &users)
	return users, err
}

// GetUser returns one user.
func (c *Client) GetUser(ctx context.Context, userID int) (db.User, error) {
	var user db.User
	err := c.Do(ctx, http.MethodGet, "/users/"+strconv.Itoa(userID), nil, &user)
	return user, err
}

// CreateUser creates a user; password may be empty for an external user.
func (c *Client) CreateUser(ctx context.Context, user db.User, password string) (db.User, error) {
	body := db.UserWithPwd{User: user, Pwd: password}
	var created db.User
	err := c.Do(ctx, http.MethodPost, "/users", body, &created)
	return created, err
}

// UpdateUser sends the whole user object; read it first.
func (c *Client) UpdateUser(ctx context.Context, user db.User) error {
	return c.Do(ctx, http.MethodPut, "/users/"+strconv.Itoa(user.ID), user, nil)
}

// SetUserPassword changes a user's password. currentPassword is required only
// when a user changes their own password.
func (c *Client) SetUserPassword(ctx context.Context, userID int, password, currentPassword string) error {
	body := struct {
		Pwd        string `json:"password"`
		CurrentPwd string `json:"current_password,omitempty"`
	}{Pwd: password, CurrentPwd: currentPassword}
	return c.Do(ctx, http.MethodPost, "/users/"+strconv.Itoa(userID)+"/password", body, nil)
}

// DeleteUser removes a user.
func (c *Client) DeleteUser(ctx context.Context, userID int) error {
	return c.Do(ctx, http.MethodDelete, "/users/"+strconv.Itoa(userID), nil, nil)
}
