// Screenshots of the AI agent UI on the agent stand: the inventory menu, the mcp inventory form
// (new, edit, token changes, validation), the agent template form and the run dialogs.
// NODE_PATH=~/.npm/_npx/9833c18b2d85bc59/node_modules node AGENTS/tools/agent-stand/shoot.cjs [out_dir] [--dark]
const path = require('path');
const fs = require('fs');
const { chromium } = require('playwright');

const BASE = process.env.BASE || 'http://127.0.0.1:3102';
const OUT = path.resolve(process.argv[2] && !process.argv[2].startsWith('--') ? process.argv[2] : path.join(process.env.AGENT_STAND_DIR || path.join(__dirname, '..', '..', 'tmp', 'agent-stand'), 'shots'));
const DARK = process.argv.includes('--dark');
const DATA = process.env.AGENT_STAND_DIR || path.join(__dirname, '..', '..', 'tmp', 'agent-stand');
const ids = Object.fromEntries(fs.readFileSync(path.join(DATA, 'ids.txt'), 'utf8').trim().split(' ').map((kv) => kv.split('=')));
const P = ids.project;

fs.mkdirSync(OUT, { recursive: true });

async function shot(page, name, locator) {
  await page.waitForTimeout(400);
  const file = path.join(OUT, `${name}${DARK ? '-dark' : ''}.png`);
  if (locator) {
    await locator.screenshot({ path: file });
  } else {
    await page.screenshot({ path: file });
  }
  console.log('shot', file);
}

const activeDialog = (page, testId) => page.locator(`[data-testid="${testId}"]`).last();

(async () => {
  const browser = await chromium.launch({ channel: 'chrome' });
  const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 }, baseURL: BASE });
  const login = await ctx.request.post('/api/auth/login', { data: { auth: 'admin', password: 'admin123' } });
  if (!login.ok()) throw new Error(`login ${login.status()}`);
  if (DARK) {
    await ctx.addInitScript(() => localStorage.setItem('darkMode', '1'));
  }
  const page = await ctx.newPage();
  const errors = [];
  page.on('console', (m) => { if (m.type() === 'error' || m.type() === 'warning') errors.push(`${m.type()}: ${m.text()}`); });
  page.on('pageerror', (e) => errors.push(`pageerror: ${e.message}`));
  page.on('response', (r) => { if (r.status() >= 400) errors.push(`http ${r.status()} ${r.request().method()} ${r.url()}`); });

  // Inventory: the New menu with the MCP kind, then the new and the edit forms.
  await page.goto(`/project/${P}/inventory`);
  await page.locator('[data-testid="inventory-table"]').waitFor();
  await page.locator('[data-testid="inventory-new"]').click();
  await page.locator('[data-testid="inventory-new-mcp"]').waitFor();
  await shot(page, '01-inventory-new-menu');

  await page.locator('[data-testid="inventory-new-mcp"]').click();
  const newDialog = activeDialog(page, 'inventoryDialog');
  await newDialog.locator('[data-testid="mcpInventoryForm-form"]').waitFor();
  await shot(page, '02-mcp-inventory-new', newDialog);
  await newDialog.locator('[data-testid="editDialog-cancel"]').click();
  await page.waitForTimeout(500);

  await page.locator('[data-testid="inventory-rowEdit"]').first().click();
  const editDialog = activeDialog(page, 'inventoryDialog');
  await editDialog.locator('[data-testid="mcpInventoryForm-server"]').nth(2).waitFor();
  await shot(page, '03-mcp-inventory-edit', editDialog);

  // Changes: remove the token of "docs", give "search" one, add a server with mistakes and try
  // to save — the form refuses with the reasons at the fields.
  const rows = editDialog.locator('[data-testid="mcpInventoryForm-server"]');
  // The first server with a token: the stand may have run verify.cjs, which removes one.
  await editDialog.locator('[data-testid="mcpInventoryForm-serverTokenRemove"]').first().click();
  await rows.nth(2).locator('[data-testid="mcpInventoryForm-serverToken"] input').fill('new-fake-token');
  await editDialog.locator('[data-testid="mcpInventoryForm-addServer"]').click();
  await rows.nth(3).locator('[data-testid="mcpInventoryForm-serverName"] input').fill('my server');
  await rows.nth(3).locator('[data-testid="mcpInventoryForm-serverUrl"] input').fill('https://user:secret@mcp.example.com/');
  await editDialog.locator('[data-testid="editDialog-save"]').click();
  await page.waitForTimeout(600);
  await shot(page, '04-mcp-inventory-edit-changes', editDialog);
  await editDialog.locator('[data-testid="editDialog-cancel"]').click();
  await page.waitForTimeout(500);
  await shot(page, '05-inventory-list');

  // Templates: the New menu with the agent apps, a new Claude Code template, the saved one.
  await page.goto(`/project/${P}/templates`);
  await page.locator('[data-testid="templates-new"]').waitFor();
  await page.waitForTimeout(500);
  await page.locator('[data-testid="templates-new"]').click();
  await page.locator('[data-testid="templates-new-claude"]').waitFor();
  await shot(page, '06-templates-new-menu');
  await page.locator('[data-testid="templates-new-claude"]').click();
  await page.locator('[data-testid="templateForm-agentPrompt"]').last().waitFor();
  await page.waitForTimeout(600);
  await shot(page, '07-template-claude-new');
  // Saving without a prompt shows the rule; with "Prompt" allowed for tasks it is optional.
  await page.locator('[data-testid="editDialog-save"]').last().click();
  await page.waitForTimeout(600);
  await shot(page, '08-template-claude-new-validation');
  await page.locator('[data-testid="editDialog-cancel"]').last().click();
  await page.waitForTimeout(500);

  await page.goto(`/project/${P}/templates/${ids.claude_template}`);
  await page.locator('[data-testid="template-edit"]').click();
  await page.locator('[data-testid="templateForm-agentPrompt"]').last().waitFor();
  await page.waitForTimeout(600);
  await shot(page, '09-template-claude-edit');
  await page.locator('[data-testid="editDialog-cancel"]').last().click();
  await page.waitForTimeout(500);

  // Running: the prompt of the template, editable because the template allows it.
  await page.locator('[data-testid="template-run"]').click();
  await page.locator('[data-testid="newTask-agentPrompt"]').waitFor();
  await page.waitForTimeout(400);
  await shot(page, '10-run-claude-prompt', activeDialog(page, 'newTaskDialog'));
  await page.locator('[data-testid="editDialog-cancel"]').last().click();
  await page.waitForTimeout(500);

  await page.goto(`/project/${P}/templates/${ids.codex_template}`);
  await page.locator('[data-testid="template-run"]').click();
  await page.locator('[data-testid="newTask-form"]').waitFor();
  await page.waitForTimeout(600);
  const codexPrompt = await page.locator('[data-testid="newTask-agentPrompt"]').count();
  console.log('codex run dialog prompt fields:', codexPrompt);
  await shot(page, '11-run-codex-no-prompt', activeDialog(page, 'newTaskDialog'));
  await page.locator('[data-testid="editDialog-cancel"]').last().click();

  await page.goto(`/project/${P}/templates`);
  await page.locator('[data-testid="templates-rowName"]').first().waitFor();
  await page.waitForTimeout(500);
  await shot(page, '12-templates-list');

  console.log(errors.length ? `console:\n${errors.join('\n')}` : 'console: clean');
  await browser.close();
})().catch((e) => { console.error(e); process.exit(1); });
