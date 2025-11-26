package sql

import (
	"github.com/Masterminds/squirrel"
	"github.com/semaphoreui/semaphore/db"
)

func (d *SqlDb) GetWorkflows(projectID int, params db.RetrieveQueryParams) ([]db.Workflow, error) {
	var workflows []db.Workflow
	err := d.getObjects(projectID, db.WorkflowProps, params, nil, &workflows)
	return workflows, err
}

func (d *SqlDb) GetWorkflow(projectID int, workflowID int) (db.Workflow, error) {
	var workflow db.Workflow
	err := d.getObject(projectID, db.WorkflowProps, workflowID, &workflow)
	return workflow, err
}

func (d *SqlDb) CreateWorkflow(workflow db.Workflow) (db.Workflow, error) {
	newWorkflow, err := d.CreateObject(db.WorkflowProps, workflow)
	if err != nil {
		return db.Workflow{}, err
	}
	return newWorkflow.(db.Workflow), nil
}

func (d *SqlDb) UpdateWorkflow(workflow db.Workflow) error {
	_, err := d.exec(
		"update workflow set name=?, description=?, updated_at=? where id=? and project_id=?",
		workflow.Name,
		workflow.Description,
		workflow.UpdatedAt,
		workflow.ID,
		workflow.ProjectID)
	return err
}

func (d *SqlDb) DeleteWorkflow(projectID int, workflowID int) error {
	return d.deleteObject(projectID, db.WorkflowProps, workflowID)
}

func (d *SqlDb) GetWorkflowNodes(projectID int, workflowID int) ([]db.WorkflowNode, error) {
	var nodes []db.WorkflowNode
    
    query, args, err := squirrel.Select("*").
        From("workflow_node").
        Where("workflow_id = ?", workflowID).
        ToSql()

    if err != nil {
        return nil, err
    }
    
    _, err = d.selectAll(&nodes, query, args...)
	return nodes, err
}

func (d *SqlDb) CreateWorkflowNode(node db.WorkflowNode) (db.WorkflowNode, error) {
	newNode, err := d.CreateObject(db.WorkflowNodeProps, node)
	if err != nil {
		return db.WorkflowNode{}, err
	}
	return newNode.(db.WorkflowNode), nil
}

func (d *SqlDb) UpdateWorkflowNode(node db.WorkflowNode) error {
     _, err := d.exec("update workflow_node set project_template_id=?, type=?, position_x=?, position_y=?, config_json=? where id=?",
        node.ProjectTemplateID, node.Type, node.PositionX, node.PositionY, node.ConfigJSON, node.ID)
    return err
}

func (d *SqlDb) DeleteWorkflowNode(projectID int, nodeID int) error {
    _, err := d.exec("delete from workflow_node where id=?", nodeID)
    return err
}

func (d *SqlDb) DeleteWorkflowNodes(projectID int, workflowID int) error {
    _, err := d.exec("delete from workflow_node where workflow_id=?", workflowID)
    return err
}

func (d *SqlDb) GetWorkflowLinks(projectID int, workflowID int) ([]db.WorkflowLink, error) {
	var links []db.WorkflowLink
    query, args, err := squirrel.Select("*").
        From("workflow_link").
        Where("workflow_id = ?", workflowID).
        ToSql()

    if err != nil {
        return nil, err
    }
    
    _, err = d.selectAll(&links, query, args...)
	return links, err
}

func (d *SqlDb) CreateWorkflowLink(link db.WorkflowLink) (db.WorkflowLink, error) {
	newLink, err := d.CreateObject(db.WorkflowLinkProps, link)
	if err != nil {
		return db.WorkflowLink{}, err
	}
	return newLink.(db.WorkflowLink), nil
}

func (d *SqlDb) DeleteWorkflowLinks(projectID int, workflowID int) error {
    _, err := d.exec("delete from workflow_link where workflow_id=?", workflowID)
    return err
}


func (d *SqlDb) CreateWorkflowRun(run db.WorkflowRun) (db.WorkflowRun, error) {
	newRun, err := d.CreateObject(db.WorkflowRunProps, run)
	if err != nil {
		return db.WorkflowRun{}, err
	}
	return newRun.(db.WorkflowRun), nil
}

func (d *SqlDb) GetWorkflowRun(projectID int, runID int) (db.WorkflowRun, error) {
    var run db.WorkflowRun
    query, args, err := squirrel.Select("r.*").
        From("workflow_run r").
        Join("workflow w on r.workflow_id = w.id").
        Where("r.id = ? AND w.project_id = ?", runID, projectID).
        ToSql()
    
    if err != nil {
        return db.WorkflowRun{}, err
    }
    
    err = d.selectOne(&run, query, args...)
	return run, err
}

func (d *SqlDb) GetWorkflowRuns(projectID int, workflowID *int, params db.RetrieveQueryParams) ([]db.WorkflowRun, error) {
    var runs []db.WorkflowRun
    
    q := squirrel.Select("r.*").
        From("workflow_run r").
        Join("workflow w on r.workflow_id = w.id").
        Where("w.project_id = ?", projectID)

    if workflowID != nil {
        q = q.Where("r.workflow_id = ?", *workflowID)
    }

    if params.Count > 0 {
        q = q.Limit(uint64(params.Count))
    }
    if params.Offset > 0 {
        q = q.Offset(uint64(params.Offset))
    }
    
    q = q.OrderBy("r.created_at DESC")

    query, args, err := q.ToSql()
    if err != nil {
        return nil, err
    }

    _, err = d.selectAll(&runs, query, args...)
    return runs, err
}

func (d *SqlDb) UpdateWorkflowRun(run db.WorkflowRun) error {
    _, err := d.exec("update workflow_run set status=?, finished_at=? where id=?", run.Status, run.FinishedAt, run.ID)
    return err
}


func (d *SqlDb) CreateWorkflowNodeRun(run db.WorkflowNodeRun) (db.WorkflowNodeRun, error) {
	newRun, err := d.CreateObject(db.WorkflowNodeRunProps, run)
	if err != nil {
		return db.WorkflowNodeRun{}, err
	}
	return newRun.(db.WorkflowNodeRun), nil
}

func (d *SqlDb) UpdateWorkflowNodeRun(run db.WorkflowNodeRun) error {
    _, err := d.exec("update workflow_node_run set status=?, finished_at=?, task_id=? where id=?", run.Status, run.FinishedAt, run.TaskID, run.ID)
    return err
}

func (d *SqlDb) GetWorkflowNodeRun(projectID int, runID int) (db.WorkflowNodeRun, error) {
    var run db.WorkflowNodeRun
    // simplified check
    err := d.selectOne(&run, "select * from workflow_node_run where id=?", runID)
	return run, err
}

func (d *SqlDb) GetWorkflowNodeRunByTaskID(taskID int) (db.WorkflowNodeRun, error) {
    var run db.WorkflowNodeRun
    err := d.selectOne(&run, "select * from workflow_node_run where task_id=?", taskID)
	return run, err
}

func (d *SqlDb) GetWorkflowNodeRuns(projectID int, workflowRunID int) ([]db.WorkflowNodeRun, error) {
    var runs []db.WorkflowNodeRun
    _, err := d.selectAll(&runs, "select * from workflow_node_run where workflow_run_id=?", workflowRunID)
	return runs, err
}
