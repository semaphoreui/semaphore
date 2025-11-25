# Workflow Feature Documentation

## Overview

The Workflow feature allows you to chain multiple tasks together into complex automation pipelines, similar to AWX/Tower Workflow Templates. This enables sequential and parallel task execution with conditional transitions.

## Features

### Core Functionality
- **Visual Workflow Editor**: Drag-and-drop interface for building workflows
- **Multiple Node Types**:
  - **Task Nodes**: Execute Semaphore templates
  - **Pause Nodes**: Add delays between tasks
  - **Approval Nodes**: Manual approval points (MVP: auto-approve)
- **Conditional Transitions**:
  - `success`: Execute next node only if current succeeds
  - `failure`: Execute next node only if current fails
  - `always`: Always execute next node
- **Parallel Execution**: Multiple nodes can run simultaneously
- **Real-time Monitoring**: Track workflow execution with live status updates
- **Execution History**: View past workflow runs and their results

### Database Schema

The workflow feature adds the following tables:

```sql
project__workflow           -- Workflow definitions
project__workflow_node      -- Nodes in the workflow
project__workflow_link      -- Connections between nodes
project__workflow_run       -- Execution history
project__workflow_node_run  -- Individual node execution tracking
```

### API Endpoints

```
GET    /api/project/{id}/workflows                    -- List workflows
POST   /api/project/{id}/workflows                    -- Create workflow
GET    /api/project/{id}/workflows/{workflow_id}      -- Get workflow
PUT    /api/project/{id}/workflows/{workflow_id}      -- Update workflow
DELETE /api/project/{id}/workflows/{workflow_id}      -- Delete workflow

POST   /api/project/{id}/workflows/{workflow_id}/run  -- Run workflow
GET    /api/project/{id}/workflows/{workflow_id}/runs -- List runs
GET    /api/project/{id}/workflows/{workflow_id}/runs/{run_id} -- Get run details
POST   /api/project/{id}/workflows/{workflow_id}/runs/{run_id}/stop -- Stop run

GET    /api/project/{id}/workflow_runs                -- All workflow runs for project
```

## Usage Guide

### Creating a Workflow

1. Navigate to **Workflows** in your project
2. Click **New Workflow**
3. Enter a name and description
4. Click **Save**

### Building the Workflow

1. **Add Nodes**:
   - Click on node types in the left sidebar (Task, Pause, Approval)
   - Nodes appear on the canvas

2. **Configure Nodes**:
   - Click a node to select it
   - Properties panel appears on the right
   - For Task nodes: Select the template to execute
   - For Pause nodes: Set duration in seconds

3. **Connect Nodes**:
   - Drag from the output connector (right side) of one node
   - Drop on the input connector (left side) of another node
   - Click the connection circle to set condition (success/failure/always)

4. **Position Nodes**:
   - Drag nodes to organize your workflow visually
   - Use the grid for alignment

5. **Save Workflow**:
   - Click the **Save** button in the toolbar

### Running a Workflow

1. From the workflow editor or list, click **Run**
2. A new workflow run is created
3. The system automatically:
   - Finds the start node (node with no incoming connections)
   - Executes nodes in order
   - Follows conditional transitions
   - Runs parallel branches simultaneously

### Monitoring Execution

1. Click **Runs** in the workflow editor
2. View all execution history
3. Click on a run to see:
   - Overall status and timing
   - Visual representation with node states
   - Detailed node execution information
   - Links to underlying task logs

## Example Workflow

### Simple CI/CD Pipeline

```
[Build] --success--> [Test] --success--> [Deploy to Staging] --success--> [Approval] --success--> [Deploy to Production]
           |                      |
           |--failure--> [Notify Team]
```

1. **Build Node**: Runs build template
2. **Test Node**: Runs test suite (only if build succeeds)
3. **Deploy to Staging**: Deploys to staging environment (only if tests pass)
4. **Approval Node**: Manual approval point (MVP: auto-approves)
5. **Deploy to Production**: Final deployment (only after approval)
6. **Notify Team**: Sends notification (runs on any failure)

### Parallel Execution Example

```
             /--> [Database Migration]
[Checkout]  ---> [Unit Tests]         --> [Integration Tests] --> [Deploy]
             \--> [Linting]
```

1. **Checkout**: Clone repository
2. **Parallel Phase**: Database Migration, Unit Tests, and Linting run simultaneously
3. **Integration Tests**: Runs after all parallel tasks complete
4. **Deploy**: Final deployment

## Architecture

### Backend Components

1. **Database Models** (`db/Workflow.go`):
   - `Workflow`, `WorkflowNode`, `WorkflowLink`
   - `WorkflowRun`, `WorkflowNodeRun`

2. **SQL Store** (`db/sql/workflow.go`):
   - CRUD operations for all workflow entities
   - Efficient queries with proper indexing

3. **API Handlers** (`api/projects/workflow.go`):
   - RESTful endpoints
   - Request validation
   - Integration with workflow engine

4. **Workflow Engine** (`services/workflows/workflow_engine.go`):
   - Workflow execution logic
   - Node scheduling and execution
   - State management
   - Parallel execution handling

### Frontend Components

1. **Workflow List** (`web/src/views/project/Workflows.vue`):
   - Display all workflows
   - Create/edit/delete operations

2. **Workflow Editor** (`web/src/views/project/WorkflowEditor.vue`):
   - Visual drag-and-drop editor
   - Node management
   - Connection management
   - Properties panel

3. **Workflow Runs** (`web/src/views/project/WorkflowRuns.vue`):
   - Execution history
   - Real-time status updates
   - Visual workflow state display
   - Link to task logs

## Execution Flow

1. **Workflow Start**:
   ```
   User clicks "Run" → API creates WorkflowRun → Engine starts execution
   ```

2. **Node Execution**:
   ```
   Find start nodes → Execute node → Create Task (if type=task) → Wait for completion
   ```

3. **Transition**:
   ```
   Node completes → Check outgoing links → Evaluate conditions → Execute next nodes
   ```

4. **Completion**:
   ```
   No more running nodes → Determine final status → Update WorkflowRun
   ```

## Configuration

### Node Configuration

**Task Node**:
```json
{
  "task_template_id": 123,
  "environment": {
    "DEPLOY_ENV": "production"
  }
}
```

**Pause Node**:
```json
{
  "duration": 30
}
```

**Approval Node**:
```json
{
  "approvers": ["user1", "user2"],
  "timeout": 3600
}
```

## MVP Limitations

Current MVP version has the following limitations:

1. **No Loops**: Cannot create circular workflows
2. **Single Start Node**: Only one node without incoming links
3. **No Advanced Logic**: No complex conditional expressions
4. **Auto-Approval**: Approval nodes auto-approve after 1 second
5. **Basic Variable Passing**: Limited support for passing data between nodes
6. **No Fan-in Logic**: Multiple paths converging don't wait for all to complete

## Future Enhancements

Planned features for future releases:

1. **Manual Approval**: Real approval workflow with notifications
2. **Advanced Conditions**: Custom expressions for transitions
3. **Variable Passing**: Full support for passing outputs between nodes
4. **Loop Support**: Repeat nodes based on conditions
5. **Fan-in/Fan-out**: Advanced parallel execution patterns
6. **Webhooks**: Trigger workflows from external events
7. **Scheduling**: Schedule workflow execution
8. **Sub-workflows**: Nest workflows within workflows
9. **Error Handling**: Retry logic and error handlers
10. **Notifications**: Alert on workflow events

## Testing

### Manual Testing Checklist

- [ ] Create a new workflow
- [ ] Add task, pause, and approval nodes
- [ ] Connect nodes with different conditions
- [ ] Save the workflow
- [ ] Run the workflow
- [ ] Monitor real-time execution
- [ ] View execution history
- [ ] Stop a running workflow
- [ ] Edit an existing workflow
- [ ] Delete a workflow

### Example Test Workflow

1. Create three templates in your project:
   - Template A: Simple echo task
   - Template B: Another echo task
   - Template C: Final echo task

2. Create workflow:
   ```
   [Task A] --success--> [Pause 5s] --always--> [Task B] --success--> [Task C]
   ```

3. Run and verify:
   - Task A executes
   - 5-second pause occurs
   - Task B executes
   - Task C executes
   - All node statuses update correctly

## Troubleshooting

### Common Issues

**Workflow won't run**:
- Ensure at least one start node exists
- Check that task nodes have templates selected
- Verify all nodes are connected properly

**Nodes stuck in pending**:
- Check task pool has available runners
- Verify templates are valid
- Check for missing dependencies (inventory, keys, etc.)

**Real-time updates not working**:
- Ensure WebSocket connection is active
- Check browser console for errors
- Refresh the page

## API Examples

### Create Workflow

```bash
curl -X POST http://localhost:3000/api/project/1/workflows \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Workflow",
    "description": "Example workflow",
    "nodes": [],
    "links": []
  }'
```

### Run Workflow

```bash
curl -X POST http://localhost:3000/api/project/1/workflows/1/run
```

### Get Run Status

```bash
curl http://localhost:3000/api/project/1/workflows/1/runs/1
```

## Integration

The workflow feature integrates seamlessly with existing Semaphore features:

- **Templates**: Workflows execute existing task templates
- **Inventory**: Task nodes use template's inventory
- **Keys**: SSH keys from template configuration
- **Environments**: Variable groups apply to workflow tasks
- **Runners**: Uses the project's runner pool
- **Permissions**: Respects project user permissions
- **Logs**: Full task logging for each node
- **Notifications**: Can trigger alerts (inherited from tasks)

## Contributing

To extend the workflow feature:

1. **Add New Node Type**:
   - Update `WorkflowNodeType` enum in `db/Workflow.go`
   - Implement execution logic in `services/workflows/workflow_engine.go`
   - Add UI components in `web/src/views/project/WorkflowEditor.vue`

2. **Add New Condition Type**:
   - Update `WorkflowLinkCondition` enum
   - Implement condition evaluation in engine
   - Update UI for condition selection

3. **Enhance Engine**:
   - Modify `executeWorkflowAsync()` for new execution patterns
   - Add new methods for advanced features
   - Ensure proper error handling and state management

## License

This feature is part of Semaphore UI and follows the same license.
