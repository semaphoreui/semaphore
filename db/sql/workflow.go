package sql

import (
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/tz"
)

// Workflow CRUD operations

func (d *SqlDb) GetWorkflows(projectID int) ([]db.Workflow, error) {
	var workflows []db.Workflow
	err := d.selectAll(
		&workflows,
		"select * from project__workflow where project_id=? order by name",
		projectID,
	)
	return workflows, err
}

func (d *SqlDb) GetWorkflow(projectID int, workflowID int) (db.Workflow, error) {
	var workflow db.Workflow
	var err error
	if projectID > 0 {
		err = d.selectOne(
			&workflow,
			"select * from project__workflow where id=? and project_id=?",
			workflowID,
			projectID,
		)
	} else {
		err = d.selectOne(
			&workflow,
			"select * from project__workflow where id=?",
			workflowID,
		)
	}
	return workflow, err
}

func (d *SqlDb) CreateWorkflow(workflow db.Workflow) (db.Workflow, error) {
	now := tz.Now()
	workflow.Created = now
	workflow.Updated = now

	insertID, err := d.insert(
		"id",
		"insert into project__workflow (project_id, name, description, created, updated) values (?, ?, ?, ?, ?)",
		workflow.ProjectID,
		workflow.Name,
		workflow.Description,
		workflow.Created,
		workflow.Updated,
	)

	if err != nil {
		return workflow, err
	}

	workflow.ID = insertID
	return workflow, nil
}

func (d *SqlDb) UpdateWorkflow(workflow db.Workflow) error {
	now := tz.Now()
	workflow.Updated = now

	_, err := d.exec(
		"update project__workflow set name=?, description=?, updated=? where id=? and project_id=?",
		workflow.Name,
		workflow.Description,
		workflow.Updated,
		workflow.ID,
		workflow.ProjectID,
	)
	return err
}

func (d *SqlDb) DeleteWorkflow(projectID int, workflowID int) error {
	_, err := d.exec(
		"delete from project__workflow where id=? and project_id=?",
		workflowID,
		projectID,
	)
	return err
}

// WorkflowNode operations

func (d *SqlDb) GetWorkflowNodes(workflowID int) ([]db.WorkflowNode, error) {
	var nodes []db.WorkflowNode
	err := d.selectAll(
		&nodes,
		"select * from project__workflow_node where workflow_id=?",
		workflowID,
	)
	return nodes, err
}

func (d *SqlDb) CreateWorkflowNode(node db.WorkflowNode) (db.WorkflowNode, error) {
	insertID, err := d.insert(
		"id",
		"insert into project__workflow_node (workflow_id, task_id, type, position_x, position_y, config_json) values (?, ?, ?, ?, ?, ?)",
		node.WorkflowID,
		node.TaskID,
		node.Type,
		node.PositionX,
		node.PositionY,
		node.ConfigJSON,
	)

	if err != nil {
		return node, err
	}

	node.ID = insertID
	return node, nil
}

func (d *SqlDb) UpdateWorkflowNode(node db.WorkflowNode) error {
	_, err := d.exec(
		"update project__workflow_node set task_id=?, type=?, position_x=?, position_y=?, config_json=? where id=? and workflow_id=?",
		node.TaskID,
		node.Type,
		node.PositionX,
		node.PositionY,
		node.ConfigJSON,
		node.ID,
		node.WorkflowID,
	)
	return err
}

func (d *SqlDb) DeleteWorkflowNode(workflowID int, nodeID int) error {
	_, err := d.exec(
		"delete from project__workflow_node where id=? and workflow_id=?",
		nodeID,
		workflowID,
	)
	return err
}

// WorkflowLink operations

func (d *SqlDb) GetWorkflowLinks(workflowID int) ([]db.WorkflowLink, error) {
	var links []db.WorkflowLink
	err := d.selectAll(
		&links,
		"select * from project__workflow_link where workflow_id=?",
		workflowID,
	)
	return links, err
}

func (d *SqlDb) CreateWorkflowLink(link db.WorkflowLink) (db.WorkflowLink, error) {
	insertID, err := d.insert(
		"id",
		"insert into project__workflow_link (workflow_id, from_node_id, to_node_id, condition) values (?, ?, ?, ?)",
		link.WorkflowID,
		link.FromNodeID,
		link.ToNodeID,
		link.Condition,
	)

	if err != nil {
		return link, err
	}

	link.ID = insertID
	return link, nil
}

func (d *SqlDb) DeleteWorkflowLink(workflowID int, linkID int) error {
	_, err := d.exec(
		"delete from project__workflow_link where id=? and workflow_id=?",
		linkID,
		workflowID,
	)
	return err
}

// WorkflowRun operations

func (d *SqlDb) GetWorkflowRuns(workflowID int, params db.RetrieveQueryParams) ([]db.WorkflowRun, error) {
	q := squirrel.Select("*").
		From("project__workflow_run").
		Where("workflow_id=?", workflowID).
		OrderBy("created desc")

	if params.Count > 0 {
		q = q.Limit(uint64(params.Count))
	}
	if params.Offset > 0 {
		q = q.Offset(uint64(params.Offset))
	}

	query, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	var runs []db.WorkflowRun
	err = d.selectAll(&runs, query, args...)
	return runs, err
}

func (d *SqlDb) GetWorkflowRun(workflowID int, runID int) (db.WorkflowRun, error) {
	var run db.WorkflowRun
	err := d.selectOne(
		&run,
		"select * from project__workflow_run where id=? and workflow_id=?",
		runID,
		workflowID,
	)
	return run, err
}

func (d *SqlDb) CreateWorkflowRun(run db.WorkflowRun) (db.WorkflowRun, error) {
	now := tz.Now()
	run.Created = now

	insertID, err := d.insert(
		"id",
		"insert into project__workflow_run (workflow_id, status, user_id, created, start, end, message) values (?, ?, ?, ?, ?, ?, ?)",
		run.WorkflowID,
		run.Status,
		run.UserID,
		run.Created,
		run.Start,
		run.End,
		run.Message,
	)

	if err != nil {
		return run, err
	}

	run.ID = insertID
	return run, nil
}

func (d *SqlDb) UpdateWorkflowRun(run db.WorkflowRun) error {
	_, err := d.exec(
		"update project__workflow_run set status=?, start=?, end=?, message=? where id=? and workflow_id=?",
		run.Status,
		run.Start,
		run.End,
		run.Message,
		run.ID,
		run.WorkflowID,
	)
	return err
}

// WorkflowRunNode operations

func (d *SqlDb) GetWorkflowRunNodes(runID int) ([]db.WorkflowRunNode, error) {
	var nodes []db.WorkflowRunNode
	err := d.selectAll(
		&nodes,
		"select * from project__workflow_run_node where workflow_run_id=? order by created",
		runID,
	)
	return nodes, err
}

func (d *SqlDb) CreateWorkflowRunNode(node db.WorkflowRunNode) (db.WorkflowRunNode, error) {
	now := tz.Now()
	node.Created = now

	insertID, err := d.insert(
		"id",
		"insert into project__workflow_run_node (workflow_run_id, workflow_node_id, task_id, status, created, start, end, message) values (?, ?, ?, ?, ?, ?, ?, ?)",
		node.WorkflowRunID,
		node.WorkflowNodeID,
		node.TaskID,
		node.Status,
		node.Created,
		node.Start,
		node.End,
		node.Message,
	)

	if err != nil {
		return node, err
	}

	node.ID = insertID
	return node, nil
}

func (d *SqlDb) UpdateWorkflowRunNode(node db.WorkflowRunNode) error {
	_, err := d.exec(
		"update project__workflow_run_node set status=?, task_id=?, start=?, end=?, message=? where id=? and workflow_run_id=?",
		node.Status,
		node.TaskID,
		node.Start,
		node.End,
		node.Message,
		node.ID,
		node.WorkflowRunID,
	)
	return err
}
