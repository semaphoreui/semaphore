# Test Plan for Semaphore UI using Playwright MCP

> For LLM/MCP execution instructions, safety rules, and reporting templates, see `AGENT.md`.

## Test Case 1: Project Lifecycle Management
**Objective**: Verify complete project creation, update, and deletion workflow
**Steps**:
1. Verify project creation and retrieve project details
2. Update project properties
3. Verify updates were applied correctly

**Expected Results**:
- Project created successfully with correct initial values
- Project details retrieved accurately
- Updates applied and reflected in project data
- Project deleted successfully
