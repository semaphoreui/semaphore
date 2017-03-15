package db

type Environment struct {
	ID        int     `db:"id" json:"id"`
	Name      string  `db:"name" json:"name" binding:"required"`
	ProjectID int     `db:"project_id" json:"project_id"`
	Password  *string `db:"password" json:"password"`
	JSON      string  `db:"json" json:"json" binding:"required"`
	Removed   bool    `db:"removed" json:"removed"`
}
