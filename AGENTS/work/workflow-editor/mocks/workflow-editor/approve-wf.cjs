const { chromium } = require('playwright');
const fs = require('fs');

const BASE = 'http://localhost:3100';
const OUT = '/tmp/semaphore-stand/shots-wf';
const [W3, R3] = fs.readFileSync('/tmp/semaphore-stand/wf3.txt', 'utf8').trim().split(' ');

(async () => {
  const browser = await chromium.launch({ channel: 'chrome' });
  const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 }, deviceScaleFactor: 2 });
  await ctx.request.post(`${BASE}/api/auth/login`, { data: { auth: 'admin', password: 'admin123' } });
  const page = await ctx.newPage();
  const errors = [];
  page.on('pageerror', (e) => errors.push(e.message));
  const out = [];

  await page.goto(`${BASE}/project/1/workflows/${W3}/runs/${R3}`);
  await page.waitForSelector('.WorkflowNodeCard');
  await page.waitForTimeout(1500);
  const pending = await page.locator('.WorkflowNodeCard--pending').count();
  out.push(`pending approval cards: ${pending}`);
  out.push(`approve button in card: ${await page.locator('.WorkflowNodeCard__actions button', { hasText: 'Approve' }).isVisible()}`);
  await page.screenshot({ path: `${OUT}/run-approval-pending.png` });

  // Remember the viewport, wait through a poll, and make sure it did not move.
  const before = await page.evaluate(() => document.querySelector('.drawflow').style.transform);
  await page.mouse.move(700, 500);
  await page.mouse.wheel(0, -120);
  await page.waitForTimeout(6000);
  const after = await page.evaluate(() => document.querySelector('.drawflow').style.transform);
  out.push(`viewport kept across a poll after panning: ${before !== after && after.length > 0} (${after})`);

  await page.locator('.WorkflowNodeCard__actions button', { hasText: 'Approve' }).click();
  await page.waitForTimeout(2500);
  const running = await page.locator('.WorkflowNodeCard--running').count();
  const sub = await page.locator('.WorkflowNodeCard--running .WorkflowNodeCard__sub').first().textContent();
  out.push(`after approve, waiting delay cards: ${running}, subtitle: ${sub.trim()}`);
  const activeEdges = await page.locator('.connection.WorkflowGraph__conn--active').count();
  out.push(`active (animated) edges: ${activeEdges}`);
  await page.waitForTimeout(2000);
  const sub2 = await page.locator('.WorkflowNodeCard--running .WorkflowNodeCard__sub').first().textContent();
  out.push(`countdown ticks: ${sub.trim()} -> ${sub2.trim()}`);
  await page.screenshot({ path: `${OUT}/run-delay-countdown.png` });

  // Clicking a task card with a task opens the log dialog (Release run #2).
  const [W1] = fs.readFileSync('/tmp/semaphore-stand/wf-ids.txt', 'utf8').trim().split(' ');
  await page.goto(`${BASE}/project/1/workflows/${W1}/runs/2`);
  await page.waitForSelector('.WorkflowNodeCard--clickable');
  await page.waitForTimeout(1200);
  await page.locator('.WorkflowNodeCard--clickable').first().click();
  await page.waitForTimeout(1200);
  out.push(`task log dialog opens on card click: ${await page.locator('.v-dialog--active').isVisible()}`);
  await page.screenshot({ path: `${OUT}/run-task-log.png` });

  console.log(out.join('\n'));
  if (errors.length) console.log('PAGE ERRORS:\n' + errors.join('\n'));
  await browser.close();
})().catch((e) => { console.error(e); process.exit(1); });
