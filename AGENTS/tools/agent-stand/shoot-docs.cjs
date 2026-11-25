// The screenshots of the docs page about AI agent templates (docs/docs/user-guide/apps/ai-agents.md),
// taken on the agent stand (stand.sh start, seed.sh; the fake CLI): the New Template menu, the
// editor of the Claude Code template, the MCP inventory form and the log of an agent task.
// Gives the seeded Claude template a prompt with ${{ … }} expressions and a model argument for
// the shots and puts the old ones back. Writes webp files (needs cwebp) at twice the size.
// NODE_PATH=~/.npm/_npx/9833c18b2d85bc59/node_modules node AGENTS/tools/agent-stand/shoot-docs.cjs [out_dir]
const path = require('path');
const fs = require('fs');
const assert = require('assert');
const { execFileSync } = require('child_process');
const { chromium } = require('playwright');

const BASE = process.env.BASE || 'http://127.0.0.1:3102';
const ROOT = path.join(__dirname, '..', '..', '..');
const DATA = process.env.AGENT_STAND_DIR || path.join(ROOT, 'AGENTS', 'tmp', 'agent-stand');
const OUT = path.resolve(process.argv[2] || path.join(ROOT, 'docs', 'static', 'assets'));
const ids = Object.fromEntries(fs.readFileSync(path.join(DATA, 'ids.txt'), 'utf8').trim().split(' ').map((kv) => kv.split('=')));
const P = ids.project;

const PROMPT = [
  'Find the flaky tests in services/api and fix them.',
  'Open a pull request against ${{ vars.extra.base_branch }} and post its link',
  'to ${{ vars.env.TRACKER_URL }} with the token ${{ secrets.env.TRACKER_TOKEN }}.',
].join('\n');

fs.mkdirSync(OUT, { recursive: true });

(async () => {
  const browser = await chromium.launch({ channel: 'chrome' });
  const ctx = await browser.newContext({
    viewport: { width: 1280, height: 940 }, deviceScaleFactor: 2, baseURL: BASE,
  });
  assert.ok((await ctx.request.post('/api/auth/login', { data: { auth: 'admin', password: 'admin123' } })).ok());
  const api = async (method, url, data) => {
    const res = await ctx.request.fetch(url, { method, data });
    assert.ok(res.ok(), `${method} ${url}: ${res.status()} ${await res.text()}`);
    return res.status() === 204 ? null : res.json().catch(() => null);
  };

  // The values the prompt names, in the Variable Group of the template.
  const env = await api('GET', `/api/project/${P}/environment/${ids.env}`);
  const has = (type, name) => (env.secrets || []).find((s) => s.type === type && s.name === name);
  const known = has('env', 'TRACKER_TOKEN');
  await api('PUT', `/api/project/${P}/environment/${ids.env}`, {
    ...env,
    json: JSON.stringify({ ...JSON.parse(env.json || '{}'), base_branch: 'main' }),
    env: JSON.stringify({ ...JSON.parse(env.env || '{}'), TRACKER_URL: 'https://tracker.example.com/api' }),
    secrets: [{
      type: 'env',
      name: 'TRACKER_TOKEN',
      secret: 'fake-tracker-token',
      ...(known ? { id: known.id, operation: 'update' } : { operation: 'create' }),
    }],
  });

  // The fake CLI leaves its own lines out of the log while FAKE_AGENT_QUIET is set.
  const quiet = async (on) => {
    const group = await api('GET', `/api/project/${P}/environment/${ids.env}`);
    const vars = JSON.parse(group.env || '{}');
    if (on) vars.FAKE_AGENT_QUIET = '1'; else delete vars.FAKE_AGENT_QUIET;
    await api('PUT', `/api/project/${P}/environment/${ids.env}`, { ...group, env: JSON.stringify(vars), secrets: [] });
  };
  await quiet(true);

  const templateUrl = `/api/project/${P}/templates/${ids.claude_template}`;
  const template = await api('GET', templateUrl);
  await api('PUT', templateUrl, {
    ...template,
    arguments: JSON.stringify(['--model', 'claude-sonnet-5-5']),
    task_params: { ...template.task_params, prompt: PROMPT },
  });

  const page = await ctx.newPage();
  const at = (testId) => page.locator(`[data-testid="${testId}"]`);

  const shot = async (name, locator) => {
    await page.waitForTimeout(500);
    const png = path.join(OUT, `${name}.png`);
    await (locator || page).screenshot({ path: png });
    execFileSync('cwebp', ['-quiet', '-q', '90', png, '-o', path.join(OUT, `${name}.webp`)]);
    fs.unlinkSync(png);
    console.log('shot', path.join(OUT, `${name}.webp`));
  };

  // The card of the dialog on top.
  const dialog = () => at('editDialog-title').last().locator('xpath=ancestor::*[contains(@class,"v-card")][1]');

  try {
    await page.goto(`/project/${P}/templates`);
    await at('templates-new').waitFor();
    await page.waitForTimeout(500);
    await at('templates-new').click();
    await at('templates-new-claude').waitFor();
    await shot('agent-template-new');
    await page.keyboard.press('Escape');

    await page.goto(`/project/${P}/templates/${ids.claude_template}`);
    await at('template-edit').click();
    await at('templateForm-agentPrompt').last().waitFor();
    await page.waitForTimeout(600);
    await shot('agent-template-editor', dialog());
    await at('editDialog-cancel').last().click();

    await page.goto(`/project/${P}/inventory`);
    await at('inventory-table').waitFor();
    await at('inventory-rowEdit').first().click();
    await at('mcpInventoryForm-server').nth(2).waitFor();
    await shot('agent-mcp-inventory', dialog());
    await at('editDialog-cancel').last().click();

    // The log of a task of the template; the fake CLI reports a made-up run.
    const task = await api('POST', `/api/project/${P}/tasks`, { template_id: Number(ids.claude_template) });
    await page.goto(`/project/${P}/templates/${ids.claude_template}?t=${task.id}`);
    await at('taskLog-agentDone').waitFor({ timeout: 60000 });
    await page.waitForTimeout(1500);
    await shot('agent-task-log', at('taskLog-page').last());
  } finally {
    await quiet(false);
    const current = await api('GET', templateUrl);
    await api('PUT', templateUrl, {
      ...current, arguments: template.arguments, task_params: template.task_params,
    });
    await browser.close();
  }
})().catch((e) => { console.error(e); process.exit(1); });
