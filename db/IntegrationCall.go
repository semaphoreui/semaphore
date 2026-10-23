package db

import "time"

// IntegrationCallOutcome is what one integration did with a webhook call.
type IntegrationCallOutcome string

const (
	// IntegrationCallOK means the integration started a task or a workflow run.
	IntegrationCallOK IntegrationCallOutcome = "ok"
	// IntegrationCallUnauthorized means the request failed the integration's auth method.
	IntegrationCallUnauthorized IntegrationCallOutcome = "unauthorized"
	// IntegrationCallSkipped means the integration's matchers did not match the request.
	IntegrationCallSkipped IntegrationCallOutcome = "skipped"
	// IntegrationCallError means the start was attempted and failed.
	IntegrationCallError IntegrationCallOutcome = "error"
)

// IntegrationCall records one webhook call of one integration. One request to an
// alias shared by several integrations records one call per integration, so the
// history can be filtered by integration as well as by alias.
//
// Alias and IntegrationName are plain strings and IntegrationID is nullable: the
// history outlives the alias and the integration the call was made through.
type IntegrationCall struct {
	ID              int                    `db:"id" json:"id"`
	ProjectID       int                    `db:"project_id" json:"project_id"`
	IntegrationID   *int                   `db:"integration_id" json:"integration_id"`
	IntegrationName string                 `db:"integration_name" json:"integration_name"`
	Alias           string                 `db:"alias" json:"alias"`
	Created         time.Time              `db:"created" json:"created"`
	SourceIP        string                 `db:"source_ip" json:"source_ip"`
	Method          string                 `db:"method" json:"method"`
	Outcome         IntegrationCallOutcome `db:"outcome" json:"outcome"`
	Reason          string                 `db:"reason" json:"reason"`
	// Duration is how long the integration took to answer the call, in milliseconds.
	Duration int  `db:"duration" json:"duration"`
	TaskID   *int `db:"task_id" json:"task_id"`
	// RunID and WorkflowID identify the workflow run a call started. Both are
	// needed to link to it, and a workflow run has no root task of its own, so
	// TaskID alone would leave a successful workflow call pointing at nothing.
	RunID      *int `db:"run_id" json:"run_id"`
	WorkflowID *int `db:"workflow_id" json:"workflow_id"`
}
