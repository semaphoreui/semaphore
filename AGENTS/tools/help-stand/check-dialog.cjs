// Built-in help of the Ansible inventory dialog: the panel follows the focus in the form, and
// the active section (in focus or under the pointer) marks its field.
// usage: NODE_PATH=<playwright> node AGENTS/tools/help-stand/check-dialog.cjs
const path = require('node:path');
const { chromium } = require('playwright');

const BASE = process.env.BASE || 'http://127.0.0.1:3100';
const OUT = process.env.SHOTS || '/tmp/semaphore-help-stand/shots';
const results = [];

function check(name, ok, detail = '') {
  results.push(ok);
  console.log(`${ok ? 'PASS' : 'FAIL'} ${name}${detail ? ` — ${detail}` : ''}`);
}

const state = (page) => page.evaluate(() => {
  const body = document.querySelector('[data-testid="editDialog-docsBody"]');
  const active = document.querySelector('.HelpPanel__section--active');
  const mark = document.querySelector('[data-help-docs-field]');
  const view = body.getBoundingClientRect();
  const box = active ? active.getBoundingClientRect() : null;
  const field = mark && mark.closest('[data-testid]');
  return {
    active: active ? active.getAttribute('data-field') : null,
    // The title of the active section is inside the view of the panel.
    inView: box ? box.top >= view.top - 1 && box.top < view.bottom : null,
    marked: field ? field.getAttribute('data-testid') : null,
    markedTag: mark ? `${mark.tagName.toLowerCase()}.${[...mark.classList][0]}` : null,
    hidden: [...document.querySelectorAll('[data-testid^="editDialog-docsSectionHidden-"]')]
      .map((el) => el.getAttribute('data-testid').replace('editDialog-docsSectionHidden-', '')),
    focused: document.activeElement.tagName.toLowerCase(),
  };
});

async function openEdit(page, rowText) {
  await page.goto(`${BASE}/project/1/inventory`);
  await page.locator('tr', { hasText: rowText }).locator('[data-testid="inventory-rowEdit"]').click();
  await page.waitForSelector('[data-testid="inventoryForm-name"]');
  await page.waitForTimeout(300);
  await page.locator('[data-testid="inventoryDialog"] [data-testid="editDialog-docs"]').click();
  await page.waitForSelector('[data-testid="editDialog-docsPanel"]');
}

const input = (id) => `[data-testid="${id}"] input`;
const section = (id) => `[data-testid="editDialog-docsSection-${id}"]`;

(async () => {
  const browser = await chromium.launch({ channel: 'chrome' });
  const ctx = await browser.newContext({ viewport: { width: 1280, height: 800 } });
  await ctx.request.post(`${BASE}/api/auth/login`, { data: { auth: 'admin', password: 'admin123' } });
  const page = await ctx.newPage();
  const errors = [];
  page.on('pageerror', (e) => errors.push(e.message));
  page.on('console', (m) => { if (m.type() === 'error') errors.push(m.text()); });

  // ---- Stage 5: an inventory of type File, focus by click and by Tab.
  await openEdit(page, 'file hosts');
  let s = await state(page);
  check('panel opens at its top with no active section', s.active === null
    && (await page.locator('[data-testid="editDialog-docsBody"]').evaluate((el) => el.scrollTop)) === 0);
  check('sections of fields the form does not show say so',
    s.hidden.join() === 'inventoryForm-runnerTag,inventoryForm-inventoryCode', s.hidden.join());

  const fileFields = ['inventoryForm-name', 'inventoryForm-sshKeyId', 'inventoryForm-becomeKeyId',
    'inventoryForm-type', 'inventoryForm-inventoryPath', 'inventoryForm-repositoryId'];

  const card = page.locator('[data-testid="inventoryDialog"]');
  const cardBefore = await card.boundingBox();
  await page.locator(input('inventoryForm-inventoryPath')).click();
  await page.waitForTimeout(700);
  s = await state(page);
  check('click in Path: the panel scrolls to its section and marks the field',
    s.active === 'inventoryForm-inventoryPath' && s.inView && s.marked === 'inventoryForm-inventoryPath'
    && s.focused === 'input', JSON.stringify(s));
  await page.screenshot({ path: path.join(OUT, 'stage5-focus-path-light.png') });

  // Typing goes on while the panel scrolls.
  await page.locator(input('inventoryForm-name')).click();
  await page.keyboard.type(' typed while scrolling');
  await page.waitForTimeout(700);
  const cardAfter = await card.boundingBox();
  check('typing is not interrupted and the dialog does not move',
    (await page.locator(input('inventoryForm-name')).inputValue()) === 'file hosts typed while scrolling'
    && cardBefore.x === cardAfter.x && cardBefore.y === cardAfter.y);

  // Tab through the form.
  const trail = [];
  for (let i = 0; i < 5; i += 1) {
    await page.keyboard.press('Tab');
    await page.waitForTimeout(600);
    const t = await state(page);
    trail.push(`${t.active}${t.inView ? '' : '(OUT OF VIEW)'}`);
  }
  check('Tab through the form: the panel follows every field',
    trail.join() === fileFields.slice(1).join(), trail.join(' → '));

  await page.keyboard.press('Tab'); // Cancel
  await page.waitForTimeout(300);
  s = await state(page);
  check('focus on Cancel: no active section, nothing marked', s.active === null && s.marked === null,
    JSON.stringify({ active: s.active, marked: s.marked, focused: s.focused }));

  // A select: its menu is outside the dialog, the field stays the active one.
  await page.locator('[data-testid="inventoryForm-type"]').click();
  await page.waitForTimeout(400);
  s = await state(page);
  check('menu of Type open: Type stays the active field', s.active === 'inventoryForm-type', s.active);
  await page.locator('.v-overlay--active .v-list-item', { hasText: /^Static$/ }).click();
  await page.waitForTimeout(500);
  s = await state(page);
  check('Type changed to Static: the hidden notes follow the form',
    s.hidden.join() === 'inventoryForm-runnerTag,inventoryForm-inventoryPath,inventoryForm-repositoryId',
    s.hidden.join());

  // The code editor.
  await page.locator('[data-testid="inventoryForm-inventoryCode"] .cm-content').click();
  await page.waitForTimeout(700);
  s = await state(page);
  check('focus in the editor: the Inventory section is active and in view, the editor is marked',
    s.active === 'inventoryForm-inventoryCode' && s.inView && s.marked === 'inventoryForm-inventoryCode',
    JSON.stringify(s));
  await page.screenshot({ path: path.join(OUT, 'stage5-focus-editor-light.png') });
  await page.keyboard.press('Escape');

  // ---- Stage 6: sections under the pointer and in focus.
  await openEdit(page, 'file hosts');
  const marks = [];
  for (const id of fileFields) {
    await page.locator(section(id)).scrollIntoViewIfNeeded();
    await page.locator(`${section(id)} h3`).hover();
    await page.waitForTimeout(250);
    const t = await state(page);
    marks.push(t.active === id && t.marked === id ? 'ok' : `${id}: ${JSON.stringify(t)}`);
    if (id === 'inventoryForm-type') {
      await page.screenshot({ path: path.join(OUT, 'stage6-hover-type-light.png') });
    }
  }
  check('pointer over each section marks its field', marks.every((m) => m === 'ok'), marks.join(' '));
  check('a frame is marked, not the root of the input', (await state(page)).markedTag === 'div.v-field',
    (await state(page)).markedTag);

  await page.mouse.move(300, 300);
  await page.waitForTimeout(250);
  s = await state(page);
  check('pointer away: nothing active, nothing marked', s.active === null && s.marked === null);

  await page.locator(section('inventoryForm-runnerTag')).scrollIntoViewIfNeeded();
  await page.locator(`${section('inventoryForm-runnerTag')} h3`).hover();
  await page.waitForTimeout(250);
  s = await state(page);
  check('section of a field that is not shown is disabled: not active, nothing marked',
    s.active === null && s.marked === null);
  await page.screenshot({ path: path.join(OUT, 'stage6-hidden-field-light.png') });

  // Focus on a section (a click), then the pointer over another one.
  await page.locator(`${section('inventoryForm-sshKeyId')} h3`).click();
  await page.mouse.move(300, 300);
  await page.waitForTimeout(300);
  s = await state(page);
  check('click on a section gives it the focus: its field is marked',
    s.active === 'inventoryForm-sshKeyId' && s.marked === 'inventoryForm-sshKeyId' && s.focused === 'section',
    JSON.stringify(s));
  await page.locator(`${section('inventoryForm-name')} h3`).hover();
  await page.waitForTimeout(250);
  s = await state(page);
  const over = s.marked;
  await page.mouse.move(300, 300);
  await page.waitForTimeout(250);
  s = await state(page);
  check('pointer wins over focus, and the focus comes back after it',
    over === 'inventoryForm-name' && s.marked === 'inventoryForm-sshKeyId', `${over} → ${s.marked}`);

  // Tab from a section to the next one.
  await page.keyboard.press('Tab');
  await page.waitForTimeout(300);
  s = await state(page);
  check('Tab moves the active section to the next one', s.active === 'inventoryForm-becomeKeyId'
    && s.marked === 'inventoryForm-becomeKeyId', JSON.stringify(s));

  // Closing the panel removes the mark.
  await page.locator('[data-testid="editDialog-docsClose"]').click();
  await page.waitForTimeout(300);
  check('panel closed: no mark left in the form',
    (await page.locator('[data-help-docs-field]').count()) === 0);
  await page.keyboard.press('Escape');

  // Dark theme, the editor.
  await page.evaluate(() => localStorage.setItem('darkMode', '1'));
  await openEdit(page, 'static hosts');
  await page.locator(section('inventoryForm-inventoryCode')).scrollIntoViewIfNeeded();
  await page.locator(`${section('inventoryForm-inventoryCode')} h3`).hover();
  await page.waitForTimeout(300);
  s = await state(page);
  check('dark theme: the editor is marked by its section', s.marked === 'inventoryForm-inventoryCode');
  await page.screenshot({ path: path.join(OUT, 'stage6-hover-editor-dark.png') });
  await page.locator(section('inventoryForm-name')).scrollIntoViewIfNeeded();
  await page.locator(`${section('inventoryForm-name')} h3`).hover();
  await page.waitForTimeout(300);
  await page.screenshot({ path: path.join(OUT, 'stage6-hover-name-dark.png') });
  await page.evaluate(() => localStorage.removeItem('darkMode'));

  await browser.close();
  console.log(errors.length ? `BROWSER ERRORS:\n${errors.join('\n')}` : 'no browser errors');
  console.log(`${results.filter(Boolean).length}/${results.length} checks passed`);
  process.exit(results.every(Boolean) && !errors.length ? 0 : 2);
})().catch((e) => { console.error(e); process.exit(1); });
