// Screenshots of the workflow inputs UI on the stand: the edge panel in explicit
// and by_name mode (editor of the "Inputs: explicit mapping" workflow), the run
// view with output counts on the cards, and the task dialog's Outputs panel.
// usage: NODE_PATH=~/.npm/_npx/9833c18b2d85bc59/node_modules node shoot-inputs.cjs [--dark]
//   WF=<workflow id, default 9>  OUT=<dir, default /tmp/semaphore-stand/shots-inputs>
const fs = require('fs');
const { chromium } = require('playwright');

const dark = process.argv.includes('--dark');
const suffix = dark ? 'dark' : 'light';
const BASE = process.env.BASE || 'http://localhost:3100';
const PID = process.env.PID || 1;
const WF = process.env.WF || 9;
const WF_BYNAME = process.env.WF_BYNAME || 1;
const OUT = process.env.OUT || '/tmp/semaphore-stand/shots-inputs';
fs.mkdirSync(OUT, { recursive: true });

// Drawflow selects a connection on a mousedown whose target is the path. A
// real click can not be used: the condition pill sits on the path's midpoint
// and most of a short edge lies under it, so the event is dispatched on the
// path element itself.
async function clickPath(page, sel) {
  return page.evaluate((selector) => {
    const path = document.querySelector(selector);
    if (!path) return false;
    ['mousedown', 'mouseup'].forEach((type) => path.dispatchEvent(
      new MouseEvent(type, { bubbles: true, cancelable: true, button: 0 }),
    ));
    return true;
  }, sel);
}

(async () => {
  const browser = await chromium.launch({ channel: 'chrome' });
  const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 }, deviceScaleFactor: 2 });
  await ctx.request.post(`${BASE}/api/auth/login`, { data: { auth: 'admin', password: 'admin123' } });
  const page = await ctx.newPage();
  const errors = [];
  page.on('pageerror', (e) => errors.push(e.message));
  page.on('console', (m) => { if (m.type() === 'error') errors.push(m.text()); });
  await page.addInitScript((d) => { if (d) localStorage.setItem('darkMode', '1'); else localStorage.removeItem('darkMode'); }, dark);

  // Editor: explicit edge panel.
  await page.goto(`${BASE}/project/${PID}/workflows/${WF}/edit`);
  await page.waitForSelector('.WorkflowNodeCard', { timeout: 15000 });
  await page.waitForTimeout(1200);
  const hadExplicit = await clickPath(page, 'svg.WorkflowGraph__conn--explicit path.main-path');
  if (!hadExplicit) errors.push('no explicit edge on the canvas');
  await page.waitForSelector('.WorkflowEdgeProperties', { timeout: 5000 });
  await page.waitForTimeout(600);
  await page.screenshot({ path: `${OUT}/editor-edge-explicit-${suffix}.png` });
  const panelText = await page.locator('.WorkflowEdgeProperties').innerText();
  const checks = {
    'explicit checkbox checked': await page.locator('.WorkflowEdgeProperties .v-input--checkbox .v-input--selection-controls__input input').isChecked(),
    'mapping rows shown': (await page.locator('.WorkflowEdgeProperties__row').count()) > 0,
    'hint source line': /Keys suggested from run|No finished run/.test(panelText),
  };

  // Editor: by_name edge panel — in the by_name workflow, the first edge whose
  // destination has survey variables (its panel shows the variable chips).
  await page.goto(`${BASE}/project/${PID}/workflows/${WF_BYNAME}/edit`);
  await page.waitForSelector('.WorkflowNodeCard', { timeout: 15000 });
  await page.waitForTimeout(1200);
  const selectByName = (idx) => {
    const path = document.querySelectorAll('svg.connection:not(.WorkflowGraph__conn--explicit) path.main-path')[idx];
    ['mousedown', 'mouseup'].forEach((type) => path.dispatchEvent(
      new MouseEvent(type, { bubbles: true, cancelable: true, button: 0 }),
    ));
  };
  const byNameCount = await page.locator('svg.connection:not(.WorkflowGraph__conn--explicit)').count();
  let firstWithChips = -1;
  let withMatch = -1;
  for (let i = 0; i < byNameCount && withMatch < 0; i += 1) {
    // eslint-disable-next-line no-await-in-loop
    await page.evaluate(selectByName, i);
    // eslint-disable-next-line no-await-in-loop
    await page.waitForTimeout(300);
    // eslint-disable-next-line no-await-in-loop
    const n = await page.locator('.WorkflowEdgeProperties__chips .v-chip').count();
    // eslint-disable-next-line no-await-in-loop
    const matched = await page.locator('.WorkflowEdgeProperties__chips .v-chip .mdi-check').count();
    if (n > 0 && firstWithChips < 0) firstWithChips = i;
    if (matched > 0) withMatch = i;
  }
  const chosen = withMatch >= 0 ? withMatch : firstWithChips;
  if (chosen >= 0) await page.evaluate(selectByName, chosen);
  await page.waitForTimeout(500);
  await page.screenshot({ path: `${OUT}/editor-edge-byname-${suffix}.png` });
  const chips = await page.locator('.WorkflowEdgeProperties__chips .v-chip').count();
  checks['by_name chips shown'] = chips > 0;
  checks['by_name chip matched a produced output'] = (await page.locator('.WorkflowEdgeProperties__chips .v-chip .mdi-check').count()) > 0;

  // Run view: latest run of the workflow, output counts on cards, task dialog.
  const runs = await (await ctx.request.get(`${BASE}/api/project/${PID}/workflows/${WF}/runs`)).json();
  const run = runs.sort((a, b) => b.id - a.id)[0];
  await page.goto(`${BASE}/project/${PID}/workflows/${WF}/runs/${run.id}`);
  await page.waitForSelector('.WorkflowNodeCard', { timeout: 15000 });
  await page.waitForTimeout(1500);
  await page.screenshot({ path: `${OUT}/run-outputs-${suffix}.png` });
  const cardText = await page.locator('.WorkflowGraph').innerText();
  checks['outputs count on a card'] = /\d+ outputs?/.test(cardText);

  const producer = page.locator('.WorkflowNodeCard', { hasText: 'produce renamed' }).first();
  await producer.click();
  await page.waitForSelector('.v-dialog--active', { timeout: 10000 });
  await page.waitForTimeout(800);
  // The dialog opens on the log; the Outputs panel is on the details tab.
  await page.locator('.v-dialog--active').getByRole('tab', { name: /details/i }).click();
  await page.waitForTimeout(1200);
  await page.locator('.v-dialog--active .TaskDetails__outputName').last().scrollIntoViewIfNeeded();
  await page.waitForTimeout(400);
  const dialogText = await page.locator('.v-dialog--active').innerText();
  checks['task dialog has Outputs panel'] = /Outputs/.test(dialogText) && /image_tag|tag/.test(dialogText);
  await page.screenshot({ path: `${OUT}/task-outputs-${suffix}.png` });

  await browser.close();
  Object.entries(checks).forEach(([k, v]) => console.log(`${v ? 'PASS' : 'FAIL'} ${k}`));
  console.log(`saved to ${OUT}${errors.length ? '\nBROWSER ERRORS:\n' + errors.join('\n') : ''}`);
  process.exit(errors.length || Object.values(checks).some((v) => !v) ? 2 : 0);
})().catch((e) => { console.error(e); process.exit(1); });
