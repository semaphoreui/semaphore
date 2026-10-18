// Screenshot the workflow editor on the stand with one node selected: the left panel (survey), the
// canvas and the node properties panel with bound survey fields (<out>.png), a crop of the task
// params card (<out>-fields.png) and the open "assign a workflow variable" menu (<out>-assign.png).
// usage: NODE_PATH=~/.npm/_npx/9833c18b2d85bc59/node_modules node shoot-editor.cjs <workflow_id> [node_title] [out.png] [--dark]
const { chromium } = require('playwright');
const args = process.argv.slice(2);
const dark = args.includes('--dark');
const [WF, TITLE = 'bash: consume', OUT = '/tmp/semaphore-stand/shot-editor.png'] = args.filter((a) => a !== '--dark');
const BASE = process.env.BASE || 'http://localhost:3100';
const PID = process.env.PID || 1;
(async () => {
  const browser = await chromium.launch({ channel: 'chrome' });
  const ctx = await browser.newContext({ viewport: { width: 1680, height: 960 }, deviceScaleFactor: 2 });
  await ctx.request.post(`${BASE}/api/auth/login`, { data: { auth: 'admin', password: 'admin123' } });
  const page = await ctx.newPage();
  const errors = [];
  page.on('pageerror', (e) => errors.push(e.message));
  page.on('console', (m) => { if (m.type() === 'error') errors.push(m.text()); });
  await page.addInitScript((d) => {
    if (d) localStorage.setItem('darkMode', '1'); else localStorage.removeItem('darkMode');
    localStorage.removeItem('workflowEditor__sideCollapsed');
  }, dark);
  await page.goto(`${BASE}/project/${PID}/workflows/${WF}/edit`);
  await page.waitForSelector('.WorkflowNodeCard', { timeout: 15000 });
  await page.waitForTimeout(800);
  const card = page.locator('.WorkflowNodeCard', { hasText: TITLE }).first();
  await card.click();
  await page.waitForSelector('.WorkflowNodeProperties', { timeout: 5000 });
  await page.waitForTimeout(800);
  const params = page.locator('.WorkflowNodeProperties__params');
  const bound = page.locator('.TaskParamsForm__bound').first();
  if (await bound.count()) {
    // The panel scrolls; bring the first bound field into the frame.
    await bound.scrollIntoViewIfNeeded();
    await page.waitForTimeout(300);
  }
  await page.screenshot({ path: OUT });
  if (await params.count()) {
    await params.screenshot({ path: OUT.replace(/\.png$/, '-fields.png') });
  }
  // The assign menu of the first unbound field.
  const assign = page.locator('.WorkflowVarAssign__btn').first();
  if (await assign.count()) {
    await assign.scrollIntoViewIfNeeded();
    await assign.click();
    await page.waitForSelector('.v-menu__content.menuable__content__active .WorkflowVarAssign__list', { timeout: 3000 });
    await page.waitForTimeout(300);
    await page.screenshot({ path: OUT.replace(/\.png$/, '-assign.png') });
    await page.keyboard.press('Escape');
  }
  await browser.close();
  console.log(`saved ${OUT}${errors.length ? '\nBROWSER ERRORS:\n' + errors.join('\n') : ''}`);
  process.exit(errors.length ? 2 : 0);
})().catch((e) => { console.error(e); process.exit(1); });
