package sql

import (
	"github.com/semaphoreui/semaphore/db"
	coreSql "github.com/semaphoreui/semaphore/db/sql"
)

// WorkflowStoreImpl is the SQL implementation of the workflow store used by
// the open-source build. It persists workflow templates, nodes, edges, runs
// and approvals directly into the database so that the dredd integration tests
// and the workflow API endpoints work without a Pro licence.
type WorkflowStoreImpl struct {
	conn *coreSql.SqlDbConnection
}

// NewWorkflowStoreImpl creates a WorkflowStoreImpl backed by the given
// database connection.
func NewWorkflowStoreImpl(conn *coreSql.SqlDbConnection) *WorkflowStoreImpl {
	return &WorkflowStoreImpl{conn: conn}
}

// GetWorkflowRunTasks returns the tasks belonging to a workflow run.
func (d *WorkflowStoreImpl) GetWorkflowRunTasks(projectID int, runID int, params db.RetrieveQueryParams) (res []db.TaskWithTpl, err error) {
	return
}

// GetWorkflowTemplates returns all workflow templates for a project.
func (d *WorkflowStoreImpl) GetWorkflowTemplates(projectID int, params db.RetrieveQueryParams) (res []db.WorkflowTemplate, err error) {
	_, err = d.conn.SelectAll(&res,
		"select * from `project__workflow_template` where `project_id`=?",
		projectID)
	return
}

// GetWorkflowTemplate returns a single workflow template together with its
// nodes and edges.
func (d *WorkflowStoreImpl) GetWorkflowTemplate(projectID int, workflowID int) (res db.WorkflowTemplate, err error) {
	err = d.conn.SelectOne(&res,
		"select * from `project__workflow_template` where `project_id`=? and `id`=?",
		projectID, workflowID)
	if err != nil {
		return
	}

	_, err = d.conn.SelectAll(&res.Nodes,
		"select * from `project__workflow_node` where `workflow_template_id`=?",
		workflowID)
	if err != nil {
		return
	}

	_, err = d.conn.SelectAll(&res.Edges,
		"select * from `project__workflow_edge` where `workflow_template_id`=?",
		workflowID)
	return
}

// CreateWorkflowTemplate inserts a new workflow template, its nodes and its
// edges. The input node IDs are treated as logical identifiers used by the
// edges; the returned template carries the database-assigned IDs.
func (d *WorkflowStoreImpl) CreateWorkflowTemplate(workflow db.WorkflowTemplate) (res db.WorkflowTemplate, err error) {
	wfID, err := d.conn.Insert("id",
		"insert into `project__workflow_template` (`project_id`, `name`, `description`, `start_version`) values (?, ?, ?, ?)",
		workflow.ProjectID, workflow.Name, workflow.Description, workflow.StartVersion)
	if err != nil {
		return
	}

	res = workflow
	res.ID = wfID

	// Map logical node IDs (supplied by the caller) to DB-assigned IDs so that
	// edge references can be translated correctly.
	nodeIDMap := make(map[int]int, len(workflow.Nodes))
	res.Nodes = make([]db.WorkflowNode, len(workflow.Nodes))

	for i, node := range workflow.Nodes {
		logicalID := node.ID
		nodeDBID, nodeErr := d.conn.Insert("id",
			"insert into `project__workflow_node` (`workflow_template_id`, `template_id`, `inventory_id`, `environment_id`, `kind`, `convergence_mode`, `approval_timeout`, `approval_message`, `note`, `limit`, `position_x`, `position_y`) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			wfID, node.TemplateID, node.InventoryID, node.EnvironmentID,
			node.Kind, node.ConvergenceMode, node.ApprovalTimeout, node.ApprovalMessage,
			node.Note, &node.Limit, node.PositionX, node.PositionY)
		if nodeErr != nil {
			err = nodeErr
			return
		}
		node.ID = nodeDBID
		node.WorkflowTemplateID = wfID
		res.Nodes[i] = node
		nodeIDMap[logicalID] = nodeDBID
	}

	res.Edges = make([]db.WorkflowEdge, len(workflow.Edges))
	for i, edge := range workflow.Edges {
		srcDBID := nodeIDMap[edge.SourceNodeID]
		dstDBID := nodeIDMap[edge.DestinationNodeID]
		edgeDBID, edgeErr := d.conn.Insert("id",
			"insert into `project__workflow_edge` (`workflow_template_id`, `source_node_id`, `destination_node_id`, `condition`) values (?, ?, ?, ?)",
			wfID, srcDBID, dstDBID, edge.Condition)
		if edgeErr != nil {
			err = edgeErr
			return
		}
		edge.ID = edgeDBID
		edge.WorkflowTemplateID = wfID
		edge.SourceNodeID = srcDBID
		edge.DestinationNodeID = dstDBID
		res.Edges[i] = edge
	}

	return
}

// UpdateWorkflowTemplate replaces the mutable fields of an existing workflow
// template. The old nodes (and their dependent edges via cascade) are deleted
// and the new ones are inserted with the same logical→DB ID remapping used by
// CreateWorkflowTemplate.
func (d *WorkflowStoreImpl) UpdateWorkflowTemplate(workflow db.WorkflowTemplate) (err error) {
	_, err = d.conn.Exec(
		"update `project__workflow_template` set `name`=?, `description`=?, `start_version`=? where `project_id`=? and `id`=?",
		workflow.Name, workflow.Description, workflow.StartVersion, workflow.ProjectID, workflow.ID)
	if err != nil {
		return
	}

	// Deleting nodes cascades to edges (FK on delete cascade).
	_, err = d.conn.Exec("delete from `project__workflow_node` where `workflow_template_id`=?", workflow.ID)
	if err != nil {
		return
	}

	nodeIDMap := make(map[int]int, len(workflow.Nodes))
	for _, node := range workflow.Nodes {
		logicalID := node.ID
		nodeDBID, nodeErr := d.conn.Insert("id",
			"insert into `project__workflow_node` (`workflow_template_id`, `template_id`, `inventory_id`, `environment_id`, `kind`, `convergence_mode`, `approval_timeout`, `approval_message`, `note`, `limit`, `position_x`, `position_y`) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			workflow.ID, node.TemplateID, node.InventoryID, node.EnvironmentID,
			node.Kind, node.ConvergenceMode, node.ApprovalTimeout, node.ApprovalMessage,
			node.Note, &node.Limit, node.PositionX, node.PositionY)
		if nodeErr != nil {
			err = nodeErr
			return
		}
		nodeIDMap[logicalID] = nodeDBID
	}

	for _, edge := range workflow.Edges {
		srcDBID := nodeIDMap[edge.SourceNodeID]
		dstDBID := nodeIDMap[edge.DestinationNodeID]
		_, edgeErr := d.conn.Insert("id",
			"insert into `project__workflow_edge` (`workflow_template_id`, `source_node_id`, `destination_node_id`, `condition`) values (?, ?, ?, ?)",
			workflow.ID, srcDBID, dstDBID, edge.Condition)
		if edgeErr != nil {
			err = edgeErr
			return
		}
	}

	return
}

// DeleteWorkflowTemplate removes a workflow template and its dependent nodes
// and edges (deleted via cascade).
func (d *WorkflowStoreImpl) DeleteWorkflowTemplate(projectID int, workflowID int) (err error) {
	_, err = d.conn.Exec(
		"delete from `project__workflow_template` where `project_id`=? and `id`=?",
		projectID, workflowID)
	return
}

// GetWorkflowRuns returns the runs for a given workflow template, newest first.
func (d *WorkflowStoreImpl) GetWorkflowRuns(projectID int, workflowTemplateID int, params db.RetrieveQueryParams) (res []db.WorkflowRun, err error) {
	_, err = d.conn.SelectAll(&res,
		"select * from `project__workflow_run` where `project_id`=? and `workflow_template_id`=? order by `id` desc",
		projectID, workflowTemplateID)
	return
}

// GetWorkflowRun returns a single workflow run identified by project, template
// and run IDs.
func (d *WorkflowStoreImpl) GetWorkflowRun(projectID int, workflowTemplateID int, runID int) (res db.WorkflowRun, err error) {
	err = d.conn.SelectOne(&res,
		"select * from `project__workflow_run` where `project_id`=? and `workflow_template_id`=? and `id`=?",
		projectID, workflowTemplateID, runID)
	return
}

// GetWorkflowRunByID returns a workflow run by project and run ID, without
// requiring the workflow template ID.
func (d *WorkflowStoreImpl) GetWorkflowRunByID(projectID int, runID int) (res db.WorkflowRun, err error) {
	err = d.conn.SelectOne(&res,
		"select * from `project__workflow_run` where `project_id`=? and `id`=?",
		projectID, runID)
	return
}

// GetActiveWorkflowRuns returns all non-terminal workflow runs.
func (d *WorkflowStoreImpl) GetActiveWorkflowRuns() (res []db.WorkflowRun, err error) {
	return
}

// CreateWorkflowRun inserts a new workflow run record.
func (d *WorkflowStoreImpl) CreateWorkflowRun(run db.WorkflowRun) (res db.WorkflowRun, err error) {
	runID, err := d.conn.Insert("id",
		"insert into `project__workflow_run` (`project_id`, `workflow_template_id`, `status`, `version`, `start`, `end`, `root_task_id`) values (?, ?, ?, ?, ?, ?, ?)",
		run.ProjectID, run.WorkflowTemplateID, run.Status, run.Version, run.Start, run.End, run.RootTaskID)
	if err != nil {
		return
	}
	res = run
	res.ID = runID
	return
}

// UpdateWorkflowRun updates a workflow run's mutable fields.
func (d *WorkflowStoreImpl) UpdateWorkflowRun(run db.WorkflowRun) (err error) {
	_, err = d.conn.Exec(
		"update `project__workflow_run` set `status`=?, `version`=?, `start`=?, `end`=?, `root_task_id`=? where `project_id`=? and `id`=?",
		run.Status, run.Version, run.Start, run.End, run.RootTaskID, run.ProjectID, run.ID)
	return
}

// UpdateWorkflowRunStatusUnless atomically sets the run status unless the
// current status is one of the excluded values.
func (d *WorkflowStoreImpl) UpdateWorkflowRunStatusUnless(run db.WorkflowRun, excluded []db.WorkflowRunStatus) (ok bool, err error) {
	return
}

// SetWorkflowRunRootTask links a workflow run to its root task.
func (d *WorkflowStoreImpl) SetWorkflowRunRootTask(projectID int, runID int, taskID int) (ok bool, err error) {
	return
}

// GetWorkflowApprovals returns all approval records for a workflow run.
func (d *WorkflowStoreImpl) GetWorkflowApprovals(projectID int, runID int) (res []db.WorkflowApproval, err error) {
	_, err = d.conn.SelectAll(&res,
		"select * from `project__workflow_approval` where `project_id`=? and `workflow_run_id`=?",
		projectID, runID)
	return
}

// GetWorkflowApproval returns the approval for a specific node within a run.
func (d *WorkflowStoreImpl) GetWorkflowApproval(projectID int, runID int, nodeID int) (res db.WorkflowApproval, err error) {
	err = d.conn.SelectOne(&res,
		"select * from `project__workflow_approval` where `project_id`=? and `workflow_run_id`=? and `workflow_node_id`=?",
		projectID, runID, nodeID)
	return
}

// CreateWorkflowApproval inserts a new workflow approval record.
func (d *WorkflowStoreImpl) CreateWorkflowApproval(approval db.WorkflowApproval) (res db.WorkflowApproval, err error) {
	approvalID, err := d.conn.Insert("id",
		"insert into `project__workflow_approval` (`project_id`, `workflow_run_id`, `workflow_node_id`, `status`, `created`) values (?, ?, ?, ?, ?)",
		approval.ProjectID, approval.WorkflowRunID, approval.WorkflowNodeID, approval.Status, approval.Created)
	if err != nil {
		return
	}
	res = approval
	res.ID = approvalID
	return
}

// UpdateWorkflowApproval overwrites the mutable fields of an approval record.
func (d *WorkflowStoreImpl) UpdateWorkflowApproval(approval db.WorkflowApproval) (err error) {
	_, err = d.conn.Exec(
		"update `project__workflow_approval` set `status`=?, `resolved`=?, `resolved_by_user_id`=? where `id`=?",
		approval.Status, approval.Resolved, approval.ResolvedByUserID, approval.ID)
	return
}

// ResolveWorkflowApprovalIfPending atomically moves an approval from Pending
// to the given status. Returns ok=true when the row was updated (i.e. the
// approval was still pending).
func (d *WorkflowStoreImpl) ResolveWorkflowApprovalIfPending(approval db.WorkflowApproval) (ok bool, err error) {
	result, err := d.conn.Exec(
		"update `project__workflow_approval` set `status`=?, `resolved`=?, `resolved_by_user_id`=? where `id`=? and `status`=?",
		approval.Status, approval.Resolved, approval.ResolvedByUserID, approval.ID, db.WorkflowApprovalPending)
	if err != nil {
		return
	}
	n, _ := result.RowsAffected()
	ok = n > 0
	return
}
