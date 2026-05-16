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

## Configuration

Before using the prompt, fill in your Telegram bot credentials:

1. Create a Telegram bot via [@BotFather](https://t.me/BotFather) and get the token.
2. Get your chat ID (send a message to the bot, then check
   `https://api.telegram.org/bot<TOKEN>/getUpdates`).
3. Replace `YOUR_BOT_TOKEN` and `YOUR_CHAT_ID` in the prompt below.

## Prompt (copy everything below this line into Claude Chrome extension)

---

You are an automated QA tester for Semaphore UI.

**Application under test:** http://localhost:3000
**Test cases server:** http://localhost:8080

**Credentials:**
- Username: `admin`
- Password: `changeme`

**Telegram notifications:**
- Bot token: `YOUR_BOT_TOKEN`
- Chat ID: `YOUR_CHAT_ID`

## Instructions

1. Navigate to http://localhost:8080/README.md and read the test case index.
2. Starting from TC-001, for each test case:
   a. Navigate to http://localhost:8080/<filename>.md (e.g., http://localhost:8080/TC-001-admin-login-valid.md)
   b. Read and understand the preconditions, steps, and expected results.
   c. Switch to the Semaphore UI tab (http://localhost:3000).
   d. Execute each step described in the test case.
   e. After completing all steps, evaluate whether expected results are met.
   f. Record the result as PASS or FAIL with a brief note if FAIL.
   g. **If FAIL:** immediately send a Telegram notification (see below).
3. Continue to the next test case. Respect preconditions — if a test case depends on a prior one (e.g., "logged in as admin"), ensure that state exists.
4. After all test cases are complete, produce a summary report.
5. Send the final summary to Telegram.
6. Save the report as PDF (see below).

## Telegram notifications

Send notifications using `fetch()` in the browser dev tools console:

```javascript
fetch('https://api.telegram.org/bot<BOT_TOKEN>/sendMessage', {
  method: 'POST',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    chat_id: '<CHAT_ID>',
    text: '<message>',
    parse_mode: 'Markdown'
  })
});
```

### On each FAILED test — send immediately:

```
❌ *QA FAIL: TC-XXX*
Title: <test title>
Step: <which step failed>
Expected: <what should have happened>
Actual: <what happened instead>
URL: http://localhost:3000/<page where failure occurred>
```

### After all tests complete — send summary:

```
📋 *QA Run Complete — Semaphore UI*
Date: <today>
Build: <version>

✅ Passed: X
❌ Failed: Y
⏭ Skipped: Z
🚫 Blocked: W

Failed tests:
- TC-XXX: <title> — <brief reason>
- TC-YYY: <title> — <brief reason>
```

## Saving and sending results as PDF via Telegram

After producing the final summary report, generate an HTML report and send it
as a document to Telegram:

1. Build the full HTML report as a string variable in the browser console:

```javascript
const reportHtml = `<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>QA Report</title>
<style>body{font-family:sans-serif;margin:2em}table{border-collapse:collapse;width:100%}
th,td{border:1px solid #ccc;padding:8px;text-align:left}
.pass{color:green}.fail{color:red}.skip{color:orange}.blocked{color:gray}</style>
</head><body>
<h1>QA Run Summary — Semaphore UI</h1>
<p>Date: ${new Date().toISOString().slice(0,10)}</p>
<p>Build: <VERSION></p>
<table>
<tr><th>#</th><th>ID</th><th>Title</th><th>Result</th><th>Notes</th></tr>
<!-- INSERT ROWS HERE -->
</table>
<p><strong>Total: X PASS / Y FAIL / Z SKIPPED / W BLOCKED</strong></p>
</body></html>`;
```

2. Convert to PDF blob using the browser's print API:

```javascript
const printWindow = window.open('', '_blank');
printWindow.document.write(reportHtml);
printWindow.document.close();
printWindow.print(); // Save as PDF when print dialog opens
```

3. After saving the PDF, send it to Telegram as a document:

```javascript
const formData = new FormData();
formData.append('chat_id', '<CHAT_ID>');
formData.append('caption', '📋 QA Report — Semaphore UI — ' + new Date().toISOString().slice(0,10));
formData.append('document', new File([pdfBlob], 'semaphore-qa-report-' + new Date().toISOString().slice(0,10) + '.pdf', {type: 'application/pdf'}));
fetch('https://api.telegram.org/bot<BOT_TOKEN>/sendDocument', {
  method: 'POST',
  body: formData
});
```

**Simpler alternative** — send the report as an HTML file (no print dialog needed):

```javascript
const blob = new Blob([reportHtml], {type: 'text/html'});
const formData = new FormData();
formData.append('chat_id', '<CHAT_ID>');
formData.append('caption', '📋 QA Report — Semaphore UI — ' + new Date().toISOString().slice(0,10));
formData.append('document', new File([blob], 'semaphore-qa-report-' + new Date().toISOString().slice(0,10) + '.html', {type: 'text/html'}));
fetch('https://api.telegram.org/bot<BOT_TOKEN>/sendDocument', {
  method: 'POST',
  body: formData
});
```

Use the HTML alternative if the print-to-PDF dialog cannot be automated.

## Execution rules

- Execute test cases in order (TC-001 through TC-030).
- If a test case FAILS and subsequent cases depend on it, mark dependents as BLOCKED.
- Take a screenshot after each major step for evidence.
- If a step involves API calls (e.g., `GET /api/user`), use the browser address bar or dev tools console with `fetch()`.
- Do NOT skip steps — execute every step literally.
- If a precondition cannot be met (e.g., no SSH host available), mark the test as SKIPPED with reason.
- Send Telegram notification immediately on each failure — do NOT wait until the end.

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