package models

type Environment struct {
	ID        int    `db:"id" json:"id"`
	ProjectID int    `db:"project_id" json:"project_id"`
	Password  string `db:"password" json:"password"`
	JSON      string `db:"json" json:"json"`
}
