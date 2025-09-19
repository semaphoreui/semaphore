package factory

import (
	"github.com/Digital-Data-Co/forge/db"
	"github.com/Digital-Data-Co/forge/pro/db/sql"
)

func NewTerraformStore(store db.Store) db.TerraformStore {
	return &sql.TerraformStoreImpl{}
}

func NewAnsibleTaskRepository(store db.Store) db.AnsibleTaskRepository {
	return &sql.AnsibleTaskStoreImpl{}
}
