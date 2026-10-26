// Checklist for the edge inputs UI on the stand: open the "Inputs: explicit
// empty list" workflow, select its explicit edge, type a key for image_tag in
// the panel, save, and read the mapping back through the API; then switch the
// edge to by_name, save, and check the fields are gone. Restores the original
// graph through the API at the end. Exits 1 on any FAIL or page error.
// usage: NODE_PATH=~/.npm/_npx/9833c18b2d85bc59/node_modules node interact-inputs.cjs   (WF=<workflow id, default 12>)
const { chromium } = require('playwright');

const BASE = process.env.BASE || 'http://localhost:3100';
const PID = process.env.PID || 1;
const WF = process.env.WF || 12;
const results = [];
const check = (name, ok, detail = '') => results.push(`${ok ? 'PASS' : 'FAIL'} ${name}${detail ? ' — ' + detail : ''}`);

const selectPath = (sel) => {
  const path = document.querySelector(sel);
  if (!path) return false;
  ['mousedown', 'mouseup'].forEach((type) => path.dispatchEvent(
    new MouseEvent(type, { bubbles: true, cancelable: true, button: 0 }),
  ));
  return true;
};

(async () => {
  const browser = await chromium.launch({ channel: 'chrome' });
  const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 } });
  await ctx.request.post(`${BASE}/api/auth/login`, { data: { auth: 'admin', password: 'admin123' } });
  const api = (p) => ctx.request.get(`${BASE}/api/project/${PID}${p}`).then((r) => r.json());
  const original = await api(`/workflows/${WF}`);
  const page = await ctx.newPage();
  const errors = [];
  page.on('pageerror', (e) => errors.push(e.message));
  page.on('console', (m) => { if (m.type() === 'error') errors.push(m.text()); });

  try {
    await page.goto(`${BASE}/project/${PID}/workflows/${WF}/edit`);
    await page.waitForSelector('.WorkflowNodeCard', { timeout: 15000 });
    await page.waitForTimeout(1000);

    check('explicit edge selected', await page.evaluate(selectPath, 'svg.WorkflowGraph__conn--explicit path.main-path'));
    await page.waitForSelector('.WorkflowEdgeProperties__row', { timeout: 5000 });
    // Save is gated by permissions and validity, not by dirtiness; the
    // unsaved-changes flag is what must stay clean until something is edited.
    const dirtyBefore = await page.evaluate(() => document.querySelector('.WorkflowEditor').__vue__.dirty);
    check('not dirty after merely selecting the edge', dirtyBefore === false);

    // Type a key for image_tag (first row) and leave the field.
    const keyInput = page.locator('.WorkflowEdgeProperties__row').first().locator('input[type="text"]');
    await keyInput.fill('image_tag');
    await keyInput.press('Tab');
    await page.waitForTimeout(400);
    check('dirty after typing a key', await page.evaluate(() => document.querySelector('.WorkflowEditor').__vue__.dirty));
    check('explicit pill icon on the canvas', (await page.locator('.WorkflowEdgeLabel__inputs').count()) > 0);

    await page.getByRole('button', { name: /^save$/i }).click();
    await page.waitForTimeout(1500);
    const saved = await api(`/workflows/${WF}`);
    const edge = saved.edges[0];
    check('saved edge is explicit with the typed pair', edge.input_mode === 'explicit'
      && JSON.stringify(edge.input_mappings) === JSON.stringify([{ var: 'image_tag', key: 'image_tag' }]),
    JSON.stringify({ input_mode: edge.input_mode, input_mappings: edge.input_mappings }));

    // Switch to by_name: the fields must disappear from the saved edge.
    await page.waitForTimeout(500);
    check('edge selected again after save', await page.evaluate(selectPath, 'svg.WorkflowGraph__conn--explicit path.main-path'));
    await page.waitForSelector('.WorkflowEdgeProperties', { timeout: 5000 });
    await page.locator('.WorkflowEdgeProperties .v-input--checkbox').click();
    await page.waitForTimeout(400);
    check('by_name chips shown after unchecking', (await page.locator('.WorkflowEdgeProperties__chips .v-chip').count()) > 0);
    check('explicit pill icon gone', (await page.locator('.WorkflowEdgeLabel__inputs').count()) === 0);
    await page.getByRole('button', { name: /^save$/i }).click();
    await page.waitForTimeout(1500);
    const saved2 = await api(`/workflows/${WF}`);
    check('saved edge is by_name without mappings', saved2.edges[0].input_mode === 'by_name' && !saved2.edges[0].input_mappings,
      JSON.stringify({ input_mode: saved2.edges[0].input_mode, input_mappings: saved2.edges[0].input_mappings }));
  } finally {
    // Restore the seeded graph (a new revision with the original edges).
    const restore = await ctx.request.put(`${BASE}/api/project/${PID}/workflows/${WF}`, { data: original });
    check('original graph restored', restore.status() === 204, `HTTP ${restore.status()}`);
    await browser.close();
  }

  results.forEach((r) => console.log(r));
  if (errors.length) console.log('BROWSER ERRORS:\n' + errors.join('\n'));
  process.exit(errors.length || results.some((r) => r.startsWith('FAIL')) ? 1 : 0);
})().catch((e) => { console.error(e); process.exit(1); });
