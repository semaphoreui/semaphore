package sql

import "github.com/ansible-semaphore/semaphore/db"

func (d *SqlDb) GetView(projectID int, viewID int) (view db.View, err error) {
	return
}

func (d *SqlDb) GetViews(projectID int) (views []db.View, err error) {
	return
}

func (d *SqlDb) UpdateView(view db.View) error {
	return nil
}

func (d *SqlDb) CreateView(view db.View) (newView db.View, err error) {
	return
}

func (d *SqlDb) DeleteView(projectID int, viewID int) error {
	return nil
}