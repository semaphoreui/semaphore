package db

// Repository is the model for code stored in a git repository
type Repository struct {
	ID        int    `db:"id" json:"id"`
	Name      string `db:"name" json:"name" binding:"required"`
	ProjectID int    `db:"project_id" json:"project_id"`
	GitURL    string `db:"git_url" json:"git_url" binding:"required"`
	GitBranch string `db:"git_branch" json:"git_branch" binding:"required"`
	SSHKeyID  int    `db:"ssh_key_id" json:"ssh_key_id" binding:"required"`
	Removed   bool   `db:"removed" json:"removed"`

	SSHKey AccessKey `db:"-" json:"-"`
}
