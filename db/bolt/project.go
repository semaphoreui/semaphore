package bolt

import (
	"github.com/ansible-semaphore/semaphore/db"
	"time"
)

func (d *BoltDb) CreateProject(project db.Project) (db.Project, error) {
	project.Created = time.Now()

	newProject, err := d.createObject(0, db.ProjectProps, project)

	if err != nil {
		return db.Project{}, err
	}

	return newProject.(db.Project), nil
}

func (d *BoltDb) GetAllProjects() (projects []db.Project, err error) {
	err = d.getObjects(0, db.ProjectProps, db.RetrieveQueryParams{}, nil, &projects)

	return
}

func (d *BoltDb) GetProjects(userID int) (projects []db.Project, err error) {
	projects = make([]db.Project, 0)

	var allProjects []db.Project

	err = d.getObjects(0, db.ProjectProps, db.RetrieveQueryParams{}, nil, &allProjects)

	if err != nil {
		return
	}

	for _, v := range allProjects {
		_, err2 := d.GetProjectUser(v.ID, userID)
		if err2 == nil {
			projects = append(projects, v)
		} else if err2 != db.ErrNotFound {
			err = err2
			return
		}
	}

	return
}

func (d *BoltDb) GetProject(projectID int) (project db.Project, err error) {
	err = d.getObject(0, db.ProjectProps, intObjectID(projectID), &project)
	return
}

func (d *BoltDb) DeleteProject(projectID int) error {
	return d.deleteObject(0, db.ProjectProps, intObjectID(projectID), nil)
}

func (d *BoltDb) UpdateProject(project db.Project) error {
	return d.updateObject(0, db.ProjectProps, project)
}
