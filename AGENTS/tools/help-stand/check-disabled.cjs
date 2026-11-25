// Built-in help of the Ansible inventory dialog: sections are cards, and the section of a field
// the form does not show now is disabled; both themes, both inventory types.
// usage: NODE_PATH=<playwright> node AGENTS/tools/help-stand/check-disabled.cjs
const path = require('node:path');
const { chromium } = require('playwright');

const BASE = process.env.BASE || 'http://127.0.0.1:3100';
const OUT = process.env.SHOTS || '/tmp/semaphore-help-stand/shots';
const results = [];
const check = (name, ok, detail = '') => {
  results.push(ok);
  console.log(`${ok ? 'PASS' : 'FAIL'} ${name}${detail ? ` — ${detail}` : ''}`);
};
const section = (id) => `[data-testid="editDialog-docsSection-inventoryForm-${id}"]`;

const sections = (page) => page.evaluate(() => Object.fromEntries(
  [...document.querySelectorAll('.HelpPanel__section')].map((el) => [
    el.getAttribute('data-field').replace('inventoryForm-', ''),
    {
      disabled: el.classList.contains('HelpPanel__section--disabled'),
      active: el.classList.contains('HelpPanel__section--active'),
      tab: el.getAttribute('tabindex'),
      opacity: getComputedStyle(el.querySelector('.HelpPanel__sectionTitle')).opacity,
      background: getComputedStyle(el).backgroundColor,
    },
  ]),
));
const disabledOf = (all) => Object.keys(all).filter((k) => all[k].disabled).join();

async function hover(page, id) {
  await page.locator(section(id)).scrollIntoViewIfNeeded();
  await page.locator(`${section(id)} h3`).hover();
  await page.waitForTimeout(300);
}

(async () => {
  const browser = await chromium.launch({ channel: 'chrome' });
  const ctx = await browser.newContext({ viewport: { width: 1280, height: 800 } });
  await ctx.request.post(`${BASE}/api/auth/login`, { data: { auth: 'admin', password: 'admin123' } });
  const page = await ctx.newPage();
  const errors = [];
  page.on('pageerror', (e) => errors.push(e.message));
  page.on('console', (m) => { if (m.type() === 'error') errors.push(m.text()); });

  for (const dark of [false, true]) {
    await page.goto(`${BASE}/project/1/inventory`);
    await page.evaluate((d) => (d ? localStorage.setItem('darkMode', '1') : localStorage.removeItem('darkMode')), dark);
    await page.goto(`${BASE}/project/1/inventory`);
    await page.locator('tr', { hasText: 'file hosts' }).locator('[data-testid="inventory-rowEdit"]').click();
    await page.waitForSelector('[data-testid="inventoryForm-name"]');
    await page.waitForTimeout(300);
    await page.locator('[data-testid="inventoryDialog"] [data-testid="editDialog-docs"]').click();
    await page.waitForSelector('[data-testid="editDialog-docsPanel"]');
    const theme = dark ? 'dark' : 'light';

    // Type = File: the editor is not on the form.
    let all = await sections(page);
    check(`${theme}, File: Runner tag and Inventory are disabled, the rest is not`,
      disabledOf(all) === 'runnerTag,inventoryCode', disabledOf(all));
    check(`${theme}, File: a disabled section is dimmed and out of the Tab order`,
      all.inventoryCode.opacity === '0.45' && all.inventoryCode.tab === null
      && all.name.opacity === '1' && all.name.tab === '0',
      `opacity ${all.inventoryCode.opacity} vs ${all.name.opacity}, tabindex ${all.inventoryCode.tab} vs ${all.name.tab}`);

    const idle = all.inventoryCode.background;
    check(`${theme}: an enabled card has a background of its own, a disabled one a paler one`,
      all.name.background !== 'rgba(0, 0, 0, 0)' && idle !== all.name.background,
      `enabled ${all.name.background}, disabled ${idle}`);
    await hover(page, 'inventoryCode');
    all = await sections(page);
    check(`${theme}, File: the pointer over the disabled Inventory section changes nothing`,
      !all.inventoryCode.active && all.inventoryCode.background === idle
      && (await page.locator('[data-help-docs-field]').count()) === 0, all.inventoryCode.background);
    await page.screenshot({ path: path.join(OUT, `stage9-file-${theme}.png`) });

    await page.locator(`${section('inventoryCode')} h3`).click();
    await page.waitForTimeout(300);
    all = await sections(page);
    check(`${theme}, File: a click on the disabled section does not activate it`,
      !all.inventoryCode.active && (await page.locator('[data-help-docs-field]').count()) === 0);

    await hover(page, 'inventoryPath');
    all = await sections(page);
    check(`${theme}, File: an enabled section still highlights`, all.inventoryPath.active
      && all.inventoryPath.background !== all.name.background
      && (await page.locator('[data-help-docs-field]').count()) === 1,
      `active ${all.inventoryPath.background}, idle ${all.name.background}`);
    const gaps = await page.evaluate(() => {
      const cards = [...document.querySelectorAll('.HelpPanel__section')];
      return cards.slice(1).map((el, i) => Math.round(
        el.getBoundingClientRect().top - cards[i].getBoundingClientRect().bottom,
      ));
    });
    check(`${theme}: 12 px between the cards`, gaps.every((gap) => gap === 12), gaps.join());
    await page.locator('[data-testid="editDialog-docsBody"]').evaluate((el) => { el.scrollTop = 0; });
    await page.mouse.move(300, 300);
    await page.waitForTimeout(300);
    await page.screenshot({ path: path.join(OUT, `stage10-top-${theme}.png`) });
    await hover(page, 'inventoryPath');

    // Type = Static: Path and Repository go, the editor comes.
    await page.locator('[data-testid="inventoryForm-type"]').click();
    await page.locator('.v-overlay--active .v-list-item', { hasText: /^Static$/ }).click();
    await page.waitForTimeout(500);
    all = await sections(page);
    check(`${theme}, Static: Runner tag, Path and Repository are disabled, Inventory is not`,
      disabledOf(all) === 'runnerTag,inventoryPath,repositoryId', disabledOf(all));
    await hover(page, 'inventoryCode');
    all = await sections(page);
    check(`${theme}, Static: the Inventory section highlights the editor`, all.inventoryCode.active
      && (await page.locator('[data-help-docs-field]').count()) === 1);
    await hover(page, 'inventoryPath');
    all = await sections(page);
    // The focus is in the Type field (its menu has just closed), so Type stays the active field.
    const markedField = await page.evaluate(() => {
      const mark = document.querySelector('[data-help-docs-field]');
      return mark ? mark.closest('[data-testid]').getAttribute('data-testid') : null;
    });
    check(`${theme}, Static: the pointer over the disabled Path section activates nothing of its own`,
      !all.inventoryPath.active && all.type.active && markedField === 'inventoryForm-type',
      `active: ${Object.keys(all).filter((k) => all[k].active).join()}, marked: ${markedField}`);
    await page.screenshot({ path: path.join(OUT, `stage9-static-${theme}.png`) });

    // Tab through the panel skips the disabled sections.
    await page.locator(`${section('name')} h3`).click();
    const trail = [];
    for (let i = 0; i < 4; i += 1) {
      await page.keyboard.press('Tab');
      trail.push(await page.evaluate(() => {
        const s = document.activeElement.closest('.HelpPanel__section');
        return s ? s.getAttribute('data-field').replace('inventoryForm-', '') : document.activeElement.tagName.toLowerCase();
      }));
    }
    check(`${theme}, Static: Tab goes over the enabled sections only`,
      trail.slice(0, 4).join() === 'sshKeyId,becomeKeyId,type,inventoryCode', trail.join(' → '));
    await page.keyboard.press('Escape');
  }
  await page.evaluate(() => localStorage.removeItem('darkMode'));

  await browser.close();
  console.log(errors.length ? `BROWSER ERRORS:\n${errors.join('\n')}` : 'no browser errors');
  console.log(`${results.filter(Boolean).length}/${results.length} checks passed`);
  process.exit(results.every(Boolean) && !errors.length ? 0 : 2);
})().catch((e) => { console.error(e); process.exit(1); });
