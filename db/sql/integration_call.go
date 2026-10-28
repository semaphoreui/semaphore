package sql

import (
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/Masterminds/squirrel"
	"github.com/semaphoreui/semaphore/db"
)

// Widths of the project__integration_call columns written from request data.
// Method and source IP reach us through the request: `handlers.ProxyHeaders`
// rewrites RemoteAddr from X-Forwarded-For without validating it, and a reason
// carries an error message. All three are truncated rather than refused —
// losing the tail of a value must not lose the call itself.
const (
	maxIntegrationCallReason   = 255
	maxIntegrationCallSourceIP = 64
	maxIntegrationCallMethod   = 10
	// A webhook body can be megabytes, and it is stored once per integration
	// the call reaches.
	maxIntegrationCallBody = 64 * 1024
	// Anyone who knows a webhook URL can send calls that fail auth, so only the
	// newest of them are kept per integration.
	maxUnauthorizedIntegrationCalls = 100
)

// integrationCallListColumns leaves request_body out: a page of the history
// would otherwise carry every body on it. GetIntegrationCall loads one with it.
const integrationCallListColumns = "id, project_id, integration_id, integration_name, alias, created, " +
	"source_ip, method, outcome, reason, duration, task_id, run_id, workflow_id, request_size"

// truncateUTF8 cuts s to at most n bytes without splitting a rune. A cut inside
// a multi-byte rune leaves invalid UTF-8, which PostgreSQL rejects outright.
func truncateUTF8(s string, n int) string {
	if len(s) <= n {
		return s
	}

	for n > 0 && !utf8.RuneStart(s[n]) {
		n--
	}

	return s[:n]
}

func (d *SqlDb) CreateIntegrationCall(call db.IntegrationCall) error {
	call.Reason = truncateUTF8(call.Reason, maxIntegrationCallReason)
	call.SourceIP = truncateUTF8(call.SourceIP, maxIntegrationCallSourceIP)
	call.Method = truncateUTF8(call.Method, maxIntegrationCallMethod)

	if call.RequestBody != nil {
		// A body need not be text at all: a text column rejects invalid UTF-8, and
		// PostgreSQL rejects NUL too. The replacement character keeps it visible
		// that bytes were lost.
		body := strings.ReplaceAll(*call.RequestBody, "\x00", "\uFFFD")
		body = truncateUTF8(strings.ToValidUTF8(body, "\uFFFD"), maxIntegrationCallBody)
		call.RequestBody = &body
	}

	_, err := d.exec(
		"insert into project__integration_call "+
			"(project_id, integration_id, integration_name, alias, created, source_ip, method, outcome, reason, duration, "+
			"task_id, run_id, workflow_id, request_size, request_body) "+
			"values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
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
		call.RunID,
		call.WorkflowID,
		call.RequestSize,
		call.RequestBody)

	if err != nil || call.Outcome != db.IntegrationCallUnauthorized || call.IntegrationID == nil {
		return err
	}

	return d.pruneUnauthorizedIntegrationCalls(call.ProjectID, *call.IntegrationID)
}

// pruneUnauthorizedIntegrationCalls deletes all but the newest
// maxUnauthorizedIntegrationCalls rejected calls of an integration. HA nodes may
// prune at once; a second delete finds nothing.
func (d *SqlDb) pruneUnauthorizedIntegrationCalls(projectID int, integrationID int) error {
	var oldest db.IntegrationCall
	err := d.selectOne(&oldest,
		"select id from project__integration_call where project_id=? and integration_id=? and outcome=? "+
			"order by id desc limit 1 offset ?",
		projectID, integrationID, db.IntegrationCallUnauthorized, maxUnauthorizedIntegrationCalls)

	if errors.Is(err, db.ErrNotFound) {
		return nil
	}

	if err != nil {
		return err
	}

	_, err = d.exec(
		"delete from project__integration_call where project_id=? and integration_id=? and outcome=? and id<=?",
		projectID, integrationID, db.IntegrationCallUnauthorized, oldest.ID)

	return err
}

// GetIntegrationCalls returns a page of the project's integration calls, newest
// first. An empty alias or a nil integrationID means no filter on that column.
// params.BeforeID pages backwards by primary key, params.Count caps the page.
// includeBody loads the request bodies, which the history page does not need.
func (d *SqlDb) GetIntegrationCalls(
	projectID int, params db.RetrieveQueryParams, alias string, integrationID *int, includeBody bool,
) (calls []db.IntegrationCall, err error) {

	columns := integrationCallListColumns
	if includeBody {
		columns = "*"
	}

	q := squirrel.Select(columns).
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

// GetIntegrationCall returns one call with its request body. The project is part
// of the lookup, so a call id from another project is not found.
func (d *SqlDb) GetIntegrationCall(projectID int, callID int) (call db.IntegrationCall, err error) {
	err = d.selectOne(&call,
		"select * from project__integration_call where project_id=? and id=?",
		projectID, callID)
	return
}
