package db_test

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/stretchr/testify/require"
)

func setupWorkflowValidationFixtures(t *testing.T) (db.Store, int, int, int, int, int) {
	t.Helper()

	store := sql.CreateTestStore()

	mkRepo := func(projectID int, name string) int {
		key, err := store.CreateAccessKey(db.AccessKey{ProjectID: &projectID, Type: db.AccessKeyNone})
		require.NoError(t, err)
		repo, err := store.CreateRepository(db.Repository{
			ProjectID: projectID,
			Name:      name,
			GitURL:    "https://example.com/repo.git",
			GitBranch: "main",
			SSHKeyID:  key.ID,
		})
		require.NoError(t, err)
		return repo.ID
	}

	project, err := store.CreateProject(db.Project{Name: "proj"})
	require.NoError(t, err)
	repoID := mkRepo(project.ID, "repo")

	tplA, err := store.CreateTemplate(db.Template{ProjectID: project.ID, RepositoryID: repoID, Name: "A", Playbook: "a.yml"})
	require.NoError(t, err)
	tplB, err := store.CreateTemplate(db.Template{ProjectID: project.ID, RepositoryID: repoID, Name: "B", Playbook: "b.yml"})
	require.NoError(t, err)
	tplC, err := store.CreateTemplate(db.Template{ProjectID: project.ID, RepositoryID: repoID, Name: "C", Playbook: "c.yml"})
	require.NoError(t, err)

	otherProject, err := store.CreateProject(db.Project{Name: "other"})
	require.NoError(t, err)
	otherRepoID := mkRepo(otherProject.ID, "other-repo")
	tplOther, err := store.CreateTemplate(db.Template{ProjectID: otherProject.ID, RepositoryID: otherRepoID, Name: "X", Playbook: "x.yml"})
	require.NoError(t, err)

	return store, project.ID, tplA.ID, tplB.ID, tplC.ID, tplOther.ID
}

// TestValidateWorkflowTemplateIgnoresNodePositions ensures the graphical
// editor's canvas coordinates are pure layout metadata and never influence
// validation (a valid graph stays valid regardless of node positions).
func TestValidateWorkflowTemplateIgnoresNodePositions(t *testing.T) {
	store, projectID, tplA, tplB, _, _ := setupWorkflowValidationFixtures(t)

	err := db.ValidateWorkflowTemplate(store, db.WorkflowTemplate{
		ProjectID: projectID,
		Name:      "wf",
		Nodes: []db.WorkflowNode{
			{ID: 1, TemplateID: tplA, PositionX: -50, PositionY: 9999},
			{ID: 2, TemplateID: tplB, PositionX: 1234, PositionY: 0},
		},
		Edges: []db.WorkflowEdge{
			{SourceNodeID: 1, DestinationNodeID: 2, Condition: db.WorkflowEdgeOnSuccess},
		},
	})

	require.NoError(t, err)
}

func TestValidateWorkflowTemplateNoteNode(t *testing.T) {
	store, projectID, tplA, tplB, _, _ := setupWorkflowValidationFixtures(t)
	note := "remember to bump the version"

	// A note node alongside a normal graph is valid: it is a standalone
	// annotation and does not count as a second root.
	err := db.ValidateWorkflowTemplate(store, db.WorkflowTemplate{
		ProjectID: projectID,
		Name:      "wf",
		Nodes: []db.WorkflowNode{
			{ID: 1, TemplateID: tplA},
			{ID: 2, TemplateID: tplB},
			{ID: 3, Kind: db.WorkflowNodeNoteKind, Note: &note},
		},
		Edges: []db.WorkflowEdge{
			{SourceNodeID: 1, DestinationNodeID: 2, Condition: db.WorkflowEdgeOnSuccess},
		},
	})
	require.NoError(t, err)

	// A workflow consisting only of note nodes has no root and is invalid.
	err = db.ValidateWorkflowTemplate(store, db.WorkflowTemplate{
		ProjectID: projectID,
		Name:      "wf",
		Nodes:     []db.WorkflowNode{{ID: 1, Kind: db.WorkflowNodeNoteKind, Note: &note}},
	})
	require.Error(t, err)

	// Note nodes can not be connected by edges.
	err = db.ValidateWorkflowTemplate(store, db.WorkflowTemplate{
		ProjectID: projectID,
		Name:      "wf",
		Nodes: []db.WorkflowNode{
			{ID: 1, TemplateID: tplA},
			{ID: 2, Kind: db.WorkflowNodeNoteKind, Note: &note},
		},
		Edges: []db.WorkflowEdge{
			{SourceNodeID: 1, DestinationNodeID: 2, Condition: db.WorkflowEdgeOnSuccess},
		},
	})
	require.Error(t, err)

	// A note node can not carry task fields.
	err = db.ValidateWorkflowTemplate(store, db.WorkflowTemplate{
		ProjectID: projectID,
		Name:      "wf",
		Nodes: []db.WorkflowNode{
			{ID: 1, TemplateID: tplA},
			{ID: 2, Kind: db.WorkflowNodeNoteKind, TemplateID: tplB},
		},
	})
	require.Error(t, err)

	// A task node can not carry note text.
	err = db.ValidateWorkflowTemplate(store, db.WorkflowTemplate{
		ProjectID: projectID,
		Name:      "wf",
		Nodes:     []db.WorkflowNode{{ID: 1, TemplateID: tplA, Note: &note}},
	})
	require.Error(t, err)
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

	err = db.ValidateWorkflowTemplate(store, db.WorkflowTemplate{
		ProjectID: projectID,
		Name:      "wf",
		Nodes: []db.WorkflowNode{
			{ID: 1, TemplateID: tplA},
			{ID: 2, Kind: db.WorkflowNodeApprovalKind, Limit: db.StringArrayField{"web"}},
		},
		Edges: []db.WorkflowEdge{
			{SourceNodeID: 1, DestinationNodeID: 2, Condition: db.WorkflowEdgeOnSuccess},
		},
	})
	if err == nil {
		t.Fatal("expected approval node with limit to fail validation")
	}
}
