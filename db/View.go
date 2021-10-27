package db

import "fmt"

type View struct {
	ID        int    `db:"id" json:"id"`
	ProjectID int    `db:"project_id" json:"project_id"`
	Title     string `db:"title" json:"title"`
	Position  int    `db:"int" json:"int"`
}

func (view *View) Validate() error {
	if view.Title == "" {
		return fmt.Errorf("title can not be empty")
	}
	return nil
}