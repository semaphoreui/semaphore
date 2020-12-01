package models

type ProjectUser struct {
	ProjectID int  `db:"project_id" json:"project_id"`
	UserID    int  `db:"user_id" json:"user_id"`
	Admin     bool `db:"admin" json:"admin"`
}
