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
			"insert into project__workflow_node (workflow_template_id, template_id, inventory_id, environment_id) values (?, ?, ?, ?)",
			node.WorkflowTemplateID,
			node.TemplateID,
			node.InventoryID,
			node.EnvironmentID,
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
		"insert into project__workflow_template (project_id, name, description) values (?, ?, ?)",
		workflow.ProjectID,
		workflow.Name,
		workflow.Description,
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
		"update project__workflow_template set name=?, description=? where project_id=? and id=?",
		workflow.Name,
		workflow.Description,
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
		"insert into project__workflow_run (project_id, workflow_template_id, status, start, end, root_task_id) values (?, ?, ?, ?, ?, ?)",
		run.ProjectID,
		run.WorkflowTemplateID,
		run.Status,
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
		"update project__workflow_run set status=?, start=?, end=?, root_task_id=? where project_id=? and workflow_template_id=? and id=?",
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
