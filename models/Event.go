package models

import (
	"time"

	"github.com/ansible-semaphore/semaphore/database"
)

type Event struct {
	ProjectID   *int      `db:"project_id" json:"project_id"`
	ObjectID    *int      `db:"object_id" json:"object_id"`
	ObjectType  *string   `db:"object_type" json:"object_type"`
	Description *string   `db:"description" json:"description"`
	Created     time.Time `db:"created" json:"created"`

	ObjectName  string  `db:"-" json:"object_name"`
	ProjectName *string `db:"project_name" json:"project_name"`
}

func (evt Event) Insert() error {
	_, err := database.Mysql.Exec("insert into event set project_id=?, object_id=?, object_type=?, description=?, created=NOW(6)", evt.ProjectID, evt.ObjectID, evt.ObjectType, evt.Description)

	return err
}
