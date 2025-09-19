package sql

import (
	"github.com/Digital-Data-Co/forge/db"
	"github.com/Digital-Data-Co/forge/db/sql"
)

type AnsibleTaskStoreImpl struct {
}

func NewAnsibleTask(connection *sql.SqlDbConnection) db.AnsibleTaskRepository {
	return &AnsibleTaskStoreImpl{}
}

func (d *AnsibleTaskStoreImpl) CreateAnsibleTaskHost(host db.AnsibleTaskHost) error {
	return nil
}

func (d *AnsibleTaskStoreImpl) CreateAnsibleTaskError(error db.AnsibleTaskError) error {
	return nil
}

func (d *AnsibleTaskStoreImpl) GetAnsibleTaskHosts(projectID int, taskID int) (res []db.AnsibleTaskHost, err error) {
	return
}

func (d *AnsibleTaskStoreImpl) GetAnsibleTaskErrors(projectID int, taskID int) (res []db.AnsibleTaskError, err error) {
	return
}
