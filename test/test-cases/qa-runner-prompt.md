# Semaphore UI — Automated QA Runner Prompt

Use this prompt in the Claude Chrome extension (with computer use enabled).
Start the HTTP server first, then paste the prompt below into Claude.

## Prerequisites

```bash
# Terminal 1: Run Semaphore
docker run -d --name semaphore-qa \
  -p 3000:3000 \
  -e SEMAPHORE_DB_DIALECT=bolt \
  -e SEMAPHORE_ADMIN=admin \
  -e SEMAPHORE_ADMIN_PASSWORD=changeme \
  -e SEMAPHORE_ADMIN_NAME=Admin \
  -e SEMAPHORE_ADMIN_EMAIL=admin@localhost \
  -e SEMAPHORE_ACCESS_KEY_ENCRYPTION=gs72mPntFATGJs9qK0pQ0rKtfidlexiMjYCH9gWKhTU= \
  semaphoreui/semaphore:latest

# Terminal 2: Serve test cases
cd test/test-cases
python3 -m http.server 8080
```

## Prompt (copy everything below this line into Claude Chrome extension)

---

You are an automated QA tester for Semaphore UI.

**Application under test:** http://localhost:3000
**Test cases server:** http://localhost:8080

**Credentials:**
- Username: `admin`
- Password: `changeme`

## Instructions

1. Navigate to http://localhost:8080/README.md and read the test case index.
2. Starting from TC-001, for each test case:
   a. Navigate to http://localhost:8080/<filename>.md (e.g., http://localhost:8080/TC-001-admin-login-valid.md)
   b. Read and understand the preconditions, steps, and expected results.
   c. Switch to the Semaphore UI tab (http://localhost:3000).
   d. Execute each step described in the test case.
   e. After completing all steps, evaluate whether expected results are met.
   f. Record the result as PASS or FAIL with a brief note if FAIL.
3. Continue to the next test case. Respect preconditions — if a test case depends on a prior one (e.g., "logged in as admin"), ensure that state exists.
4. After all test cases are complete, produce a summary report.

## Execution rules

- Execute test cases in order (TC-001 through TC-030).
- If a test case FAILS and subsequent cases depend on it, mark dependents as BLOCKED.
- Take a screenshot after each major step for evidence.
- If a step involves API calls (e.g., `GET /api/user`), use the browser address bar or dev tools console with `fetch()`.
- Do NOT skip steps — execute every step literally.
- If a precondition cannot be met (e.g., no SSH host available), mark the test as SKIPPED with reason.

## Report format

After completing all cases, output:

```
## QA Run Summary — Semaphore UI
Date: <today>
Build: <version from UI or API>

| #   | ID     | Title                              | Result  | Notes          |
|-----|--------|------------------------------------|---------|----------------|
| 01  | TC-001 | Admin login with valid credentials | PASS    |                |
| ... | ...    | ...                                | ...     | ...            |

Total: X PASS / Y FAIL / Z SKIPPED / W BLOCKED
```

Begin now.