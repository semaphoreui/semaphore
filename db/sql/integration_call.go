package sql

import (
	"github.com/Masterminds/squirrel"
	"github.com/semaphoreui/semaphore/db"
)

// maxIntegrationCallReason is the width of project__integration_call.reason.
// A reason comes from an error message, so it is truncated rather than refused:
// losing the tail of an explanation must not lose the call itself.
const maxIntegrationCallReason = 255

func (d *SqlDb) CreateIntegrationCall(call db.IntegrationCall) error {
	if len(call.Reason) > maxIntegrationCallReason {
		call.Reason = call.Reason[:maxIntegrationCallReason]
	}

	_, err := d.exec(
		"insert into project__integration_call "+
			"(project_id, integration_id, integration_name, alias, created, source_ip, method, outcome, reason, duration, task_id, run_id) "+
			"values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		call.ProjectID,
		call.IntegrationID,
		call.IntegrationName,
		call.Alias,
		call.Created,
		call.SourceIP,
		call.Method,
		call.Outcome,
		call.Reason,
		call.Duration,
		call.TaskID,
		call.RunID)

	return err
}

// GetIntegrationCalls returns a page of the project's integration calls, newest
// first. An empty alias or a nil integrationID means no filter on that column.
// params.BeforeID pages backwards by primary key, params.Count caps the page.
func (d *SqlDb) GetIntegrationCalls(
	projectID int, params db.RetrieveQueryParams, alias string, integrationID *int,
) (calls []db.IntegrationCall, err error) {

	q := squirrel.Select("*").
		From("project__integration_call").
		Where("project_id=?", projectID).
		OrderBy("id desc")

	if alias != "" {
		q = q.Where("alias=?", alias)
	}

	if integrationID != nil {
		q = q.Where("integration_id=?", *integrationID)
	}

	// Keyset pagination: the index on (project_id, id) walks backwards whatever
	// the depth, unlike OFFSET which scans and discards the skipped rows.
	if params.BeforeID > 0 {
		q = q.Where("id < ?", params.BeforeID)
	}

	if params.Count > 0 {
		q = q.Limit(uint64(params.Count))
	}

	query, args, err := q.ToSql()
	if err != nil {
		return
	}

	calls = []db.IntegrationCall{}
	_, err = d.selectAll(&calls, query, args...)

	return
}
