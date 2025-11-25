package sql

import (
	"fmt"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/tz"
)

// GetWorkflows retrieves all workflows for a project
func (d *SqlDb) GetWorkflows(projectID int, params db.RetrieveQueryParams) (workflows []db.Workflow, err error) {
	query := "select * from project__workflow where project_id=?"
	order := "order by name"

	if params.SortBy != "" {
		order = fmt.Sprintf("order by %s", params.SortBy)
		if params.SortInverted {
			order += " desc"
		}
	}

	query += " " + order

	if params.Count > 0 {
		query += fmt.Sprintf(" limit %d", params.Count)
	}

	if params.Offset > 0 {
		query += fmt.Sprintf(" offset %d", params.Offset)
	}

	_, err = d.selectAll(&workflows, query, projectID)
	return
}

// GetWorkflow retrieves a specific workflow with its nodes and links
func (d *SqlDb) GetWorkflow(projectID int, workflowID int) (workflow db.Workflow, err error) {
	err = d.selectOne(&workflow, "select * from project__workflow where project_id=? and id=?", projectID, workflowID)

	if err != nil {
		return
	}

	// Load nodes
	workflow.Nodes, err = d.GetWorkflowNodes(workflowID)
	if err != nil {
		return
	}

	// Load links
	workflow.Links, err = d.GetWorkflowLinks(workflowID)

	return
}

// CreateWorkflow creates a new workflow
func (d *SqlDb) CreateWorkflow(workflow db.Workflow) (newWorkflow db.Workflow, err error) {
	err = workflow.Validate()
	if err != nil {
		return
	}

	now := tz.Now()
	workflow.CreatedAt = now
	workflow.UpdatedAt = now

	insertID, err := d.insert("id", "insert into project__workflow (project_id, name, description, created_at, updated_at) values (?, ?, ?, ?, ?)",
		workflow.ProjectID,
		workflow.Name,
		workflow.Description,
		workflow.CreatedAt,
		workflow.UpdatedAt,
	)

	if err != nil {
		return
	}

	newWorkflow = workflow
	newWorkflow.ID = insertID
	return
}

// UpdateWorkflow updates an existing workflow
func (d *SqlDb) UpdateWorkflow(workflow db.Workflow) error {
	err := workflow.Validate()
	if err != nil {
		return err
	}

	workflow.UpdatedAt = tz.Now()

	_, err = d.exec("update project__workflow set name=?, description=?, updated_at=? where id=?",
		workflow.Name,
		workflow.Description,
		workflow.UpdatedAt,
		workflow.ID,
	)

	return err
}

// DeleteWorkflow deletes a workflow
func (d *SqlDb) DeleteWorkflow(projectID int, workflowID int) error {
	_, err := d.exec("delete from project__workflow where project_id=? and id=?", projectID, workflowID)
	return err
}

// GetWorkflowNodes retrieves all nodes for a workflow
func (d *SqlDb) GetWorkflowNodes(workflowID int) (nodes []db.WorkflowNode, err error) {
	_, err = d.selectAll(&nodes, "select * from project__workflow_node where workflow_id=? order by id", workflowID)
	return
}

// GetWorkflowNode retrieves a specific workflow node
func (d *SqlDb) GetWorkflowNode(workflowID int, nodeID int) (node db.WorkflowNode, err error) {
	err = d.selectOne(&node, "select * from project__workflow_node where workflow_id=? and id=?", workflowID, nodeID)
	return
}

// CreateWorkflowNode creates a new workflow node
func (d *SqlDb) CreateWorkflowNode(node db.WorkflowNode) (newNode db.WorkflowNode, err error) {
	err = node.Validate()
	if err != nil {
		return
	}

	insertID, err := d.insert("id", "insert into project__workflow_node (workflow_id, task_template_id, type, name, position_x, position_y, config) values (?, ?, ?, ?, ?, ?, ?)",
		node.WorkflowID,
		node.TaskTemplateID,
		node.Type,
		node.Name,
		node.PositionX,
		node.PositionY,
		node.Config,
	)

	if err != nil {
		return
	}

	newNode = node
	newNode.ID = insertID
	return
}

// UpdateWorkflowNode updates an existing workflow node
func (d *SqlDb) UpdateWorkflowNode(node db.WorkflowNode) error {
	err := node.Validate()
	if err != nil {
		return err
	}

	_, err = d.exec("update project__workflow_node set task_template_id=?, type=?, name=?, position_x=?, position_y=?, config=? where id=?",
		node.TaskTemplateID,
		node.Type,
		node.Name,
		node.PositionX,
		node.PositionY,
		node.Config,
		node.ID,
	)

	return err
}

// DeleteWorkflowNode deletes a workflow node
func (d *SqlDb) DeleteWorkflowNode(workflowID int, nodeID int) error {
	_, err := d.exec("delete from project__workflow_node where workflow_id=? and id=?", workflowID, nodeID)
	return err
}

// GetWorkflowLinks retrieves all links for a workflow
func (d *SqlDb) GetWorkflowLinks(workflowID int) (links []db.WorkflowLink, err error) {
	_, err = d.selectAll(&links, "select * from project__workflow_link where workflow_id=? order by id", workflowID)
	return
}

// CreateWorkflowLink creates a new workflow link
func (d *SqlDb) CreateWorkflowLink(link db.WorkflowLink) (newLink db.WorkflowLink, err error) {
	insertID, err := d.insert("id", "insert into project__workflow_link (workflow_id, from_node_id, to_node_id, condition) values (?, ?, ?, ?)",
		link.WorkflowID,
		link.FromNodeID,
		link.ToNodeID,
		link.Condition,
	)

	if err != nil {
		return
	}

	newLink = link
	newLink.ID = insertID
	return
}

// DeleteWorkflowLink deletes a workflow link
func (d *SqlDb) DeleteWorkflowLink(workflowID int, linkID int) error {
	_, err := d.exec("delete from project__workflow_link where workflow_id=? and id=?", workflowID, linkID)
	return err
}

// GetWorkflowRuns retrieves all runs for a workflow
func (d *SqlDb) GetWorkflowRuns(workflowID int, params db.RetrieveQueryParams) (runs []db.WorkflowRunWithWorkflow, err error) {
	query := `select wr.*, w.name as workflow_name 
		from project__workflow_run wr 
		join project__workflow w on wr.workflow_id = w.id 
		where wr.workflow_id=? 
		order by wr.id desc`

	if params.Count > 0 {
		query += fmt.Sprintf(" limit %d", params.Count)
	}

	if params.Offset > 0 {
		query += fmt.Sprintf(" offset %d", params.Offset)
	}

	_, err = d.selectAll(&runs, query, workflowID)
	return
}

// GetProjectWorkflowRuns retrieves all workflow runs for a project
func (d *SqlDb) GetProjectWorkflowRuns(projectID int, params db.RetrieveQueryParams) (runs []db.WorkflowRunWithWorkflow, err error) {
	query := `select wr.*, w.name as workflow_name 
		from project__workflow_run wr 
		join project__workflow w on wr.workflow_id = w.id 
		where wr.project_id=? 
		order by wr.id desc`

	if params.Count > 0 {
		query += fmt.Sprintf(" limit %d", params.Count)
	}

	if params.Offset > 0 {
		query += fmt.Sprintf(" offset %d", params.Offset)
	}

	_, err = d.selectAll(&runs, query, projectID)
	return
}

// GetWorkflowRun retrieves a specific workflow run
func (d *SqlDb) GetWorkflowRun(workflowRunID int) (run db.WorkflowRun, err error) {
	err = d.selectOne(&run, "select * from project__workflow_run where id=?", workflowRunID)
	return
}

// CreateWorkflowRun creates a new workflow run
func (d *SqlDb) CreateWorkflowRun(run db.WorkflowRun) (newRun db.WorkflowRun, err error) {
	now := tz.Now()
	run.Start = &now

	insertID, err := d.insert("id", "insert into project__workflow_run (workflow_id, project_id, user_id, status, start, message) values (?, ?, ?, ?, ?, ?)",
		run.WorkflowID,
		run.ProjectID,
		run.UserID,
		run.Status,
		run.Start,
		run.Message,
	)

	if err != nil {
		return
	}

	newRun = run
	newRun.ID = insertID
	return
}

// UpdateWorkflowRun updates an existing workflow run
func (d *SqlDb) UpdateWorkflowRun(run db.WorkflowRun) error {
	_, err := d.exec("update project__workflow_run set status=?, end=?, message=? where id=?",
		run.Status,
		run.End,
		run.Message,
		run.ID,
	)

	return err
}

// DeleteWorkflowRun deletes a workflow run
func (d *SqlDb) DeleteWorkflowRun(workflowRunID int) error {
	_, err := d.exec("delete from project__workflow_run where id=?", workflowRunID)
	return err
}

// GetWorkflowNodeRuns retrieves all node runs for a workflow run
func (d *SqlDb) GetWorkflowNodeRuns(workflowRunID int) (nodeRuns []db.WorkflowNodeRun, err error) {
	_, err = d.selectAll(&nodeRuns, "select * from project__workflow_node_run where workflow_run_id=? order by id", workflowRunID)
	return
}

// GetWorkflowNodeRun retrieves a specific workflow node run
func (d *SqlDb) GetWorkflowNodeRun(nodeRunID int) (nodeRun db.WorkflowNodeRun, err error) {
	err = d.selectOne(&nodeRun, "select * from project__workflow_node_run where id=?", nodeRunID)
	return
}

// CreateWorkflowNodeRun creates a new workflow node run
func (d *SqlDb) CreateWorkflowNodeRun(nodeRun db.WorkflowNodeRun) (newNodeRun db.WorkflowNodeRun, err error) {
	now := tz.Now()
	nodeRun.Start = &now

	insertID, err := d.insert("id", "insert into project__workflow_node_run (workflow_run_id, node_id, task_id, status, start, message) values (?, ?, ?, ?, ?, ?)",
		nodeRun.WorkflowRunID,
		nodeRun.NodeID,
		nodeRun.TaskID,
		nodeRun.Status,
		nodeRun.Start,
		nodeRun.Message,
	)

	if err != nil {
		return
	}

	newNodeRun = nodeRun
	newNodeRun.ID = insertID
	return
}

// UpdateWorkflowNodeRun updates an existing workflow node run
func (d *SqlDb) UpdateWorkflowNodeRun(nodeRun db.WorkflowNodeRun) error {
	_, err := d.exec("update project__workflow_node_run set task_id=?, status=?, end=?, message=? where id=?",
		nodeRun.TaskID,
		nodeRun.Status,
		nodeRun.End,
		nodeRun.Message,
		nodeRun.ID,
	)

	return err
}
