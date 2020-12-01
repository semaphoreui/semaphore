package db

import (
	"github.com/ansible-semaphore/semaphore/models"
	"github.com/go-gorp/gorp/v3"
)


type RetrieveQueryParams struct {
	Offset       int
	Count        int
	SortBy       string
	SortInverted bool
}


type Store interface {
	Connect() error
	Close() error
	Migrate() error

	GetUsers(params RetrieveQueryParams) ([]models.User, error)
	GetAllUsers() ([]models.User, error)
	CreateUser(user models.User) (models.User, error)
	DeleteUser(userID int) error
	UpdateUser(userID int, user models.User) error
	SetUserPassword(userID int, password string) error

	//UpdateUser(user User) error
	GetUserById(userID int) (models.User, error)
	CreateProject(project models.Project) (models.Project, error)
	//DeleteProject(projectId int) error
	//UpdateProject(project Project) error
	//GetProjectById(projectId int) (Project, error)
	//GetProjects(userId int) ([]Project, error)
	//
	//CreateTemplate(template Template) error
	//DeleteTemplate(projectId, templateId int) error
	//UpdateTemplate(template Template) error
	CreateProjectUser(projectUser models.ProjectUser) (models.ProjectUser, error)
	DeleteProjectUser(projectID, userID int) error
	CreateEvent(event models.Event) (models.Event, error)

	CreateAPIToken(token models.APIToken) (models.APIToken, error)
	Sql() *gorp.DbMap
}
