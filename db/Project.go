package db

import (
	"time"
)

// Project is the top level structure in Semaphore
type Project struct {
	ID        int       `db:"id" json:"id"`
	Name      string    `db:"name" json:"name" binding:"required"`
	Created   time.Time `db:"created" json:"created"`
	Alert     bool      `db:"alert" json:"alert"`
	AlertChat string    `db:"alert_chat" json:"alert_chat"`
	Admin 	  bool		`db:"admin" json:"admin"`
}

// CreateProject writes a project to the database
func (project *Project) CreateProject() error {
	project.Created = time.Now()

	res, err := Mysql.Exec("insert into project set name=?, created=?", project.Name, project.Created)
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
