# Semaphore UI — Automated QA Runner Prompt

Use this prompt in the Claude Chrome extension (with computer use enabled).
Start the HTTP server first, then paste the prompt below into Claude.

## Prerequisites

```bash
# Run Semaphore
docker run -d --name semaphore-qa \
  -p 3000:3000 \
  -e SEMAPHORE_DB_DIALECT=sqlite \
  -e SEMAPHORE_ADMIN=admin \
  -e SEMAPHORE_ADMIN_PASSWORD=changeme \
  -e SEMAPHORE_ADMIN_NAME=Admin \
  -e SEMAPHORE_ADMIN_EMAIL=admin@localhost \
  -e SEMAPHORE_ACCESS_KEY_ENCRYPTION=gs72mPntFATGJs9qK0pQ0rKtfidlexiMjYCH9gWKhTU= \
  semaphoreui/semaphore:latest
```

No local server needed — test cases are read directly from GitHub. But if you want to test local test cases,
you can use following command:

cd test/test-cases
python3 -m http.server 8080

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
**Test cases:** https://raw.githubusercontent.com/semaphoreui/semaphore/refs/heads/develop/test/test-cases/

**Credentials:**
- Username: `admin`
- Password: `changeme`

**Telegram notifications:**
- Bot token: `YOUR_BOT_TOKEN`
- Chat ID: `YOUR_CHAT_ID`

## Instructions

1. Fetch the test case index from:
   https://raw.githubusercontent.com/semaphoreui/semaphore/refs/heads/develop/test/test-cases/README.md
2. Parse the index table to extract all test case filenames (e.g., `TC-001-admin-login-valid.md`).
3. Starting from TC-001, for each test case:
   a. Fetch the test case from GitHub raw URL:
      `https://raw.githubusercontent.com/semaphoreui/semaphore/refs/heads/develop/test/test-cases/<filename>`
   b. Read and understand the preconditions, steps, and expected results.
   c. Switch to the Semaphore UI tab (http://localhost:3000).
   d. Execute each step described in the test case.
   e. After completing all steps, evaluate whether expected results are met.
   f. Record the result as PASS or FAIL with a brief note if FAIL.
   g. **If FAIL:** immediately send a Telegram notification (see below).
4. Continue to the next test case. Respect preconditions — if a test case depends on a prior one (e.g., "logged in as admin"), ensure that state exists.
5. After all test cases are complete, produce a summary report.
6. Send the final summary to Telegram.
7. Send the report as a document to Telegram (see below).

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

## Saving and sending results via Telegram

After producing the final summary report, generate an HTML report, convert it to
both a screenshot (JPG) and a PDF, then send both to Telegram.

### Step 1 — Build the HTML report

```javascript
const reportHtml = `<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>QA Report</title>
<style>body{font-family:sans-serif;margin:2em;background:#fff}
table{border-collapse:collapse;width:100%}
th,td{border:1px solid #ccc;padding:8px;text-align:left}
th{background:#f5f5f5}
.pass{color:#2e7d32;font-weight:bold}.fail{color:#c62828;font-weight:bold}
.skip{color:#f57c00;font-weight:bold}.blocked{color:#757575;font-weight:bold}</style>
</head><body>
<h1>QA Run Summary — Semaphore UI</h1>
<p>Date: ${new Date().toISOString().slice(0,10)}</p>
<p>Build: <VERSION></p>
<table>
<tr><th>#</th><th>ID</th><th>Title</th><th>Result</th><th>Notes</th></tr>
<!-- INSERT ROWS HERE with class="pass|fail|skip|blocked" on Result cells -->
</table>
<p><strong>Total: X PASS / Y FAIL / Z SKIPPED / W BLOCKED</strong></p>
</body></html>`;
```

### Step 2 — Render to JPG and send as photo with the summary message

```javascript
// Render HTML in an iframe and capture as image using html2canvas approach
const iframe = document.createElement('iframe');
iframe.style.cssText = 'position:fixed;top:0;left:0;width:1200px;height:900px;opacity:0;z-index:-1';
document.body.appendChild(iframe);
iframe.contentDocument.open();
iframe.contentDocument.write(reportHtml);
iframe.contentDocument.close();

// Wait for render, then capture via canvas
await new Promise(r => setTimeout(r, 500));
const canvas = document.createElement('canvas');
const rect = iframe.contentDocument.body.getBoundingClientRect();
canvas.width = 1200;
canvas.height = Math.max(rect.height + 40, 900);
const ctx = canvas.getContext('2d');
ctx.fillStyle = '#fff';
ctx.fillRect(0, 0, canvas.width, canvas.height);

// Draw using SVG foreignObject
const svgData = `<svg xmlns="http://www.w3.org/2000/svg" width="${canvas.width}" height="${canvas.height}">
  <foreignObject width="100%" height="100%">
    ${new XMLSerializer().serializeToString(iframe.contentDocument.documentElement)}
  </foreignObject>
</svg>`;
const img = new Image();
img.src = 'data:image/svg+xml;charset=utf-8,' + encodeURIComponent(svgData);
await new Promise(r => { img.onload = r; });
ctx.drawImage(img, 0, 0);
document.body.removeChild(iframe);

// Convert canvas to JPG blob
const jpgBlob = await new Promise(r => canvas.toBlob(r, 'image/jpeg', 0.9));

// Send photo to Telegram with summary caption
const photoForm = new FormData();
photoForm.append('chat_id', '<CHAT_ID>');
photoForm.append('caption', '📋 *QA Run Complete — Semaphore UI*\n' +
  'Date: ' + new Date().toISOString().slice(0,10) + '\n' +
  '✅ Passed: X | ❌ Failed: Y | ⏭ Skipped: Z | 🚫 Blocked: W');
photoForm.append('parse_mode', 'Markdown');
photoForm.append('photo', new File([jpgBlob], 'qa-report.jpg', {type: 'image/jpeg'}));
await fetch('https://api.telegram.org/bot<BOT_TOKEN>/sendPhoto', {
  method: 'POST',
  body: photoForm
});
```

### Step 3 — Print to PDF and send as document

1. Open a new tab with the HTML report content.
2. Use Ctrl+P (Cmd+P on Mac) to open the print dialog.
3. Select "Save as PDF" as the destination.
4. Click "Save" — save as `semaphore-qa-report-<YYYY-MM-DD>.pdf`.

Then send the PDF to Telegram:

```javascript
const input = document.createElement('input');
input.type = 'file';
input.accept = '.pdf';
input.onchange = async () => {
  const file = input.files[0];
  const formData = new FormData();
  formData.append('chat_id', '<CHAT_ID>');
  formData.append('caption', '📋 QA Report — Semaphore UI — ' + new Date().toISOString().slice(0,10));
  formData.append('document', file);
  await fetch('https://api.telegram.org/bot<BOT_TOKEN>/sendDocument', {
    method: 'POST',
    body: formData
  });
};
input.click(); // Select the PDF you just saved
```

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