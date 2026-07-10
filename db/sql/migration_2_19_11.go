package sql

import (
	"database/sql"
	"encoding/json"

	"github.com/go-gorp/gorp/v3"
)

type migration_2_19_11 struct {
	db *SqlDb
}

// PostApply moves per-node run parameters (limit, inventory_id, environment_id)
// of workflow nodes into project__task_params rows referenced by the freshly
// added task_params_id column.
//
// The legacy inventory_id/environment_id/limit columns are intentionally kept
// (unused): they participate in table-level foreign key definitions, and SQLite
// can not drop such columns without rebuilding the table.
func (m migration_2_19_11) PostApply(tx *gorp.Transaction) error {
	type nodeOverrides struct {
		ID            int     `db:"id"`
		ProjectID     int     `db:"project_id"`
		Limit         *string `db:"limit"`
		InventoryID   *int    `db:"inventory_id"`
		EnvironmentID *int    `db:"environment_id"`
	}

	var nodes []nodeOverrides
	_, err := tx.Select(&nodes, m.db.PrepareQuery(
		"select n.id as id, wt.project_id as project_id, n.`limit` as `limit`, "+
			"n.inventory_id as inventory_id, n.environment_id as environment_id "+
			"from project__workflow_node n "+
			"join project__workflow_template wt on wt.id = n.workflow_template_id "+
			"where (n.`limit` is not null and n.`limit` <> '' and n.`limit` <> '[]') "+
			"or n.inventory_id is not null or n.environment_id is not null"))
	if err != nil {
		return err
	}

	for _, n := range nodes {
		if err = upsertWorkflowNodeTaskParams(tx, m.db, n.ID, n.ProjectID, n.Limit, n.InventoryID, n.EnvironmentID, nil); err != nil {
			return err
		}
	}

	return nil
}

// upsertWorkflowNodeTaskParams copies legacy per-node overrides into
// project__task_params. When existingParamsID is nil and the node already has
// task_params_id, the existing row is updated instead of creating a duplicate.
func upsertWorkflowNodeTaskParams(
	tx *gorp.Transaction,
	db *SqlDb,
	nodeID int,
	projectID int,
	limit *string,
	inventoryID *int,
	environmentID *int,
	existingParamsID *int64,
) error {
	paramsArg, inventoryArg, environment, err := workflowNodeOverrideValues(
		tx, db, limit, inventoryID, environmentID)
	if err != nil {
		return err
	}

	if paramsArg == nil && inventoryArg == nil && environment == "" {
		return nil
	}

	var paramsID int64
	if existingParamsID != nil {
		paramsID = *existingParamsID
	} else {
		var currentID sql.NullInt64
		currentID, err = tx.SelectNullInt(
			db.PrepareQuery("select task_params_id from project__workflow_node where id = ?"),
			nodeID)
		if err != nil {
			return err
		}
		if currentID.Valid {
			paramsID = currentID.Int64
		}
	}

	if paramsID == 0 {
		insertQuery := "insert into project__task_params (project_id, params, environment, message, inventory_id) values (?, ?, ?, '', ?)"
		switch db.Sql().Dialect.(type) {
		case gorp.PostgresDialect:
			paramsID, err = tx.SelectInt(
				db.PrepareQuery(insertQuery+" returning id"),
				projectID, paramsArg, environment, inventoryArg)
			if err != nil {
				return err
			}
		default:
			res, err := tx.Exec(
				db.PrepareQuery(insertQuery),
				projectID, paramsArg, environment, inventoryArg)
			if err != nil {
				return err
			}
			paramsID, err = res.LastInsertId()
			if err != nil {
				return err
			}
		}

		_, err = tx.Exec(
			db.PrepareQuery("update project__workflow_node set task_params_id=? where id=?"),
			paramsID, nodeID)
		if err != nil {
			return err
		}
		return nil
	}

	updateQuery := "update project__task_params set params = coalesce(?, params), " +
		"environment = case when ? <> '' then ? else environment end, " +
		"inventory_id = coalesce(?, inventory_id) where id = ?"
	_, err = tx.Exec(
		db.PrepareQuery(updateQuery),
		paramsArg, environment, environment, inventoryArg, paramsID)
	return err
}

func workflowNodeOverrideValues(
	tx *gorp.Transaction,
	db *SqlDb,
	limit *string,
	inventoryID *int,
	environmentID *int,
) (paramsArg any, inventoryArg any, environment string, err error) {
	if limit != nil && *limit != "" && *limit != "[]" {
		var limitVals []string
		if json.Unmarshal([]byte(*limit), &limitVals) == nil && len(limitVals) > 0 {
			var params []byte
			params, err = json.Marshal(map[string]any{"limit": limitVals})
			if err != nil {
				return
			}
			paramsArg = string(params)
		}
	}

	if environmentID != nil {
		environment, err = tx.SelectStr(
			db.PrepareQuery("select json from project__environment where id = ?"),
			*environmentID)
		if err != nil {
			return
		}
	}

	if inventoryID != nil {
		inventoryArg = *inventoryID
	}

	return
}
