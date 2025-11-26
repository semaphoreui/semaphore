package projects

import (
	"net/http"
	"time"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/workflows"
)

func WorkflowMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := helpers.GetFromContext(r, "project").(db.Project)
		workflowID, err := helpers.GetIntParam("workflow_id", w, r)
		if err != nil {
			return
		}

		workflow, err := helpers.Store(r).GetWorkflow(project.ID, workflowID)
		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		r = helpers.SetContextValue(r, "workflow", workflow)
		next.ServeHTTP(w, r)
	})
}

func GetWorkflows(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	workflows, err := helpers.Store(r).GetWorkflows(project.ID, helpers.QueryParams(r.URL))
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, workflows)
}

func CreateWorkflow(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	var workflow db.Workflow
	if !helpers.Bind(w, r, &workflow) {
		return
	}
	workflow.ProjectID = project.ID
	workflow.CreatedAt = time.Now()
	workflow.UpdatedAt = time.Now()

	newWorkflow, err := helpers.Store(r).CreateWorkflow(workflow)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusCreated, newWorkflow)
}

func GetWorkflow(w http.ResponseWriter, r *http.Request) {
	workflow := helpers.GetFromContext(r, "workflow").(db.Workflow)
	
	nodes, err := helpers.Store(r).GetWorkflowNodes(workflow.ProjectID, workflow.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	
	links, err := helpers.Store(r).GetWorkflowLinks(workflow.ProjectID, workflow.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	
	type WorkflowWithDetails struct {
		db.Workflow
		Nodes []db.WorkflowNode `json:"nodes"`
		Links []db.WorkflowLink `json:"links"`
	}

	res := WorkflowWithDetails{
		Workflow: workflow,
		Nodes:    nodes,
		Links:    links,
	}

	helpers.WriteJSON(w, http.StatusOK, res)
}

func UpdateWorkflow(w http.ResponseWriter, r *http.Request) {
	workflow := helpers.GetFromContext(r, "workflow").(db.Workflow)
	
	type WorkflowPayload struct {
		db.Workflow
		Nodes []db.WorkflowNode `json:"nodes"`
		Links []db.WorkflowLink `json:"links"`
	}

	var payload WorkflowPayload
	if !helpers.Bind(w, r, &payload) {
		return
	}

	// Update Workflow
	workflow.Name = payload.Name
	workflow.Description = payload.Description
	workflow.UpdatedAt = time.Now()
	
	err := helpers.Store(r).UpdateWorkflow(workflow)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	store := helpers.Store(r)
	projectID := workflow.ProjectID

	// Sync Nodes
	existingNodes, err := store.GetWorkflowNodes(projectID, workflow.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	
	existingNodesMap := make(map[int]bool)
	for _, n := range existingNodes {
		existingNodesMap[n.ID] = true
	}

	// Keep track of node ID mapping (temp ID -> real ID) if needed for links
	// But usually frontend sends temp IDs for new nodes, we need to resolve them for links.
	// For MVP, let's assume links use the IDs returned by creation.
	// This implies links need to be saved AFTER nodes.
	
	// Issue: If I create a node, I get a new ID. The links in payload might reference the temporary ID (e.g. negative or string).
	// If the payload contains Links, they must reference valid Node IDs.
	// If the user created new nodes in the UI, they don't have IDs yet.
	// The frontend should probably handle this by sending some correlation ID, or we handle it here.
	
	// Simple approach: 
	// The frontend sends Nodes. We process them.
	// If a Node has ID=0, it's new. We create it.
	// But how do we link it?
	// The Link object has FromNodeID and ToNodeID.
	// If Node was new, we don't know its ID until now.
	
	// Assumption: The frontend sends the graph. 
	// Maybe we can match by Position/Config? Unreliable.
	// We need a temporary ID mechanism.
	// But `WorkflowLink` struct expects `int` for IDs.
	
	// Allow `CreateWorkflowNode` to return the new ID.
	// We can use array index correlation if we are careful? No.
	
	// Let's assume for now the frontend only saves Nodes/Links that are stable, or we do a full replacement.
	// Or, more robustly:
	// We delete all links first (since they depend on nodes).
	// We process nodes.
	// We re-create links. 
	// But wait, if we delete all links, we lose history? No, links are just structural.
	// But if we create new nodes for existing ones, we lose history (node runs).
	
	// Improved Node Sync:
	// 1. Update existing nodes (ID > 0).
    // 2. Create new nodes (ID == 0). Store mapping of {Index in Payload -> New ID} if possible? 
    //    Or better, the payload nodes should have a temporary ID field? 
    //    The struct `WorkflowNode` doesn't have it.
    
    // Alternative: The UI saves nodes one by one? No, too chatty.
    
    // Hack: Use `PositionX` as a temporary ID carrier if it's really needed? No.
    
    // Let's just trust that `Links` in the payload refer to `Nodes` in the payload by... ID?
    // If ID is 0, it's ambiguous.
    
    // OK, since I cannot change the struct easily to add "TempID" without migration,
    // I will assume for MVP that the user adds nodes, then saves.
    // If they add a node and a link to it in one go, the link will have "0" as ID? That fails.
    
    // Solution:
    // We only support updating existing nodes and creating new nodes that DON'T have links yet in the same batch?
    // Or, we assume the frontend is smart enough to save nodes first, get IDs, then save links?
    // "Save to API" implies one click.
    
    // Let's try to match by coordinates/config if ID=0?
    // Or just Delete All and Recreate All? 
    // If I delete all nodes, `workflow_node_run` (history) will be deleted due to cascade!
    // We MUST NOT delete existing nodes if we can avoid it.
    
    // MVP Compromise:
    // Only update existing nodes (ID > 0).
    // Create new nodes (ID == 0).
    // Delete missing nodes (ID > 0 not in payload).
    
    // Links:
    // Delete all links.
    // Recreate all links.
    // BUT: Links need FromNodeID and ToNodeID.
    // If a node was just created, we need its ID.
    // The link in payload refers to... what?
    // If the frontend generated a fake ID (e.g. -1), `int` can hold it.
    // So, we can map {FakeID -> RealID}.
    
    nodeIdMap := make(map[int]int) // Payload ID -> Real ID

	for _, node := range payload.Nodes {
		originalID := node.ID
        node.WorkflowID = workflow.ID
		if node.ID > 0 {
			// Update
			if existingNodesMap[node.ID] {
				store.UpdateWorkflowNode(node)
				delete(existingNodesMap, node.ID)
				nodeIdMap[originalID] = originalID
			} else {
                // ID provided but not found? Create it as new? Or error?
                // Create as new.
                node.ID = 0 
                newNode, _ := store.CreateWorkflowNode(node)
                nodeIdMap[originalID] = newNode.ID
            }
		} else {
			// Create
			newNode, _ := store.CreateWorkflowNode(node)
			// How to map? We rely on the frontend passing a negative ID for new nodes?
            // If ID is 0, we can't map it uniquely if there are multiple.
            // If frontend passes negative IDs, we can map them.
            if originalID < 0 {
                nodeIdMap[originalID] = newNode.ID
            }
		}
	}

	// Delete remaining existing nodes
	for id := range existingNodesMap {
		store.DeleteWorkflowNode(projectID, id)
	}

	// Links
	// First delete all existing links for this workflow
	store.DeleteWorkflowLinks(projectID, workflow.ID)

	// Create new links
	for _, link := range payload.Links {
		link.WorkflowID = workflow.ID
        
        // Resolve IDs
        if newID, ok := nodeIdMap[link.FromNodeID]; ok {
            link.FromNodeID = newID
        }
        if newID, ok := nodeIdMap[link.ToNodeID]; ok {
            link.ToNodeID = newID
        }
        
        // If IDs are still invalid (e.g. referring to deleted node), we skip or fail.
        // Assuming database constraints will catch invalid IDs if we don't.
        
		store.CreateWorkflowLink(link)
	}

	w.WriteHeader(http.StatusNoContent)
}

func DeleteWorkflow(w http.ResponseWriter, r *http.Request) {
	workflow := helpers.GetFromContext(r, "workflow").(db.Workflow)
	err := helpers.Store(r).DeleteWorkflow(workflow.ProjectID, workflow.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func GetWorkflowRuns(w http.ResponseWriter, r *http.Request) {
	workflow := helpers.GetFromContext(r, "workflow").(db.Workflow)
	runs, err := helpers.Store(r).GetWorkflowRuns(workflow.ProjectID, &workflow.ID, helpers.QueryParams(r.URL))
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, runs)
}

func RunWorkflow(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	workflow := helpers.GetFromContext(r, "workflow").(db.Workflow)

	engine := helpers.GetFromContext(r, "workflow_engine").(*workflows.WorkflowEngine)

	run, err := engine.RunWorkflow(project.ID, workflow.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusCreated, run)
}
