// Drives the agent UI on the agent stand and checks the result through the API: saving an mcp
// inventory, creating one, editing an agent template, running agent tasks with and without a
// prompt of their own (the stand runs fake-agent.sh instead of the real CLIs).
// NODE_PATH=~/.npm/_npx/9833c18b2d85bc59/node_modules node AGENTS/tools/agent-stand/verify.cjs
const path = require('path');
const fs = require('fs');
const assert = require('assert');
const { chromium } = require('playwright');

const BASE = process.env.BASE || 'http://127.0.0.1:3102';
const DATA = process.env.AGENT_STAND_DIR || path.join(__dirname, '..', '..', 'tmp', 'agent-stand');
const ids = Object.fromEntries(fs.readFileSync(path.join(DATA, 'ids.txt'), 'utf8').trim().split(' ').map((kv) => kv.split('=')));
const P = ids.project;

(async () => {
  const browser = await chromium.launch({ channel: 'chrome' });
  const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 }, baseURL: BASE });
  assert.ok((await ctx.request.post('/api/auth/login', { data: { auth: 'admin', password: 'admin123' } })).ok());
  const api = async (url) => (await ctx.request.get(url)).json();
  const page = await ctx.newPage();
  const failed = [];
  page.on('response', (r) => { if (r.status() >= 400 && !r.url().endsWith('/branches')) failed.push(`${r.status()} ${r.request().method()} ${r.url()}`); });

  // 0. The agent templates open in the wide editor of Ansible and Terraform.
  await page.goto(`/project/${P}/templates`);
  await page.locator('[data-testid="templates-new"]').waitFor();
  await page.waitForTimeout(500);
  const widths = {};
  for (const app of ['ansible', 'terraform', 'claude', 'codex']) {
    // eslint-disable-next-line no-await-in-loop
    await page.locator('[data-testid="templates-new"]').click();
    // eslint-disable-next-line no-await-in-loop
    await page.locator(`[data-testid="templates-new-${app}"]`).click();
    const editor = page.locator('[data-testid="templateDialog"]').last();
    // eslint-disable-next-line no-await-in-loop
    await editor.locator('[data-testid="templateForm-form"]').waitFor();
    // eslint-disable-next-line no-await-in-loop
    await page.waitForTimeout(400);
    // eslint-disable-next-line no-await-in-loop
    widths[app] = Math.round((await editor.boundingBox()).width);
    // eslint-disable-next-line no-await-in-loop
    await editor.locator('[data-testid="editDialog-cancel"]').click();
    // eslint-disable-next-line no-await-in-loop
    await page.waitForTimeout(400);
  }
  assert.deepStrictEqual([widths.claude, widths.codex], [widths.ansible, widths.terraform], JSON.stringify(widths));
  console.log('0 editor width as Ansible / Terraform: ok', JSON.stringify(widths));

  // 1. Edit the seeded inventory: remove the token of "docs", give "search" one.
  await page.goto(`/project/${P}/inventory`);
  await page.locator('[data-testid="inventory-rowEdit"]').first().click();
  let dialog = page.locator('[data-testid="inventoryDialog"]').last();
  const rows = dialog.locator('[data-testid="mcpInventoryForm-server"]');
  await rows.nth(2).waitFor();
  await rows.nth(1).locator('[data-testid="mcpInventoryForm-serverTokenRemove"]').click();
  await rows.nth(2).locator('[data-testid="mcpInventoryForm-serverToken"] input').fill('fake-token-search');
  const put = page.waitForRequest((r) => r.method() === 'PUT' && r.url().endsWith(`/inventory/${ids.inventory}`));
  await dialog.locator('[data-testid="editDialog-save"]').click();
  const putBody = JSON.parse((await put).postData());
  await page.waitForTimeout(800);
  let inv = await api(`/api/project/${P}/inventory/${ids.inventory}`);
  const byName = Object.fromEntries(inv.mcp_servers.map((s) => [s.name, s]));
  assert.deepStrictEqual(
    [byName.github.has_token, byName.docs.has_token, byName.search.has_token, byName.search.enabled],
    [true, false, true, false],
  );
  assert.ok(!JSON.stringify(inv).includes('fake-token'), 'a GET must not return a token');
  console.log('1 edit inventory: ok', JSON.stringify(putBody.mcp_servers.map((s) => ({ n: s.name, t: s.token && s.token.operation }))));

  // 2. Create an mcp inventory from the menu.
  await page.locator('[data-testid="inventory-new"]').click();
  await page.locator('[data-testid="inventory-new-mcp"]').click();
  dialog = page.locator('[data-testid="inventoryDialog"]').last();
  await dialog.locator('[data-testid="mcpInventoryForm-form"]').waitFor();
  await dialog.locator('[data-testid="mcpInventoryForm-name"] input').fill('Docs MCP');
  const first = dialog.locator('[data-testid="mcpInventoryForm-server"]').first();
  await first.locator('[data-testid="mcpInventoryForm-serverName"] input').fill('context7');
  await first.locator('[data-testid="mcpInventoryForm-serverUrl"] input').fill('https://mcp.example.com/context7');
  await first.locator('[data-testid="mcpInventoryForm-serverToken"] input').fill('fake-token-ctx');
  await dialog.locator('[data-testid="editDialog-save"]').click();
  await page.waitForTimeout(800);
  const list = await api(`/api/project/${P}/inventory`);
  const created = list.find((i) => i.name === 'Docs MCP');
  assert.ok(created && created.type === 'mcp', 'the new inventory is listed as mcp');
  inv = await api(`/api/project/${P}/inventory/${created.id}`);
  assert.deepStrictEqual(inv.mcp_servers.map((s) => [s.name, s.url, s.enabled, s.has_token]), [['context7', 'https://mcp.example.com/context7', true, true]]);
  console.log('2 create inventory: ok', created.id);

  // 3. Edit the Claude Code template: a new prompt is saved in task_params.
  await page.goto(`/project/${P}/templates/${ids.claude_template}`);
  await page.locator('[data-testid="template-edit"]').click();
  const prompt = page.locator('[data-testid="templateForm-agentPrompt"] textarea').first();
  await prompt.waitFor();
  await prompt.fill('Find the flaky tests in services/api and fix them.');
  await page.locator('[data-testid="editDialog-save"]').last().click();
  await page.waitForTimeout(1000);
  let tpl = await api(`/api/project/${P}/templates/${ids.claude_template}`);
  assert.deepStrictEqual(tpl.task_params, { prompt: 'Find the flaky tests in services/api and fix them.', allow_override_prompt: true });
  console.log('3 edit template: ok', JSON.stringify(tpl.task_params));

  // 4. Run it with a prompt of its own.
  await page.locator('[data-testid="template-run"]').click();
  const taskPrompt = page.locator('[data-testid="newTask-agentPrompt"] textarea').first();
  await taskPrompt.waitFor();
  assert.strictEqual(await taskPrompt.inputValue(), 'Find the flaky tests in services/api and fix them.');
  await taskPrompt.fill('Only list the flaky tests, change nothing.');
  const post = page.waitForRequest((r) => r.method() === 'POST' && r.url().endsWith(`/project/${P}/tasks`));
  await page.locator('[data-testid="editDialog-save"]').last().click();
  const postBody = JSON.parse((await post).postData());
  assert.deepStrictEqual(postBody.params.prompt, 'Only list the flaky tests, change nothing.');

  async function finished(templateId) {
    for (let i = 0; i < 60; i += 1) {
      // eslint-disable-next-line no-await-in-loop
      const tasks = await api(`/api/project/${P}/templates/${templateId}/tasks/last`);
      const t = tasks[0];
      if (t && ['success', 'error', 'stopped'].includes(t.status)) return t;
      // eslint-disable-next-line no-await-in-loop
      await page.waitForTimeout(500);
    }
    throw new Error('task did not finish');
  }
  const outputOf = async (task) => (await api(`/api/project/${P}/tasks/${task.id}/output`)).map((o) => o.output).join('\n');

  let task = await finished(ids.claude_template);
  let out = await outputOf(task);
  assert.strictEqual(task.status, 'success', out);
  assert.ok(out.includes('prompt: Only list the flaky tests, change nothing.'), out);
  assert.ok(out.includes('mcp token variables: 1'), out); // github only: docs lost its token, search is disabled
  assert.ok(out.includes('first token: ********'), out);
  assert.ok(!out.includes('fake-token'), 'a token must not reach the log');
  assert.ok(out.includes('anthropic key set: yes'), out);
  console.log('4 run with own prompt: ok, task', task.id);

  // The task's own prompt is shown in its details.
  await page.goto(`/project/${P}/history?t=${task.id}`);
  await page.locator('[data-testid="taskLog-tab-details"]').click();
  const shownPrompt = page.locator('[data-testid="taskDetails-prompt"]');
  await shownPrompt.waitFor();
  assert.strictEqual((await shownPrompt.innerText()).trim(), 'Only list the flaky tests, change nothing.');
  await page.waitForTimeout(400);
  fs.mkdirSync(path.join(DATA, 'shots'), { recursive: true });
  await page.screenshot({ path: path.join(DATA, 'shots', '13-task-details-prompt.png') });
  console.log('4 details show the prompt: ok');

  // 5. Run the Codex template: no prompt field, the template's prompt is used.
  await page.goto(`/project/${P}/templates/${ids.codex_template}`);
  await page.locator('[data-testid="template-run"]').click();
  await page.locator('[data-testid="newTask-form"]').waitFor();
  await page.waitForTimeout(500);
  assert.strictEqual(await page.locator('[data-testid="newTask-agentPrompt"]').count(), 0);
  await page.locator('[data-testid="editDialog-save"]').last().click();
  task = await finished(ids.codex_template);
  out = await outputOf(task);
  assert.strictEqual(task.status, 'success', out);
  assert.ok(out.includes('prompt: Review the last commit and list risky changes.'), out);
  assert.ok(/fake agent args: exec --skip-git-repo-check.* -$/m.test(out), out);
  console.log('5 run with template prompt: ok, task', task.id);

  await page.goto(`/project/${P}/history?t=${task.id}`);
  await page.locator('[data-testid="taskLog-tab-details"]').click();
  // A task without params of its own has no params card at all.
  await page.locator('[data-testid="taskDetails-created"]').waitFor();
  assert.strictEqual(await page.locator('[data-testid="taskDetails-prompt"]').count(), 0);
  console.log('5 details without a prompt of its own: ok');

  tpl = await api(`/api/project/${P}/templates/${ids.claude_template}`);
  console.log(failed.length ? `failed requests:\n${failed.join('\n')}` : 'no failed requests');
  await browser.close();
})().catch((e) => { console.error(e); process.exit(1); });
