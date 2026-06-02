package bolt

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
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
			{ID: 1, TemplateID: tplA.ID},
			{ID: 2, TemplateID: tplB.ID},
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

	if len(workflow.Nodes) != 2 || len(workflow.Edges) != 1 {
		t.Fatal("unexpected persisted workflow graph")
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

	if err = store.DeleteWorkflowTemplate(project.ID, workflow.ID); err != nil {
		t.Fatal(err)
	}

	if _, err = store.GetWorkflowTemplate(project.ID, workflow.ID); err == nil {
		t.Fatal("expected workflow to be deleted")
	}
}
