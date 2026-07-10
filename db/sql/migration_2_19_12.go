package sql

import (
	"github.com/go-gorp/gorp/v3"
)

type migration_2_19_12 struct {
	db *SqlDb
}

// PostApply repairs workflow nodes whose per-node inventory/environment overrides
// were not copied into project__task_params by the initial 2.19.11 migration.
// It is idempotent: nodes already fully migrated are left unchanged.
func (m migration_2_19_12) PostApply(tx *gorp.Transaction) error {
	type nodeOverrides struct {
		ID            int     `db:"id"`
		ProjectID     int     `db:"project_id"`
		TaskParamsID  *int64  `db:"task_params_id"`
		Limit         *string `db:"limit"`
		InventoryID   *int    `db:"inventory_id"`
		EnvironmentID *int    `db:"environment_id"`
	}

	var nodes []nodeOverrides
	_, err := tx.Select(&nodes, m.db.PrepareQuery(
		"select n.id as id, wt.project_id as project_id, n.task_params_id as task_params_id, "+
			"n.`limit` as `limit`, n.inventory_id as inventory_id, n.environment_id as environment_id "+
			"from project__workflow_node n "+
			"join project__workflow_template wt on wt.id = n.workflow_template_id "+
			"where n.inventory_id is not null or n.environment_id is not null"))
	if err != nil {
		return err
	}

	for _, n := range nodes {
		if err = upsertWorkflowNodeTaskParams(
			tx, m.db, n.ID, n.ProjectID, n.Limit, n.InventoryID, n.EnvironmentID, n.TaskParamsID); err != nil {
			return err
		}
	}

	return nil
}
