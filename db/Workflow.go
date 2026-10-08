package db

import (
	"regexp"
	"time"

	"github.com/semaphoreui/semaphore/pkg/common_errors"
)

type WorkflowEdgeCondition string

const (
	WorkflowEdgeOnSuccess WorkflowEdgeCondition = "on_success"
	WorkflowEdgeOnFailure WorkflowEdgeCondition = "on_failure"
	WorkflowEdgeAlways    WorkflowEdgeCondition = "always"
)

type WorkflowNodeKind string

const (
	WorkflowNodeTaskKind     WorkflowNodeKind = "task"
	WorkflowNodeApprovalKind WorkflowNodeKind = "approval"
	WorkflowNodeNoteKind     WorkflowNodeKind = "note"
	WorkflowNodeDelayKind    WorkflowNodeKind = "delay"
)

// WorkflowEdgeInputMode says how the outputs of the edge's source node feed
// the survey variables of its destination task node.
type WorkflowEdgeInputMode string

const (
	// WorkflowEdgeInputByName delivers every output whose name equals the name
	// of a non-secret survey variable of the destination template. The default;
	// an empty mode means by_name.
	WorkflowEdgeInputByName WorkflowEdgeInputMode = "by_name"
	// WorkflowEdgeInputExplicit delivers only the pairs listed in InputMappings.
	WorkflowEdgeInputExplicit WorkflowEdgeInputMode = "explicit"
)

// WorkflowInputMapping feeds one survey variable of the destination template
// (Var) from one top-level output of the source node (Key).
type WorkflowInputMapping struct {
	Var string `json:"var" backup:"var"`
	Key string `json:"key" backup:"key"`
}

// workflowOutputNamePattern is the shape of an output name. The hyphen is
// allowed because Terraform output names may contain it.
var workflowOutputNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_-]*$`)

// IsValidWorkflowOutputName reports whether name may be the key of an output
// or of an input mapping.
func IsValidWorkflowOutputName(name string) bool {
	return workflowOutputNamePattern.MatchString(name)
}

type WorkflowConvergenceMode string

const (
	WorkflowConvergenceAll WorkflowConvergenceMode = "all"
	WorkflowConvergenceAny WorkflowConvergenceMode = "any"
)

type WorkflowTemplate struct {
	ID int `db:"id" json:"id" backup:"-"`

	ProjectID int    `db:"project_id" json:"project_id" backup:"-"`
	Name      string `db:"name" json:"name" backup:"name"`

	Description *string `db:"description" json:"description,omitempty" backup:"description"`

	StartVersion *string `db:"start_version" json:"start_version,omitempty" backup:"start_version"`

	Nodes []WorkflowNode `db:"-" bolt:"include" json:"nodes" backup:"-"`
	Edges []WorkflowEdge `db:"-" bolt:"include" json:"edges" backup:"edges"`

	// RevisionID and Revision describe the graph revision the Nodes/Edges were
	// loaded from: the latest one for a template read through
	// GetWorkflowTemplate, the run's own for a graph read for a run. Ignored on
	// write — every PUT creates a new revision.
	RevisionID int `db:"-" json:"revision_id,omitempty" backup:"-"`
	Revision   int `db:"-" json:"revision,omitempty" backup:"-"`

	// RevisionAuthorID is set by the API layer before Create/Update so the
	// revision records who saved it. Not part of the JSON contract.
	RevisionAuthorID *int `db:"-" json:"-" backup:"-"`

	LastRun *WorkflowRun `db:"-" json:"last_run,omitempty" backup:"-"`
}

// WorkflowRevision is one immutable snapshot of a workflow template's graph.
// Nodes and edges belong to a revision and are never rewritten: saving a
// template appends a revision, a run pins the revision it started from, so an
// edit can neither change a running run nor orphan the node ids of past runs.
// Revisions no run refers to are deleted when a newer one is saved.
type WorkflowRevision struct {
	ID int `db:"id" json:"id"`

	ProjectID          int `db:"project_id" json:"project_id"`
	WorkflowTemplateID int `db:"workflow_template_id" json:"workflow_template_id"`

	// Number is the 1-based ordinal of the revision within its template; the
	// highest number is the current graph.
	Number int `db:"number" json:"number"`

	Created         time.Time `db:"created" json:"created"`
	CreatedByUserID *int      `db:"created_by_user_id" json:"created_by_user_id,omitempty"`

	// HasRuns is filled by list queries: true when at least one run pins this
	// revision (such a revision survives later saves).
	HasRuns bool `db:"-" json:"has_runs"`
}

type WorkflowNode struct {
	ID int `db:"id" json:"id" backup:"id"`

	WorkflowTemplateID int `db:"workflow_template_id" json:"workflow_template_id" backup:"-"`
	RevisionID         int `db:"revision_id" json:"-" backup:"-"`

	TemplateID      int                     `db:"template_id" json:"template_id,omitempty" backup:"-"`
	Kind            WorkflowNodeKind        `db:"kind" json:"kind,omitempty" backup:"kind"`
	ConvergenceMode WorkflowConvergenceMode `db:"convergence_mode" json:"convergence_mode,omitempty" backup:"convergence_mode"`
	ApprovalTimeout *int                    `db:"approval_timeout" json:"approval_timeout,omitempty" backup:"approval_timeout"`
	ApprovalMessage *string                 `db:"approval_message" json:"approval_message,omitempty" backup:"approval_message"`

	TaskParamsID *int        `db:"task_params_id" json:"-" backup:"-"`
	TaskParams   *TaskParams `db:"-" json:"task_params,omitempty" backup:"task_params"`

	Note         *string `db:"note" json:"note,omitempty" backup:"note"`
	DelaySeconds *int    `db:"delay_seconds" json:"delay_seconds,omitempty" backup:"delay_seconds"`

	PositionX int `db:"position_x" json:"position_x" backup:"position_x"`
	PositionY int `db:"position_y" json:"position_y" backup:"position_y"`
}

type WorkflowEdge struct {
	ID int `db:"id" json:"id" backup:"-"`

	WorkflowTemplateID int `db:"workflow_template_id" json:"workflow_template_id" backup:"-"`
	RevisionID         int `db:"revision_id" json:"-" backup:"-"`
	SourceNodeID       int `db:"source_node_id" json:"source_node_id" backup:"source_node_id"`
	DestinationNodeID  int `db:"destination_node_id" json:"destination_node_id" backup:"destination_node_id"`

	Condition WorkflowEdgeCondition `db:"condition" json:"condition" backup:"condition"`

	// InputMode and InputMappings say how the outputs of the source node feed
	// the survey variables of the destination task node; see
	// AGENTS/work/workflow-editor/artifacts.md § Consumer. Only an edge into a
	// task node may carry them.
	InputMode WorkflowEdgeInputMode `db:"input_mode" json:"input_mode,omitempty" backup:"input_mode"`

	// InputMappingsJSON used internally for read from database.
	// It is not used for store input mappings to database.
	// Do not use it in your code. Use InputMappings instead.
	InputMappingsJSON *string                `db:"input_mappings" json:"-" backup:"-"`
	InputMappings     []WorkflowInputMapping `db:"-" json:"input_mappings,omitempty" backup:"input_mappings"`
}

func (mode WorkflowEdgeInputMode) Validate() error {
	switch mode {
	case WorkflowEdgeInputByName, WorkflowEdgeInputExplicit:
		return nil
	default:
		return common_errors.NewValidationError("workflow edge input mode is invalid")
	}
}

func (edge WorkflowEdge) EffectiveInputMode() WorkflowEdgeInputMode {
	if edge.InputMode == "" {
		return WorkflowEdgeInputByName
	}
	return edge.InputMode
}

type WorkflowDelayStatus string

const (
	WorkflowDelayWaiting WorkflowDelayStatus = "waiting"
	WorkflowDelaySuccess WorkflowDelayStatus = "success"
	WorkflowDelayStopped WorkflowDelayStatus = "stopped"
)

func (status WorkflowDelayStatus) Validate() error {
	switch status {
	case WorkflowDelayWaiting, WorkflowDelaySuccess, WorkflowDelayStopped:
		return nil
	default:
		return common_errors.NewValidationError("workflow delay status is invalid")
	}
}

type WorkflowDelay struct {
	ID int `db:"id" json:"id" backup:"-"`

	ProjectID      int                 `db:"project_id" json:"project_id" backup:"-"`
	WorkflowRunID  int                 `db:"workflow_run_id" json:"workflow_run_id" backup:"workflow_run_id"`
	WorkflowNodeID int                 `db:"workflow_node_id" json:"workflow_node_id" backup:"workflow_node_id"`
	Status         WorkflowDelayStatus `db:"status" json:"status" backup:"status"`
	ResumeAt       time.Time           `db:"resume_at" json:"resume_at" backup:"resume_at"`
	Created        time.Time           `db:"created" json:"created" backup:"created"`
	Resolved       *time.Time          `db:"resolved" json:"resolved,omitempty" backup:"resolved"`
}

type WorkflowRunStatus string

const (
	WorkflowRunRunning  WorkflowRunStatus = "running"
	WorkflowRunApproval WorkflowRunStatus = "approval"
	WorkflowRunSuccess  WorkflowRunStatus = "success"
	WorkflowRunStopped  WorkflowRunStatus = "stopped"
	WorkflowRunFailed   WorkflowRunStatus = "failed"
)

func (status WorkflowRunStatus) IsFinished() bool {
	return status == WorkflowRunSuccess || status == WorkflowRunStopped || status == WorkflowRunFailed
}

type WorkflowRun struct {
	ID int `db:"id" json:"id" backup:"-"`

	ProjectID          int `db:"project_id" json:"project_id" backup:"-"`
	WorkflowTemplateID int `db:"workflow_template_id" json:"workflow_template_id" backup:"workflow_template_id"`

	// RevisionID pins the graph revision the run executes; the engine and the
	// run view read nodes and edges through it, never through the template.
	RevisionID int `db:"revision_id" json:"revision_id,omitempty" backup:"-"`

	Status WorkflowRunStatus `db:"status" json:"status" backup:"status"`

	Version *string `db:"version" json:"version,omitempty" backup:"version"`

	Start *time.Time `db:"start" json:"start,omitempty" backup:"start"`
	End   *time.Time `db:"end" json:"end,omitempty" backup:"end"`

	RootTaskID *int `db:"root_task_id" json:"root_task_id,omitempty" backup:"root_task_id"`
}

type WorkflowApprovalStatus string

const (
	WorkflowApprovalPending  WorkflowApprovalStatus = "pending"
	WorkflowApprovalApproved WorkflowApprovalStatus = "approved"
	WorkflowApprovalRejected WorkflowApprovalStatus = "rejected"
)

type WorkflowApproval struct {
	ID int `db:"id" json:"id" backup:"-"`

	ProjectID        int                    `db:"project_id" json:"project_id" backup:"-"`
	WorkflowRunID    int                    `db:"workflow_run_id" json:"workflow_run_id" backup:"workflow_run_id"`
	WorkflowNodeID   int                    `db:"workflow_node_id" json:"workflow_node_id" backup:"workflow_node_id"`
	Status           WorkflowApprovalStatus `db:"status" json:"status" backup:"status"`
	Created          time.Time              `db:"created" json:"created" backup:"created"`
	Resolved         *time.Time             `db:"resolved" json:"resolved,omitempty" backup:"resolved"`
	ResolvedByUserID *int                   `db:"resolved_by_user_id" json:"resolved_by_user_id,omitempty" backup:"resolved_by_user_id"`
}

func (condition WorkflowEdgeCondition) Validate() error {
	switch condition {
	case WorkflowEdgeOnSuccess, WorkflowEdgeOnFailure, WorkflowEdgeAlways:
		return nil
	default:
		return common_errors.NewValidationError("workflow edge condition is invalid")
	}
}

func (kind WorkflowNodeKind) Validate() error {
	switch kind {
	case WorkflowNodeTaskKind,
		WorkflowNodeApprovalKind,
		WorkflowNodeNoteKind,
		WorkflowNodeDelayKind:
		return nil
	default:
		return common_errors.NewValidationError("workflow node kind is invalid")
	}
}

func (node WorkflowNode) EffectiveKind() WorkflowNodeKind {
	if node.Kind == "" {
		return WorkflowNodeTaskKind
	}
	return node.Kind
}

func (mode WorkflowConvergenceMode) Validate() error {
	switch mode {
	case WorkflowConvergenceAll, WorkflowConvergenceAny:
		return nil
	default:
		return common_errors.NewValidationError("workflow node convergence mode is invalid")
	}
}

func (node WorkflowNode) EffectiveConvergenceMode() WorkflowConvergenceMode {
	if node.ConvergenceMode == "" {
		return WorkflowConvergenceAll
	}
	return node.ConvergenceMode
}

func (status WorkflowApprovalStatus) Validate() error {
	switch status {
	case WorkflowApprovalPending, WorkflowApprovalApproved, WorkflowApprovalRejected:
		return nil
	default:
		return common_errors.NewValidationError("workflow approval status is invalid")
	}
}

// WorkflowTemplateValidationStore is the slice of Store workflow template
// validation depends on. Narrowing the dependency lets callers (and tests)
// validate against a minimal mock instead of a full store.
type WorkflowTemplateValidationStore interface {
	GetTemplate(projectID int, templateID int) (Template, error)
}
