//go:build !pro

package bolt

import (
	"github.com/pkg/errors"
	"github.com/semaphoreui/semaphore/db"
)

func (d *BoltDb) CreateAnsibleTaskHost(host db.AnsibleTaskHost) error {
	return errors.New("not implemented")
}

func (d *BoltDb) CreateAnsibleTaskError(error db.AnsibleTaskError) error {
	return errors.New("not implemented")
}

func (d *BoltDb) GetAnsibleTaskHosts(projectID int, taskID int) (res []db.AnsibleTaskHost, err error) {
	err = errors.New("not implemented")
	return
}

func (d *BoltDb) GetAnsibleTaskErrors(projectID int, taskID int) (res []db.AnsibleTaskError, err error) {
	err = errors.New("not implemented")
	return
}
