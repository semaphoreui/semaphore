package db

import "fmt"

type AccessKey struct {
	ID   int    `db:"id" json:"id"`
	Name string `db:"name" json:"name" binding:"required"`
	// 'aws/do/gcloud/ssh',
	Type string `db:"type" json:"type" binding:"required"`

	ProjectID *int    `db:"project_id" json:"project_id"`
	Key       *string `db:"key" json:"key"`
	Secret    *string `db:"secret" json:"secret"`

	Removed bool `db:"removed" json:"removed"`
}

func (key AccessKey) GetPath(tmpdir string) string {
	return fmt.Sprintf("%s/%s_%d", tmpdir, key.Type, key.ID)
}
