const { chromium } = require('playwright');
const fs = require('fs');

const BASE = 'http://localhost:3100';
const OUT = process.env.OUT || '/tmp/semaphore-stand/shots-wf';
const PID = 1;
const [W1, W2] = fs.readFileSync('/tmp/semaphore-stand/wf-ids.txt', 'utf8').trim().split(' ');
const RUN = process.env.RUN || '1';

(async () => {
  fs.mkdirSync(OUT, { recursive: true });
  const browser = await chromium.launch({ channel: 'chrome' });
  const errors = [];

  async function session(dark) {
    const ctx = await browser.newContext({
      viewport: { width: 1440, height: 900 }, deviceScaleFactor: 2,
    });
    await ctx.request.post(`${BASE}/api/auth/login`, { data: { auth: 'admin', password: 'admin123' } });
    const page = await ctx.newPage();
    page.on('pageerror', (e) => errors.push(`${dark ? 'dark' : 'light'}: ${e.message}`));
    page.on('console', (m) => { if (m.type() === 'error') errors.push(`${dark ? 'dark' : 'light'} console: ${m.text()}`); });
    await page.addInitScript((d) => { if (d) localStorage.setItem('darkMode', '1'); else localStorage.removeItem('darkMode'); }, dark);
    return { ctx, page };
  }

  const settle = (page, ms = 900) => page.waitForTimeout(ms);
  const suffix = (dark) => (dark ? 'dark' : 'light');

  for (const dark of [false, true]) {
    const { ctx, page } = await session(dark);

    // Editor with the Release workflow.
    await page.goto(`${BASE}/project/${PID}/workflows/${W1}/edit`);
    await page.waitForSelector('.WorkflowNodeCard');
    await settle(page, 1200);
    await page.screenshot({ path: `${OUT}/editor-${suffix(dark)}.png` });

    // Select the first node -> properties drawer + hover plus.
    const first = page.locator('.WorkflowNodeCard').first();
    await first.click();
    await settle(page, 600);
    await first.hover();
    await settle(page, 300);
    await page.screenshot({ path: `${OUT}/editor-selected-${suffix(dark)}.png` });

    // Quick-add menu from the plus handle.
    await first.locator('.WorkflowNodeCard__plus').click({ force: true });
    await settle(page, 600);
    await page.screenshot({ path: `${OUT}/editor-quickadd-${suffix(dark)}.png` });
    await page.keyboard.press('Escape');
    await settle(page, 300);

    // Legacy workflow (auto layout) with hostile names.
    await page.goto(`${BASE}/project/${PID}/workflows/${W2}/edit`);
    await page.waitForSelector('.WorkflowNodeCard');
    await settle(page, 1200);
    await page.screenshot({ path: `${OUT}/editor-legacy-${suffix(dark)}.png` });

    // Run view.
    await page.goto(`${BASE}/project/${PID}/workflows/${W1}/runs/${RUN}`);
    await page.waitForSelector('.WorkflowNodeCard');
    await settle(page, 1500);
    await page.screenshot({ path: `${OUT}/run-${suffix(dark)}.png` });

    // Empty editor.
    await page.goto(`${BASE}/project/${PID}/workflows/new`);
    await page.waitForSelector('.WorkflowGraph');
    await settle(page, 900);
    await page.screenshot({ path: `${OUT}/editor-empty-${suffix(dark)}.png` });

    await ctx.close();
  }

  await browser.close();
  if (errors.length) {
    console.log('BROWSER ERRORS:\n' + errors.join('\n'));
    process.exit(2);
  }
  console.log('ok');
})().catch((e) => { console.error(e); process.exit(1); });
