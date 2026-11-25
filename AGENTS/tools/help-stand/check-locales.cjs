// Built-in help in the languages of the interface: for each language given, the dialog of an
// Ansible inventory and the project Settings page open their help in that language, the active
// section marks its element and the sections of missing fields are disabled.
// usage: NODE_PATH=<playwright> node AGENTS/tools/help-stand/check-locales.cjs [locale ...]
//        (default: ru ja)
const fs = require('node:fs');
const path = require('node:path');
const { chromium } = require('playwright');

const BASE = process.env.BASE || 'http://127.0.0.1:3100';
const OUT = process.env.SHOTS || '/tmp/semaphore-help-stand/shots';
const LOCALES = process.argv.slice(2).length ? process.argv.slice(2) : ['ru', 'ja'];
const ROOT = path.resolve(__dirname, '../../..');
const results = [];

const check = (name, ok, detail = '') => {
  results.push(ok);
  console.log(`${ok ? 'PASS' : 'FAIL'} ${name}${detail ? ` — ${detail}` : ''}`);
};

// The title and the section titles of a document, read from its file.
function expected(file) {
  const text = fs.readFileSync(path.join(ROOT, 'web/src/help', file), 'utf8');
  return {
    title: /^# (.+)$/m.exec(text)[1].trim(),
    sections: [...text.matchAll(/^## (.+?)\s+\{#[\w-]+\}$/gm)].map((m) => m[1]),
  };
}

const panelText = (page) => page.evaluate(() => ({
  title: document.querySelector('[data-testid="editDialog-docsTitle"]').textContent,
  sections: [...document.querySelectorAll('.HelpPanel__section h3')].map((h) => h.textContent),
  disabled: [...document.querySelectorAll('.HelpPanel__section--disabled')].map((el) => el.getAttribute('data-field')),
}));

const marked = (page) => page.evaluate(() => {
  const mark = document.querySelector('[data-help-docs-field]');
  return mark ? mark.closest('[data-testid]').getAttribute('data-testid') : null;
});

(async () => {
  const browser = await chromium.launch({ channel: 'chrome' });
  const errors = [];

  for (const locale of LOCALES) {
    const ctx = await browser.newContext({ viewport: { width: 1280, height: 800 } });
    await ctx.addInitScript((l) => localStorage.setItem('lang', l), locale);
    await ctx.request.post(`${BASE}/api/auth/login`, { data: { auth: 'admin', password: 'admin123' } });
    const page = await ctx.newPage();
    page.on('pageerror', (e) => errors.push(`${locale}: ${e.message}`));
    page.on('console', (m) => { if (m.type() === 'error') errors.push(`${locale}: ${m.text()}`); });
    const requests = [];
    page.on('request', (r) => requests.push(r.url()));

    // The dialog.
    const dialog = expected(`dialogs/inventory-ansible.${locale}.md`);
    await page.goto(`${BASE}/project/1/inventory`);
    await page.locator('tr', { hasText: 'file hosts' }).locator('[data-testid="inventory-rowEdit"]').click();
    await page.waitForSelector('[data-testid="inventoryForm-name"]');
    await page.waitForTimeout(300);
    const dialogTitle = (await page.locator('[data-testid="inventoryDialog"] [data-testid="editDialog-title"]').innerText()).trim();
    await page.locator('[data-testid="inventoryDialog"] [data-testid="editDialog-docs"]').click();
    await page.waitForSelector('[data-testid="editDialog-docsPanel"]');
    let shown = await panelText(page);
    check(`${locale}: the dialog shows its help in this language, titled as the dialog`,
      shown.title === dialog.title && shown.title === dialogTitle
      && shown.sections.join('|') === dialog.sections.join('|'),
      `dialog "${dialogTitle}", panel "${shown.title}": ${shown.sections.join(' | ')}`);
    check(`${locale}: the editor of a File inventory is disabled in the panel`,
      shown.disabled.join() === 'inventoryForm-runnerTag,inventoryForm-inventoryCode', shown.disabled.join());
    const chunk = requests.find((url) => url.includes(`/assets/inventory-ansible.${locale}-`));
    check(`${locale}: the translation is a chunk of its own from the server`, Boolean(chunk)
      && requests.every((url) => url.startsWith(`${BASE}/`)), chunk ? chunk.replace(BASE, '') : 'no chunk');

    await page.locator('[data-testid="editDialog-docsSection-inventoryForm-sshKeyId"] h3').hover();
    await page.waitForTimeout(300);
    check(`${locale}: a section under the pointer marks its field`,
      (await marked(page)) === 'inventoryForm-sshKeyId');
    await page.screenshot({ path: path.join(OUT, `locale-${locale}-dialog.png`) });
    await page.keyboard.press('Escape');

    // The page.
    const settings = expected(`pages/project-settings.${locale}.md`);
    await page.goto(`${BASE}/project/1/settings`);
    await page.waitForSelector('[data-testid="newProject-name"]');
    await page.locator('[data-testid="settings-docs"]').click();
    await page.waitForSelector('[data-testid="editDialog-docsPanel"]');
    shown = await panelText(page);
    check(`${locale}: the Settings page shows its help in this language`,
      shown.title === settings.title && shown.sections.join('|') === settings.sections.join('|'),
      `"${shown.title}": ${shown.sections.join(' | ')}`);
    await page.locator('[data-testid="newProject-tg"] input').click();
    await page.waitForTimeout(600);
    check(`${locale}: focus in a field of the page marks it and its section`,
      (await marked(page)) === 'newProject-tg');
    await page.screenshot({ path: path.join(OUT, `locale-${locale}-settings.png`) });
    await ctx.close();
  }

  await browser.close();
  console.log(errors.length ? `BROWSER ERRORS:\n${errors.join('\n')}` : 'no browser errors');
  console.log(`${results.filter(Boolean).length}/${results.length} checks passed`);
  process.exit(results.every(Boolean) && !errors.length ? 0 : 2);
})().catch((e) => { console.error(e); process.exit(1); });
