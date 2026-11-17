// Built-in help and fields the page keeps but does not show, and fields repeated in rows:
//   - a tab of a dialog the user has left (UserForm, EnvironmentForm: `v-window-item` hides it
//     with `display: none` and keeps it in the page) has its sections disabled, also after the
//     user has visited it and come back;
//   - focus in a field of the second row of a table (the variables of a variable group) marks
//     that row, not the first one.
// usage: NODE_PATH=<playwright> node AGENTS/tools/help-stand/check-hidden-rows.cjs
const path = require('node:path');
const { chromium } = require('playwright');

const BASE = process.env.BASE || 'http://127.0.0.1:3100';
const OUT = process.env.SHOTS || '/tmp/semaphore-help-stand/shots';
const results = [];

const check = (name, ok, detail = '') => {
  results.push(ok);
  console.log(`${ok ? 'PASS' : 'FAIL'} ${name}${detail ? ` — ${detail}` : ''}`);
};

// The state of the sections of the panel on top: which are disabled, which is active, and the
// `data-testid` of the element marked in the form, with whether it is shown.
const panelState = (page) => page.evaluate(() => {
  const panels = [...document.querySelectorAll('[data-testid="editDialog-docsPanel"]')];
  const p = panels[panels.length - 1];
  const disabled = [...p.querySelectorAll('section.HelpPanel__section--disabled')]
    .map((s) => s.getAttribute('data-field'));
  const marks = [...document.querySelectorAll('[data-help-docs-field]')].map((el) => ({
    field: el.closest('[data-testid]').getAttribute('data-testid'),
    shown: el.checkVisibility(),
    row: el.closest('tr') ? [...el.closest('tbody').children].indexOf(el.closest('tr')) : null,
  }));
  return { disabled, marks };
});

async function openPanel(page) {
  const button = page.locator('[data-testid="editDialog-docs"]').last();
  await button.waitFor({ timeout: 5000 });
  await button.click();
  await page.locator('[data-testid="editDialog-docsPanel"]').last().waitFor({ timeout: 5000 });
  await page.waitForTimeout(300);
}

async function tab(page, id) {
  await page.locator(`[data-testid="${id}"]`).click();
  // The transition of v-window ends with `display: none` on the item left.
  await page.waitForTimeout(800);
}

(async () => {
  const browser = await chromium.launch({ channel: 'chrome' });
  const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 } });
  await ctx.addInitScript(() => localStorage.setItem('lang', 'en'));
  await ctx.request.post(`${BASE}/api/auth/login`, { data: { auth: 'admin', password: 'admin123' } });
  const page = await ctx.newPage();
  const errors = [];
  page.on('pageerror', (e) => errors.push(e.message));
  page.on('console', (m) => { if (m.type() === 'error') errors.push(m.text()); });

  // The variable group dialog: tabs Variables and Secrets, rows of variables. (The user dialog
  // has tabs too, but the dialog of an existing user does not open: UserForm reads
  // `item.external` before the user is loaded, as on develop.)
  await page.goto(`${BASE}/project/1/environment`);
  await page.waitForTimeout(700);
  await page.locator('[data-testid="varGroup-new"]').click();
  await page.waitForTimeout(700);
  await openPanel(page);

  // Three rows of variables; the focus goes into the value of the second, the third, the first.
  const add = page.locator('[data-testid="varGroup-addVar"]');
  while (await page.locator('[data-testid="varGroup-varRow"]').count() < 3) {
    await add.click();
    await page.waitForTimeout(200);
  }
  const values = page.locator('[data-testid="varGroup-varValue"] input');
  let s;
  for (const row of [1, 2, 0]) {
    await values.nth(row).click();
    await page.waitForTimeout(300);
    s = await panelState(page);
    check(`variable group: focus in the value of row ${row + 1} marks that row`,
      s.marks.length === 1 && s.marks[0].field === 'varGroup-varValue' && s.marks[0].row === row,
      `marks: ${JSON.stringify(s.marks)}`);
  }
  await values.nth(1).click();
  await page.waitForTimeout(300);
  await page.screenshot({ path: path.join(OUT, 'rows-vargroup.png') });

  s = await panelState(page);
  check('variable group, Variables tab: the sections of Secrets are disabled (never visited)',
    s.disabled.includes('varGroup-secretVarName') && !s.disabled.includes('varGroup-varName'),
    `disabled: ${s.disabled.join(', ')}`);

  await tab(page, 'varGroup-tab-secrets');
  await page.locator('[data-testid="varGroup-addSecretVar"]').click();
  await page.waitForTimeout(300);
  s = await panelState(page);
  check('variable group, Secrets tab: the sections of Variables are disabled',
    s.disabled.includes('varGroup-varName') && !s.disabled.includes('varGroup-secretVarName'),
    `disabled: ${s.disabled.join(', ')}`);

  // A disabled section takes no pointer events; the check is that nothing hidden is marked.
  await page.locator('[data-testid="editDialog-docsSection-varGroup-varName"]').hover();
  await page.waitForTimeout(300);
  s = await panelState(page);
  check('variable group, Secrets tab: the pointer over a Variables section marks nothing hidden',
    s.marks.every((m) => m.shown), `marks: ${JSON.stringify(s.marks)}`);
  await page.screenshot({ path: path.join(OUT, 'hidden-vargroup-secrets.png') });

  await tab(page, 'varGroup-tab-variables');
  s = await panelState(page);
  check('variable group, back on Variables: the sections of Secrets are disabled again',
    s.disabled.includes('varGroup-secretVarName') && !s.disabled.includes('varGroup-varName'),
    `disabled: ${s.disabled.join(', ')}`);

  await browser.close();
  console.log(errors.length ? `BROWSER ERRORS:\n${errors.join('\n')}` : 'no browser errors');
  console.log(`${results.filter(Boolean).length}/${results.length} checks passed`);
  process.exit(results.every(Boolean) && !errors.length ? 0 : 2);
})().catch((e) => { console.error(e); process.exit(1); });
