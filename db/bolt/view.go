package bolt

import "github.com/ansible-semaphore/semaphore/db"

func (d *BoltDb) GetView(projectID int, viewID int) (view db.View, err error) {
	return
}

func (d *BoltDb) GetViews(projectID int) (views []db.View, err error) {
	return
}

func (d *BoltDb) UpdateView(view db.View) error {
	return nil
}

func (d *BoltDb) CreateView(view db.View) (newView db.View, err error) {
	return
}

func (d *BoltDb) DeleteView(projectID int, viewID int) error {
	return nil
}