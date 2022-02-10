package bolt

import "github.com/ansible-semaphore/semaphore/db"

func (d *BoltDb) GetView(projectID int, viewID int) (view db.View, err error) {
	err = d.getObject(projectID, db.ViewProps, intObjectID(viewID), &view)
	return
}

func (d *BoltDb) GetViews(projectID int) (views []db.View, err error) {
	err = d.getObjects(projectID, db.ViewProps, db.RetrieveQueryParams{}, nil, &views)
	return
}

func (d *BoltDb) UpdateView(view db.View) error {
	return d.updateObject(view.ProjectID, db.ViewProps, view)
}

func (d *BoltDb) CreateView(view db.View) (db.View, error) {
	newView, err := d.createObject(view.ProjectID, db.ViewProps, view)
	return newView.(db.View), err
}

func (d *BoltDb) DeleteView(projectID int, viewID int) error {
	return d.deleteObject(projectID, db.ViewProps, intObjectID(viewID), nil)
}

func (d *BoltDb) SetViewPositions(projectID int, positions map[int]int) error {
	for id, position := range positions {
		view, err := d.GetView(projectID, id)
		if err != nil {
			return err
		}
		view.Position = position
		err = d.UpdateView(view)
		if err != nil {
			return err
		}
	}
	return nil
}
