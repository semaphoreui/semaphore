package tasks

import (
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db/bolt"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
)

func TestWorkflowTraversal_OnFailureDoesNotTraverseOnSuccess(t *testing.T) {
	pool := &TaskPool{}
	workflow := db.WorkflowTemplate{
		Nodes: []db.WorkflowNode{{ID: 1}, {ID: 2}, {ID: 4}},
		Edges: []db.WorkflowEdge{
			{SourceNodeID: 1, DestinationNodeID: 2, Condition: db.WorkflowEdgeOnSuccess},
			{SourceNodeID: 1, DestinationNodeID: 4, Condition: db.WorkflowEdgeOnFailure},
		},
	}

	statusByNodeID := map[int]task_logger.TaskStatus{1: task_logger.TaskFailStatus}

	readyB, blockedB := pool.isWorkflowNodeReady(workflow, workflow.Nodes[1], statusByNodeID)
	if readyB || !blockedB {
		t.Fatal("node B should not be traversed when A fails")
	}

	readyD, blockedD := pool.isWorkflowNodeReady(workflow, workflow.Nodes[2], statusByNodeID)
	if !readyD || blockedD {
		t.Fatal("node D should be traversed when A fails via on_failure")
	}
}

func TestWorkflowConvergenceAll(t *testing.T) {
	pool := &TaskPool{}
	workflow := db.WorkflowTemplate{
		Nodes: []db.WorkflowNode{{ID: 1}, {ID: 2}, {ID: 3, ConvergenceMode: db.WorkflowConvergenceAll}},
		Edges: []db.WorkflowEdge{
			{SourceNodeID: 1, DestinationNodeID: 3, Condition: db.WorkflowEdgeOnSuccess},
			{SourceNodeID: 2, DestinationNodeID: 3, Condition: db.WorkflowEdgeOnSuccess},
		},
	}

	ready, blocked := pool.isWorkflowNodeReady(workflow, workflow.Nodes[2], map[int]task_logger.TaskStatus{
		1: task_logger.TaskSuccessStatus,
	})
	if ready || blocked {
		t.Fatal("node should wait for all parents")
	}

	ready, blocked = pool.isWorkflowNodeReady(workflow, workflow.Nodes[2], map[int]task_logger.TaskStatus{
		1: task_logger.TaskSuccessStatus,
		2: task_logger.TaskFailStatus,
	})
	if ready || !blocked {
		t.Fatal("node should be blocked when a parent does not satisfy the edge condition")
	}

	ready, blocked = pool.isWorkflowNodeReady(workflow, workflow.Nodes[2], map[int]task_logger.TaskStatus{
		1: task_logger.TaskSuccessStatus,
		2: task_logger.TaskSuccessStatus,
	})
	if !ready || blocked {
		t.Fatal("node should run after all converging parents satisfy conditions")
	}
}

func TestWorkflowConvergenceAny(t *testing.T) {
	pool := &TaskPool{}
	workflow := db.WorkflowTemplate{
		Nodes: []db.WorkflowNode{{ID: 1}, {ID: 2}, {ID: 3, ConvergenceMode: db.WorkflowConvergenceAny}},
		Edges: []db.WorkflowEdge{
			{SourceNodeID: 1, DestinationNodeID: 3, Condition: db.WorkflowEdgeOnSuccess},
			{SourceNodeID: 2, DestinationNodeID: 3, Condition: db.WorkflowEdgeOnSuccess},
		},
	}

	ready, blocked := pool.isWorkflowNodeReady(workflow, workflow.Nodes[2], map[int]task_logger.TaskStatus{
		1: task_logger.TaskSuccessStatus,
	})
	if !ready || blocked {
		t.Fatal("node should run when any parent satisfies condition")
	}

	ready, blocked = pool.isWorkflowNodeReady(workflow, workflow.Nodes[2], map[int]task_logger.TaskStatus{
		1: task_logger.TaskFailStatus,
	})
	if ready || blocked {
		t.Fatal("node should wait when a matching parent could still finish")
	}

	ready, blocked = pool.isWorkflowNodeReady(workflow, workflow.Nodes[2], map[int]task_logger.TaskStatus{
		1: task_logger.TaskFailStatus,
		2: task_logger.TaskFailStatus,
	})
	if ready || !blocked {
		t.Fatal("node should be blocked when all parents finished and none matched")
	}
}

func TestWorkflowApprovalProgressionAndEdgeMatching(t *testing.T) {
	store := bolt.CreateTestStore()

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
	tplRoot, err := store.CreateTemplate(db.Template{ProjectID: project.ID, RepositoryID: repo.ID, Name: "A", Playbook: "a.yml"})
	if err != nil {
		t.Fatal(err)
	}
	tplSuccess, err := store.CreateTemplate(db.Template{ProjectID: project.ID, RepositoryID: repo.ID, Name: "B", Playbook: "b.yml"})
	if err != nil {
		t.Fatal(err)
	}
	tplFailure, err := store.CreateTemplate(db.Template{ProjectID: project.ID, RepositoryID: repo.ID, Name: "C", Playbook: "c.yml"})
	if err != nil {
		t.Fatal(err)
	}

	workflow, err := store.CreateWorkflowTemplate(db.WorkflowTemplate{
		ProjectID: project.ID,
		Name:      "wf",
		Nodes: []db.WorkflowNode{
			{ID: 1, TemplateID: tplRoot.ID},
			{ID: 2, Kind: db.WorkflowNodeApprovalKind},
			{ID: 3, TemplateID: tplSuccess.ID},
			{ID: 4, TemplateID: tplFailure.ID},
		},
		Edges: []db.WorkflowEdge{
			{SourceNodeID: 1, DestinationNodeID: 2, Condition: db.WorkflowEdgeOnSuccess},
			{SourceNodeID: 2, DestinationNodeID: 3, Condition: db.WorkflowEdgeOnSuccess},
			{SourceNodeID: 2, DestinationNodeID: 4, Condition: db.WorkflowEdgeOnFailure},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	pool := TaskPool{
		register:               make(chan *TaskRunner, 10),
		store:                  store,
		state:                  NewMemoryTaskStateStore(),
		inventoryService:       &InventoryServiceMock{},
		encryptionService:      &EncryptionServiceMock{},
		keyInstallationService: &KeyInstallerMock{},
		logWriteService:        &mockLogWriteService{},
	}

	run, err := pool.StartWorkflow(workflow, nil)
	if err != nil {
		t.Fatal(err)
	}

	tasks, err := store.GetProjectTasks(project.ID, db.RetrieveQueryParams{})
	if err != nil {
		t.Fatal(err)
	}
	var rootTask db.TaskWithTpl
	for _, task := range tasks {
		if task.WorkflowRunID != nil && *task.WorkflowRunID == run.ID && task.WorkflowNodeID != nil && *task.WorkflowNodeID == workflow.Nodes[0].ID {
			rootTask = task
			break
		}
	}
	if rootTask.ID == 0 {
		t.Fatal("expected root workflow task")
	}
	rootTask.Status = task_logger.TaskSuccessStatus
	if err = store.UpdateTask(rootTask.Task); err != nil {
		t.Fatal(err)
	}
	if err = pool.HandleWorkflowTaskCompletion(rootTask.Task); err != nil {
		t.Fatal(err)
	}

	approvalNodeID := workflow.Nodes[1].ID
	approval, err := store.GetWorkflowApproval(project.ID, run.ID, approvalNodeID)
	if err != nil || approval.Status != db.WorkflowApprovalPending {
		t.Fatal("expected pending approval after root completion")
	}

	approval.Status = db.WorkflowApprovalApproved
	now := time.Now().UTC()
	approval.Resolved = &now
	if err = store.UpdateWorkflowApproval(approval); err != nil {
		t.Fatal(err)
	}
	if err = pool.ProgressWorkflowRun(project.ID, run.ID, nil); err != nil {
		t.Fatal(err)
	}

	tasks, err = store.GetProjectTasks(project.ID, db.RetrieveQueryParams{})
	if err != nil {
		t.Fatal(err)
	}
	foundSuccessBranch := false
	foundFailureBranch := false
	for _, task := range tasks {
		if task.WorkflowRunID == nil || *task.WorkflowRunID != run.ID || task.WorkflowNodeID == nil {
			continue
		}
		if *task.WorkflowNodeID == workflow.Nodes[2].ID {
			foundSuccessBranch = true
		}
		if *task.WorkflowNodeID == workflow.Nodes[3].ID {
			foundFailureBranch = true
		}
	}
	if !foundSuccessBranch || foundFailureBranch {
		t.Fatal("approved node should only enqueue on_success branch")
	}
}
