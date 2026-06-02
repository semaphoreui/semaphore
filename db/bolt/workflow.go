package bolt

import (
	"github.com/semaphoreui/semaphore/db"
)

func (d *BoltDb) GetWorkflowTemplates(projectID int, params db.RetrieveQueryParams) (workflows []db.WorkflowTemplate, err error) {
	workflows = make([]db.WorkflowTemplate, 0)
	err = d.getObjects(projectID, db.WorkflowTemplateProps, params, nil, &workflows)
	return
}

func (d *BoltDb) GetWorkflowTemplate(projectID int, workflowID int) (workflow db.WorkflowTemplate, err error) {
	err = d.getObject(projectID, db.WorkflowTemplateProps, intObjectID(workflowID), &workflow)
	return
}

func (d *BoltDb) CreateWorkflowTemplate(workflow db.WorkflowTemplate) (newWorkflow db.WorkflowTemplate, err error) {
	err = db.ValidateWorkflowTemplate(d, workflow)
	if err != nil {
		return
	}

	res, err := d.createObject(workflow.ProjectID, db.WorkflowTemplateProps, workflow)
	if err != nil {
		return
	}

	newWorkflow = res.(db.WorkflowTemplate)
	return
}

func (d *BoltDb) UpdateWorkflowTemplate(workflow db.WorkflowTemplate) (err error) {
	err = db.ValidateWorkflowTemplate(d, workflow)
	if err != nil {
		return
	}

	return d.updateObject(workflow.ProjectID, db.WorkflowTemplateProps, workflow)
}

func (d *BoltDb) DeleteWorkflowTemplate(projectID int, workflowID int) error {
	runs, err := d.GetWorkflowRuns(projectID, workflowID, db.RetrieveQueryParams{})
	if err != nil {
		return err
	}
	for _, run := range runs {
		err = d.deleteObject(projectID, db.WorkflowRunProps, intObjectID(run.ID), nil)
		if err != nil {
			return err
		}
	}

	return d.deleteObject(projectID, db.WorkflowTemplateProps, intObjectID(workflowID), nil)
}

func (d *BoltDb) GetWorkflowRuns(projectID int, workflowTemplateID int, params db.RetrieveQueryParams) (runs []db.WorkflowRun, err error) {
	runs = make([]db.WorkflowRun, 0)
	err = d.getObjects(projectID, db.WorkflowRunProps, params, func(i any) bool {
		r := i.(db.WorkflowRun)
		return r.WorkflowTemplateID == workflowTemplateID
	}, &runs)
	return
}

func (d *BoltDb) GetWorkflowRun(projectID int, workflowTemplateID int, runID int) (run db.WorkflowRun, err error) {
	run, err = d.GetWorkflowRunByID(projectID, runID)
	if err != nil {
		return
	}
	if run.WorkflowTemplateID != workflowTemplateID {
		err = db.ErrNotFound
	}
	return
}

func (d *BoltDb) GetWorkflowRunByID(projectID int, runID int) (run db.WorkflowRun, err error) {
	err = d.getObject(projectID, db.WorkflowRunProps, intObjectID(runID), &run)
	return
}

func (d *BoltDb) CreateWorkflowRun(run db.WorkflowRun) (newRun db.WorkflowRun, err error) {
	res, err := d.createObject(run.ProjectID, db.WorkflowRunProps, run)
	if err != nil {
		return
	}

	newRun = res.(db.WorkflowRun)
	return
}

func (d *BoltDb) UpdateWorkflowRun(run db.WorkflowRun) error {
	return d.updateObject(run.ProjectID, db.WorkflowRunProps, run)
}
