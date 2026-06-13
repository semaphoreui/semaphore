# Workflow Feature Implementation

This document describes the Workflow feature implementation for Semaphore UI, similar to AWX Workflow Templates.

## Overview

The Workflow feature allows users to:
- Chain multiple tasks together into pipelines
- Create sequential and parallel execution flows
- Define conditional transitions (success/failure/always)
- Visual drag-and-drop workflow editor
- Launch workflows as a single unit
- Track execution of each step in real-time
- Re-run individual steps
- Pass variables between steps
- Add special workflow nodes (task, pause, approval)

## Architecture

### Backend Components

#### 1. Database Models (`db/Workflow.go`)
- `Workflow`: Main workflow template
- `WorkflowNode`: Nodes in the workflow (task/pause/approval)
- `WorkflowLink`: Connections between nodes with conditions
- `WorkflowRun`: Execution instance of a workflow
- `WorkflowRunNode`: Execution state of each node

#### 2. Database Migration (`db/sql/migrations/v2.18.0.sql`)
Creates tables:
- `project__workflow`
- `project__workflow_node`
- `project__workflow_link`
- `project__workflow_run`
- `project__workflow_run_node`

#### 3. Database Access Layer
- **SQL Implementation** (`db/sql/workflow.go`): Full CRUD operations for all workflow entities
- **BoltDB Implementation** (`db/bolt/workflow.go`): BoltDB support for workflows
- **Store Interface** (`db/Store.go`): `WorkflowManager` interface added

#### 4. Workflow Execution Engine (`services/workflows/workflow_engine.go`)
- `WorkflowEngine`: Manages workflow execution
- `WorkflowRunState`: Tracks running workflow state
- `NodeState`: Tracks individual node execution
- Features:
  - Finds start node (no incoming links)
  - Executes nodes sequentially/parallel based on dependencies
  - Handles conditional transitions (success/failure/always)
  - Tracks node states (pending/running/success/error/skipped)
  - Integrates with existing TaskPool for task execution

#### 5. REST API (`api/projects/workflows.go`)
Endpoints:
- `GET /project/:id/workflows` - List workflows
- `POST /project/:id/workflows` - Create workflow
- `GET /project/:id/workflows/:workflow_id` - Get workflow with nodes/links
- `PUT /project/:id/workflows/:workflow_id` - Update workflow
- `DELETE /project/:id/workflows/:workflow_id` - Delete workflow
- `PUT /project/:id/workflows/:workflow_id/nodes` - Update nodes and links
- `POST /project/:id/workflows/:workflow_id/run` - Run workflow
- `GET /project/:id/workflows/:workflow_id/runs` - List workflow runs
- `GET /project/:id/workflows/:workflow_id/runs/:run_id` - Get workflow run with node states

### Frontend Components

#### 1. Workflows List (`web/src/views/project/Workflows.vue`)
- Lists all workflows in a project
- Create/edit/delete workflows
- Navigate to workflow editor

#### 2. Workflow Editor (`web/src/views/project/workflow/WorkflowEditor.vue`)
- Visual drag-and-drop canvas
- Add nodes (task/pause/approval) via right-click
- Drag nodes to reposition
- Create links between nodes
- Configure nodes (select template for task nodes)
- Save workflow structure
- Run workflow

#### 3. Workflow Run View (`web/src/views/project/workflow/WorkflowRun.vue`)
- Visual representation of workflow execution
- Real-time status updates (polls every 2 seconds)
- Color-coded node states:
  - Blue: Running
  - Green: Success
  - Red: Error
  - Grey: Pending/Skipped
- Click nodes to view details
- Shows task links for task nodes

## Usage

### Creating a Workflow

1. Navigate to `/project/:id/workflows`
2. Click "New Workflow"
3. Enter name and description
4. Click on the canvas to add nodes
5. Select node type (Task/Pause/Approval)
6. For task nodes, select a template
7. Click "Create Link" on a node, then click target node
8. Save workflow

### Running a Workflow

1. Open workflow editor
2. Click "Run" button
3. View execution in workflow run view
4. Monitor node states in real-time

## MVP Limitations

As per requirements, the MVP has these limitations:
- No advanced logic (only success/failure conditions)
- No loops
- No advanced fan-in/fan-out logic
- Only one start node supported
- UI is minimalistic but functional
- Approval nodes auto-approve (no manual approval UI yet)
- Pause nodes have fixed 5-second duration

## Future Enhancements

Potential improvements:
- Manual approval UI for approval nodes
- Configurable pause duration
- Variable passing between nodes
- Advanced conditions and logic
- Loops and advanced control flow
- Better error handling and retry logic
- Workflow templates/cloning
- Workflow scheduling

## Files Modified/Created

### Backend
- `db/Workflow.go` - Models
- `db/sql/migrations/v2.18.0.sql` - Migration
- `db/sql/workflow.go` - SQL implementation
- `db/bolt/workflow.go` - BoltDB implementation
- `db/Store.go` - Interface definition
- `services/workflows/workflow_engine.go` - Execution engine
- `api/projects/workflows.go` - API handlers
- `api/router.go` - Route registration

### Frontend
- `web/src/views/project/Workflows.vue` - List view
- `web/src/views/project/workflow/WorkflowEditor.vue` - Editor
- `web/src/views/project/workflow/WorkflowRun.vue` - Run view
- `web/src/router/index.js` - Routes
- `web/src/lang/en.js` - Translations

## Testing

To test the feature:
1. Run database migrations
2. Start the backend server
3. Navigate to a project
4. Go to Workflows section
5. Create a workflow with multiple nodes
6. Run the workflow
7. Monitor execution

## Notes

- The workflow engine integrates seamlessly with the existing task execution system
- All workflow data is stored per-project
- Workflow runs are tracked independently from regular tasks
- The execution engine handles parallel node execution when dependencies allow
