package db

type SecretStorageType string

const (
	SecretStorageTypeLocal SecretStorageType = "local"
	SecretStorageTypeVault SecretStorageType = "vault"
	SecretStorageTypeDvls  SecretStorageType = "dvls"
)

type SecretStorage struct {
	ID        int               `db:"id" json:"id" backup:"-"`
	ProjectID int               `db:"project_id" json:"project_id" backup:"-"`
	Name      string            `db:"name" json:"name"`
	Type      SecretStorageType `db:"type" json:"type"`
	Params    MapStringAnyField `db:"params" json:"params"`
	ReadOnly  bool              `db:"readonly" json:"readonly"`

	SourceStorageType *AccessKeySourceStorageType `db:"-" json:"source_storage_type,omitempty" backup:"-"`
	// Secret is a source value: literal secret for local storage,
	// env var name for "env", or file path for "file".
	Secret string `db:"-" json:"secret,omitempty" backup:"-"`
}
