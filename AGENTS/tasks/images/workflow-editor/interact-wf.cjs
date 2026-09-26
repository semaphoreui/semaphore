// Drives the editor through the criterion-3 checklist and reports each step.
const { chromium } = require('playwright');
const fs = require('fs');

const BASE = 'http://localhost:3100';
const OUT = '/tmp/semaphore-stand/shots-wf';
const [W1] = fs.readFileSync('/tmp/semaphore-stand/wf-ids.txt', 'utf8').trim().split(' ');

(async () => {
  const browser = await chromium.launch({ channel: 'chrome' });
  const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 }, deviceScaleFactor: 2 });
  await ctx.request.post(`${BASE}/api/auth/login`, { data: { auth: 'admin', password: 'admin123' } });
  const page = await ctx.newPage();
  const errors = [];
  page.on('pageerror', (e) => errors.push(e.message));
  const results = [];
  const check = (name, ok, extra = '') => { results.push(`${ok ? 'PASS' : 'FAIL'} ${name} ${extra}`); };
  const model = () => page.evaluate(() => {
    const vm = document.querySelector('.WorkflowGraph').__vue__;
    return vm.exportModel();
  });
  const settle = (ms = 500) => page.waitForTimeout(ms);
  const nodeBox = async (index) => page.locator('.drawflow-node').nth(index).boundingBox();

  await page.goto(`${BASE}/project/1/workflows/${W1}/edit`);
  await page.waitForSelector('.WorkflowNodeCard');
  await settle(1500);
  const initial = await model();
  check('graph loads', initial.nodes.length === 6 && initial.edges.length === 4, `${initial.nodes.length} nodes, ${initial.edges.length} edges`);

  // 1. Quick-add from the "+" handle of the last task (Notify on-call, node 5).
  const notify = page.locator('.drawflow-node', { hasText: 'Notify on-call' });
  await notify.hover();
  await notify.locator('.WorkflowNodeCard__plus').click({ force: true });
  await page.waitForSelector('.WorkflowQuickAddMenu');
  await page.locator('.WorkflowQuickAddMenu .v-list-item', { hasText: 'Approval' }).first().click();
  await settle();
  let m = await model();
  const added = m.nodes.find((n) => n.kind === 'approval' && n.id === 7);
  const link = m.edges.find((e) => e.source_node_id === 5 && e.destination_node_id === 7);
  check('quick-add from plus creates a connected node to the right', !!added && !!link && added.position_x > 700, JSON.stringify({ added, link }));
  check('properties panel opened for the new node', await page.locator('.WorkflowNodeProperties').isVisible());

  // 2. Escape closes the panel and clears the selection.
  await page.locator('.WorkflowGraph__canvas').focus();
  await page.keyboard.press('Escape');
  await settle();
  check('Escape closes properties', !(await page.locator('.WorkflowNodeProperties').isVisible()));
  check('Escape clears Drawflow selection', (await page.locator('.drawflow-node.selected').count()) === 0);

  // 3. Palette click adds a node in the centre.
  await page.locator('.WorkflowEditor__paletteItem--delay').click();
  await settle();
  m = await model();
  check('palette click adds a delay node', m.nodes.some((n) => n.kind === 'delay' && n.id === 8));

  // 4. Connect by dragging from the new delay's output to the new approval's input.
  const src = page.locator('#node-8 .output');
  const dst = page.locator('#node-7 .input');
  const sb = await src.boundingBox();
  const db = await dst.boundingBox();
  await page.mouse.move(sb.x + sb.width / 2, sb.y + sb.height / 2);
  await page.mouse.down();
  await page.mouse.move(db.x + db.width / 2 - 40, db.y + 20, { steps: 8 });
  await page.mouse.move(db.x + db.width / 2, db.y + db.height / 2, { steps: 8 });
  await page.mouse.up();
  await settle();
  m = await model();
  check('drag creates an edge', m.edges.some((e) => e.source_node_id === 8 && e.destination_node_id === 7));

  // 5. Cycle guard: 7 -> 8 would close a cycle (8 -> 7 exists).
  const src2 = page.locator('#node-7 .output');
  const dst2 = page.locator('#node-8 .input');
  const s2 = await src2.boundingBox();
  const d2 = await dst2.boundingBox();
  await page.mouse.move(s2.x + s2.width / 2, s2.y + s2.height / 2);
  await page.mouse.down();
  await page.mouse.move(d2.x + d2.width / 2, d2.y + d2.height / 2, { steps: 12 });
  await page.mouse.up();
  await settle();
  m = await model();
  check('cycle is blocked', !m.edges.some((e) => e.source_node_id === 7 && e.destination_node_id === 8));

  // 6. Pill click changes the condition of edge 1 -> 2 (its midpoint is clear of nodes).
  const pill = page.locator('.WorkflowEdgeLabel').first();
  await pill.locator('.WorkflowEdgeLabel__pill').click();
  await settle();
  await page.locator('.v-menu__content.menuable__content__active .v-list-item', { hasText: 'Always' }).click();
  await settle();
  m = await model();
  const e12 = m.edges.find((e) => e.source_node_id === 1 && e.destination_node_id === 2);
  check('pill menu changes the condition', e12 && e12.condition === 'always', JSON.stringify(e12));

  // 7. Snap: drag the delay node and check its position lands on the grid.
  const nb = await page.locator('#node-8').boundingBox();
  await page.mouse.move(nb.x + 100, nb.y + 30);
  await page.mouse.down();
  await page.mouse.move(nb.x + 137, nb.y + 53, { steps: 6 });
  await page.mouse.up();
  await settle();
  m = await model();
  const moved = m.nodes.find((n) => n.id === 8);
  check('moved node snaps to 20 px grid', moved.position_x % 20 === 0 && moved.position_y % 20 === 0, JSON.stringify({ x: moved.position_x, y: moved.position_y }));

  // 8. Delete key removes the selected node (the delay is selected after drag).
  await page.locator('.WorkflowGraph__canvas').focus();
  await page.keyboard.press('Delete');
  await settle();
  m = await model();
  check('Delete removes the selected node and its edges', !m.nodes.some((n) => n.id === 8) && !m.edges.some((e) => e.source_node_id === 8));

  // 9. Undo / redo over the steps.
  const before = await model();
  let steps = 0;
  for (let i = 0; i < 20; i += 1) {
    const canUndo = await page.evaluate(() => !document.querySelector('.mdi-undo').closest('button').disabled);
    if (!canUndo) break;
    await page.keyboard.press('Meta+z');
    await settle(250);
    steps += 1;
  }
  const afterUndo = await model();
  check('undo walks back to the loaded graph', afterUndo.nodes.length === 6 && afterUndo.edges.length === 4, `${steps} steps`);
  for (let i = 0; i < steps; i += 1) {
    await page.keyboard.press('Meta+Shift+z');
    await settle(250);
  }
  const afterRedo = await model();
  check('redo restores the edited graph', JSON.stringify(afterRedo) === JSON.stringify(before));

  // 10. Problems badge: add a task node without a template.
  await page.locator('.WorkflowEditor__paletteItem--task').click();
  await settle();
  check('problem badge appears on an incomplete task', (await page.locator('.WorkflowNodeCard__badge').count()) >= 1);
  check('save is disabled with problems', await page.locator('button:has-text("Save")').isDisabled());
  await page.screenshot({ path: `${OUT}/editor-interactions.png` });

  // 11. Dirty guard on navigation.
  await page.locator('a:has-text("Workflows")').first().click();
  await settle();
  check('leaving with unsaved changes asks for confirmation', await page.locator('.v-dialog--active').isVisible());

  console.log(results.join('\n'));
  if (errors.length) console.log('PAGE ERRORS:\n' + errors.join('\n'));
  await browser.close();
  process.exit(results.some((r) => r.startsWith('FAIL')) || errors.length ? 1 : 0);
})().catch((e) => { console.error(e); process.exit(1); });
