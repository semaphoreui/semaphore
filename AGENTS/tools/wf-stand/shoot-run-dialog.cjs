// Start a run of a workflow with a survey through the UI and screenshot the survey dialog and
// the run view with its "Variables" strip. Fills `image_tag` (the required variable of the
// seed-survey.sh workflow) and picks the first enum option.
// usage: NODE_PATH=~/.npm/_npx/9833c18b2d85bc59/node_modules node shoot-run-dialog.cjs <workflow_id> [out_dir] [--dark]
const { chromium } = require('playwright');
const args = process.argv.slice(2);
const dark = args.includes('--dark');
const [WF, OUT_DIR = '/tmp/semaphore-stand/shots-survey'] = args.filter((a) => a !== '--dark');
const BASE = process.env.BASE || 'http://localhost:3100';
const PID = process.env.PID || 1;
const IMAGE_TAG = process.env.IMAGE_TAG || '1.2.3';
(async () => {
  const browser = await chromium.launch({ channel: 'chrome' });
  const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 }, deviceScaleFactor: 2 });
  await ctx.request.post(`${BASE}/api/auth/login`, { data: { auth: 'admin', password: 'admin123' } });
  const page = await ctx.newPage();
  const errors = [];
  page.on('pageerror', (e) => errors.push(e.message));
  page.on('console', (m) => { if (m.type() === 'error') errors.push(m.text()); });
  await page.addInitScript((d) => { if (d) localStorage.setItem('darkMode', '1'); else localStorage.removeItem('darkMode'); }, dark);
  await page.goto(`${BASE}/project/${PID}/workflows/${WF}`);
  await page.click('[data-testid="workflow-run"]');
  const dialog = page.locator('.v-dialog--active');
  await dialog.waitFor({ timeout: 5000 });
  await page.waitForTimeout(500);
  // Required field left empty: the dialog must refuse.
  await dialog.locator('[data-testid="workflow-run-dialog-run"]').click();
  await page.waitForTimeout(300);
  console.log('validation message shown:', await dialog.locator('.v-messages__message').count() > 0);
  await page.screenshot({ path: `${OUT_DIR}/run-dialog-invalid.png` });
  await dialog.getByLabel(/Image tag/).fill(IMAGE_TAG);
  const select = dialog.locator('.v-select').first();
  await select.click();
  await page.getByRole('option').first().click();
  await page.waitForTimeout(300);
  await page.screenshot({ path: `${OUT_DIR}/run-dialog.png` });
  await dialog.locator('[data-testid="workflow-run-dialog-run"]').click();
  await page.waitForURL(/\/runs\/\d+$/, { timeout: 10000 });
  const runId = page.url().match(/\/runs\/(\d+)$/)[1];
  await page.waitForSelector('.WorkflowNodeCard', { timeout: 15000 });
  await page.waitForTimeout(1500);
  const vars = await page.locator('.WorkflowRun__var').allInnerTexts();
  console.log(`run ${runId} variables: ${vars.join(' | ')}`);
  await page.screenshot({ path: `${OUT_DIR}/run-view-variables.png` });
  await browser.close();
  console.log(`saved ${OUT_DIR}/run-dialog{,-invalid}.png, run-view-variables.png${errors.length ? '\nBROWSER ERRORS:\n' + errors.join('\n') : ''}`);
  process.exit(errors.length ? 2 : 0);
})().catch((e) => { console.error(e); process.exit(1); });
