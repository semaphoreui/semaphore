# Semaphore UI Test Plan

> For LLM/MCP execution instructions, safety rules, and reporting templates, see `AGENT.md`.

## Test Case 1: Project Lifecycle Management
**Objective**: Verify complete project creation, update, and deletion workflow
**Steps**:
1. Create a new project with specific configuration
2. Verify project creation and retrieve project details
3. Update project properties (name, max_parallel_tasks)
4. Verify updates were applied correctly
5. Clean up: Delete the test project

**Expected Results**:
- Project created successfully with correct initial values
- Project details retrieved accurately
- Updates applied and reflected in project data
- Project deleted successfully

---

## Test Case 2: Template Execution and Task Monitoring
**Objective**: Execute a template task and monitor its execution lifecycle
**Steps**:
1. Identify an existing successful template (e.g., "Ping semaphoreui.com")
2. Execute the template task
3. Monitor task execution status
4. Retrieve task details and verify completion
5. Verify task output/logs are accessible

**Expected Results**:
- Task starts successfully
- Task status transitions correctly (running → success/error)
- Task details are accurate and complete
- Task output is accessible

---

## Test Case 3: Failed Task Analysis
**Objective**: Verify error handling and failure analysis capabilities
**Steps**:
1. Identify a template with recent failures
2. Retrieve failed task details
3. Analyze task failure using analysis tool
4. Verify error information is comprehensive
5. Check if failure patterns can be identified

**Expected Results**:
- Failed tasks are identified correctly
- Failure analysis provides detailed error context
- Error messages are clear and actionable
- Failure patterns are detectable

---

## Test Case 4: Environment and Inventory Management
**Objective**: Test CRUD operations for environments and inventory
**Steps**:
1. Create a new environment with environment variables
2. Verify environment creation
3. Update environment variables
4. Create a new inventory item
5. Verify inventory creation
6. Update inventory content
7. Clean up: Delete created resources

**Expected Results**:
- Environment created with correct variables
- Environment updates work correctly
- Inventory created with correct content
- Inventory updates work correctly
- Resources deleted successfully

---

## Test Case 5: Bulk Task Operations and Filtering
**Objective**: Test task filtering, bulk operations, and task management
**Steps**:
1. List tasks with different status filters
2. Filter tasks by status (success, error, running)
3. Get waiting tasks (if any)
4. Verify task filtering works correctly
5. Test bulk task analysis (if multiple failures exist)

**Expected Results**:
- Task listing works with filters
- Status filtering returns correct tasks
- Waiting tasks are identified
- Bulk operations work correctly
