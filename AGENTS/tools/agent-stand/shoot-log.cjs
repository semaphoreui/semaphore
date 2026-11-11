// Runs the Claude template of the agent stand (the fake CLI reports a made-up run, see
// fake-agent.sh) and checks the task log in the browser: the agent's events are rows of a
// transcript, not JSON; a result sits in the row of its call; a call opens. Takes screenshots
// while the task runs, when it has ended and with calls open, and prints the raw log.
// NODE_PATH=~/.npm/_npx/9833c18b2d85bc59/node_modules node AGENTS/tools/agent-stand/shoot-log.cjs [out_dir]
const path = require('path');
const fs = require('fs');
const assert = require('assert');
const { chromium } = require('playwright');

const BASE = process.env.BASE || 'http://127.0.0.1:3102';
const DATA = process.env.AGENT_STAND_DIR || path.join(__dirname, '..', '..', 'tmp', 'agent-stand');
const OUT = path.resolve(process.argv[2] || path.join(DATA, 'shots-log'));
const ids = Object.fromEntries(fs.readFileSync(path.join(DATA, 'ids.txt'), 'utf8').trim().split(' ').map((kv) => kv.split('=')));
const P = ids.project;
const T = ids.claude_template;

fs.mkdirSync(OUT, { recursive: true });

(async () => {
  const browser = await chromium.launch({ channel: 'chrome' });
  const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 }, baseURL: BASE });
  assert.ok((await ctx.request.post('/api/auth/login', { data: { auth: 'admin', password: 'admin123' } })).ok());
  const page = await ctx.newPage();
  const errors = [];
  page.on('console', (m) => { if (m.type() === 'error') errors.push(`console: ${m.text()}`); });
  page.on('pageerror', (e) => errors.push(`pageerror: ${e.message}`));

  const shot = async (name) => {
    const file = path.join(OUT, `${name}.png`);
    await page.locator('[data-testid="taskLog-page"]').last().screenshot({ path: file });
    console.log('shot', file);
  };

  const created = await ctx.request.post(`/api/project/${P}/tasks`, { data: { template_id: Number(T) } });
  assert.ok(created.ok(), `start task: ${created.status()}`);
  const task = await created.json();

  await page.goto(`/project/${P}/templates/${T}?t=${task.id}`);
  const log = page.locator('[data-testid="taskLog-output"]').last();
  await log.waitFor();

  // While it runs: rows appear as the agent goes, a call without a result is running.
  await page.locator('[data-testid="taskLog-agentTool"]').first().waitFor({ timeout: 20000 });
  await page.waitForTimeout(1200);
  await shot('01-running');

  for (let i = 0; i < 120; i += 1) {
    // eslint-disable-next-line no-await-in-loop
    const t = await (await ctx.request.get(`/api/project/${P}/tasks/${task.id}`)).json();
    if (['success', 'error', 'stopped'].includes(t.status)) {
      assert.strictEqual(t.status, 'success');
      break;
    }
    // eslint-disable-next-line no-await-in-loop
    await page.waitForTimeout(500);
  }
  await page.locator('[data-testid="taskLog-agentDone"]').waitFor({ timeout: 20000 });
  await page.waitForTimeout(1500);
  await shot('02-finished');

  const text = await log.innerText();
  assert.ok(!text.includes('@agent'), 'no event is shown as JSON');
  assert.ok(!text.includes('"type"'), 'no event of the CLI is shown');
  assert.ok(!text.includes('fake-token'), 'the token is masked');
  assert.ok(text.includes('The token of the server is ********'), text);
  assert.ok(text.includes('list_issues returned both.'), 'a subagent does not break the line of the agent');

  const states = await page.locator('[data-testid="taskLog-agentToolState"]').evaluateAll((els) => els.map((e) => e.dataset.state));
  // ToolSearch, list_issues, Agent, Read (of the subagent), Bash (failed), get_issue (no result).
  assert.deepStrictEqual(states, ['done', 'done', 'done', 'done', 'failed', 'interrupted']);
  assert.strictEqual(await page.locator('[data-testid="taskLog-agentDone"]').count(), 1, 'one summary');
  assert.strictEqual(await page.locator('[data-testid="taskLog-agentInit"]').count(), 1, 'one init');
  assert.strictEqual(await page.locator('[data-testid="taskLog-agentResult"]').count(), 0, 'every result is in its call');
  // The failed call is open by itself.
  assert.ok((await page.locator('[data-testid="taskLog-agentToolResult"]').first().innerText()).includes('No such file'));

  // Open the calls: arguments and results in a readable form.
  const tools = page.locator('[data-testid="taskLog-agentTool"]');
  await tools.nth(0).click();
  await tools.nth(1).click();
  await page.waitForTimeout(500);
  const details = await page.locator('[data-testid="taskLog-agentToolDetails"]').allInnerTexts();
  assert.ok(details.some((d) => d.includes('mcp__github__list_issues\nmcp__github__get_issue')), details.join('\n---\n'));
  assert.ok(details.some((d) => d.includes('repo: semaphoreui/semaphore') && d.includes('- number: 4410') && d.includes('See util/config_removed.go.')), details.join('\n---\n'));
  await shot('03-calls-open');

  await tools.nth(0).click();
  await tools.nth(1).click();
  await tools.nth(2).click();
  await tools.nth(3).click();
  await page.waitForTimeout(500);
  await log.evaluate((el) => { el.scrollTop = el.scrollHeight; });
  await page.waitForTimeout(500);
  await shot('04-subagent-open');

  const raw = await (await ctx.request.get(`/api/project/${P}/tasks/${task.id}/raw_output`)).text();
  assert.ok(!raw.includes('@agent'), 'the raw log is plain text');
  assert.ok(raw.includes('> mcp__github__list_issues {"repo":"semaphoreui/semaphore"'), raw);
  console.log('--- raw log ---');
  console.log(raw);

  assert.deepStrictEqual(errors, []);
  console.log('ok: task', task.id);
  await browser.close();
})().catch((e) => { console.error(e); process.exit(1); });
