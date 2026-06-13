package bolt

import (
	"encoding/json"
	"errors"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/tz"
)

// Workflow operations

func (d *BoltDb) GetWorkflows(projectID int) ([]db.Workflow, error) {
	var workflows []db.Workflow
	err := d.getObjects(projectID, "workflows", &workflows)
	return workflows, err
}

func (d *BoltDb) GetWorkflow(projectID int, workflowID int) (db.Workflow, error) {
	var workflow db.Workflow
	err := d.getObject(projectID, "workflows", workflowID, &workflow)
	if err != nil {
		return workflow, err
	}
	if projectID > 0 && workflow.ProjectID != projectID {
		return workflow, db.ErrNotFound
	}
	return workflow, nil
}

func (d *BoltDb) CreateWorkflow(workflow db.Workflow) (db.Workflow, error) {
	now := tz.Now()
	workflow.Created = now
	workflow.Updated = now

	workflow.ID = d.getNextID(workflow.ProjectID, "workflows")
	return workflow, d.saveObject(workflow.ProjectID, "workflows", workflow.ID, workflow)
}

func (d *BoltDb) UpdateWorkflow(workflow db.Workflow) error {
	now := tz.Now()
	workflow.Updated = now

	existing, err := d.GetWorkflow(workflow.ProjectID, workflow.ID)
	if err != nil {
		return err
	}

	workflow.Created = existing.Created
	return d.saveObject(workflow.ProjectID, "workflows", workflow.ID, workflow)
}

func (d *BoltDb) DeleteWorkflow(projectID int, workflowID int) error {
	return d.deleteObject(projectID, "workflows", workflowID)
}

// WorkflowNode operations

func (d *BoltDb) GetWorkflowNodes(workflowID int) ([]db.WorkflowNode, error) {
	workflow, err := d.GetWorkflow(0, workflowID)
	if err != nil {
		return nil, err
	}

	var nodes []db.WorkflowNode
	err = d.getObjects(workflow.ProjectID, "workflow_nodes", &nodes)
	if err != nil {
		return nil, err
	}

	// Filter by workflow_id
	var filteredNodes []db.WorkflowNode
	for _, node := range nodes {
		if node.WorkflowID == workflowID {
			filteredNodes = append(filteredNodes, node)
		}
	}

	return filteredNodes, nil
}

func (d *BoltDb) CreateWorkflowNode(node db.WorkflowNode) (db.WorkflowNode, error) {
	workflow, err := d.GetWorkflow(0, node.WorkflowID)
	if err != nil {
		return node, err
	}

	node.ID = d.getNextID(workflow.ProjectID, "workflow_nodes")
	return node, d.saveObject(workflow.ProjectID, "workflow_nodes", node.ID, node)
}

func (d *BoltDb) UpdateWorkflowNode(node db.WorkflowNode) error {
	workflow, err := d.GetWorkflow(0, node.WorkflowID)
	if err != nil {
		return err
	}

	return d.saveObject(workflow.ProjectID, "workflow_nodes", node.ID, node)
}

func (d *BoltDb) DeleteWorkflowNode(workflowID int, nodeID int) error {
	workflow, err := d.GetWorkflow(0, workflowID)
	if err != nil {
		return err
	}

	return d.deleteObject(workflow.ProjectID, "workflow_nodes", nodeID)
}

// WorkflowLink operations

func (d *BoltDb) GetWorkflowLinks(workflowID int) ([]db.WorkflowLink, error) {
	workflow, err := d.GetWorkflow(0, workflowID)
	if err != nil {
		return nil, err
	}

	var links []db.WorkflowLink
	err = d.getObjects(workflow.ProjectID, "workflow_links", &links)
	if err != nil {
		return nil, err
	}

	// Filter by workflow_id
	var filteredLinks []db.WorkflowLink
	for _, link := range links {
		if link.WorkflowID == workflowID {
			filteredLinks = append(filteredLinks, link)
		}
	}

	return filteredLinks, nil
}

func (d *BoltDb) CreateWorkflowLink(link db.WorkflowLink) (db.WorkflowLink, error) {
	workflow, err := d.GetWorkflow(0, link.WorkflowID)
	if err != nil {
		return link, err
	}

	link.ID = d.getNextID(workflow.ProjectID, "workflow_links")
	return link, d.saveObject(workflow.ProjectID, "workflow_links", link.ID, link)
}

func (d *BoltDb) DeleteWorkflowLink(workflowID int, linkID int) error {
	workflow, err := d.GetWorkflow(0, workflowID)
	if err != nil {
		return err
	}

	return d.deleteObject(workflow.ProjectID, "workflow_links", linkID)
}

// WorkflowRun operations

func (d *BoltDb) GetWorkflowRuns(workflowID int, params db.RetrieveQueryParams) ([]db.WorkflowRun, error) {
	workflow, err := d.GetWorkflow(0, workflowID)
	if err != nil {
		return nil, err
	}

	var runs []db.WorkflowRun
	err = d.getObjects(workflow.ProjectID, "workflow_runs", &runs)
	if err != nil {
		return nil, err
	}

	// Filter by workflow_id
	var filteredRuns []db.WorkflowRun
	for _, run := range runs {
		if run.WorkflowID == workflowID {
			filteredRuns = append(filteredRuns, run)
		}
	}

	// Apply pagination
	start := params.Offset
	end := start + params.Count
	if params.Count <= 0 {
		end = len(filteredRuns)
	}
	if start > len(filteredRuns) {
		return []db.WorkflowRun{}, nil
	}
	if end > len(filteredRuns) {
		end = len(filteredRuns)
	}

	return filteredRuns[start:end], nil
}

func (d *BoltDb) GetWorkflowRun(workflowID int, runID int) (db.WorkflowRun, error) {
	workflow, err := d.GetWorkflow(0, workflowID)
	if err != nil {
		return db.WorkflowRun{}, err
	}

	var run db.WorkflowRun
	err = d.getObject(workflow.ProjectID, "workflow_runs", runID, &run)
	if err != nil {
		return run, err
	}

	if run.WorkflowID != workflowID {
		return run, db.ErrNotFound
	}

	return run, nil
}

func (d *BoltDb) CreateWorkflowRun(run db.WorkflowRun) (db.WorkflowRun, error) {
	workflow, err := d.GetWorkflow(0, run.WorkflowID)
	if err != nil {
		return run, err
	}

	now := tz.Now()
	run.Created = now

	run.ID = d.getNextID(workflow.ProjectID, "workflow_runs")
	return run, d.saveObject(workflow.ProjectID, "workflow_runs", run.ID, run)
}

func (d *BoltDb) UpdateWorkflowRun(run db.WorkflowRun) error {
	workflow, err := d.GetWorkflow(0, run.WorkflowID)
	if err != nil {
		return err
	}

	existing, err := d.GetWorkflowRun(run.WorkflowID, run.ID)
	if err != nil {
		return err
	}

	run.Created = existing.Created
	return d.saveObject(workflow.ProjectID, "workflow_runs", run.ID, run)
}

// WorkflowRunNode operations

func (d *BoltDb) GetWorkflowRunNodes(runID int) ([]db.WorkflowRunNode, error) {
	// Get run to find projectID
	var runs []db.WorkflowRun
	projects, err := d.GetAllProjects()
	if err != nil {
		return nil, err
	}

	var run db.WorkflowRun
	for _, project := range projects {
		err = d.getObjects(project.ID, "workflow_runs", &runs)
		if err == nil {
			for _, r := range runs {
				if r.ID == runID {
					run = r
					break
				}
			}
		}
		if run.ID > 0 {
			break
		}
	}

	if run.ID == 0 {
		return nil, db.ErrNotFound
	}

	workflow, err := d.GetWorkflow(0, run.WorkflowID)
	if err != nil {
		return nil, err
	}

	var nodes []db.WorkflowRunNode
	err = d.getObjects(workflow.ProjectID, "workflow_run_nodes", &nodes)
	if err != nil {
		return nil, err
	}

	// Filter by run_id
	var filteredNodes []db.WorkflowRunNode
	for _, node := range nodes {
		if node.WorkflowRunID == runID {
			filteredNodes = append(filteredNodes, node)
		}
	}

	return filteredNodes, nil
}

func (d *BoltDb) CreateWorkflowRunNode(node db.WorkflowRunNode) (db.WorkflowRunNode, error) {
	// Get run to find projectID
	var runs []db.WorkflowRun
	projects, err := d.GetAllProjects()
	if err != nil {
		return node, err
	}

	var run db.WorkflowRun
	for _, project := range projects {
		err = d.getObjects(project.ID, "workflow_runs", &runs)
		if err == nil {
			for _, r := range runs {
				if r.ID == node.WorkflowRunID {
					run = r
					break
				}
			}
		}
		if run.ID > 0 {
			break
		}
	}

	if run.ID == 0 {
		return node, db.ErrNotFound
	}

	workflow, err := d.GetWorkflow(0, run.WorkflowID)
	if err != nil {
		return node, err
	}

	now := tz.Now()
	node.Created = now

	node.ID = d.getNextID(workflow.ProjectID, "workflow_run_nodes")
	return node, d.saveObject(workflow.ProjectID, "workflow_run_nodes", node.ID, node)
}

func (d *BoltDb) UpdateWorkflowRunNode(node db.WorkflowRunNode) error {
	// Get run to find projectID
	var runs []db.WorkflowRun
	projects, err := d.GetAllProjects()
	if err != nil {
		return err
	}

	var run db.WorkflowRun
	for _, project := range projects {
		err = d.getObjects(project.ID, "workflow_runs", &runs)
		if err == nil {
			for _, r := range runs {
				if r.ID == node.WorkflowRunID {
					run = r
					break
				}
			}
		}
		if run.ID > 0 {
			break
		}
	}

	if run.ID == 0 {
		return db.ErrNotFound
	}

	workflow, err := d.GetWorkflow(0, run.WorkflowID)
	if err != nil {
		return err
	}

	existing, err := d.GetWorkflowRunNodes(node.WorkflowRunID)
	if err != nil {
		return err
	}

	for _, n := range existing {
		if n.ID == node.ID {
			node.Created = n.Created
			break
		}
	}

	return d.saveObject(workflow.ProjectID, "workflow_run_nodes", node.ID, node)
}
