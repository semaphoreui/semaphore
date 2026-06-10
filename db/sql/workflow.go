package sql

import (
	"github.com/Masterminds/squirrel"
	"github.com/semaphoreui/semaphore/db"
)

func (d *SqlDb) getWorkflowNodes(workflowID int) (nodes []db.WorkflowNode, err error) {
	nodes = make([]db.WorkflowNode, 0)
	_, err = d.selectAll(&nodes, "select * from project__workflow_node where workflow_template_id=? order by id", workflowID)
	return
}

func (d *SqlDb) getWorkflowEdges(workflowID int) (edges []db.WorkflowEdge, err error) {
	edges = make([]db.WorkflowEdge, 0)
	_, err = d.selectAll(&edges, "select * from project__workflow_edge where workflow_template_id=? order by id", workflowID)
	return
}

func (d *SqlDb) fillWorkflow(workflow *db.WorkflowTemplate) (err error) {
	workflow.Nodes, err = d.getWorkflowNodes(workflow.ID)
	if err != nil {
		return
	}

	workflow.Edges, err = d.getWorkflowEdges(workflow.ID)
	return
}

func (d *SqlDb) writeWorkflowGraph(workflow db.WorkflowTemplate) (err error) {
	_, err = d.exec("delete from project__workflow_edge where workflow_template_id=?", workflow.ID)
	if err != nil {
		return
	}

	_, err = d.exec("delete from project__workflow_node where workflow_template_id=?", workflow.ID)
	if err != nil {
		return
	}

	nodeIDMap := make(map[int]int)
	for i, node := range workflow.Nodes {
		node.WorkflowTemplateID = workflow.ID

		insertID, err2 := d.insert(
			"id",
			"insert into project__workflow_node (workflow_template_id, template_id, inventory_id, environment_id, kind, convergence_mode, approval_timeout, approval_message, note, `limit`, position_x, position_y) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			node.WorkflowTemplateID,
			node.TemplateID,
			node.InventoryID,
			node.EnvironmentID,
			node.EffectiveKind(),
			node.EffectiveConvergenceMode(),
			node.ApprovalTimeout,
			node.ApprovalMessage,
			node.Note,
			&node.Limit,
			node.PositionX,
			node.PositionY,
		)
		if err2 != nil {
			err = err2
			return
		}

		var effectiveID int
		if node.ID != 0 {
			effectiveID = node.ID
		} else {
			effectiveID = -(i + 1)
		}
		nodeIDMap[effectiveID] = insertID
	}

	for _, edge := range workflow.Edges {
		sourceNodeID := edge.SourceNodeID
		destinationNodeID := edge.DestinationNodeID

		mappedSourceNodeID, ok := nodeIDMap[sourceNodeID]
		if !ok {
			return db.NewValidationError("workflow edge source node does not belong to workflow")
		}
		mappedDestinationNodeID, ok := nodeIDMap[destinationNodeID]
		if !ok {
			return db.NewValidationError("workflow edge destination node does not belong to workflow")
		}

		_, err = d.insert(
			"id",
			"insert into project__workflow_edge (workflow_template_id, source_node_id, destination_node_id, condition) values (?, ?, ?, ?)",
			workflow.ID,
			mappedSourceNodeID,
			mappedDestinationNodeID,
			edge.Condition,
		)
		if err != nil {
			return
		}
	}

	return
}

func (d *SqlDb) GetWorkflowTemplates(projectID int, params db.RetrieveQueryParams) (workflows []db.WorkflowTemplate, err error) {
	pp, err := params.Validate(db.WorkflowTemplateProps)
	if err != nil {
		return
	}

	workflows = make([]db.WorkflowTemplate, 0)
	q := squirrel.Select("*").From("project__workflow_template").Where("project_id=?", projectID)

	if pp.Count > 0 {
		q = q.Limit(uint64(pp.Count))
	}
	if pp.Offset > 0 {
		q = q.Offset(uint64(pp.Offset))
	}

	q = q.OrderBy("id")

	query, args, err := q.ToSql()
	if err != nil {
		return
	}

	_, err = d.selectAll(&workflows, query, args...)
	if err != nil {
		return
	}

	for i := range workflows {
		err = d.fillWorkflow(&workflows[i])
		if err != nil {
			return
		}

		// Attach the most recent run (ordered by id desc) so the list can show
		// last status/version, mirroring templates' last_task.
		var runs []db.WorkflowRun
		runs, err = d.GetWorkflowRuns(projectID, workflows[i].ID, db.RetrieveQueryParams{Count: 1})
		if err != nil {
			return
		}
		if len(runs) > 0 {
			workflows[i].LastRun = &runs[0]
		}
	}

	return
}

func (d *SqlDb) GetWorkflowTemplate(projectID int, workflowID int) (workflow db.WorkflowTemplate, err error) {
	err = d.selectOne(&workflow, "select * from project__workflow_template where project_id=? and id=?", projectID, workflowID)
	if err != nil {
		return
	}

	err = d.fillWorkflow(&workflow)
	return
}

func (d *SqlDb) CreateWorkflowTemplate(workflow db.WorkflowTemplate) (newWorkflow db.WorkflowTemplate, err error) {
	err = db.ValidateWorkflowTemplate(d, workflow)
	if err != nil {
		return
	}

	insertID, err := d.insert(
		"id",
		"insert into project__workflow_template (project_id, name, description, start_version) values (?, ?, ?, ?)",
		workflow.ProjectID,
		workflow.Name,
		workflow.Description,
		workflow.StartVersion,
	)
	if err != nil {
		return
	}

	workflow.ID = insertID
	err = d.writeWorkflowGraph(workflow)
	if err != nil {
		return
	}

	newWorkflow, err = d.GetWorkflowTemplate(workflow.ProjectID, workflow.ID)
	return
}

func (d *SqlDb) UpdateWorkflowTemplate(workflow db.WorkflowTemplate) (err error) {
	err = db.ValidateWorkflowTemplate(d, workflow)
	if err != nil {
		return
	}

	_, err = d.exec(
		"update project__workflow_template set name=?, description=?, start_version=? where project_id=? and id=?",
		workflow.Name,
		workflow.Description,
		workflow.StartVersion,
		workflow.ProjectID,
		workflow.ID,
	)
	if err != nil {
		return
	}

	return d.writeWorkflowGraph(workflow)
}

func (d *SqlDb) DeleteWorkflowTemplate(projectID int, workflowID int) error {
	_, err := d.exec("delete from project__workflow_template where project_id=? and id=?", projectID, workflowID)
	return err
}

func (d *SqlDb) GetWorkflowRuns(projectID int, workflowTemplateID int, params db.RetrieveQueryParams) (runs []db.WorkflowRun, err error) {
	runs = make([]db.WorkflowRun, 0)
	q := squirrel.Select("*").
		From("project__workflow_run").
		Where("project_id=? and workflow_template_id=?", projectID, workflowTemplateID).
		OrderBy("id desc")

	if params.Count > 0 {
		q = q.Limit(uint64(params.Count))
	}
	if params.Offset > 0 {
		q = q.Offset(uint64(params.Offset))
	}

	query, args, err := q.ToSql()
	if err != nil {
		return
	}

	_, err = d.selectAll(&runs, query, args...)
	return
}

func (d *SqlDb) GetWorkflowRun(projectID int, workflowTemplateID int, runID int) (run db.WorkflowRun, err error) {
	err = d.selectOne(
		&run,
		"select * from project__workflow_run where project_id=? and workflow_template_id=? and id=?",
		projectID,
		workflowTemplateID,
		runID,
	)
	return
}

func (d *SqlDb) GetWorkflowRunByID(projectID int, runID int) (run db.WorkflowRun, err error) {
	err = d.selectOne(&run, "select * from project__workflow_run where project_id=? and id=?", projectID, runID)
	return
}

func (d *SqlDb) CreateWorkflowRun(run db.WorkflowRun) (newRun db.WorkflowRun, err error) {
	insertID, err := d.insert(
		"id",
		"insert into project__workflow_run (project_id, workflow_template_id, status, version, start, `end`, root_task_id) values (?, ?, ?, ?, ?, ?, ?)",
		run.ProjectID,
		run.WorkflowTemplateID,
		run.Status,
		run.Version,
		run.Start,
		run.End,
		run.RootTaskID,
	)
	if err != nil {
		return
	}

	run.ID = insertID
	newRun = run
	return
}

func (d *SqlDb) UpdateWorkflowRun(run db.WorkflowRun) error {
	_, err := d.exec(
		"update project__workflow_run set status=?, start=?, `end`=?, root_task_id=? where project_id=? and workflow_template_id=? and id=?",
		run.Status,
		run.Start,
		run.End,
		run.RootTaskID,
		run.ProjectID,
		run.WorkflowTemplateID,
		run.ID,
	)
	return err
}

func (d *SqlDb) GetActiveWorkflowRuns() (runs []db.WorkflowRun, err error) {
	runs = make([]db.WorkflowRun, 0)
	_, err = d.selectAll(
		&runs,
		"select * from project__workflow_run where status in (?, ?) order by id",
		db.WorkflowRunRunning,
		db.WorkflowRunApproval,
	)
	return
}

func (d *SqlDb) UpdateWorkflowRunStatusUnless(run db.WorkflowRun, excluded []db.WorkflowRunStatus) (bool, error) {
	q := squirrel.Update("project__workflow_run").
		Set("status", run.Status).
		Set("`end`", run.End).
		Where("project_id=? and id=?", run.ProjectID, run.ID)

	if len(excluded) > 0 {
		q = q.Where(squirrel.NotEq{"status": excluded})
	}

	query, args, err := q.ToSql()
	if err != nil {
		return false, err
	}

	res, err := d.exec(query, args...)
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	return affected > 0, err
}

func (d *SqlDb) SetWorkflowRunRootTask(projectID int, runID int, taskID int) (bool, error) {
	res, err := d.exec(
		"update project__workflow_run set root_task_id=? where project_id=? and id=? and root_task_id is null",
		taskID,
		projectID,
		runID,
	)
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	return affected > 0, err
}

func (d *SqlDb) GetWorkflowApprovals(projectID int, runID int) (approvals []db.WorkflowApproval, err error) {
	approvals = make([]db.WorkflowApproval, 0)
	_, err = d.selectAll(
		&approvals,
		"select * from project__workflow_approval where project_id=? and workflow_run_id=? order by id",
		projectID,
		runID,
	)
	return
}

func (d *SqlDb) GetWorkflowApproval(projectID int, runID int, nodeID int) (approval db.WorkflowApproval, err error) {
	err = d.selectOne(
		&approval,
		"select * from project__workflow_approval where project_id=? and workflow_run_id=? and workflow_node_id=?",
		projectID,
		runID,
		nodeID,
	)
	return
}

func (d *SqlDb) CreateWorkflowApproval(approval db.WorkflowApproval) (newApproval db.WorkflowApproval, err error) {
	insertID, err := d.insert(
		"id",
		"insert into project__workflow_approval (project_id, workflow_run_id, workflow_node_id, status, created, resolved, resolved_by_user_id) values (?, ?, ?, ?, ?, ?, ?)",
		approval.ProjectID,
		approval.WorkflowRunID,
		approval.WorkflowNodeID,
		approval.Status,
		approval.Created,
		approval.Resolved,
		approval.ResolvedByUserID,
	)
	if err != nil {
		return
	}

	approval.ID = insertID
	newApproval = approval
	return
}

func (d *SqlDb) UpdateWorkflowApproval(approval db.WorkflowApproval) error {
	_, err := d.exec(
		"update project__workflow_approval set status=?, resolved=?, resolved_by_user_id=? where project_id=? and workflow_run_id=? and workflow_node_id=? and id=?",
		approval.Status,
		approval.Resolved,
		approval.ResolvedByUserID,
		approval.ProjectID,
		approval.WorkflowRunID,
		approval.WorkflowNodeID,
		approval.ID,
	)
	return err
}

func (d *SqlDb) ResolveWorkflowApprovalIfPending(approval db.WorkflowApproval) (bool, error) {
	res, err := d.exec(
		"update project__workflow_approval set status=?, resolved=?, resolved_by_user_id=? "+
			"where project_id=? and workflow_run_id=? and workflow_node_id=? and id=? and status=?",
		approval.Status,
		approval.Resolved,
		approval.ResolvedByUserID,
		approval.ProjectID,
		approval.WorkflowRunID,
		approval.WorkflowNodeID,
		approval.ID,
		db.WorkflowApprovalPending,
	)
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	return affected > 0, err
}
