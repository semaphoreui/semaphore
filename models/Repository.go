package models

type Repository struct {
	ID        int    `db:"id" json:"id"`
	ProjectID int    `db:"project_id" json:"project_id"`
	GitUrl    string `db:"git_url" json:"git_url" binding:"required"`
	SshKeyID  int    `db:"ssh_key_id" json:"ssh_key_id" binding:"required"`
}
