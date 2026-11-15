// Built-in help on the built binary: the panel asks the Semaphore server for its document and
// nothing else, works with every other host blocked and under a sub-path; dialogs without a
// document have no button (since 2026-11-16 every dialog with input has one); the neighbours of the help (inline hints, Save, Escape) still work.
// usage: NODE_PATH=<playwright> node AGENTS/tools/help-stand/check-binary.cjs
const path = require('node:path');
const { chromium } = require('playwright');

const BASE = process.env.BASE || 'http://127.0.0.1:3100';
const SUB = process.env.SUB || 'http://127.0.0.1:3102/sem';
const OUT = process.env.SHOTS || '/tmp/semaphore-help-stand/shots';
const results = [];
const check = (name, ok, detail = '') => {
  results.push(ok);
  console.log(`${ok ? 'PASS' : 'FAIL'} ${name}${detail ? ` — ${detail}` : ''}`);
};

const DOCS = '[data-testid="editDialog-docs"]';
const PANEL = '[data-testid="editDialog-docsPanel"]';

async function login(ctx, base) {
  const res = await ctx.request.post(`${base}/api/auth/login`, { data: { auth: 'admin', password: 'admin123' } });
  return res.status();
}

async function openInventory(page, base, rowText) {
  await page.goto(`${base}/project/1/inventory`);
  await page.locator('tr', { hasText: rowText }).locator('[data-testid="inventory-rowEdit"]').click();
  await page.waitForSelector('[data-testid="inventoryForm-name"]');
  await page.waitForTimeout(300);
}

(async () => {
  const browser = await chromium.launch({ channel: 'chrome' });
  const errors = [];
  const watch = (page) => {
    page.on('pageerror', (e) => errors.push(e.message));
    page.on('console', (m) => { if (m.type() === 'error') errors.push(m.text()); });
  };

  // ---- Criterion 5, part 1: what the page asks for while the help is used.
  {
    const ctx = await browser.newContext({ viewport: { width: 1280, height: 800 } });
    await login(ctx, BASE);
    const page = await ctx.newPage();
    watch(page);
    await openInventory(page, BASE, 'file hosts');

    const requests = [];
    page.on('request', (r) => requests.push(r.url()));
    await page.locator(`[data-testid="inventoryDialog"] ${DOCS}`).click();
    await page.waitForSelector(PANEL);
    await page.locator('[data-testid="inventoryForm-inventoryPath"] input').click();
    await page.locator('[data-testid="editDialog-docsSection-inventoryForm-type"] h3').hover();
    await page.waitForTimeout(500);
    await page.locator('[data-testid="editDialog-docsClose"]').click();

    const foreign = requests.filter((url) => !url.startsWith(`${BASE}/`));
    const chunk = requests.find((url) => /\/assets\/inventory-ansible-[\w-]+\.js$/.test(url));
    check('dialog: from opening the panel to closing it, every request goes to the Semaphore server',
      foreign.length === 0 && Boolean(chunk), `${requests.length} request(s): ${requests.map((u) => u.replace(BASE, '')).join(', ')}`);

    // The same for the page.
    await page.keyboard.press('Escape');
    await page.goto(`${BASE}/project/1/settings`);
    await page.waitForSelector('[data-testid="newProject-name"]');
    const pageRequests = [];
    page.on('request', (r) => pageRequests.push(r.url()));
    await page.locator('[data-testid="settings-docs"]').click();
    await page.waitForSelector(PANEL);
    await page.waitForTimeout(400);
    const pageChunk = pageRequests.find((url) => /\/assets\/project-settings-[\w-]+\.js$/.test(url));
    check('page: opening the panel asks the Semaphore server for the document and nothing else',
      pageRequests.every((url) => url.startsWith(`${BASE}/`)) && Boolean(pageChunk),
      pageRequests.map((u) => u.replace(BASE, '')).join(', '));
    await ctx.close();
  }

  // ---- Criterion 5, part 2: every host but the server is blocked.
  {
    const ctx = await browser.newContext({ viewport: { width: 1280, height: 800 } });
    const blocked = [];
    await ctx.route('**/*', (route) => {
      const url = route.request().url();
      if (url.startsWith(`${BASE}/`)) {
        return route.continue();
      }
      blocked.push(url);
      return route.abort();
    });
    await login(ctx, BASE);
    const page = await ctx.newPage();
    watch(page);
    await openInventory(page, BASE, 'static hosts');
    await page.locator(`[data-testid="inventoryDialog"] ${DOCS}`).click();
    await page.waitForSelector(PANEL);
    const text = await page.locator(PANEL).innerText();
    check('with every other host blocked the panel of the dialog shows its document',
      text.includes('Edit Ansible Inventory') && text.includes('User Credentials') && text.includes('ansible_port'),
      `${text.length} characters of text`);
    await page.keyboard.press('Escape');

    await page.goto(`${BASE}/project/1/settings`);
    await page.waitForSelector('[data-testid="newProject-name"]');
    await page.locator('[data-testid="settings-docs"]').click();
    await page.waitForSelector(PANEL);
    const pageText = await page.locator(PANEL).innerText();
    check('with every other host blocked the panel of the page shows its document',
      pageText.includes('Project Settings') && pageText.includes('Backup Project'));
    check('the application asked for nothing outside the server at all', blocked.length === 0, blocked.join(', '));
    await ctx.close();
  }

  // ---- A server under a sub-path: the chunk of the document is asked for under the path.
  {
    const ctx = await browser.newContext({ viewport: { width: 1280, height: 800 } });
    const status = await login(ctx, SUB);
    const page = await ctx.newPage();
    watch(page);
    const chunks = [];
    page.on('response', (r) => { if (/\/assets\/(inventory-ansible|project-settings)-/.test(r.url())) chunks.push(`${r.status()} ${r.url()}`); });
    await openInventory(page, SUB, 'file hosts');
    await page.locator(`[data-testid="inventoryDialog"] ${DOCS}`).click();
    await page.waitForSelector(PANEL);
    await page.waitForTimeout(300);
    const title = await page.locator('[data-testid="editDialog-docsTitle"]').innerText();
    check('sub-path: the panel opens and its chunk comes from under the path',
      status === 204 && title === 'Edit Ansible Inventory' && chunks.length === 1
      && chunks[0].startsWith(`200 ${SUB}/assets/inventory-ansible-`), chunks.join(', '));
    await ctx.close();
  }

  // ---- Dialogs and pages without a document; neighbours of the help.
  const ctx = await browser.newContext({ viewport: { width: 1280, height: 800 } });
  await login(ctx, BASE);
  const page = await ctx.newPage();
  watch(page);

  // Criterion 1.
  await openInventory(page, BASE, 'static hosts');
  let order = await page.evaluate(() => [...document.querySelectorAll('[data-testid="inventoryDialog"] .item-dialog__title-actions [data-testid]')].map((el) => el.getAttribute('data-testid')));
  const editTitle = (await page.locator('[data-testid="inventoryDialog"] [data-testid="editDialog-title"]').innerText()).trim();
  await page.keyboard.press('Escape');
  await page.locator('[data-testid="inventory-new"]').click();
  await page.locator('[data-testid="inventory-new-ansible"]').click();
  await page.waitForSelector('[data-testid="inventoryForm-name"]');
  const newTitle = (await page.locator('[data-testid="inventoryDialog"] [data-testid="editDialog-title"]').innerText()).trim();
  const newHasButton = await page.locator(`[data-testid="inventoryDialog"] ${DOCS}`).count();
  check('Edit and New Ansible Inventory have the help button, left of the close button',
    editTitle === 'Edit Ansible Inventory' && newTitle === 'New Ansible Inventory' && newHasButton === 1
    && order.join() === 'editDialog-docs,editDialog-close', `${editTitle}: ${order.join(' → ')}; ${newTitle}: ${newHasButton}`);
  await page.keyboard.press('Escape');

  // The workspace of Terraform is attached, not edited, on this page: its dialog has no help.
  await page.locator('tr', { hasText: 'tf workspace' }).locator('[data-testid="inventory-rowAttach"]').click();
  await page.waitForSelector('[data-testid="inventoryAttachDialog"]');
  await page.waitForTimeout(300);
  const attach = {
    title: (await page.locator('[data-testid="inventoryAttachDialog"] [data-testid="editDialog-title"]').innerText()).trim(),
    button: await page.locator(`[data-testid="inventoryAttachDialog"] ${DOCS}`).count(),
  };
  check('the attach dialog of a Terraform workspace has its own help button', attach.button === 1, `"${attach.title}"`);
  await page.screenshot({ path: path.join(OUT, 'accept-terraform-no-button.png') });
  await page.keyboard.press('Escape');

  const without = [];
  for (const [url, open, name] of [
    ['keys', '[data-testid="keystore-new"]', 'Key'],
    ['repositories', '[data-testid="repositories-new"]', 'Repository'],
    ['environment', '[data-testid="varGroup-new"]', 'Variable Group'],
  ]) {
    await page.goto(`${BASE}/project/1/${url}`);
    await page.locator(open).first().click();
    await page.waitForSelector('.v-overlay--active [data-testid="editDialog-title"]');
    await page.waitForTimeout(400);
    without.push(`${name}: ${await page.locator(`.v-overlay--active ${DOCS}`).count()}`);
    if (name !== 'Variable Group') {
      await page.keyboard.press('Escape');
    }
  }
  check('Key, Repository and Variable Group dialogs have a help button',
    without.every((w) => w.endsWith(': 1')), without.join(', '));

  // The Variable Group dialog is still open: its own button of inline hints works as before.
  const hints = page.locator('.v-overlay--active [data-testid="editDialog-help"]');
  const opacity = () => hints.evaluate((el) => el.style.opacity);
  const before = await opacity();
  await hints.click();
  await page.waitForTimeout(300);
  const after = await opacity();
  check('the button of inline hints of the Variable Group dialog still toggles them, and opens no panel',
    before === '0.3' && after === '1' && (await page.locator(PANEL).count()) === 0, `opacity ${before} → ${after}`);
  await page.keyboard.press('Escape');

  // Criterion 2: the form is edited and saved with the panel open.
  await openInventory(page, BASE, 'static hosts');
  await page.locator(`[data-testid="inventoryDialog"] ${DOCS}`).click();
  await page.waitForSelector(PANEL);
  await page.locator('[data-testid="inventoryForm-name"] input').fill('static hosts (edited)');
  await page.locator('[data-testid="inventoryDialog"] [data-testid="editDialog-save"]').click();
  await page.waitForTimeout(700);
  const saved = await (await ctx.request.get(`${BASE}/api/project/1/inventory/1`)).json();
  check('Save works with the panel open and closes the dialog with the panel',
    saved.name === 'static hosts (edited)' && (await page.locator(PANEL).count()) === 0
    && !(await page.locator('[data-testid="inventoryDialog"]').isVisible()), saved.name);
  await ctx.request.put(`${BASE}/api/project/1/inventory/1`, { data: { ...saved, name: 'static hosts' } });

  // Escape: a menu of the form first, then the dialog with its panel.
  await openInventory(page, BASE, 'file hosts');
  await page.locator(`[data-testid="inventoryDialog"] ${DOCS}`).click();
  await page.waitForSelector(PANEL);
  await page.locator('[data-testid="inventoryForm-type"]').click();
  await page.waitForSelector('.v-overlay--active .v-list-item');
  await page.keyboard.press('Escape');
  await page.waitForTimeout(400);
  const afterFirst = (await page.locator('[data-testid="inventoryDialog"]').isVisible()) && (await page.locator(PANEL).count()) === 1;
  await page.keyboard.press('Escape');
  await page.waitForTimeout(400);
  const afterSecond = !(await page.locator('[data-testid="inventoryDialog"]').isVisible()) && (await page.locator(PANEL).count()) === 0;
  check('Escape closes the open menu first and the dialog with its panel next', afterFirst && afterSecond,
    `after the first: dialog and panel ${afterFirst ? 'stay' : 'GONE'}; after the second: ${afterSecond ? 'closed' : 'STILL OPEN'}`);

  // A project without the backup and cache buttons: their sections are disabled.
  await page.goto(`${BASE}/project/2/settings`);
  await page.waitForSelector('[data-testid="newProject-name"]');
  await page.waitForTimeout(400);
  const buttons = await page.locator('[data-testid="settings-exportProject"], [data-testid="settings-clearCache"]').count();
  await page.locator('[data-testid="settings-docs"]').click();
  await page.waitForSelector(PANEL);
  await page.waitForTimeout(400);
  const disabled = await page.evaluate(() => [...document.querySelectorAll('.HelpPanel__section--disabled')].map((el) => el.getAttribute('data-field')));
  check('a project of another type: no Backup and Clear cache buttons, their sections are disabled',
    buttons === 0 && disabled.join() === 'settings-exportProject,settings-clearCache', disabled.join());
  await page.locator('[data-testid="editDialog-docsSection-settings-clearCache"]').scrollIntoViewIfNeeded();
  await page.screenshot({ path: path.join(OUT, 'accept-settings-typed-project.png') });

  // Criterion 8: both themes, a field marked by its section.
  for (const dark of [false, true]) {
    await page.evaluate((d) => (d ? localStorage.setItem('darkMode', '1') : localStorage.removeItem('darkMode')), dark);
    await openInventory(page, BASE, 'file hosts');
    await page.locator(`[data-testid="inventoryDialog"] ${DOCS}`).click();
    await page.waitForSelector(PANEL);
    await page.locator('[data-testid="editDialog-docsSection-inventoryForm-sshKeyId"] h3').hover();
    await page.waitForTimeout(300);
    await page.screenshot({ path: path.join(OUT, `accept-binary-${dark ? 'dark' : 'light'}.png`) });
    await page.keyboard.press('Escape');
  }
  await page.evaluate(() => localStorage.removeItem('darkMode'));

  await browser.close();
  console.log(errors.length ? `BROWSER ERRORS:\n${errors.join('\n')}` : 'no browser errors');
  console.log(`${results.filter(Boolean).length}/${results.length} checks passed`);
  process.exit(results.every(Boolean) && !errors.length ? 0 : 2);
})().catch((e) => { console.error(e); process.exit(1); });
