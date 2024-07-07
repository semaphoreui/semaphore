package sql

import "github.com/ansible-semaphore/semaphore/db"

func (d *SqlDb) GetView(projectID int, viewID int) (view db.View, err error) {
	err = d.getObject(projectID, db.ViewProps, viewID, &view)
	return
}

func (d *SqlDb) GetViews(projectID int) (views []db.View, err error) {
	err = d.getObjects(projectID, db.ViewProps, db.RetrieveQueryParams{}, nil, &views)
	return
}

func (d *SqlDb) UpdateView(view db.View) error {
	_, err := d.exec(
		"update project__view set title=?, position=?, project_id=? where id=?",
		view.Title,
		view.Position,
		view.ProjectID,
		view.ID)

	return err
}

func (d *SqlDb) CreateView(view db.View) (newView db.View, err error) {
	insertID, err := d.insert(
		"id",
		"insert into project__view (project_id, title, position) values (?, ?, ?)",
		view.ProjectID,
		view.Title,
		view.Position)

	if err != nil {
		return
	}

	newView = view
	newView.ID = insertID
	return
}

func (d *SqlDb) DeleteView(projectID int, viewID int) error {
	return d.deleteObject(projectID, db.ViewProps, viewID)
}

func (d *SqlDb) SetViewPositions(projectID int, positions map[int]int) error {
	for id, position := range positions {
		_, err := d.exec("update project__view set position=? where project_id=? and id=?",
			position,
			projectID,
			id)
		if err != nil {
			return err
		}
	}
	return nil
}
