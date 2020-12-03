package db

import (
	"errors"
	"github.com/ansible-semaphore/semaphore/models"
	"github.com/go-gorp/gorp/v3"
)


type RetrieveQueryParams struct {
	Offset       int
	Count        int
	SortBy       string
	SortInverted bool
}

var ErrNotFound = errors.New("sql: no rows in result set")

type Store interface {
	Connect() error
	Close() error
	Migrate() error

	GetEnvironment(projectID int, environmentID int) (models.Environment, error)
	GetEnvironments(projectID int, params RetrieveQueryParams) ([]models.Environment, error)
	UpdateEnvironment(env models.Environment) error

	GetUsers(params RetrieveQueryParams) ([]models.User, error)
	CreateUser(user models.User) (models.User, error)
	DeleteUser(userID int) error
	UpdateUser(user models.User) error
	SetUserPassword(userID int, password string) error
	GetUser(userID int) (models.User, error)

	CreateProject(project models.Project) (models.Project, error)
	//DeleteProject(projectId int) error
	//UpdateProject(project Project) error
	//GetProjectById(projectId int) (Project, error)
	//GetProjects(userId int) ([]Project, error)
	//

	GetTemplates(projectID int, params RetrieveQueryParams) ([]models.Template, error)
	CreateTemplate(template models.Template) (models.Template, error)
	UpdateTemplate(template models.Template) error
	GetTemplate(projectID int, templateID int) (models.Template, error)
	DeleteTemplate(projectID int, templateID int) error

	CreateProjectUser(projectUser models.ProjectUser) (models.ProjectUser, error)
	DeleteProjectUser(projectID, userID int) error

	CreateEvent(event models.Event) (models.Event, error)

	GetAPITokens(userID int) ([]models.APIToken, error)
	CreateAPIToken(token models.APIToken) (models.APIToken, error)
	GetAPIToken(tokenID string) (models.APIToken, error)
	ExpireAPIToken(userID int, tokenID string) error

	GetSession(userID int, sessionID int) (models.Session, error)
	ExpireSession(userID int, sessionID int) error
	TouchSession(userID int, sessionID int) error

	Sql() *gorp.DbMap
}
