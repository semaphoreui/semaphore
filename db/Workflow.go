package db

import (
	"fmt"
	"time"

	"github.com/semaphoreui/semaphore/pkg/task_logger"
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
	// WorkflowNodeNoteKind is a pure annotation node: it holds free-form text,
	// does not execute, and is excluded from the run graph (no edges, never a
	// root, ignored by the runner).
	WorkflowNodeNoteKind WorkflowNodeKind = "note"
)

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

	// StartVersion seeds run versioning, mirroring a build template's
	// start_version. The first run takes this value; later runs bump its numeric
	// segment via GetNextBuildVersion.
	StartVersion *string `db:"start_version" json:"start_version,omitempty" backup:"start_version"`

	// Nodes carries the backup:"-" tag because the export/import layer wraps each
	// node (see project.BackupWorkflowNode) to translate template/inventory/
	// environment references from IDs to names. Edges only reference node IDs,
	// which are kept verbatim in the backup, so they are exported as-is.
	Nodes []WorkflowNode `db:"-" bolt:"include" json:"nodes" backup:"-"`
	Edges []WorkflowEdge `db:"-" bolt:"include" json:"edges" backup:"edges"`

	// LastRun is the most recent run of this workflow, attached when listing
	// workflows so the list can show status/version (nil otherwise). Read-only,
	// not persisted.
	LastRun *WorkflowRun `db:"-" json:"last_run,omitempty" backup:"-"`
}

type WorkflowNode struct {
	// ID is kept in the backup (backup:"id") so that edges, which reference
	// nodes by ID, stay resolvable after export/import. The ID is only
	// meaningful within a single workflow and is remapped to a fresh value when
	// the graph is persisted (see SqlDb.writeWorkflowGraph).
	ID int `db:"id" json:"id" backup:"id"`

	WorkflowTemplateID int `db:"workflow_template_id" json:"workflow_template_id" backup:"-"`
	// TemplateID, InventoryID and EnvironmentID are excluded from the backup
	// (backup:"-") because they are project-scoped IDs. The export/import layer
	// translates them to/from names (see project.BackupWorkflowNode).
	TemplateID      int                     `db:"template_id" json:"template_id,omitempty" backup:"-"`
	Kind            WorkflowNodeKind        `db:"kind" json:"kind,omitempty" backup:"kind"`
	ConvergenceMode WorkflowConvergenceMode `db:"convergence_mode" json:"convergence_mode,omitempty" backup:"convergence_mode"`
	ApprovalTimeout *int                    `db:"approval_timeout" json:"approval_timeout,omitempty" backup:"approval_timeout"`
	ApprovalMessage *string                 `db:"approval_message" json:"approval_message,omitempty" backup:"approval_message"`

	InventoryID   *int             `db:"inventory_id" json:"inventory_id,omitempty" backup:"-"`
	EnvironmentID *int             `db:"environment_id" json:"environment_id,omitempty" backup:"-"`
	Limit         StringArrayField `db:"limit" json:"limit,omitempty" backup:"limit"`

	// Note holds the free-form text of a "note" node. It is only valid on note
	// nodes and is ignored during execution.
	Note *string `db:"note" json:"note,omitempty" backup:"note"`

	// PositionX/PositionY are the node's coordinates on the graphical editor
	// canvas. They are pure layout metadata and do not participate in validation
	// or execution. Stored as integer pixels for cross-dialect safety.
	PositionX int `db:"position_x" json:"position_x" backup:"position_x"`
	PositionY int `db:"position_y" json:"position_y" backup:"position_y"`
}

type WorkflowEdge struct {
	ID int `db:"id" json:"id" backup:"-"`

	WorkflowTemplateID int `db:"workflow_template_id" json:"workflow_template_id" backup:"-"`
	SourceNodeID       int `db:"source_node_id" json:"source_node_id" backup:"source_node_id"`
	DestinationNodeID  int `db:"destination_node_id" json:"destination_node_id" backup:"destination_node_id"`

	Condition WorkflowEdgeCondition `db:"condition" json:"condition" backup:"condition"`
}

type WorkflowRunStatus string

const (
	WorkflowRunRunning WorkflowRunStatus = "running"
	// WorkflowRunApproval marks a run that is paused waiting for human input:
	// it has at least one pending approval. Like running, it is non-terminal
	// (the run has not ended).
	WorkflowRunApproval WorkflowRunStatus = "approval"
	WorkflowRunSuccess  WorkflowRunStatus = "success"
	WorkflowRunFailed   WorkflowRunStatus = "failed"
	// WorkflowRunStopped marks a run that a user manually stopped: its in-flight
	// tasks were signalled to stop and no further nodes are launched. Like
	// success/failed it is terminal.
	WorkflowRunStopped WorkflowRunStatus = "stopped"
)

// IsFinished reports whether the run has reached a terminal state. Running and
// approval are non-terminal: the run is still in progress (approval is merely
// blocked on a human decision). Stopped is terminal — a stopped run is never
// progressed or re-evaluated.
func (status WorkflowRunStatus) IsFinished() bool {
	return status == WorkflowRunSuccess || status == WorkflowRunFailed || status == WorkflowRunStopped
}

type WorkflowRun struct {
	ID int `db:"id" json:"id" backup:"-"`

	ProjectID          int `db:"project_id" json:"project_id" backup:"-"`
	WorkflowTemplateID int `db:"workflow_template_id" json:"workflow_template_id" backup:"workflow_template_id"`

	Status WorkflowRunStatus `db:"status" json:"status" backup:"status"`

	// Version is the build version assigned to this run (derived from the
	// template's StartVersion) and propagated to every task the run launches.
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

func workflowNodeID(node WorkflowNode, idx int) int {
	if node.ID != 0 {
		return node.ID
	}
	// Validation needs a stable per-node identifier even before persistence assigns IDs.
	// Use deterministic negative placeholders for nodes without explicit IDs.
	return -(idx + 1)
}

func (condition WorkflowEdgeCondition) Validate() error {
	switch condition {
	case WorkflowEdgeOnSuccess, WorkflowEdgeOnFailure, WorkflowEdgeAlways:
		return nil
	default:
		return NewValidationError("workflow edge condition is invalid")
	}
}

func (kind WorkflowNodeKind) Validate() error {
	switch kind {
	case WorkflowNodeTaskKind, WorkflowNodeApprovalKind, WorkflowNodeNoteKind:
		return nil
	default:
		return NewValidationError("workflow node kind is invalid")
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
		return NewValidationError("workflow node convergence mode is invalid")
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
		return NewValidationError("workflow approval status is invalid")
	}
}

func WorkflowConditionMatches(status task_logger.TaskStatus, condition WorkflowEdgeCondition) bool {
	switch condition {
	case WorkflowEdgeOnSuccess:
		return status == task_logger.TaskSuccessStatus
	case WorkflowEdgeOnFailure:
		return status.IsFinished() && status != task_logger.TaskSuccessStatus
	case WorkflowEdgeAlways:
		return status.IsFinished()
	default:
		return false
	}
}

func ValidateWorkflowTemplate(d Store, workflow WorkflowTemplate) error {
	if workflow.Name == "" {
		return NewValidationError("workflow name can not be empty")
	}

	if len(workflow.Nodes) == 0 {
		return NewValidationError("workflow must contain at least one node")
	}

	nodeByID := make(map[int]WorkflowNode)
	for i, node := range workflow.Nodes {
		nodeID := workflowNodeID(node, i)
		if _, ok := nodeByID[nodeID]; ok {
			return NewValidationError("workflow contains duplicate node ids")
		}
		kind := node.EffectiveKind()
		if err := kind.Validate(); err != nil {
			return err
		}
		mode := node.EffectiveConvergenceMode()
		if err := mode.Validate(); err != nil {
			return err
		}

		hasTaskFields := node.TemplateID != 0 || node.InventoryID != nil ||
			node.EnvironmentID != nil || (node.Limit != nil && len(node.Limit) > 0)
		hasApprovalFields := node.ApprovalTimeout != nil || node.ApprovalMessage != nil

		switch kind {
		case WorkflowNodeTaskKind:
			if node.TemplateID == 0 {
				return NewValidationError("workflow task node requires template_id")
			}
			if hasApprovalFields {
				return NewValidationError("workflow task node can not contain approval fields")
			}
			if node.Note != nil {
				return NewValidationError("only note nodes can contain note text")
			}

			tpl, err := d.GetTemplate(workflow.ProjectID, node.TemplateID)
			if err != nil {
				return NewValidationError("workflow node references a missing template")
			}
			if tpl.ProjectID != workflow.ProjectID {
				return NewValidationError("workflow node template must belong to workflow project")
			}
		case WorkflowNodeApprovalKind:
			if hasTaskFields {
				return NewValidationError("workflow approval node can not contain task node fields")
			}
			if node.ApprovalTimeout != nil && *node.ApprovalTimeout <= 0 {
				return NewValidationError("workflow approval timeout must be greater than zero")
			}
			if node.Note != nil {
				return NewValidationError("only note nodes can contain note text")
			}
		case WorkflowNodeNoteKind:
			if hasTaskFields {
				return NewValidationError("workflow note node can not contain task node fields")
			}
			if hasApprovalFields {
				return NewValidationError("workflow note node can not contain approval fields")
			}
		}

		nodeByID[nodeID] = node
	}

	incomingCount := make(map[int]int)
	adjacency := make(map[int][]int)

	for _, edge := range workflow.Edges {
		if edge.SourceNodeID == 0 || edge.DestinationNodeID == 0 {
			return NewValidationError("workflow edge node id can not be empty")
		}
		if edge.SourceNodeID == edge.DestinationNodeID {
			return NewValidationError("workflow can not contain self-referencing edges")
		}
		if _, ok := nodeByID[edge.SourceNodeID]; !ok {
			return NewValidationError("workflow edge source node does not belong to workflow")
		}
		if _, ok := nodeByID[edge.DestinationNodeID]; !ok {
			return NewValidationError("workflow edge destination node does not belong to workflow")
		}
		if nodeByID[edge.SourceNodeID].EffectiveKind() == WorkflowNodeNoteKind ||
			nodeByID[edge.DestinationNodeID].EffectiveKind() == WorkflowNodeNoteKind {
			return NewValidationError("workflow note node can not be connected by edges")
		}
		if err := edge.Condition.Validate(); err != nil {
			return err
		}

		incomingCount[edge.DestinationNodeID]++
		adjacency[edge.SourceNodeID] = append(adjacency[edge.SourceNodeID], edge.DestinationNodeID)
	}

	// Note nodes are pure annotations: they never execute and are excluded from
	// the run graph, so they do not count toward the single-root requirement.
	roots := 0
	for id, node := range nodeByID {
		if node.EffectiveKind() == WorkflowNodeNoteKind {
			continue
		}
		if incomingCount[id] == 0 {
			roots++
		}
	}
	if roots != 1 {
		return NewValidationError("workflow must have exactly one root node")
	}

	const (
		markNone = iota
		markVisiting
		markDone
	)
	// Three-color DFS cycle detection:
	// visiting->visiting means a back-edge and therefore a cycle.

	marks := make(map[int]int)
	var visit func(id int) bool
	visit = func(id int) bool {
		switch marks[id] {
		case markVisiting:
			return true
		case markDone:
			return false
		}

		marks[id] = markVisiting
		for _, next := range adjacency[id] {
			if visit(next) {
				return true
			}
		}
		marks[id] = markDone
		return false
	}

	for id, node := range nodeByID {
		if node.EffectiveKind() == WorkflowNodeNoteKind {
			continue
		}
		if visit(id) {
			return NewValidationError("workflow graph must be a DAG")
		}
	}

	return nil
}

func WorkflowRootNode(workflow WorkflowTemplate) (WorkflowNode, error) {
	incoming := make(map[int]int)

	for _, edge := range workflow.Edges {
		incoming[edge.DestinationNodeID]++
	}

	var root *WorkflowNode
	for i := range workflow.Nodes {
		node := &workflow.Nodes[i]
		if node.EffectiveKind() == WorkflowNodeNoteKind {
			continue
		}
		if incoming[node.ID] == 0 {
			if root != nil {
				return WorkflowNode{}, fmt.Errorf("workflow has multiple root nodes")
			}
			root = node
		}
	}

	if root == nil {
		return WorkflowNode{}, fmt.Errorf("workflow has no root node")
	}

	return *root, nil
}
