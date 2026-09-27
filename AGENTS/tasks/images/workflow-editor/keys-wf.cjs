// Verifies keyboard navigation on the canvas: arrows pan, +/- zoom, 0 fits, 1 resets.
const { chromium } = require('playwright');
const fs = require('fs');
const BASE = 'http://localhost:3100';
const [W1] = fs.readFileSync('/tmp/semaphore-stand/wf-ids.txt', 'utf8').trim().split(' ');

(async () => {
  const browser = await chromium.launch({ channel: 'chrome' });
  const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 } });
  await ctx.request.post(`${BASE}/api/auth/login`, { data: { auth: 'admin', password: 'admin123' } });
  const page = await ctx.newPage();
  const out = [];
  const vp = () => page.evaluate(() => {
    const vm = document.querySelector('.WorkflowGraph').__vue__;
    return vm.getViewport();
  });
  const check = (name, ok, extra = '') => out.push(`${ok ? 'PASS' : 'FAIL'} ${name} ${extra}`);

  for (const url of [`/project/1/workflows/${W1}/edit`, `/project/1/workflows/${W1}/runs/2`]) {
    await page.goto(BASE + url);
    await page.waitForSelector('.WorkflowNodeCard');
    await page.waitForTimeout(1800);
    const label = url.includes('/runs/') ? 'run view' : 'editor';

    // Focus the canvas by clicking empty space.
    await page.mouse.click(1300, 800);
    const start = await vp();

    await page.keyboard.press('ArrowRight');
    let v = await vp();
    check(`${label}: ArrowRight pans by 40`, Math.abs((start.x - v.x) - 40) < 0.01, `${start.x} -> ${v.x}`);
    await page.keyboard.press('Shift+ArrowDown');
    v = await vp();
    check(`${label}: Shift+ArrowDown pans by 200`, Math.abs((start.y - v.y) - 200) < 0.01, `${start.y} -> ${v.y}`);

    const z0 = v.zoom;
    await page.keyboard.press('+');
    v = await vp();
    check(`${label}: + zooms in`, v.zoom > z0, `${z0} -> ${v.zoom}`);
    await page.keyboard.press('-');
    v = await vp();
    check(`${label}: - zooms out`, Math.abs(v.zoom - z0) < 1e-9, `${v.zoom}`);

    await page.keyboard.press('1');
    v = await vp();
    check(`${label}: 1 resets zoom to 100%`, Math.abs(v.zoom - 1) < 1e-9, `${v.zoom}`);

    await page.keyboard.press('0');
    v = await vp();
    check(`${label}: 0 fits the graph`, Math.abs(v.zoom - start.zoom) < 1e-6 && Math.abs(v.x - start.x) < 1 && Math.abs(v.y - start.y) < 1, `${JSON.stringify(v)} vs ${JSON.stringify(start)}`);

    // Wheel pan and ctrl+wheel zoom still work.
    await page.mouse.move(1000, 500);
    await page.mouse.wheel(0, 100);
    v = await vp();
    check(`${label}: wheel pans`, Math.abs((start.y - v.y) - 100) < 0.01, `${v.y}`);
    await page.keyboard.down('Control');
    await page.mouse.wheel(0, -100);
    await page.keyboard.up('Control');
    const v2 = await vp();
    check(`${label}: ctrl+wheel zooms in`, v2.zoom > v.zoom, `${v.zoom} -> ${v2.zoom}`);
  }

  // Keys typed in the properties panel must not move the canvas.
  await page.goto(`${BASE}/project/1/workflows/${W1}/edit`);
  await page.waitForSelector('.WorkflowNodeCard');
  await page.waitForTimeout(1800);
  await page.locator('.WorkflowNodeCard').first().click();
  await page.waitForSelector('.WorkflowNodeProperties');
  const before = await vp();
  await page.locator('.WorkflowNodeProperties input').first().focus();
  await page.keyboard.press('ArrowLeft');
  await page.keyboard.press('0');
  const after = await vp();
  check('typing in the properties panel does not move the canvas', JSON.stringify(before) === JSON.stringify(after));

  // Tab reaches a card; Enter selects it (editor) and opens the properties panel.
  await page.goto(`${BASE}/project/1/workflows/${W1}/edit`);
  await page.waitForSelector('.WorkflowNodeCard');
  await page.waitForTimeout(1800);
  await page.mouse.click(1300, 800);
  await page.keyboard.press('Tab');
  const focused = await page.evaluate(() => document.activeElement && document.activeElement.className);
  check('Tab from the canvas focuses a node card', /WorkflowNodeCard/.test(focused || ''), focused);
  await page.keyboard.press('Enter');
  await page.waitForTimeout(500);
  check('Enter on a focused card opens the properties panel', await page.locator('.WorkflowNodeProperties').isVisible());
  check('Enter marks the card selected in Drawflow', (await page.locator('.drawflow-node.selected').count()) === 1);

  console.log(out.join('\n'));
  await browser.close();
  process.exit(out.some((l) => l.startsWith('FAIL')) ? 1 : 0);
})().catch((e) => { console.error(e); process.exit(1); });
