// Crops one card per node kind from the Release workflow editor (light theme).
const { chromium } = require('playwright');
const fs = require('fs');
const BASE = 'http://localhost:3100';
const OUT = '/tmp/semaphore-stand/shots-nodes';
const W1 = process.env.WF || '4';

(async () => {
  fs.mkdirSync(OUT, { recursive: true });
  const browser = await chromium.launch({ channel: 'chrome' });
  const ctx = await browser.newContext({ viewport: { width: 1600, height: 1000 }, deviceScaleFactor: 2 });
  await ctx.request.post(`${BASE}/api/auth/login`, { data: { auth: 'admin', password: 'admin123' } });
  const page = await ctx.newPage();
  await page.goto(`${BASE}/project/1/workflows/${W1}/edit`);
  await page.waitForSelector('.WorkflowNodeCard');
  await page.waitForTimeout(1800);
  // Edge pills would intrude into the crops; hide them for the shots.
  await page.addStyleTag({ content: '.WorkflowEdgeLabel { display: none !important; }' });
  // 100 % zoom so every crop has the same scale.
  await page.mouse.click(1400, 900);
  await page.keyboard.press('1');
  await page.waitForTimeout(400);

  const targets = [
    { name: 'node-task', text: 'Deploy website', hover: true },
    { name: 'node-approval', text: 'Approval', hover: false },
    { name: 'node-delay', text: 'Delay', hover: false },
    { name: 'node-note', text: 'Prod apply needs', hover: false },
  ];
  for (const t of targets) {
    const node = page.locator('.drawflow-node', { hasText: t.text }).first();
    await node.scrollIntoViewIfNeeded();
    const box = await node.boundingBox();
    if (t.hover) {
      await page.mouse.move(box.x + box.width / 2, box.y + box.height / 2);
      await page.waitForTimeout(250);
    } else {
      await page.mouse.move(10, 10);
      await page.waitForTimeout(250);
    }
    const pad = { l: 20, t: 20, r: t.hover ? 64 : 20, b: 20 };
    await page.screenshot({
      path: `${OUT}/${t.name}.png`,
      clip: {
        x: box.x - pad.l, y: box.y - pad.t, width: box.width + pad.l + pad.r, height: box.height + pad.t + pad.b,
      },
    });
    console.log(t.name, JSON.stringify(box));
  }
  await browser.close();
})().catch((e) => { console.error(e); process.exit(1); });
