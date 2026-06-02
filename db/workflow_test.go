package db_test

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/bolt"
)

func setupWorkflowValidationFixtures(t *testing.T) (db.Store, int, int, int, int, int) {
	t.Helper()

	store := bolt.CreateTestStore()

	project, err := store.CreateProject(db.Project{Name: "proj"})
	if err != nil {
		t.Fatal(err)
	}

	invID := 0
	tplA, err := store.CreateTemplate(db.Template{ProjectID: project.ID, Name: "A", Playbook: "a.yml", InventoryID: &invID})
	if err != nil {
		t.Fatal(err)
	}
	tplB, err := store.CreateTemplate(db.Template{ProjectID: project.ID, Name: "B", Playbook: "b.yml", InventoryID: &invID})
	if err != nil {
		t.Fatal(err)
	}
	tplC, err := store.CreateTemplate(db.Template{ProjectID: project.ID, Name: "C", Playbook: "c.yml", InventoryID: &invID})
	if err != nil {
		t.Fatal(err)
	}

	otherProject, err := store.CreateProject(db.Project{Name: "other"})
	if err != nil {
		t.Fatal(err)
	}
	tplOther, err := store.CreateTemplate(db.Template{ProjectID: otherProject.ID, Name: "X", Playbook: "x.yml", InventoryID: &invID})
	if err != nil {
		t.Fatal(err)
	}

	return store, project.ID, tplA.ID, tplB.ID, tplC.ID, tplOther.ID
}

func TestValidateWorkflowTemplateRejectsCycle(t *testing.T) {
	store, projectID, tplA, tplB, _, _ := setupWorkflowValidationFixtures(t)

	err := db.ValidateWorkflowTemplate(store, db.WorkflowTemplate{
		ProjectID: projectID,
		Name:      "wf",
		Nodes: []db.WorkflowNode{
			{ID: 1, TemplateID: tplA},
			{ID: 2, TemplateID: tplB},
		},
		Edges: []db.WorkflowEdge{
			{SourceNodeID: 1, DestinationNodeID: 2, Condition: db.WorkflowEdgeOnSuccess},
			{SourceNodeID: 2, DestinationNodeID: 1, Condition: db.WorkflowEdgeOnSuccess},
		},
	})

	if err == nil {
		t.Fatal("expected cycle validation error")
	}
}

func TestValidateWorkflowTemplateRequiresSingleRoot(t *testing.T) {
	store, projectID, tplA, tplB, _, _ := setupWorkflowValidationFixtures(t)

	err := db.ValidateWorkflowTemplate(store, db.WorkflowTemplate{
		ProjectID: projectID,
		Name:      "wf",
		Nodes: []db.WorkflowNode{
			{ID: 1, TemplateID: tplA},
			{ID: 2, TemplateID: tplB},
		},
	})

	if err == nil {
		t.Fatal("expected root validation error")
	}
}

func TestValidateWorkflowTemplateRejectsCrossProjectTemplate(t *testing.T) {
	store, projectID, tplA, _, _, tplOther := setupWorkflowValidationFixtures(t)
	invalidTemplateID := tplOther + 100000

	err := db.ValidateWorkflowTemplate(store, db.WorkflowTemplate{
		ProjectID: projectID,
		Name:      "wf",
		Nodes: []db.WorkflowNode{
			{ID: 1, TemplateID: tplA},
			{ID: 2, TemplateID: invalidTemplateID},
		},
		Edges: []db.WorkflowEdge{
			{SourceNodeID: 1, DestinationNodeID: 2, Condition: db.WorkflowEdgeOnSuccess},
		},
	})

	if err == nil {
		t.Fatal("expected cross-project validation error")
	}
}

func TestValidateWorkflowTemplateAllowsApprovalNodeWithoutTemplate(t *testing.T) {
	store, projectID, tplA, _, _, _ := setupWorkflowValidationFixtures(t)
	timeout := 60

	err := db.ValidateWorkflowTemplate(store, db.WorkflowTemplate{
		ProjectID: projectID,
		Name:      "wf",
		Nodes: []db.WorkflowNode{
			{ID: 1, TemplateID: tplA},
			{ID: 2, Kind: db.WorkflowNodeApprovalKind, ApprovalTimeout: &timeout},
		},
		Edges: []db.WorkflowEdge{
			{SourceNodeID: 1, DestinationNodeID: 2, Condition: db.WorkflowEdgeOnSuccess},
		},
	})

	if err != nil {
		t.Fatalf("expected valid approval workflow, got %v", err)
	}
}

func TestValidateWorkflowTemplateRejectsInvalidNodeKindCombinations(t *testing.T) {
	store, projectID, tplA, _, _, _ := setupWorkflowValidationFixtures(t)
	timeout := 60

	err := db.ValidateWorkflowTemplate(store, db.WorkflowTemplate{
		ProjectID: projectID,
		Name:      "wf",
		Nodes: []db.WorkflowNode{
			{ID: 1, TemplateID: tplA},
			{ID: 2, Kind: db.WorkflowNodeApprovalKind, TemplateID: tplA},
		},
		Edges: []db.WorkflowEdge{
			{SourceNodeID: 1, DestinationNodeID: 2, Condition: db.WorkflowEdgeOnSuccess},
		},
	})
	if err == nil {
		t.Fatal("expected approval node with template_id to fail validation")
	}

	err = db.ValidateWorkflowTemplate(store, db.WorkflowTemplate{
		ProjectID: projectID,
		Name:      "wf",
		Nodes: []db.WorkflowNode{
			{ID: 1, TemplateID: tplA},
			{ID: 2, Kind: db.WorkflowNodeTaskKind, TemplateID: tplA, ApprovalTimeout: &timeout},
		},
		Edges: []db.WorkflowEdge{
			{SourceNodeID: 1, DestinationNodeID: 2, Condition: db.WorkflowEdgeOnSuccess},
		},
	})
	if err == nil {
		t.Fatal("expected task node with approval fields to fail validation")
	}
}
