// Runs the Claude and the Codex templates of the agent stand (the fake CLI reports a made-up run
// in the events of each CLI, see fake-agent.sh) and checks the task log in the browser: the
// agent's events are rows of a transcript, not JSON; a result sits in the row of its call; a call
// opens. Takes screenshots while the task runs, when it has ended and with calls open, and
// prints the raw logs.
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

  const at = (testId) => page.locator(`[data-testid="${testId}"]`);
  const toolStates = () => at('taskLog-agentToolState').evaluateAll((els) => els.map((e) => e.dataset.state));

  // Starts a task of the template and opens its log; returns the task and the log.
  const start = async (templateId) => {
    const created = await ctx.request.post(`/api/project/${P}/tasks`, { data: { template_id: Number(templateId) } });
    assert.ok(created.ok(), `start task: ${created.status()}`);
    const task = await created.json();
    await page.goto(`/project/${P}/templates/${templateId}?t=${task.id}`);
    const log = at('taskLog-output').last();
    await log.waitFor();
    return { task, log };
  };

  const finished = async (task) => {
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
    await at('taskLog-agentDone').waitFor({ timeout: 20000 });
    await page.waitForTimeout(1500);
  };

  // What holds for the log of either CLI.
  const checkLog = async (log) => {
    const text = await log.innerText();
    assert.ok(!text.includes('@agent'), 'no event is shown as JSON');
    assert.ok(!text.includes('"type"'), 'no event of the CLI is shown');
    assert.ok(!text.includes('fake-token'), 'the token is masked');
    assert.ok(text.includes('The token of the server is ********'), text);
    assert.strictEqual(await at('taskLog-agentDone').count(), 1, 'one summary');
    assert.strictEqual(await at('taskLog-agentResult').count(), 0, 'every result is in its call');
    // The failed call is open by itself.
    assert.ok((await at('taskLog-agentToolResult').first().innerText()).includes('No such file'));
    return text;
  };

  const rawLog = async (task, call) => {
    const raw = await (await ctx.request.get(`/api/project/${P}/tasks/${task.id}/raw_output`)).text();
    assert.ok(!raw.includes('@agent'), 'the raw log is plain text');
    assert.ok(raw.includes(call), raw);
    console.log(`--- raw log of task ${task.id} ---`);
    console.log(raw);
  };

  const tools = at('taskLog-agentTool');
  const details = () => at('taskLog-agentToolDetails').allInnerTexts();

  // --- Claude Code -------------------------------------------------------------------------
  const claude = await start(ids.claude_template);

  // While it runs: rows appear as the agent goes, a call without a result is running.
  await tools.first().waitFor({ timeout: 20000 });
  await page.waitForTimeout(1200);
  await shot('01-running');

  await finished(claude.task);
  await shot('02-finished');

  let text = await checkLog(claude.log);
  assert.ok(text.includes('list_issues returned both.'), 'a subagent does not break the line of the agent');
  // ToolSearch, list_issues, Agent, Read (of the subagent), Bash (failed), get_issue (no result).
  assert.deepStrictEqual(await toolStates(), ['done', 'done', 'done', 'done', 'failed', 'interrupted']);
  assert.strictEqual(await at('taskLog-agentInit').count(), 1, 'one init');

  // Open the calls: arguments and results in a readable form.
  await tools.nth(0).click();
  await tools.nth(1).click();
  await page.waitForTimeout(500);
  let open = await details();
  assert.ok(open.some((d) => d.includes('mcp__github__list_issues\nmcp__github__get_issue')), open.join('\n---\n'));
  assert.ok(open.some((d) => d.includes('repo: semaphoreui/semaphore') && d.includes('- number: 4410') && d.includes('See util/config_removed.go.')), open.join('\n---\n'));
  await shot('03-calls-open');

  await tools.nth(0).click();
  await tools.nth(1).click();
  await tools.nth(2).click();
  await tools.nth(3).click();
  await page.waitForTimeout(500);
  await claude.log.evaluate((el) => { el.scrollTop = el.scrollHeight; });
  await page.waitForTimeout(500);
  await shot('04-subagent-open');

  await rawLog(claude.task, '> mcp__github__list_issues {"repo":"semaphoreui/semaphore"');

  // --- Codex -------------------------------------------------------------------------------
  const codex = await start(ids.codex_template);
  await tools.first().waitFor({ timeout: 20000 });
  await finished(codex.task);
  await shot('05-codex-finished');

  text = await checkLog(codex.log);
  // The plan (one row, ticked off), list_issues, git log, cat (failed), the change of files,
  // get_issue (no result).
  assert.deepStrictEqual(await toolStates(), ['done', 'done', 'done', 'failed', 'done', 'interrupted']);
  assert.strictEqual(await at('taskLog-agentThinking').count(), 2, 'the reasoning');
  assert.ok(text.includes('ERROR rmcp::transport::worker'), 'the log of the CLI stays');
  assert.ok(text.includes('Error: Reconnecting... 1/5'), text);
  assert.ok((await at('taskLog-agentDone').innerText()).includes('tokens: 91.7k in / 191 out'));

  await tools.nth(0).click();
  await tools.nth(1).click();
  await tools.nth(4).click();
  await page.waitForTimeout(500);
  open = await details();
  assert.ok(open.some((d) => d.includes('progress: 3/3') && d.includes('- [x] Read the last commit')), open.join('\n---\n'));
  assert.ok(open.some((d) => d.includes('repo: semaphoreui/semaphore') && d.includes('- number: 4410') && d.includes('See util/config_removed.go.')), open.join('\n---\n'));
  assert.ok(open.some((d) => d.includes('- path: services/api/REVIEW.md') && d.includes('kind: add')), open.join('\n---\n'));
  await shot('06-codex-calls-open');

  await rawLog(codex.task, '> mcp__github__list_issues {"repo":"semaphoreui/semaphore"');

  assert.deepStrictEqual(errors, []);
  console.log('ok: tasks', claude.task.id, codex.task.id);
  await browser.close();
})().catch((e) => { console.error(e); process.exit(1); });
