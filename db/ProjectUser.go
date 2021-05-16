package db

type ProjectUser struct {
	ID        int  `db:"id" json:"-"`
	ProjectID int  `db:"project_id" json:"project_id"`
	UserID    int  `db:"user_id" json:"user_id"`
	Admin     bool `db:"admin" json:"admin"`
}
