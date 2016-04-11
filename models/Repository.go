package models

import "github.com/ansible-semaphore/semaphore/database"

type Repository struct {
	ID        int    `db:"id" json:"id"`
	Name      string `db:"name" json:"name" binding:"required"`
	ProjectID int    `db:"project_id" json:"project_id"`
	GitUrl    string `db:"git_url" json:"git_url" binding:"required"`
	SshKeyID  int    `db:"ssh_key_id" json:"ssh_key_id" binding:"required"`

	SshKey AccessKey `db:"-" json:"-"`
}

func init() {
	database.Mysql.AddTableWithName(Repository{}, "project__repository").SetKeys(true, "id")
}
