package sql

import (
	"github.com/semaphoreui/semaphore/db"
)

type HostStoreImpl struct {
}

func (d *HostStoreImpl) GetHosts(projectID *int, params db.HostListParams) ([]db.Host, int, error) {
	return nil, 0, nil
}

func (d *HostStoreImpl) GetHost(hostID int, projectID *int) (db.HostDetails, error) {
	return db.HostDetails{}, db.ErrNotFound
}

func (d *HostStoreImpl) GetHostTasks(hostID int, projectID *int, count int, beforeTaskID int) ([]db.HostTask, error) {
	return nil, nil
}

func (d *HostStoreImpl) CountActiveHosts() (int, error) {
	return 0, nil
}

func (d *HostStoreImpl) DecommissionHost(hostID int) error {
	return db.ErrNotFound
}
