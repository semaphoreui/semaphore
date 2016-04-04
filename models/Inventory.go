package models

type Inventory struct {
	ID        int    `db:"id" json:"id"`
	ProjectID int    `db:"project_id" json:"project_id"`
	KeyID     *int   `db:"key_id" json:"key_id"`
	Inventory string `db:"inventory" json:"inventory"`

	// static/aws/do/gcloud
	Type string `db:"type" json:"type"`
}
