package pro_interfaces

import (
	"net/http"
)

// WorkflowController serves the workflow HTTP API. The open-source build wires a
// no-op stub (see pro/api/projects); the licensed build provides the real
// implementation (see pro_impl/api/projects). The handlers delegate workflow
// orchestration to a WorkflowService.
type WorkflowController interface {
	GetWorkflows(w http.ResponseWriter, r *http.Request)
	AddWorkflow(w http.ResponseWriter, r *http.Request)
	GetWorkflow(w http.ResponseWriter, r *http.Request)
	UpdateWorkflow(w http.ResponseWriter, r *http.Request)
	RemoveWorkflow(w http.ResponseWriter, r *http.Request)
	RunWorkflow(w http.ResponseWriter, r *http.Request)
	GetWorkflowRuns(w http.ResponseWriter, r *http.Request)
	GetWorkflowRun(w http.ResponseWriter, r *http.Request)
	GetWorkflowRunArtifacts(w http.ResponseWriter, r *http.Request)
	GetWorkflowApprovals(w http.ResponseWriter, r *http.Request)
	ResolveWorkflowApproval(w http.ResponseWriter, r *http.Request)
}
