// Built-in help of the project Settings page: the button, the layout next to the panel, focus
// and pointer, saving with the panel open.
// usage: NODE_PATH=<playwright> node AGENTS/tools/help-stand/check-settings.cjs
const path = require('node:path');
const { chromium } = require('playwright');

const BASE = process.env.BASE || 'http://127.0.0.1:3100';
const OUT = process.env.SHOTS || '/tmp/semaphore-help-stand/shots';
const results = [];
const check = (name, ok, detail = '') => {
  results.push(ok);
  console.log(`${ok ? 'PASS' : 'FAIL'} ${name}${detail ? ` — ${detail}` : ''}`);
};
const section = (id) => `[data-testid="editDialog-docsSection-${id}"]`;

const state = (page) => page.evaluate(() => {
  const body = document.querySelector('[data-testid="editDialog-docsBody"]');
  const active = document.querySelector('.HelpPanel__section--active');
  const mark = document.querySelector('[data-help-docs-field]');
  const view = body.getBoundingClientRect();
  const box = active ? active.getBoundingClientRect() : null;
  const field = mark && mark.closest('[data-testid]');
  return {
    active: active ? active.getAttribute('data-field') : null,
    inView: box ? box.top >= view.top - 1 && box.top < view.bottom : null,
    marked: field ? field.getAttribute('data-testid') : null,
    disabled: [...document.querySelectorAll('.HelpPanel__section--disabled')].map((el) => el.getAttribute('data-field')),
  };
});

async function openSettings(page) {
  await page.goto(`${BASE}/project/1/settings`);
  await page.waitForSelector('[data-testid="newProject-name"]');
  await page.waitForTimeout(400);
}

async function openPanel(page) {
  await page.locator('[data-testid="settings-docs"]').click();
  await page.waitForSelector('[data-testid="editDialog-docsPanel"]');
  await page.waitForTimeout(300);
}

(async () => {
  const browser = await chromium.launch({ channel: 'chrome' });
  const ctx = await browser.newContext({ viewport: { width: 1280, height: 900 } });
  await ctx.request.post(`${BASE}/api/auth/login`, { data: { auth: 'admin', password: 'admin123' } });
  const page = await ctx.newPage();
  const errors = [];
  page.on('pageerror', (e) => errors.push(e.message));
  page.on('console', (m) => { if (m.type() === 'error') errors.push(m.text()); });

  await openSettings(page);
  check('the page has a help button and no panel until it is pressed',
    (await page.locator('[data-testid="settings-docs"]').count()) === 1
    && (await page.locator('[data-testid="editDialog-docsPanel"]').count()) === 0);
  await page.screenshot({ path: path.join(OUT, 'settings-closed-light.png') });

  await openPanel(page);
  const geo = await page.evaluate(() => {
    const p = document.querySelector('[data-testid="editDialog-docsPanel"]').getBoundingClientRect();
    return { top: p.top, right: p.right, height: p.height, w: window.innerWidth, h: window.innerHeight };
  });
  check('the panel is fixed to the right edge of the window, full height',
    geo.top === 0 && geo.right === geo.w && geo.height === geo.h, JSON.stringify(geo));
  const doc = await page.evaluate(() => ({
    title: document.querySelector('[data-testid="editDialog-docsTitle"]').textContent,
    sections: [...document.querySelectorAll('.HelpPanel__section h3')].map((h) => h.textContent),
  }));
  check('the panel shows the document of the page', doc.title === 'Project Settings' && doc.sections.length === 8,
    `${doc.title}: ${doc.sections.join(' | ')}`);
  let s = await state(page);
  check('nothing is disabled on a project with backup and cache buttons', s.disabled.length === 0, s.disabled.join());
  await page.screenshot({ path: path.join(OUT, 'settings-open-light.png') });

  // Layout: the content of the page next to the panel.
  for (const width of [1440, 1280, 1100, 1024]) {
    await page.setViewportSize({ width, height: 900 });
    await page.waitForTimeout(300);
    const o = await page.evaluate(() => {
      const c = document.querySelector('.project-form').getBoundingClientRect();
      const t = document.querySelector('[data-testid="settings-docs"]').getBoundingClientRect();
      const p = document.querySelector('[data-testid="editDialog-docsPanel"]').getBoundingClientRect();
      return { form: `${Math.round(c.left)}..${Math.round(c.right)}`, formWidth: Math.round(c.width),
        overlap: Math.max(0, Math.round(c.right - p.left)), button: Math.round(t.right), panel: Math.round(p.left) };
    });
    const beside = width >= 1100;
    check(`width ${width}: ${beside ? 'the form and the help button are beside the panel' : 'the panel lies over the page'}`,
      beside ? o.overlap === 0 && o.button <= o.panel : o.overlap > 0,
      `form ${o.form} (${o.formWidth} px), help button ends at ${o.button}, panel from ${o.panel}`);
  }
  await page.setViewportSize({ width: 1280, height: 900 });
  await page.waitForTimeout(300);

  // Focus in the fields and on the buttons: the panel follows.
  const input = (id) => `[data-testid="${id}"] input`;
  await page.locator(input('newProject-tg')).click();
  await page.waitForTimeout(600);
  s = await state(page);
  check('click in Telegram Chat ID: its section is active, in view, the field is marked',
    s.active === 'newProject-tg' && s.inView && s.marked === 'newProject-tg', JSON.stringify(s));

  await page.locator(input('newProject-name')).click();
  const trail = [];
  for (let i = 0; i < 4; i += 1) {
    await page.keyboard.press('Tab');
    await page.waitForTimeout(600);
    const t = await state(page);
    trail.push(`${t.active}${t.inView ? '' : '(OUT OF VIEW)'}${t.marked === t.active ? '' : `(marked ${t.marked})`}`);
  }
  check('Tab from Project Name: fields, the checkbox, then the Test Alerts button',
    trail.join() === 'newProject-maxParallelTasks,newProject-tg,newProject-alert,settings-testAlerts', trail.join(' → '));
  await page.screenshot({ path: path.join(OUT, 'settings-focus-testAlerts-light.png') });

  for (const id of ['settings-exportProject', 'settings-clearCache', 'settings-deleteProject']) {
    await page.locator(`[data-testid="${id}"]`).focus();
    await page.waitForTimeout(600);
    s = await state(page);
    check(`focus on ${id}: its section is active, in view, the button is marked`,
      s.active === id && s.inView && s.marked === id, JSON.stringify(s));
  }
  await page.screenshot({ path: path.join(OUT, 'settings-focus-delete-light.png') });

  // Sections under the pointer.
  const marks = [];
  for (const id of ['newProject-name', 'newProject-maxParallelTasks', 'newProject-tg', 'newProject-alert',
    'settings-testAlerts', 'settings-exportProject', 'settings-clearCache', 'settings-deleteProject']) {
    await page.locator(section(id)).scrollIntoViewIfNeeded();
    await page.locator(`${section(id)} h3`).hover();
    await page.waitForTimeout(250);
    const t = await state(page);
    marks.push(t.active === id && t.marked === id ? 'ok' : `${id}: ${JSON.stringify(t)}`);
    if (id === 'newProject-alert') {
      await page.screenshot({ path: path.join(OUT, 'settings-hover-alert-light.png') });
    }
  }
  check('the pointer over each of the eight sections marks its element', marks.every((m) => m === 'ok'), marks.join(' '));
  await page.mouse.move(400, 500);
  await page.waitForTimeout(300);

  // A click on a section of an element far down the page scrolls the page to it.
  await page.setViewportSize({ width: 1280, height: 500 });
  await page.evaluate(() => window.scrollTo(0, 0));
  await page.locator(section('settings-deleteProject')).scrollIntoViewIfNeeded();
  await page.locator(`${section('settings-deleteProject')} h3`).click();
  await page.waitForTimeout(500);
  const visible = await page.evaluate(() => {
    const r = document.querySelector('[data-testid="settings-deleteProject"]').getBoundingClientRect();
    return r.top >= 0 && r.bottom <= window.innerHeight;
  });
  check('a click on the Delete Project section brings the button into view', visible);
  await page.setViewportSize({ width: 1280, height: 900 });

  // The form is saved with the panel open.
  await page.locator(input('newProject-name')).fill('Help QA renamed');
  await page.locator('[data-testid="settings-save"]').click();
  await page.waitForTimeout(800);
  const saved = await (await ctx.request.get(`${BASE}/api/project/1`)).json();
  check('Save works with the panel open', saved.name === 'Help QA renamed'
    && (await page.locator('[data-testid="editDialog-docsPanel"]').count()) === 1, saved.name);
  await ctx.request.put(`${BASE}/api/project/1`, { data: { ...saved, name: 'Help QA' } });

  // Closing.
  await page.locator('[data-testid="editDialog-docsClose"]').click();
  await page.waitForTimeout(300);
  const padding = await page.evaluate(() => getComputedStyle(document.querySelector('[data-testid="settings-page"]')).paddingRight);
  check('the panel closes with its button: no mark, the page takes its width back',
    (await page.locator('[data-testid="editDialog-docsPanel"]').count()) === 0
    && (await page.locator('[data-help-docs-field]').count()) === 0 && padding === '0px', padding);

  // Dark theme.
  await page.evaluate(() => localStorage.setItem('darkMode', '1'));
  await openSettings(page);
  await openPanel(page);
  await page.locator(section('settings-exportProject')).scrollIntoViewIfNeeded();
  await page.locator(`${section('settings-exportProject')} h3`).hover();
  await page.waitForTimeout(300);
  s = await state(page);
  check('dark theme: the Backup section marks its button', s.marked === 'settings-exportProject');
  await page.screenshot({ path: path.join(OUT, 'settings-hover-backup-dark.png') });
  await page.evaluate(() => localStorage.removeItem('darkMode'));

  await browser.close();
  console.log(errors.length ? `BROWSER ERRORS:\n${errors.join('\n')}` : 'no browser errors');
  console.log(`${results.filter(Boolean).length}/${results.length} checks passed`);
  process.exit(results.every(Boolean) && !errors.length ? 0 : 2);
})().catch((e) => { console.error(e); process.exit(1); });
