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

// TestWorkflowNoteNodeRoundTrip verifies that a note node's kind and text are
// persisted and read back.
func TestWorkflowNoteNodeRoundTrip(t *testing.T) {
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
	tpl, err := store.CreateTemplate(db.Template{ProjectID: project.ID, RepositoryID: repo.ID, Name: "A", Playbook: "a.yml"})
	require.NoError(t, err)

	note := "deploy checklist"
	workflow, err := store.CreateWorkflowTemplate(db.WorkflowTemplate{
		ProjectID: project.ID,
		Name:      "wf",
		Nodes: []db.WorkflowNode{
			{ID: 1, TemplateID: tpl.ID, PositionX: 40, PositionY: 40},
			{ID: 2, Kind: db.WorkflowNodeNoteKind, Note: &note, PositionX: 300, PositionY: 80},
		},
	})
	require.NoError(t, err)
	require.Len(t, workflow.Nodes, 2)

	var noteNode db.WorkflowNode
	for _, n := range workflow.Nodes {
		if n.EffectiveKind() == db.WorkflowNodeNoteKind {
			noteNode = n
		}
	}
	require.Equal(t, db.WorkflowNodeNoteKind, noteNode.EffectiveKind())
	require.NotNil(t, noteNode.Note)
	assert.Equal(t, note, *noteNode.Note)
	assert.Equal(t, 300, noteNode.PositionX)
}

// TestWorkflowVersioningRoundTrip verifies the template start_version and the
// run version columns are persisted and read back.
func TestWorkflowVersioningRoundTrip(t *testing.T) {
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
	tpl, err := store.CreateTemplate(db.Template{ProjectID: project.ID, RepositoryID: repo.ID, Name: "A", Playbook: "a.yml"})
	require.NoError(t, err)

	startVersion := "2.0.0"
	workflow, err := store.CreateWorkflowTemplate(db.WorkflowTemplate{
		ProjectID:    project.ID,
		Name:         "wf",
		StartVersion: &startVersion,
		Nodes:        []db.WorkflowNode{{ID: 1, TemplateID: tpl.ID}},
	})
	require.NoError(t, err)
	require.NotNil(t, workflow.StartVersion)
	assert.Equal(t, "2.0.0", *workflow.StartVersion)

	got, err := store.GetWorkflowTemplate(project.ID, workflow.ID)
	require.NoError(t, err)
	require.NotNil(t, got.StartVersion)
	assert.Equal(t, "2.0.0", *got.StartVersion)

	version := "2.0.0"
	run, err := store.CreateWorkflowRun(db.WorkflowRun{
		ProjectID:          project.ID,
		WorkflowTemplateID: workflow.ID,
		Status:             db.WorkflowRunRunning,
		Version:            &version,
	})
	require.NoError(t, err)

	gotRun, err := store.GetWorkflowRunByID(project.ID, run.ID)
	require.NoError(t, err)
	require.NotNil(t, gotRun.Version)
	assert.Equal(t, "2.0.0", *gotRun.Version)
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

// createWorkflowRunFixture builds the minimal project/template/workflow/run
// scaffolding the conditional-update tests need.
func createWorkflowRunFixture(t *testing.T, store *SqlDb) (db.Project, db.WorkflowTemplate, db.WorkflowRun) {
	t.Helper()

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
	tpl, err := store.CreateTemplate(db.Template{ProjectID: project.ID, RepositoryID: repo.ID, Name: "A", Playbook: "a.yml"})
	require.NoError(t, err)

	workflow, err := store.CreateWorkflowTemplate(db.WorkflowTemplate{
		ProjectID: project.ID,
		Name:      "wf",
		Nodes:     []db.WorkflowNode{{ID: 1, TemplateID: tpl.ID}},
	})
	require.NoError(t, err)

	run, err := store.CreateWorkflowRun(db.WorkflowRun{
		ProjectID:          project.ID,
		WorkflowTemplateID: workflow.ID,
		Status:             db.WorkflowRunRunning,
	})
	require.NoError(t, err)

	return project, workflow, run
}

func TestUpdateWorkflowRunStatusUnless(t *testing.T) {
	store := CreateTestStore()
	project, _, run := createWorkflowRunFixture(t, store)

	// A non-excluded current status is updated.
	now := time.Now().UTC()
	run.Status = db.WorkflowRunFailed
	run.End = &now
	updated, err := store.UpdateWorkflowRunStatusUnless(run, []db.WorkflowRunStatus{db.WorkflowRunStopped})
	require.NoError(t, err)
	assert.True(t, updated)

	reloaded, err := store.GetWorkflowRunByID(project.ID, run.ID)
	require.NoError(t, err)
	assert.Equal(t, db.WorkflowRunFailed, reloaded.Status)
	require.NotNil(t, reloaded.End)

	// Move the run to stopped, then try a stale "running" write excluding
	// stopped: the fence must hold.
	reloaded.Status = db.WorkflowRunStopped
	require.NoError(t, store.UpdateWorkflowRun(reloaded))

	stale := reloaded
	stale.Status = db.WorkflowRunRunning
	stale.End = nil
	updated, err = store.UpdateWorkflowRunStatusUnless(stale, []db.WorkflowRunStatus{db.WorkflowRunStopped})
	require.NoError(t, err)
	assert.False(t, updated, "a stopped run must not be revived by a conditional write")

	reloaded, err = store.GetWorkflowRunByID(project.ID, run.ID)
	require.NoError(t, err)
	assert.Equal(t, db.WorkflowRunStopped, reloaded.Status)
}

func TestSetWorkflowRunRootTask(t *testing.T) {
	store := CreateTestStore()
	project, workflow, run := createWorkflowRunFixture(t, store)

	task1, err := store.CreateTask(db.Task{ProjectID: project.ID, TemplateID: workflow.Nodes[0].TemplateID}, 0)
	require.NoError(t, err)
	task2, err := store.CreateTask(db.Task{ProjectID: project.ID, TemplateID: workflow.Nodes[0].TemplateID}, 0)
	require.NoError(t, err)

	// First write wins.
	updated, err := store.SetWorkflowRunRootTask(project.ID, run.ID, task1.ID)
	require.NoError(t, err)
	assert.True(t, updated)

	// A second writer must not overwrite the recorded root.
	updated, err = store.SetWorkflowRunRootTask(project.ID, run.ID, task2.ID)
	require.NoError(t, err)
	assert.False(t, updated)

	reloaded, err := store.GetWorkflowRunByID(project.ID, run.ID)
	require.NoError(t, err)
	require.NotNil(t, reloaded.RootTaskID)
	assert.Equal(t, task1.ID, *reloaded.RootTaskID)
}

func TestResolveWorkflowApprovalIfPending(t *testing.T) {
	store := CreateTestStore()
	project, workflow, run := createWorkflowRunFixture(t, store)

	approval, err := store.CreateWorkflowApproval(db.WorkflowApproval{
		ProjectID:      project.ID,
		WorkflowRunID:  run.ID,
		WorkflowNodeID: workflow.Nodes[0].ID,
		Status:         db.WorkflowApprovalPending,
		Created:        time.Now().UTC(),
	})
	require.NoError(t, err)

	// The first resolution wins.
	now := time.Now().UTC()
	approval.Status = db.WorkflowApprovalApproved
	approval.Resolved = &now
	resolved, err := store.ResolveWorkflowApprovalIfPending(approval)
	require.NoError(t, err)
	assert.True(t, resolved)

	// A concurrent (now stale) rejection loses without overwriting.
	approval.Status = db.WorkflowApprovalRejected
	resolved, err = store.ResolveWorkflowApprovalIfPending(approval)
	require.NoError(t, err)
	assert.False(t, resolved, "a resolved approval must not be re-resolved")

	reloaded, err := store.GetWorkflowApproval(project.ID, run.ID, workflow.Nodes[0].ID)
	require.NoError(t, err)
	assert.Equal(t, db.WorkflowApprovalApproved, reloaded.Status)
}

func TestGetActiveWorkflowRuns(t *testing.T) {
	store := CreateTestStore()
	project, workflow, run := createWorkflowRunFixture(t, store)

	statuses := []db.WorkflowRunStatus{
		db.WorkflowRunApproval,
		db.WorkflowRunSuccess,
		db.WorkflowRunFailed,
		db.WorkflowRunStopped,
	}
	for _, status := range statuses {
		_, err := store.CreateWorkflowRun(db.WorkflowRun{
			ProjectID:          project.ID,
			WorkflowTemplateID: workflow.ID,
			Status:             status,
		})
		require.NoError(t, err)
	}

	active, err := store.GetActiveWorkflowRuns()
	require.NoError(t, err)
	require.Len(t, active, 2, "only running and approval runs are active")
	assert.Equal(t, db.WorkflowRunRunning, active[0].Status)
	assert.Equal(t, run.ID, active[0].ID)
	assert.Equal(t, db.WorkflowRunApproval, active[1].Status)
}

// TestCreateWorkflowApprovalDuplicate documents the unique index on
// (workflow_run_id, workflow_node_id): a concurrent duplicate insert fails
// instead of producing two pending approvals for one node.
func TestCreateWorkflowApprovalDuplicate(t *testing.T) {
	store := CreateTestStore()
	project, workflow, run := createWorkflowRunFixture(t, store)

	approval := db.WorkflowApproval{
		ProjectID:      project.ID,
		WorkflowRunID:  run.ID,
		WorkflowNodeID: workflow.Nodes[0].ID,
		Status:         db.WorkflowApprovalPending,
		Created:        time.Now().UTC(),
	}

	_, err := store.CreateWorkflowApproval(approval)
	require.NoError(t, err)

	_, err = store.CreateWorkflowApproval(approval)
	assert.Error(t, err, "the (run, node) unique index must reject duplicates")
}
