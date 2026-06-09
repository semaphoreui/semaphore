package db

type Notification struct {
	ID        int    `db:"id" json:"id"`
	ProjectID int    `db:"project_id" json:"project_id"`
	Name      string `db:"name" json:"name"`
	Type      string `db:"type" json:"type"`
	Config    string `db:"config" json:"config"`
}
