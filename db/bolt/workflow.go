package bolt

import (
	"github.com/semaphoreui/semaphore/db"
)

// Mock implementation for Workflows
func (d *BoltDb) GetWorkflows(projectID int, params db.RetrieveQueryParams) ([]db.Workflow, error) {
	// TODO: Implement
	return nil, nil
}

func (d *BoltDb) GetWorkflow(projectID int, workflowID int) (db.Workflow, error) {
	// TODO: Implement
	return db.Workflow{}, nil
}

func (d *BoltDb) CreateWorkflow(workflow db.Workflow) (db.Workflow, error) {
	// TODO: Implement
	return db.Workflow{}, nil
}

func (d *BoltDb) UpdateWorkflow(workflow db.Workflow) error {
	// TODO: Implement
	return nil
}

func (d *BoltDb) DeleteWorkflow(projectID int, workflowID int) error {
	// TODO: Implement
	return nil
}

// Mock implementation for Workflow nodes
func (d *BoltDb) GetWorkflowNodes(workflowID int) ([]db.WorkflowNode, error) {
	// TODO: Implement
	return nil, nil
}

func (d *BoltDb) GetWorkflowNode(workflowID int, nodeID int) (db.WorkflowNode, error) {
	// TODO: Implement
	return db.WorkflowNode{}, nil
}

func (d *BoltDb) CreateWorkflowNode(node db.WorkflowNode) (db.WorkflowNode, error) {
	// TODO: Implement
	return db.WorkflowNode{}, nil
}

func (d *BoltDb) UpdateWorkflowNode(node db.WorkflowNode) error {
	// TODO: Implement
	return nil
}

func (d *BoltDb) DeleteWorkflowNode(workflowID int, nodeID int) error {
	// TODO: Implement
	return nil
}

// Mock implementation for Workflow links
func (d *BoltDb) GetWorkflowLinks(workflowID int) ([]db.WorkflowLink, error) {
	// TODO: Implement
	return nil, nil
}

func (d *BoltDb) CreateWorkflowLink(link db.WorkflowLink) (db.WorkflowLink, error) {
	// TODO: Implement
	return db.WorkflowLink{}, nil
}

func (d *BoltDb) DeleteWorkflowLink(workflowID int, linkID int) error {
	// TODO: Implement
	return nil
}

// Mock implementation for Workflow runs
func (d *BoltDb) GetWorkflowRuns(workflowID int, params db.RetrieveQueryParams) ([]db.WorkflowRunWithWorkflow, error) {
	// TODO: Implement
	return nil, nil
}

func (d *BoltDb) GetProjectWorkflowRuns(projectID int, params db.RetrieveQueryParams) ([]db.WorkflowRunWithWorkflow, error) {
	// TODO: Implement
	return nil, nil
}

func (d *BoltDb) GetWorkflowRun(workflowRunID int) (db.WorkflowRun, error) {
	// TODO: Implement
	return db.WorkflowRun{}, nil
}

func (d *BoltDb) CreateWorkflowRun(run db.WorkflowRun) (db.WorkflowRun, error) {
	// TODO: Implement
	return db.WorkflowRun{}, nil
}

func (d *BoltDb) UpdateWorkflowRun(run db.WorkflowRun) error {
	// TODO: Implement
	return nil
}

func (d *BoltDb) DeleteWorkflowRun(workflowRunID int) error {
	// 'TODO: Implement
	return nil
}

// Mock implementation for Workflow node runs
func (d *BoltDb) GetWorkflowNodeRuns(workflowRunID int) ([]db.WorkflowNodeRun, error) {
	// TODO: Implement
	return nil, nil
}

func (d *BoltDb) GetWorkflowNodeRun(nodeRunID int) (db.WorkflowNodeRun, error) {
	// TODO: Implement
	return db.WorkflowNodeRun{}, nil
}

func (d *BoltDb) CreateWorkflowNodeRun(nodeRun db.WorkflowNodeRun) (db.WorkflowNodeRun, error) {
	// TODO: Implement
	return db.WorkflowNodeRun{}, nil
}

func (d *BoltDb) UpdateWorkflowNodeRun(nodeRun db.WorkflowNodeRun) error {
	// TODO: Implement
	return nil
}
