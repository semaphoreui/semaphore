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


## Test Case 2: Create a new user
**Objective**: Verify complete user creation, update, and deletion workflow
**Steps**:
1. Go to /users
2. Click the button "New User" to open the form
3. Fill required fields in the form
4. Click "Save"
5. Verify user created successfully
6. Delete the user

**Expected Results**:
- User created successfully with correct initial values
- User details retrieved accurately
- Updates applied and reflected in user data
- User deleted successfully


