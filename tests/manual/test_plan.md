# Semaphore UI Test Plan

> For LLM/MCP execution instructions, safety rules, and reporting templates, see `AGENT.md`.

## Test Case 1: Project Lifecycle Management
**Objective**: Verify complete project creation, update, and deletion workflow
**Steps**:
1. Verify project creation and retrieve project details
2. Update project properties (name, max_parallel_tasks)
3. Verify updates were applied correctly

**Expected Results**:
- Project created successfully with correct initial values
- Project details retrieved accurately
- Updates applied and reflected in project data
- Project deleted successfully


## Test Case 2: Template Execution and Task Monitoring
**Objective**: Execute a template task and monitor its execution lifecycle
**Steps**:
1. Execute the template "Ping semaphoreui.com"
2. Monitor task execution status
3. Retrieve task details and verify completion
4. Verify task output/logs are accessible

**Expected Results**:
- Task starts successfully
- Task status transitions correctly (running → success/error)
- Task details are accurate and complete
- Task output is accessible


## Test Case 3: Template simple pipeline
**Objective**: Verify that build task triggers deploy task
**Steps**:
1. Run "Build demo app" template
2. Wait until it finished and check if it is ok
3. Check if template "Deploy demo app to Dev" automatically triggered after "Build demo app" complete in 10 seconds.
4. Wait until it finished and check if it is ok.

**Expected Results**:
- "Deploy demo app to Dev" successfully executed