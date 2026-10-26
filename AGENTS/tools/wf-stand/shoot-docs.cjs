// Screenshots for docs/docs/user-guide/workflows.md § Outputs and inputs, taken on the stand
// project seeded by seed-docs.sh after one successful run of "Build and deploy":
//   workflow-inputs-explicit.png  editor, the Provision -> Deploy connection (explicit mapping)
//   workflow-inputs-by-name.png   editor, the Build -> Deploy connection (by name, matched chips)
//   workflow-run-outputs.png      run view with the "N outputs" badges
//   task-outputs.png              task dialog of "Build image", Details tab, clipped to the dialog
// usage: PID=<project> WF=<workflow> OUT=<dir> NODE_PATH=~/.npm/_npx/9833c18b2d85bc59/node_modules node shoot-docs.cjs
const fs = require('fs');
const { chromium } = require('playwright');

const BASE = process.env.BASE || 'http://localhost:3100';
const PID = process.env.PID || 2;
const WF = process.env.WF || 1;
const OUT = process.env.OUT || '/tmp/semaphore-stand/shots-docs';
fs.mkdirSync(OUT, { recursive: true });

// Drawflow selects a connection on a mousedown on the path element; a real click lands on the
// condition pill, so the events are dispatched on the path itself.
const selectEdge = (selector) => {
  const path = document.querySelector(selector);
  if (!path) return false;
  ['mousedown', 'mouseup'].forEach((type) => path.dispatchEvent(
    new MouseEvent(type, { bubbles: true, cancelable: true, button: 0 }),
  ));
  return true;
};

(async () => {
  const browser = await chromium.launch({ channel: 'chrome' });
  const ctx = await browser.newContext({ viewport: { width: 1440, height: 960 }, deviceScaleFactor: 2 });
  await ctx.request.post(`${BASE}/api/auth/login`, { data: { auth: 'admin', password: 'admin123' } });
  const page = await ctx.newPage();
  const errors = [];
  page.on('pageerror', (e) => errors.push(e.message));
  await page.addInitScript(() => localStorage.removeItem('darkMode'));
  const checks = {};

  await page.goto(`${BASE}/project/${PID}/workflows/${WF}/edit`);
  await page.waitForSelector('.WorkflowNodeCard', { timeout: 15000 });
  await page.waitForTimeout(1200);
  // Collapse the palette so the canvas is wide enough for the graph next to the side panel;
  // after a connection is selected the graph is fitted again into the remaining canvas.
  await page.locator('.WorkflowEditor__sideToggle').click();
  await page.mouse.move(1000, 200);
  await page.waitForTimeout(500);
  const fitGraph = async () => {
    await page.locator('.WorkflowCanvasControls__btn').nth(2).click();
    await page.mouse.move(1000, 200);
    await page.waitForTimeout(500);
  };
  checks['explicit edge selected'] = await page.evaluate(selectEdge, 'svg.WorkflowGraph__conn--explicit path.main-path');
  await page.waitForSelector('.WorkflowEdgeProperties', { timeout: 5000 });
  await fitGraph();
  await page.screenshot({ path: `${OUT}/workflow-inputs-explicit.png` });
  checks['mapping row shown'] = (await page.locator('.WorkflowEdgeProperties__row').count()) > 0;

  // The by-name connection to show is the one whose destination has survey variables that the
  // last run matched (Build -> Deploy), not the first one on the canvas (Provision -> Build).
  const byName = await page.locator('svg.connection:not(.WorkflowGraph__conn--explicit)').count();
  checks['by-name edge selected'] = false;
  for (let i = 0; i < byName; i += 1) {
    // eslint-disable-next-line no-await-in-loop
    await page.evaluate((idx) => {
      const path = document.querySelectorAll('svg.connection:not(.WorkflowGraph__conn--explicit) path.main-path')[idx];
      ['mousedown', 'mouseup'].forEach((type) => path.dispatchEvent(
        new MouseEvent(type, { bubbles: true, cancelable: true, button: 0 }),
      ));
    }, i);
    // eslint-disable-next-line no-await-in-loop
    await page.waitForTimeout(500);
    // eslint-disable-next-line no-await-in-loop
    if ((await page.locator('.WorkflowEdgeProperties__chips .v-chip .mdi-check').count()) > 0) {
      checks['by-name edge selected'] = true;
      break;
    }
  }
  await fitGraph();
  await page.screenshot({ path: `${OUT}/workflow-inputs-by-name.png` });
  checks['by-name chips matched'] = (await page.locator('.WorkflowEdgeProperties__chips .v-chip .mdi-check').count()) > 0;

  const runs = await (await ctx.request.get(`${BASE}/api/project/${PID}/workflows/${WF}/runs`)).json();
  const run = runs.sort((a, b) => b.id - a.id)[0];
  await page.goto(`${BASE}/project/${PID}/workflows/${WF}/runs/${run.id}`);
  await page.waitForSelector('.WorkflowNodeCard', { timeout: 15000 });
  await page.waitForTimeout(1500);
  await page.screenshot({ path: `${OUT}/workflow-run-outputs.png` });
  checks['outputs badge on a card'] = /\d+ outputs?/.test(await page.locator('.WorkflowGraph').innerText());

  // Dialog: tall viewport, Details tab, clipped to the dialog plus some context.
  await page.setViewportSize({ width: 1440, height: 1700 });
  await page.locator('.WorkflowNodeCard', { hasText: 'Build image' }).first().click();
  await page.waitForSelector('.v-dialog--active', { timeout: 10000 });
  await page.waitForTimeout(800);
  await page.locator('.v-dialog--active').getByRole('tab', { name: /details/i }).click();
  await page.waitForTimeout(1200);
  const dialog = page.locator('.v-dialog--active');
  checks['task dialog has Outputs'] = /Outputs/.test(await dialog.innerText()) && /image_tag/.test(await dialog.innerText());
  const box = await dialog.boundingBox();
  const pad = 48;
  await page.screenshot({
    path: `${OUT}/task-outputs.png`,
    clip: { x: Math.max(0, box.x - pad), y: Math.max(0, box.y - pad), width: box.width + 2 * pad, height: box.height + 2 * pad },
  });

  await browser.close();
  Object.entries(checks).forEach(([k, v]) => console.log(`${v ? 'PASS' : 'FAIL'} ${k}`));
  console.log(`saved to ${OUT}${errors.length ? '\nBROWSER ERRORS:\n' + errors.join('\n') : ''}`);
  process.exit(errors.length || Object.values(checks).some((v) => !v) ? 2 : 0);
})().catch((e) => { console.error(e); process.exit(1); });
