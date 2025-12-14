# Semaphore UI Test Plan

> For LLM/MCP execution instructions, safety rules, and reporting templates, see `AGENT.md`.

1. For each test case create a new project with demo flag.
2. After each test case delete the project.

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

---

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
