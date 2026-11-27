// Built-in help of the dialogs: opens a set of dialogs of every kind (EditDialog, HelpDialog, a
// HelpDialog nested over an EditDialog), and in each one checks the help button, the panel, the
// section of the field in focus and the mark on that field, then shoots the dialog.
// usage: NODE_PATH=<playwright> node AGENTS/tools/help-stand/check-all-dialogs.cjs [locale]
const path = require('node:path');
const { chromium } = require('playwright');

const BASE = process.env.BASE || 'http://127.0.0.1:3100';
const OUT = process.env.SHOTS || '/tmp/semaphore-help-stand/shots';
const LOCALE = process.argv[2] || 'en';
const results = [];

const check = (name, ok, detail = '') => {
  results.push(ok);
  console.log(`${ok ? 'PASS' : 'FAIL'} ${name}${detail ? ` — ${detail}` : ''}`);
};

// How to open each dialog: the page, the clicks, the card, the help button, a field to focus.
const DIALOGS = [
  {
    key: 'key', page: '/project/1/keys', open: ['[data-testid="keystore-new"]'],
    card: 'keystoreDialog', button: 'editDialog-docs', field: 'key-name',
  },
  {
    key: 'repository', page: '/project/1/repositories', open: ['[data-testid="repositories-new"]'],
    card: 'repositoriesDialog', button: 'editDialog-docs', field: 'repository-name',
  },
  {
    key: 'variable-group', page: '/project/1/environment', open: ['[data-testid="varGroup-new"]'],
    card: 'varGroupDialog', button: 'editDialog-docs', field: 'varGroup-name',
  },
  {
    key: 'host-config', page: '/project/1/host_config', open: ['[data-testid="hostConfig-new"]'],
    card: 'hostConfigDialog', button: 'editDialog-docs', field: 'hostConfigForm-type',
  },
  {
    key: 'team-member', page: '/project/1/team', open: ['[data-testid="team-new"]'],
    card: 'teamMemberDialog', button: 'editDialog-docs', field: null,
  },
  {
    key: 'template', page: '/project/1/templates',
    open: ['[data-testid="templates-new"]', '[data-testid="templates-newMenu"] .v-list-item >> nth=0'],
    card: null, button: 'editDialog-docs', field: 'templateForm-name',
  },
  {
    key: 'survey-variable', nested: true, page: '/project/1/templates',
    open: ['[data-testid="templates-new"]', '[data-testid="templates-newMenu"] .v-list-item >> nth=0',
      '[data-testid="surveyVars-add"]'],
    card: 'surveyVarDialog', button: 'surveyVarDialog-docs', field: 'surveyVarDialog-name',
  },
  {
    key: 'token-new', page: '/tokens', open: ['[data-testid="tokens-new"]'],
    card: 'tokensNewDialog', button: 'tokensNewDialog-docs', field: 'token-name',
  },
  {
    key: 'runner', page: '/runners', open: ['[data-testid="runners-new"]'],
    card: 'runnerDialog', button: 'editDialog-docs', field: 'runner-name',
  },
  {
    key: 'user', page: '/users', open: ['[data-testid="users-new"]'],
    card: 'userDialog', button: 'editDialog-docs', field: null,
  },
];

(async () => {
  const browser = await chromium.launch({ channel: 'chrome' });
  const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 } });
  await ctx.addInitScript((l) => localStorage.setItem('lang', l), LOCALE);
  await ctx.request.post(`${BASE}/api/auth/login`, { data: { auth: 'admin', password: 'admin123' } });
  const page = await ctx.newPage();
  const errors = [];
  page.on('pageerror', (e) => errors.push(e.message));
  page.on('console', (m) => { if (m.type() === 'error') errors.push(m.text()); });

  for (const d of DIALOGS) {
    try {
      await page.goto(`${BASE}${d.page}`);
      await page.waitForTimeout(700);
      for (const step of d.open) {
        await page.locator(step).first().click();
        await page.waitForTimeout(500);
      }
      // The button of the dialog on top: the last one of its kind in the page.
      const button = page.locator(`[data-testid="${d.button}"]`).last();
      await button.waitFor({ timeout: 5000 });
      await button.click();
      const panel = page.locator('[data-testid="editDialog-docsPanel"]').last();
      await panel.waitFor({ timeout: 5000 });
      await page.waitForTimeout(300);

      const state = await page.evaluate(() => {
        const panels = [...document.querySelectorAll('[data-testid="editDialog-docsPanel"]')];
        const p = panels[panels.length - 1];
        const r = p.getBoundingClientRect();
        return {
          panels: panels.length,
          title: p.querySelector('[data-testid="editDialog-docsTitle"]').textContent,
          sections: p.querySelectorAll('section').length,
          fixed: r.top === 0 && Math.round(r.right) === window.innerWidth && Math.round(r.height) === window.innerHeight,
        };
      });
      let detail = `"${state.title}", ${state.sections} sections, panels open: ${state.panels}`;
      let ok = state.fixed && state.sections > 0;

      if (d.field) {
        const field = page.locator(`[data-testid="${d.field}"]`).last();
        await field.locator('input, textarea, .v-field').first().click();
        await page.waitForTimeout(500);
        const focus = await page.evaluate((id) => {
          const panels = [...document.querySelectorAll('[data-testid="editDialog-docsPanel"]')];
          const p = panels[panels.length - 1];
          const active = p.querySelector('.HelpPanel__section--active');
          const marks = [...document.querySelectorAll('[data-help-docs-field]')]
            .map((el) => el.closest('[data-testid]').getAttribute('data-testid'));
          return { active: active && active.getAttribute('data-field'), marked: marks.includes(id) };
        }, d.field);
        ok = ok && focus.active === d.field && focus.marked;
        detail += `; focus in ${d.field}: section ${focus.active}, marked ${focus.marked}`;
      }
      check(`${d.key}${d.nested ? ' (nested over the template dialog)' : ''}`, ok, detail);
      await page.screenshot({ path: path.join(OUT, `dialog-${LOCALE}-${d.key}.png`) });
    } catch (e) {
      check(d.key, false, e.message.split('\n')[0]);
    }
  }

  await browser.close();
  console.log(errors.length ? `BROWSER ERRORS:\n${errors.join('\n')}` : 'no browser errors');
  console.log(`${results.filter(Boolean).length}/${results.length} checks passed`);
  process.exit(results.every(Boolean) && !errors.length ? 0 : 2);
})().catch((e) => { console.error(e); process.exit(1); });
