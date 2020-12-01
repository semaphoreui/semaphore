package models

import (
	"strconv"

	"github.com/ansible-semaphore/semaphore/util"
)

// AccessKey represents a key used to access a machine with ansible from semaphore
type AccessKey struct {
	ID   int    `db:"id" json:"id"`
	Name string `db:"name" json:"name" binding:"required"`
	// 'aws/do/gcloud/ssh'
	Type string `db:"type" json:"type" binding:"required"`

	ProjectID *int    `db:"project_id" json:"project_id"`
	Key       *string `db:"key" json:"key"`
	Secret    *string `db:"secret" json:"secret"`

	Removed bool `db:"removed" json:"removed"`
}

// GetPath returns the location of the access key once written to disk
func (key AccessKey) GetPath() string {
	return util.Config.TmpPath + "/access_key_" + strconv.Itoa(key.ID)
}
