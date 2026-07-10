package sql

import (
	"encoding/json"

	"github.com/go-gorp/gorp/v3"
)

type migration_2_19_11 struct {
	db *SqlDb
}

// PostApply moves per-node run parameters (previously only `limit`) of workflow
// nodes into project__task_params rows referenced by the freshly added
// task_params_id column.
//
// The legacy inventory_id/environment_id/limit columns are intentionally kept
// (unused): they participate in table-level foreign key definitions, and SQLite
// can not drop such columns without rebuilding the table.
func (m migration_2_19_11) PostApply(tx *gorp.Transaction) error {
	type nodeLimit struct {
		ID        int    `db:"id"`
		ProjectID int    `db:"project_id"`
		Limit     string `db:"limit"`
	}

	var nodes []nodeLimit
	_, err := tx.Select(&nodes, m.db.PrepareQuery(
		"select n.id as id, wt.project_id as project_id, n.`limit` as `limit` "+
			"from project__workflow_node n "+
			"join project__workflow_template wt on wt.id = n.workflow_template_id "+
			"where n.`limit` is not null and n.`limit` <> '' and n.`limit` <> '[]'"))
	if err != nil {
		return err
	}

	for _, n := range nodes {
		var limit []string
		if json.Unmarshal([]byte(n.Limit), &limit) != nil || len(limit) == 0 {
			continue
		}

		params, err2 := json.Marshal(map[string]any{"limit": limit})
		if err2 != nil {
			return err2
		}

		var paramsID int64
		// environment and message map to non-pointer string fields — write ''
		// instead of NULL so the rows scan back like gorp-inserted ones.
		insertQuery := "insert into project__task_params (project_id, params, environment, message) values (?, ?, '', '')"
		switch m.db.Sql().Dialect.(type) {
		case gorp.PostgresDialect:
			paramsID, err = tx.SelectInt(m.db.PrepareQuery(insertQuery+" returning id"), n.ProjectID, string(params))
			if err != nil {
				return err
			}
		default:
			res, err2 := tx.Exec(m.db.PrepareQuery(insertQuery), n.ProjectID, string(params))
			if err2 != nil {
				return err2
			}
			paramsID, err = res.LastInsertId()
			if err != nil {
				return err
			}
		}

		_, err = tx.Exec(
			m.db.PrepareQuery("update project__workflow_node set task_params_id=? where id=?"),
			paramsID, n.ID)
		if err != nil {
			return err
		}
	}

	return nil
}
