package db

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/models"
	"github.com/go-gorp/gorp/v3"
	"time"
)


const databaseTimeFormat = "2006-01-02T15:04:05:99Z"

// GetParsedTime returns the timestamp as it will retrieved from the database
// This allows us to create timestamp consistency on return values from create requests
func GetParsedTime(t time.Time) time.Time {
	parsedTime, err := time.Parse(databaseTimeFormat, t.Format(databaseTimeFormat))
	if err != nil {
		log.Error(err)
	}
	return parsedTime
}

type RetrieveQueryParams struct {
	Offset       int
	Count        int
	SortBy       string
	SortInverted bool
}

var ErrNotFound = errors.New("sql: no rows in result set")
var ErrInvalidOperation = errors.New("sql: no rows in result set")

type Store interface {
	Connect() error
	Close() error
	Migrate() error

	GetEnvironment(projectID int, environmentID int) (models.Environment, error)
	GetEnvironments(projectID int, params RetrieveQueryParams) ([]models.Environment, error)
	UpdateEnvironment(env models.Environment) error
	CreateEnvironment(env models.Environment) (models.Environment, error)
	DeleteEnvironment(projectID int, templateID int) error
	DeleteEnvironmentSoft(projectID int, templateID int) error

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
