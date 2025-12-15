# Test Plan for Semaphore UI using Playwright MCP

> For LLM/MCP execution instructions, safety rules, and reporting templates, see `AGENT.md`.


## Test Case 1: Create a new user
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


