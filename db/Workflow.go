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

	Nodes []WorkflowNode `db:"-" bolt:"include" json:"nodes" backup:"nodes"`
	Edges []WorkflowEdge `db:"-" bolt:"include" json:"edges" backup:"edges"`
}

type WorkflowNode struct {
	ID int `db:"id" json:"id" backup:"-"`

	WorkflowTemplateID int                     `db:"workflow_template_id" json:"workflow_template_id" backup:"-"`
	TemplateID         int                     `db:"template_id" json:"template_id,omitempty" backup:"template_id"`
	Kind               WorkflowNodeKind        `db:"kind" json:"kind,omitempty" backup:"kind"`
	ConvergenceMode    WorkflowConvergenceMode `db:"convergence_mode" json:"convergence_mode,omitempty" backup:"convergence_mode"`
	ApprovalTimeout    *int                    `db:"approval_timeout" json:"approval_timeout,omitempty" backup:"approval_timeout"`
	ApprovalMessage    *string                 `db:"approval_message" json:"approval_message,omitempty" backup:"approval_message"`

	InventoryID   *int `db:"inventory_id" json:"inventory_id,omitempty" backup:"inventory_id"`
	EnvironmentID *int `db:"environment_id" json:"environment_id,omitempty" backup:"environment_id"`
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
	WorkflowRunSuccess WorkflowRunStatus = "success"
	WorkflowRunFailed  WorkflowRunStatus = "failed"
)

type WorkflowRun struct {
	ID int `db:"id" json:"id" backup:"-"`

	ProjectID          int `db:"project_id" json:"project_id" backup:"-"`
	WorkflowTemplateID int `db:"workflow_template_id" json:"workflow_template_id" backup:"workflow_template_id"`

	Status WorkflowRunStatus `db:"status" json:"status" backup:"status"`

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
	case WorkflowNodeTaskKind, WorkflowNodeApprovalKind:
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

		if kind == WorkflowNodeTaskKind {
			if node.TemplateID == 0 {
				return NewValidationError("workflow task node requires template_id")
			}
			if node.ApprovalTimeout != nil || node.ApprovalMessage != nil {
				return NewValidationError("workflow task node can not contain approval fields")
			}

			tpl, err := d.GetTemplate(workflow.ProjectID, node.TemplateID)
			if err != nil {
				return NewValidationError("workflow node references a missing template")
			}
			if tpl.ProjectID != workflow.ProjectID {
				return NewValidationError("workflow node template must belong to workflow project")
			}
		} else {
			if node.TemplateID != 0 || node.InventoryID != nil || node.EnvironmentID != nil {
				return NewValidationError("workflow approval node can not contain task node fields")
			}
			if node.ApprovalTimeout != nil && *node.ApprovalTimeout <= 0 {
				return NewValidationError("workflow approval timeout must be greater than zero")
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
		if err := edge.Condition.Validate(); err != nil {
			return err
		}

		incomingCount[edge.DestinationNodeID]++
		adjacency[edge.SourceNodeID] = append(adjacency[edge.SourceNodeID], edge.DestinationNodeID)
	}

	roots := 0
	for id := range nodeByID {
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

	for id := range nodeByID {
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
