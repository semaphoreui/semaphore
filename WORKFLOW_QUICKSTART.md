# Workflow Feature - Quick Start Guide

## 🚀 Getting Started (5 Minutes)

### Step 1: Run Database Migration

The workflow tables will be created automatically on next server start, or you can manually apply the migration:

```bash
# Backup your database first!
# Migration file: db/sql/migrations/v2.18.0.sql
```

### Step 2: Start Semaphore Server

```bash
# Build and run
go build
./semaphore
```

The workflow engine will initialize automatically with the task pool.

### Step 3: Access Workflows

1. Open Semaphore UI in your browser
2. Navigate to any project
3. Look for **"Workflows"** in the project navigation menu
4. Click to see the workflows page

## 📝 Create Your First Workflow

### Example: Simple CI/CD Pipeline

**Goal**: Build → Test → Deploy

**Steps**:

1. **Create Workflow**:
   - Click "New Workflow"
   - Name: "CI/CD Pipeline"
   - Description: "Build, test, and deploy"
   - Click "Save"

2. **Add Nodes**:
   - Click "Task" in left sidebar (3 times)
   - You'll see 3 task nodes on canvas

3. **Configure Nodes**:
   - Click first node
   - In properties panel:
     - Name: "Build"
     - Template: Select your build template
   - Click second node:
     - Name: "Test"
     - Template: Select your test template
   - Click third node:
     - Name: "Deploy"
     - Template: Select your deploy template

4. **Connect Nodes**:
   - Drag from Build's output connector (right side)
   - Drop on Test's input connector (left side)
   - Repeat: Test → Deploy
   - Click connection circles to set conditions
   - Set all to "success" (green)

5. **Save**:
   - Click "Save" in toolbar

6. **Run**:
   - Click "Run" button
   - View execution in Runs page

## 🎯 Common Workflow Patterns

### Pattern 1: Linear Pipeline

```
[Step 1] → [Step 2] → [Step 3]
```

Use case: Sequential tasks that depend on each other

### Pattern 2: Conditional Branching

```
[Task] --success--> [On Success Task]
       --failure--> [On Failure Task]
```

Use case: Different actions based on task result

### Pattern 3: Parallel Execution

```
         /→ [Task A]
[Start] → → [Task B]
         \→ [Task C]
```

Use case: Independent tasks that can run simultaneously

### Pattern 4: With Notifications

```
[Deploy] --success--> [Success Notification]
         --failure--> [Failure Notification]
         --always---> [Cleanup]
```

Use case: Notifications and cleanup regardless of outcome

### Pattern 5: Staged Deployment

```
[Build] → [Test] → [Deploy Staging] → [Approval] → [Deploy Production]
```

Use case: Multi-stage deployment with approval gate

## 🔧 Editor Tips

### Keyboard Shortcuts
- **Click** - Select node
- **Drag** - Move node
- **Escape** - Deselect

### Node Colors
- **Green** - Task node
- **Orange** - Pause node
- **Blue** - Approval node

### Connection Colors
- **Green** - On success
- **Red** - On failure
- **Blue** - Always execute

### Best Practices
1. **Name nodes clearly** - Use descriptive names
2. **Organize layout** - Keep workflow readable
3. **Use grid** - Align nodes for cleaner look
4. **Save often** - Click Save regularly
5. **Test incrementally** - Build and test in stages

## 📊 Monitoring Workflows

### Real-time Monitoring

1. After running a workflow, click "Runs"
2. See execution history table
3. Running workflows show blue "running" chip
4. Click any row to see details

### Run Details

The details dialog shows:
- ✅ Overall status (success/failure/running/stopped)
- ⏱️ Start time and duration
- 🎨 Visual workflow diagram with colored nodes
- 📋 Node execution table
- 🔗 Links to task logs

### Node Status Colors

In the visual diagram:
- **Gray** - Not started yet
- **Orange** - Pending
- **Blue** - Running (pulsing border)
- **Green** - Success
- **Red** - Failure
- **Dark Gray** - Stopped
- **Light Gray** - Skipped

## 🛠️ Troubleshooting

### Workflow Won't Run

**Problem**: Click "Run" but nothing happens

**Solutions**:
1. Ensure workflow has at least one start node (no incoming connections)
2. Check browser console for errors
3. Verify task templates are valid
4. Check that runners are available

### Nodes Stuck in Pending

**Problem**: Nodes stay orange forever

**Solutions**:
1. Check if runners are online
2. Verify template configuration (inventory, keys, etc.)
3. Look at task logs for errors
4. Ensure task pool isn't at capacity

### Can't Connect Nodes

**Problem**: Connections won't create

**Solutions**:
1. Drag from output (right) to input (left)
2. Don't try to connect to the same node
3. Ensure you're not creating loops (MVP limitation)

### Visual Editor Issues

**Problem**: Nodes disappear or connections look wrong

**Solutions**:
1. Refresh the page
2. Clear browser cache
3. Try different browser
4. Check console for errors

## 📚 API Examples

### Create Workflow via API

```bash
curl -X POST http://localhost:3000/api/project/1/workflows \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "API Workflow",
    "description": "Created via API",
    "nodes": [
      {
        "name": "Task 1",
        "type": "task",
        "task_template_id": 123,
        "position_x": 100,
        "position_y": 100
      },
      {
        "name": "Task 2",
        "type": "task",
        "task_template_id": 124,
        "position_x": 300,
        "position_y": 100
      }
    ],
    "links": [
      {
        "from_node_id": 1,
        "to_node_id": 2,
        "condition": "success"
      }
    ]
  }'
```

### Run Workflow via API

```bash
curl -X POST http://localhost:3000/api/project/1/workflows/1/run \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Get Run Status

```bash
curl http://localhost:3000/api/project/1/workflows/1/runs/1 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 🎓 Advanced Topics

### Pause Nodes

Use pause nodes to add delays:

1. Add "Pause" node from toolbar
2. Select the node
3. Set duration in properties panel (seconds)
4. Connect to workflow

Example use case: Wait for external system to process

### Approval Nodes (MVP)

Current MVP auto-approves after 1 second:

1. Add "Approval" node
2. Connect to workflow
3. Node will pause briefly then continue

Note: Manual approval coming in future release

### Parallel Execution

Create parallel branches:

1. Have one node connect to multiple nodes
2. All target nodes start simultaneously
3. Each runs independently

Example:
```
[Task 1] → [Task 2A]
         → [Task 2B]
         → [Task 2C]
```

All three tasks (2A, 2B, 2C) run at the same time after Task 1 completes.

### Always Execute Nodes

Use "always" condition for cleanup:

1. Create connection
2. Click the connection circle
3. Select "Always"

Node will execute regardless of previous node's success/failure.

## 📖 Further Reading

- **Full Documentation**: See `WORKFLOW_FEATURE.md`
- **Implementation Details**: See `WORKFLOW_IMPLEMENTATION_SUMMARY.md`
- **API Reference**: See API section in `WORKFLOW_FEATURE.md`

## 🆘 Getting Help

1. Check troubleshooting section above
2. Review full documentation
3. Check browser console for errors
4. Check server logs
5. Open GitHub issue with:
   - Steps to reproduce
   - Expected vs actual behavior
   - Screenshots if applicable
   - Browser and version
   - Semaphore version

## ✅ Quick Checklist

Before reporting issues:
- [ ] Database migration applied
- [ ] Server restarted after migration
- [ ] At least one task template exists in project
- [ ] Workflow has start node (no incoming links)
- [ ] Task nodes have templates selected
- [ ] Runners are online and available
- [ ] Browser console checked for errors
- [ ] Page refreshed

---

**Happy Workflow Building! 🎉**

For more examples and detailed explanations, see the full documentation in `WORKFLOW_FEATURE.md`.
