// Screenshot the run view (or the editor) of a workflow on the stand.
// usage: NODE_PATH=~/.npm/_npx/9833c18b2d85bc59/node_modules node shoot-run.cjs <workflow_id> [run_id] [out.png] [--dark]
const { chromium } = require('playwright');
const args = process.argv.slice(2);
const dark = args.includes('--dark');
const [WF, RUN, OUT = '/tmp/semaphore-stand/shot-run.png'] = args.filter((a) => a !== '--dark');
const BASE = process.env.BASE || 'http://localhost:3100';
const PID = process.env.PID || 1;
(async () => {
  const browser = await chromium.launch({ channel: 'chrome' });
  const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 }, deviceScaleFactor: 2 });
  await ctx.request.post(`${BASE}/api/auth/login`, { data: { auth: 'admin', password: 'admin123' } });
  const page = await ctx.newPage();
  const errors = [];
  page.on('pageerror', (e) => errors.push(e.message));
  page.on('console', (m) => { if (m.type() === 'error') errors.push(m.text()); });
  await page.addInitScript((d) => { if (d) localStorage.setItem('darkMode', '1'); else localStorage.removeItem('darkMode'); }, dark);
  const url = RUN ? `${BASE}/project/${PID}/workflows/${WF}/runs/${RUN}` : `${BASE}/project/${PID}/workflows/${WF}/edit`;
  await page.goto(url);
  await page.waitForSelector('.WorkflowNodeCard', { timeout: 15000 });
  await page.waitForTimeout(1500);
  await page.screenshot({ path: OUT });
  await browser.close();
  console.log(`saved ${OUT}${errors.length ? '\nBROWSER ERRORS:\n' + errors.join('\n') : ''}`);
  process.exit(errors.length ? 2 : 0);
})().catch((e) => { console.error(e); process.exit(1); });
