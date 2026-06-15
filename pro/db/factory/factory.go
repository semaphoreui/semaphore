package factory

import (
	"github.com/semaphoreui/semaphore/db"
	coreSql "github.com/semaphoreui/semaphore/db/sql"
	proSql "github.com/semaphoreui/semaphore/pro/db/sql"
)

func NewTerraformStore(store db.Store) db.TerraformStore {
	return &proSql.TerraformStoreImpl{}
}

func NewAnsibleTaskRepository(store db.Store) db.AnsibleTaskRepository {
	return &proSql.AnsibleTaskStoreImpl{}
}

func NewWorkflowStore(store db.Store) db.WorkflowManager {
	sqlDb := store.(*coreSql.SqlDb)
	return proSql.NewWorkflowStoreImpl(sqlDb.GetConnection())
}
