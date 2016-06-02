package models

import (
	"time"

	database "github.com/ansible-semaphore/semaphore/db"
)

type Project struct {
	ID      int       `db:"id" json:"id"`
	Name    string    `db:"name" json:"name" binding:"required"`
	Created time.Time `db:"created" json:"created"`
}

func (project *Project) CreateProject() error {
	project.Created = time.Now()

	res, err := database.Mysql.Exec("insert into project set name=?, created=?", project.Name, project.Created)
	if err != nil {
		return err
	}

	projectID, err := res.LastInsertId()
	if err != nil {
		return err
	}

	project.ID = int(projectID)

	return nil
}
