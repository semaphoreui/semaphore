// Checks that an MCP server token reaches only the URL it was saved with when agent tasks run on a
// remote runner. Two tasks are queued while the runner is down, the server's URL and token are
// changed in the inventory form between them, then the runner starts and gets both jobs in one
// poll. The fake CLI (FAKE_AGENT_SCENARIO=tokens) prints each URL with a hash of the token it got.
// Needs `use_remote_runner: true` in the stand's config.json and runner.sh. Takes the runner's
// registration away at the end, so the other checks run on the server again.
// NODE_PATH=~/.npm/_npx/9833c18b2d85bc59/node_modules node AGENTS/tools/agent-stand/check-runner-tokens.cjs
const path = require('path');
const fs = require('fs');
const crypto = require('crypto');
const assert = require('assert');
const { execFileSync } = require('child_process');
const { chromium } = require('playwright');

const BASE = process.env.BASE || 'http://127.0.0.1:3102';
const DATA = process.env.AGENT_STAND_DIR || path.join(__dirname, '..', '..', 'tmp', 'agent-stand');
const ids = Object.fromEntries(fs.readFileSync(path.join(DATA, 'ids.txt'), 'utf8').trim().split(' ').map((kv) => kv.split('=')));
const P = ids.project;
const runner = (cmd) => console.log(execFileSync(path.join(__dirname, 'runner.sh'), [cmd], { encoding: 'utf8' }).trim());
const sha = (s) => crypto.createHash('sha256').update(s).digest('hex').slice(0, 12);

const OLD = { url: 'https://old-mcp.example.com/mcp', token: 'tok-old-1111' };
const NEW = { url: 'https://new-mcp.example.com/mcp', token: 'tok-new-2222' };

(async () => {
  const browser = await chromium.launch({ channel: 'chrome' });
  const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 }, baseURL: BASE });
  assert.ok((await ctx.request.post('/api/auth/login', { data: { auth: 'admin', password: 'admin123' } })).ok());
  const api = async (method, url, data) => {
    const res = await ctx.request.fetch(url, { method, data });
    assert.ok(res.ok(), `${method} ${url}: ${res.status()} ${await res.text()}`);
    return res.status() === 204 ? null : res.json().catch(() => null);
  };
  const page = await ctx.newPage();

  const envUrl = `/api/project/${P}/environment/${ids.env}`;
  const setScenario = async (on) => {
    const env = await api('GET', envUrl);
    const vars = JSON.parse(env.env || '{}');
    if (on) vars.FAKE_AGENT_SCENARIO = 'tokens'; else delete vars.FAKE_AGENT_SCENARIO;
    await api('PUT', envUrl, { ...env, env: JSON.stringify(vars), secrets: [] });
  };

  // Gives the server "github" (the first row) a URL and a token through the inventory form.
  const setGithub = async ({ url, token }) => {
    await page.goto(`/project/${P}/inventory`);
    await page.locator('[data-testid="inventory-rowEdit"]').first().click();
    const dialog = page.locator('[data-testid="inventoryDialog"]').last();
    const row = dialog.locator('[data-testid="mcpInventoryForm-server"]').first();
    await row.waitFor();
    await row.locator('[data-testid="mcpInventoryForm-serverUrl"] input').fill(url);
    await row.locator('[data-testid="mcpInventoryForm-serverToken"] input').fill(token);
    const saved = page.waitForResponse((r) => r.request().method() === 'PUT' && r.url().endsWith(`/inventory/${ids.inventory}`));
    await dialog.locator('[data-testid="editDialog-save"]').click();
    assert.ok((await saved).ok());
    await page.waitForTimeout(300);
  };

  const start = () => api('POST', `/api/project/${P}/tasks`, { template_id: Number(ids.claude_template) });
  const finished = async (task) => {
    for (let i = 0; i < 120; i += 1) {
      // eslint-disable-next-line no-await-in-loop
      const t = await api('GET', `/api/project/${P}/tasks/${task.id}`);
      if (['success', 'error', 'stopped'].includes(t.status)) return t;
      // eslint-disable-next-line no-await-in-loop
      await page.waitForTimeout(500);
    }
    throw new Error(`task ${task.id} did not end`);
  };

  // Tasks of one template wait for each other unless it allows parallel tasks; the runner must get
  // both jobs at once.
  const tplUrl = `/api/project/${P}/templates/${ids.claude_template}`;
  const tpl = await api('GET', tplUrl);
  const problems = [];
  try {
    await api('PUT', tplUrl, { ...tpl, allow_parallel_tasks: true });
    await setScenario(true);
    runner('start');
    runner('stop');

    await setGithub(OLD);
    const first = await start();
    await setGithub(NEW);
    const second = await start();
    console.log(`queued tasks ${first.id} (old URL and token) and ${second.id} (new URL and token) while the runner is down`);

    runner('start');
    for (const [task, want] of [[first, OLD], [second, NEW]]) {
      // eslint-disable-next-line no-await-in-loop
      const t = await finished(task);
      // eslint-disable-next-line no-await-in-loop
      await page.goto(`/project/${P}/history?t=${task.id}`);
      const log = page.locator('[data-testid="taskLog-output"]').last();
      // eslint-disable-next-line no-await-in-loop
      await log.locator('text=mcp server github').first().waitFor({ timeout: 20000 });
      // eslint-disable-next-line no-await-in-loop
      const line = (await log.innerText()).split('\n').find((l) => l.includes('mcp server github')) || '';
      const got = { url: (line.match(/url=(\S+)/) || [])[1], sha: (line.match(/token-sha=(\S+)/) || [])[1] };
      const ok = got.url === want.url && got.sha === sha(want.token);
      const sent = got.sha === sha(OLD.token) ? 'the old token' : got.sha === sha(NEW.token) ? 'the new token' : got.sha;
      console.log(`task ${task.id} (${t.status}): ${got.url} got ${sent} — ${ok ? 'ok' : 'WRONG'}`);
      if (!ok) problems.push(`task ${task.id}: ${got.url} got ${sent}`);
    }
    fs.mkdirSync(path.join(DATA, 'shots-runner'), { recursive: true });
    await page.locator('[data-testid="taskLog-page"]').last().screenshot({ path: path.join(DATA, 'shots-runner', 'second.png') });
  } finally {
    runner('stop');
    await api('PUT', tplUrl, { ...(await api('GET', tplUrl)), allow_parallel_tasks: tpl.allow_parallel_tasks });
    const rows = await api('GET', '/api/runners');
    for (const r of rows.filter((x) => x.name === 'agent-runner')) {
      // eslint-disable-next-line no-await-in-loop
      await api('DELETE', `/api/runners/${r.id}`);
    }
    await setScenario(false);
    await browser.close();
  }

  if (problems.length) {
    console.log(`A TOKEN WENT TO ANOTHER URL:\n${problems.join('\n')}`);
    process.exit(1);
  }
  console.log('ok: each URL got its own token');
})().catch((e) => { console.error(e); process.exit(1); });
