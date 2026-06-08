package sql

import (
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkflowTemplateCRUD(t *testing.T) {
	store := CreateTestStore()

	project, err := store.CreateProject(db.Project{Name: "proj"})
	if err != nil {
		t.Fatal(err)
	}

	key, err := store.CreateAccessKey(db.AccessKey{
		ProjectID: &project.ID,
		Type:      db.AccessKeyNone,
	})
	if err != nil {
		t.Fatal(err)
	}

	repo, err := store.CreateRepository(db.Repository{
		ProjectID: project.ID,
		Name:      "repo",
		GitURL:    "https://example.com/repo.git",
		GitBranch: "main",
		SSHKeyID:  key.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	tplA, err := store.CreateTemplate(db.Template{ProjectID: project.ID, RepositoryID: repo.ID, Name: "A", Playbook: "a.yml"})
	if err != nil {
		t.Fatal(err)
	}
	tplB, err := store.CreateTemplate(db.Template{ProjectID: project.ID, RepositoryID: repo.ID, Name: "B", Playbook: "b.yml"})
	if err != nil {
		t.Fatal(err)
	}

	workflow, err := store.CreateWorkflowTemplate(db.WorkflowTemplate{
		ProjectID: project.ID,
		Name:      "wf",
		Nodes: []db.WorkflowNode{
			{ID: 1, TemplateID: tplA.ID, Limit: db.StringArrayField{"web*", "db"}},
			{ID: 2, Kind: db.WorkflowNodeApprovalKind, ApprovalTimeout: intPtr(30), ApprovalMessage: strPtr("need approval"), ConvergenceMode: db.WorkflowConvergenceAny},
			{ID: 3, TemplateID: tplB.ID, ConvergenceMode: db.WorkflowConvergenceAll},
		},
		Edges: []db.WorkflowEdge{
			{
				SourceNodeID:      1,
				DestinationNodeID: 2,
				Condition:         db.WorkflowEdgeOnSuccess,
			},
			{
				SourceNodeID:      2,
				DestinationNodeID: 3,
				Condition:         db.WorkflowEdgeOnSuccess,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(workflow.Nodes) != 3 || len(workflow.Edges) != 2 {
		t.Fatal("unexpected persisted workflow graph")
	}
	if workflow.Nodes[1].Kind != db.WorkflowNodeApprovalKind || workflow.Nodes[1].ConvergenceMode != db.WorkflowConvergenceAny {
		t.Fatal("workflow node fields were not persisted")
	}

	workflow.Name = "wf-updated"
	if err = store.UpdateWorkflowTemplate(workflow); err != nil {
		t.Fatal(err)
	}

	updated, err := store.GetWorkflowTemplate(project.ID, workflow.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "wf-updated" {
		t.Fatal("workflow update was not persisted")
	}
	if updated.Nodes[1].ApprovalTimeout == nil || *updated.Nodes[1].ApprovalTimeout != 30 {
		t.Fatal("approval timeout was not persisted")
	}
	if len(updated.Nodes[0].Limit) != 2 || updated.Nodes[0].Limit[0] != "web*" || updated.Nodes[0].Limit[1] != "db" {
		t.Fatal("workflow node limit was not persisted")
	}
	if updated.Nodes[2].ConvergenceMode != db.WorkflowConvergenceAll {
		t.Fatal("convergence mode was not persisted")
	}

	if err = store.DeleteWorkflowTemplate(project.ID, workflow.ID); err != nil {
		t.Fatal(err)
	}

	if _, err = store.GetWorkflowTemplate(project.ID, workflow.ID); err == nil {
		t.Fatal("expected workflow to be deleted")
	}
}

// TestWorkflowNodePositionsRoundTrip verifies that the graphical editor's
// per-node canvas coordinates survive both the initial insert and the
// delete-and-reinsert performed by UpdateWorkflowTemplate (which reassigns node
// IDs). The position must travel with the node row in the same INSERT.
func TestWorkflowNodePositionsRoundTrip(t *testing.T) {
	store := CreateTestStore()

	project, err := store.CreateProject(db.Project{Name: "proj"})
	require.NoError(t, err)

	key, err := store.CreateAccessKey(db.AccessKey{ProjectID: &project.ID, Type: db.AccessKeyNone})
	require.NoError(t, err)

	repo, err := store.CreateRepository(db.Repository{
		ProjectID: project.ID,
		Name:      "repo",
		GitURL:    "https://example.com/repo.git",
		GitBranch: "main",
		SSHKeyID:  key.ID,
	})
	require.NoError(t, err)

	tplA, err := store.CreateTemplate(db.Template{ProjectID: project.ID, RepositoryID: repo.ID, Name: "A", Playbook: "a.yml"})
	require.NoError(t, err)
	tplB, err := store.CreateTemplate(db.Template{ProjectID: project.ID, RepositoryID: repo.ID, Name: "B", Playbook: "b.yml"})
	require.NoError(t, err)

	workflow, err := store.CreateWorkflowTemplate(db.WorkflowTemplate{
		ProjectID: project.ID,
		Name:      "wf",
		Nodes: []db.WorkflowNode{
			{ID: 1, TemplateID: tplA.ID, PositionX: 120, PositionY: 40},
			{ID: 2, TemplateID: tplB.ID, PositionX: 360, PositionY: 240},
		},
		Edges: []db.WorkflowEdge{
			{SourceNodeID: 1, DestinationNodeID: 2, Condition: db.WorkflowEdgeOnSuccess},
		},
	})
	require.NoError(t, err)
	require.Len(t, workflow.Nodes, 2)

	// Positions are persisted on initial insert and returned by GetWorkflowTemplate.
	assert.Equal(t, 120, workflow.Nodes[0].PositionX)
	assert.Equal(t, 40, workflow.Nodes[0].PositionY)
	assert.Equal(t, 360, workflow.Nodes[1].PositionX)
	assert.Equal(t, 240, workflow.Nodes[1].PositionY)

	// Simulate the editor moving a node and re-saving. UpdateWorkflowTemplate
	// deletes and reinserts every node (new IDs); positions must survive because
	// they ride along in the node INSERT.
	workflow.Nodes[1].PositionX = 500
	workflow.Nodes[1].PositionY = 300
	require.NoError(t, store.UpdateWorkflowTemplate(workflow))

	updated, err := store.GetWorkflowTemplate(project.ID, workflow.ID)
	require.NoError(t, err)
	require.Len(t, updated.Nodes, 2)
	assert.Equal(t, 120, updated.Nodes[0].PositionX)
	assert.Equal(t, 40, updated.Nodes[0].PositionY)
	assert.Equal(t, 500, updated.Nodes[1].PositionX)
	assert.Equal(t, 300, updated.Nodes[1].PositionY)
}

func TestWorkflowApprovalCRUD(t *testing.T) {
	store := CreateTestStore()

	project, err := store.CreateProject(db.Project{Name: "proj"})
	if err != nil {
		t.Fatal(err)
	}
	key, err := store.CreateAccessKey(db.AccessKey{ProjectID: &project.ID, Type: db.AccessKeyNone})
	if err != nil {
		t.Fatal(err)
	}
	repo, err := store.CreateRepository(db.Repository{
		ProjectID: project.ID,
		Name:      "repo",
		GitURL:    "https://example.com/repo.git",
		GitBranch: "main",
		SSHKeyID:  key.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	tpl, err := store.CreateTemplate(db.Template{ProjectID: project.ID, RepositoryID: repo.ID, Name: "A", Playbook: "a.yml"})
	if err != nil {
		t.Fatal(err)
	}
	workflow, err := store.CreateWorkflowTemplate(db.WorkflowTemplate{
		ProjectID: project.ID,
		Name:      "wf",
		Nodes: []db.WorkflowNode{
			{ID: 1, TemplateID: tpl.ID},
			{ID: 2, Kind: db.WorkflowNodeApprovalKind},
		},
		Edges: []db.WorkflowEdge{{
			SourceNodeID:      1,
			DestinationNodeID: 2,
			Condition:         db.WorkflowEdgeOnSuccess,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	run, err := store.CreateWorkflowRun(db.WorkflowRun{
		ProjectID:          project.ID,
		WorkflowTemplateID: workflow.ID,
		Status:             db.WorkflowRunRunning,
	})
	if err != nil {
		t.Fatal(err)
	}

	created := time.Now().UTC()
	approval, err := store.CreateWorkflowApproval(db.WorkflowApproval{
		ProjectID:      project.ID,
		WorkflowRunID:  run.ID,
		WorkflowNodeID: workflow.Nodes[1].ID,
		Status:         db.WorkflowApprovalPending,
		Created:        created,
	})
	if err != nil {
		t.Fatal(err)
	}

	loaded, err := store.GetWorkflowApproval(project.ID, approval.WorkflowRunID, approval.WorkflowNodeID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Status != db.WorkflowApprovalPending {
		t.Fatal("workflow approval status mismatch")
	}

	now := time.Now().UTC()
	loaded.Status = db.WorkflowApprovalApproved
	loaded.Resolved = &now
	if err = store.UpdateWorkflowApproval(loaded); err != nil {
		t.Fatal(err)
	}

	approvals, err := store.GetWorkflowApprovals(project.ID, approval.WorkflowRunID)
	if err != nil {
		t.Fatal(err)
	}
	if len(approvals) != 1 || approvals[0].Status != db.WorkflowApprovalApproved {
		t.Fatal("workflow approval update was not persisted")
	}
}

func intPtr(v int) *int { return &v }

func strPtr(v string) *string { return &v }
