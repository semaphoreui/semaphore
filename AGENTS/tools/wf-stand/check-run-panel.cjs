// Walks the workflow run panel (?w=<workflow>/<run>) on the stand: links, width, task dialog over it,
// back/forward, Run, approve, Stop, the old run page. Prints PASS/FAIL per check, shoots into
// /tmp/semaphore-stand/shots-panel, exits 1 on a failure.
// usage: NODE_PATH=~/.npm/_npx/9833c18b2d85bc59/node_modules node check-run-panel.cjs
const { chromium } = require('playwright');
const BASE = process.env.BASE || 'http://localhost:3100';
const PID = process.env.PID || 1; // the project seeded by seed.sh
const WF = process.env.WF || 1; // "Exec happy": task -> approval -> delay -> task
const RUN = process.env.RUN || 1; // a finished run of WF
const OTHER_WF = process.env.OTHER_WF || 2; // any other workflow of the project
const OTHER_PID = process.env.OTHER_PID || 2; // another project, without the workflow WF
const OUT = '/tmp/semaphore-stand/shots-panel';
const results = [];
function check(name, ok, detail = '') {
  results.push({ name, ok, detail });
  console.log(`${ok ? 'PASS' : 'FAIL'}  ${name}${detail ? `  [${detail}]` : ''}`);
}
(async () => {
  const browser = await chromium.launch({ channel: 'chrome' });
  const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 } });
  await ctx.request.post(`${BASE}/api/auth/login`, { data: { auth: 'admin', password: 'admin123' } });
  const page = await ctx.newPage();
  const errors = [];
  page.on('pageerror', (e) => errors.push(e.message));
  page.on('console', (m) => { if (m.type() === 'error') errors.push(m.text()); });

  const panel = page.locator('[data-testid="workflowRunPanel"]');
  const dialog = page.locator('[data-testid="taskLogDialog"]');
  const url = () => new URL(page.url());
  const panelWidth = async () => Math.round((await panel.boundingBox()).width);
  const settle = (ms = 500) => page.waitForTimeout(ms);

  // 1. a link in the list of workflows
  await page.goto(`${BASE}/project/${PID}/workflows`);
  await page.waitForSelector('[data-testid="workflows-rowLastRun"]');
  await page.locator('[data-testid="workflows-rowLastRun"]').first().click();
  await panel.waitFor();
  await page.waitForSelector('[data-testid="workflowRunPanel"] .WorkflowNodeCard');
  await settle(800);
  check('1 list link: stays on the list', url().pathname === `/project/${PID}/workflows`, url().pathname);
  check('1 list link: w=<workflow>/<run> in the URL, slash not encoded', /[?&]w=\d+\/\d+(&|$)/.test(page.url()), url().search);
  check('1 list link: list still visible', await page.locator('[data-testid="workflows-rowLastRun"]').first().isVisible());
  const half = await panelWidth();
  check('3 default width is half of the window', Math.abs(half - 720) <= 1, `${half}px of 1440`);
  const box = await panel.boundingBox();
  check('3 panel is at the right edge, full height', Math.round(box.x + box.width) === 1440 && Math.round(box.height) === 900, JSON.stringify(box));
  await page.screenshot({ path: `${OUT}/01-panel-over-list.png` });

  // 3. resize
  const handle = page.locator('[data-testid="workflowRunPanel-resize"]');
  const hb = await handle.boundingBox();
  const hx = hb.x + hb.width / 2; const hy = 450;
  await page.mouse.move(hx, hy); await page.mouse.down(); await page.mouse.move(hx - 200, hy, { steps: 5 }); await page.mouse.up();
  await settle(200);
  const wide = await panelWidth();
  check('3 drag left edge 200px left: +200px', Math.abs(wide - half - 200) <= 2, `${half} -> ${wide}`);
  check('3 width stored', (await page.evaluate(() => localStorage.getItem('WorkflowRunPanel__width'))) === String(wide));
  // limits
  const hb2 = await handle.boundingBox();
  await page.mouse.move(hb2.x + 6, hy); await page.mouse.down(); await page.mouse.move(1435, hy, { steps: 5 }); await page.mouse.up();
  await settle(200);
  check('3 narrowest width 400px', (await panelWidth()) === 400, `${await panelWidth()}`);
  const hb3 = await handle.boundingBox();
  await page.mouse.move(hb3.x + 6, hy); await page.mouse.down(); await page.mouse.move(-300, hy, { steps: 5 }); await page.mouse.up();
  await settle(200);
  check('3 widest: the window', (await panelWidth()) === 1440, `${await panelWidth()}`);
  const hb4 = await handle.boundingBox();
  await page.mouse.move(hb4.x + 8, hy); await page.mouse.down(); await page.mouse.move(540, hy, { steps: 5 }); await page.mouse.up();
  await settle(200);
  const kept = await panelWidth();
  await page.reload();
  await panel.waitFor();
  await page.waitForSelector('[data-testid="workflowRunPanel"] .WorkflowNodeCard');
  await settle(600);
  check('3/4 reload with ?w=: panel is back with the stored width', (await panelWidth()) === kept, `${kept} -> ${await panelWidth()}`);
  await page.screenshot({ path: `${OUT}/02-panel-resized-after-reload.png` });
  await page.locator('[data-testid="workflowRunPanel-resize"]').dblclick();
  await settle(200);
  check('3 double click on the edge: half again', Math.abs((await panelWidth()) - 720) <= 1, `${await panelWidth()}`);

  // 5. a node with a task opens the task dialog over the panel
  const runId = url().searchParams.get('w');
  await page.locator('[data-testid="workflowRunPanel"] [data-testid="workflowNode-task"]').first().click();
  await dialog.waitFor();
  await settle(800);
  check('5 node click: t added, w kept', url().searchParams.get('w') === runId && /^\d+$/.test(url().searchParams.get('t') || ''), url().search);
  check('5 both visible', await panel.isVisible() && await dialog.isVisible());
  const z = await page.evaluate(() => {
    const p = document.querySelector('[data-testid="workflowRunPanel"]');
    const d = document.querySelector('[data-testid="taskLogDialog"]').closest('.v-overlay');
    const r = d.querySelector('.v-overlay__content').getBoundingClientRect();
    const top = document.elementFromPoint(r.left + r.width / 2, r.top + 20);
    return { panel: getComputedStyle(p).zIndex, dialog: getComputedStyle(d).zIndex, topInDialog: !!top.closest('.v-overlay') };
  });
  check('5 task dialog lies over the panel', Number(z.dialog) > Number(z.panel) && z.topInDialog, JSON.stringify(z));
  await page.screenshot({ path: `${OUT}/03-panel-and-task-dialog.png` });
  const both = page.url();

  // Escape closes the dialog only
  await page.keyboard.press('Escape');
  await dialog.waitFor({ state: 'hidden' });
  await settle(400);
  check('5 closing the task dialog keeps the panel', await panel.isVisible() && url().searchParams.get('w') === runId && !url().searchParams.has('t'), url().search);

  // 4. Escape, back, forward
  await page.keyboard.press('Escape');
  await panel.waitFor({ state: 'hidden' });
  check('4 Escape closes the panel and drops w', !url().searchParams.has('w') && url().pathname === `/project/${PID}/workflows`, page.url());
  await page.locator('[data-testid="workflows-rowLastRun"]').first().click();
  await panel.waitFor();
  await page.goBack();
  await panel.waitFor({ state: 'hidden' });
  check('4 browser Back closes the panel', !url().searchParams.has('w'), page.url());
  await page.goForward();
  await panel.waitFor();
  check('4 browser Forward opens it again', url().searchParams.get('w') === runId, page.url());
  await page.locator('[data-testid="workflowRun-close"]').click();
  await panel.waitFor({ state: 'hidden' });
  check('4 close button closes the panel and drops w', !url().searchParams.has('w'), page.url());

  // 5. direct URL with both
  await page.goto(both);
  await panel.waitFor(); await dialog.waitFor();
  await settle(800);
  check('5 direct URL ?w=&t= opens both', await panel.isVisible() && await dialog.isVisible(), new URL(both).search);
  // Escape twice in a row closes the dialog, then the panel; neither comes back
  await page.keyboard.press('Escape');
  await page.keyboard.press('Escape');
  await settle(1500);
  check('5 Escape twice in a row closes both for good', !(await dialog.isVisible()) && !(await panel.isVisible()) && url().search === '', url().search);

  // 4. other query parameters survive
  await page.goto(`${BASE}/project/${PID}/workflows?foo=bar&w=${runId}`);
  await panel.waitFor();
  await page.locator('[data-testid="workflowRun-close"]').click();
  await panel.waitFor({ state: 'hidden' });
  check('4 closing keeps the other query parameters', url().search === '?foo=bar', url().search);

  // 7. a run that is not there
  await page.goto(`${BASE}/project/${PID}/workflows?w=${WF}/999999`);
  await page.waitForSelector('[data-testid="workflowRun-error"]');
  check('7 unknown run: the panel says so', (await page.locator('[data-testid="workflowRun-error"]').innerText()).includes('Workflow run not found'), await page.locator('[data-testid="workflowRun-error"]').innerText());
  await page.screenshot({ path: `${OUT}/04-panel-run-not-found.png` });
  // a run under a workflow it does not belong to, and a value in the old one-number form
  await page.goto(`${BASE}/project/${PID}/workflows?w=${OTHER_WF}/${RUN}`);
  await page.waitForSelector('[data-testid="workflowRun-error"]');
  check('7 a run under another workflow is not shown', (await page.locator('[data-testid="workflowRunPanel"] .WorkflowNodeCard').count()) === 0);
  await page.goto(`${BASE}/project/${PID}/workflows?w=${RUN}`);
  await page.waitForSelector('[data-testid="workflows-rowLastRun"]');
  await settle(500);
  check('4 w without a workflow opens nothing', (await panel.count()) === 0, url().search);

  // 1. the list of runs of a workflow; 2. Run from the workflow page; 6. approval, polling
  await page.goto(`${BASE}/project/${PID}/workflows/${WF}/runs`);
  await page.waitForSelector('[data-testid="workflowRuns-rowId"]');
  const rowsBefore = await page.locator('[data-testid="workflowRuns-rowId"]').count();
  await page.locator('[data-testid="workflowRuns-rowId"]').last().click();
  await panel.waitFor();
  check('1 runs list link: panel over the workflow page', url().pathname === `/project/${PID}/workflows/${WF}/runs` && url().searchParams.has('w'), page.url());
  const first = url().searchParams.get('w');
  await page.locator('[data-testid="workflowRuns-rowId"]').first().click();
  await settle(600);
  check('1 another run in the list swaps the run of the panel', url().searchParams.get('w') !== first && `${await panel.getAttribute('data-workflow-id')}/${await panel.getAttribute('data-run-id')}` === url().searchParams.get('w'), `${first} -> ${url().searchParams.get('w')}`);
  await page.locator('[data-testid="workflowRun-close"]').click();
  await panel.waitFor({ state: 'hidden' });

  await page.locator('[data-testid="workflow-run"]').click();
  await panel.waitFor();
  await settle(1500);
  const started = url().searchParams.get('w');
  check('2 Run on the workflow page: opens the new run in the panel', url().pathname === `/project/${PID}/workflows/${WF}/runs` && !!started, page.url());
  const rowsAfter = await page.locator('[data-testid="workflowRuns-rowId"]').count();
  check('2 the list under the panel shows the new run', rowsAfter === rowsBefore + 1, `${rowsBefore} -> ${rowsAfter}`);
  // 6. the run reaches the approval by itself (polling), then is approved from the node card
  await page.waitForSelector('[data-testid="workflowRunPanel"] [data-testid="workflowRun-status-approval"]', { timeout: 30000 });
  check('6 status follows the run without a reload (approval)', true);
  await page.screenshot({ path: `${OUT}/05-panel-approval.png` });
  await page.locator('[data-testid="workflowRunPanel"] [data-testid="workflowCanvas-nodeApprove"]').first().click();
  await page.waitForSelector('[data-testid="workflowRunPanel"] [data-testid="workflowRun-status-success"]', { timeout: 60000 });
  check('6 approve in the panel, run ends in success', true);
  await settle(800);
  check('6 the list under the panel follows the status', (await page.locator(`[data-testid="workflowRuns-rowStatus-success"]`).count()) >= 2, `${await page.locator('[data-testid="workflowRuns-rowStatus-success"]').count()} success rows`);
  await page.locator('[data-testid="workflowRun-close"]').click();
  await panel.waitFor({ state: 'hidden' });

  // 2. Run in the list of workflows; 6. Stop
  await page.goto(`${BASE}/project/${PID}/workflows`);
  await page.waitForSelector('[data-testid="workflows-rowRun"]');
  await page.locator('[data-testid="workflows-rowRun"]').nth(1).click();
  await panel.waitFor();
  await settle(1500);
  check('2 Run in the list of workflows: panel over the list', url().pathname === `/project/${PID}/workflows` && url().searchParams.has('w'), page.url());
  const stopRun = url().searchParams.get('w').split('/')[1];
  const lastRuns = await page.locator('[data-testid="workflows-rowLastRun"]').allInnerTexts();
  check('2 the list shows the new run as the last one', lastRuns.some((t) => t.trim() === `#${stopRun}`), lastRuns.join(' '));
  await page.locator('[data-testid="workflowRunPanel"] [data-testid="workflowRun-stop"]').click();
  await page.waitForSelector('[data-testid="workflowRunPanel"] [data-testid="workflowRun-status-stopped"]', { timeout: 30000 });
  check('6 Stop in the panel', true);
  await page.screenshot({ path: `${OUT}/06-panel-stopped.png` });

  // 8. the old address
  await page.goto(`${BASE}/project/${PID}/workflows/${WF}/runs/${RUN}`);
  await page.waitForSelector('[data-testid="workflowRun-page"] .WorkflowNodeCard');
  await settle(800);
  check('8 old URL opens the run as a page', (await panel.count()) === 0 && await page.locator('[data-testid="workflowRun-backToList"]').isVisible() && (await page.locator('[data-testid="workflowRun-close"]').count()) === 0);
  check('8 page run view shows the nodes', (await page.locator('.WorkflowNodeCard').count()) === 4, `${await page.locator('.WorkflowNodeCard').count()} nodes`);
  await page.screenshot({ path: `${OUT}/07-old-run-page.png` });

  // 9. another project: the run of the panel belongs to the project it was opened in
  const router = () => page.evaluate((to) => document.querySelector('#app').__vue_app__.config.globalProperties.$router.push(to), `/project/${OTHER_PID}/workflows?w=${WF}/${RUN}`);
  await page.goto(`${BASE}/project/${PID}/workflows?w=${WF}/${RUN}`);
  await page.waitForSelector('[data-testid="workflowRunPanel"] .WorkflowNodeCard');
  await router();
  await settle(1500);
  check('9 same ?w= in another project: the view starts anew (no graph of the old project)',
    (await page.locator('[data-testid="workflowRunPanel"] .WorkflowNodeCard').count()) === 0
    && (await page.locator('[data-testid="workflowRunPanel"] [data-testid="workflowRun-error"]').count()) === 1,
    `${await page.locator('[data-testid="workflowRunPanel"] .WorkflowNodeCard').count()} nodes`);
  await page.goto(`${BASE}/project/${PID}/workflows?w=${WF}/${RUN}&foo=bar`);
  await page.waitForSelector('[data-testid="workflowRunPanel"] .WorkflowNodeCard');
  await page.locator('[data-testid="sidebar-currentProject"]').click();
  await page.locator('[data-testid="sidebar-newProject"]').click();
  await page.locator('[data-testid="newProject-name"] input').fill(`panel check ${Date.now()}`);
  await page.locator('[data-testid="appNewProjectDialog"] [data-testid="editDialog-save"]').click();
  await page.waitForURL((u) => !u.pathname.startsWith(`/project/${PID}/`) && /^\/project\/\d+/.test(u.pathname), { timeout: 10000 });
  await settle(1500);
  const newPid = url().pathname.split('/')[2];
  check('9 creating a project closes the panel and drops w, keeps the rest', (await panel.count()) === 0 && !url().searchParams.has('w') && url().searchParams.get('foo') === 'bar', page.url());
  await page.screenshot({ path: `${OUT}/09-new-project-no-panel.png` });
  await ctx.request.delete(`${BASE}/api/project/${newPid}`);

  // 10. a run that ended while the list was on screen: opening it brings the list up to date
  const api = (p) => `${BASE}/api/project/${PID}/workflows/${WF}${p}`;
  const runOf = async (id) => (await (await ctx.request.get(api(`/runs/${id}`))).json());
  const waitRun = async (id, ok) => { for (let i = 0; i < 120; i += 1) { const r = await runOf(id); if (ok(r)) return r; await settle(500); } throw new Error(`run ${id}: timeout`); };
  const stale = (await (await ctx.request.post(api('/runs'), { data: {} })).json()).id;
  const atApproval = await waitRun(stale, (r) => r.run.status === 'approval');
  await page.goto(`${BASE}/project/${PID}/workflows`);
  const row = page.locator('[data-testid="workflows-rowLastRun"]', { hasText: `#${stale}` }).locator('xpath=ancestor::tr');
  await row.locator('[data-testid="workflows-rowStatus-approval"]').waitFor();
  const pending = atApproval.nodes.find((n) => n.approval && n.approval.status === 'pending');
  await ctx.request.post(api(`/runs/${stale}/approvals/${pending.node.id}`), { data: { status: 'approved' } });
  await waitRun(stale, (r) => r.run.status === 'success');
  await settle(1000);
  check('10 precondition: the list still shows the ended run as approval', (await row.locator('[data-testid="workflows-rowStatus-approval"]').count()) === 1);
  await page.locator('[data-testid="workflows-rowLastRun"]', { hasText: `#${stale}` }).click();
  await page.waitForSelector('[data-testid="workflowRunPanel"] [data-testid="workflowRun-status-success"]');
  await settle(1500);
  check('10 opening the ended run updates the list under the panel', (await row.locator('[data-testid="workflows-rowStatus-success"]').count()) === 1);
  await page.locator('[data-testid="workflowRun-close"]').click();
  await panel.waitFor({ state: 'hidden' });
  check('10 the list keeps the status after the panel closes', (await row.locator('[data-testid="workflows-rowStatus-success"]').count()) === 1);
  await page.screenshot({ path: `${OUT}/10-list-after-ended-run.png` });

  // dark theme, narrow screen
  await page.evaluate(() => localStorage.setItem('darkMode', '1'));
  await page.goto(`${BASE}/project/${PID}/workflows?w=${WF}/${RUN}`);
  await panel.waitFor(); await page.waitForSelector('[data-testid="workflowRunPanel"] .WorkflowNodeCard'); await settle(800);
  await page.screenshot({ path: `${OUT}/08-panel-dark.png` });
  await page.evaluate(() => localStorage.removeItem('darkMode'));
  await page.setViewportSize({ width: 600, height: 800 });
  await settle(500);
  check('narrow screen: the panel takes the whole width, no handle', (await panelWidth()) === 600 && (await page.locator('[data-testid="workflowRunPanel-resize"]').count()) === 0, `${await panelWidth()}`);

  await browser.close();
  const failed = results.filter((r) => !r.ok);
  console.log(`\n${results.length - failed.length}/${results.length} passed`);
  if (errors.length) console.log(`BROWSER ERRORS:\n${[...new Set(errors)].join('\n')}`);
  process.exit(failed.length ? 1 : 0);
})().catch((e) => { console.error(e); process.exit(1); });
