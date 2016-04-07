package models

import (
	"time"

	"github.com/ansible-semaphore/semaphore/database"
)

type Project struct {
	ID      int       `db:"id" json:"id"`
	Name    string    `db:"name" json:"name" binding:"required"`
	Created time.Time `db:"created" json:"created"`
}

func (project *Project) CreateProject() error {
	res, err := database.Mysql.Exec("insert into project set name=?", project.Name)
	if err != nil {
		return err
	}

	projectID, err := res.LastInsertId()
	if err != nil {
		return err
	}

	project.ID = int(projectID)
	project.Created = time.Now()

	return nil
}

func init() {
	database.Mysql.AddTableWithName(Project{}, "project").SetKeys(true, "id")
}
