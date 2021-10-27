package sql

import "github.com/ansible-semaphore/semaphore/db"

func (d *SqlDb) GetView(projectID int, viewID int) (view db.View, err error) {
	return
}

func (d *SqlDb) GetViews(projectID int) (views []db.View, err error) {
	views = make([]db.View, 0)
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

func (d *SqlDb) SetViewPositions(projectID int, positions map[int]int) error {
	return nil
}
