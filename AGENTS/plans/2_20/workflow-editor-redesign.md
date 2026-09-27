# Workflow editor redesign (n8n-style canvas on Drawflow)

Status: implemented in `feat/refactor_workflows` (2026-09-26). Workbench task
`TASK@88179c9bc1`, research `RESEARCH@5bf4fc406f`, approved static mock
`AGENTS/tasks/2026-09-26-workflow-node-mock.html`.

## Why

The first graphical editor (`AGENTS/plans/2_19/graphical-workflow-editor.md`)
rendered nodes as HTML strings inside Drawflow, had no fit view, no undo, no
edge labels, hard-coded colours and a broken dark theme. Users compared it
unfavourably with n8n / Kestra / Airflow.

## Decision: keep Drawflow, redo the visual layer

The industry standard for a workflow canvas is the xyflow family (Vue Flow /
React Flow) plus an auto-layout engine. Vue Flow is Vue 3 only and the web
app is Vue 2.7 + Vuetify 2. Writing our own canvas and swapping the library
(Rete.js 2) were both rejected by the product owner. Drawflow 0.0.60 therefore
stays the engine (pan, node drag, connection drag, SVG paths, export); every
pixel the user sees is ours.

Four techniques make that possible without patching Drawflow:

1. **Node = Vue component.** `addNode()` gets an empty mount point; the graph
   mounts `WorkflowNodeCard` into it with `parent: this` (Vuetify and i18n are
   inherited) and a reactive `store` shared by all cards. Statuses, countdowns
   and problem badges repaint in place. No `innerHTML`, no `escapeHtml`.
2. **Edge overlays.** After every change `refreshEdges()` colours each
   `svg.connection` by condition, sets an SVG `marker-end` arrow and places a
   condition pill (`WorkflowEdgeLabel`) at `path.getPointAtLength(len / 2)`.
   The overlay element lives inside Drawflow's `precanvas`, so it pans and
   zooms with the graph; it is re-attached after every `clear()`.
3. **Own viewport control.** A capture-phase `wheel` listener replaces
   Drawflow's ctrl-only zoom: plain wheel pans, Ctrl/Cmd (and trackpad pinch)
   zooms towards the cursor. `transform-origin` is forced to `0 0`. Fit view
   computes the bounding box of the DOM nodes; a `ResizeObserver` keeps the
   graph fitted for 1.5 s after a build (Vuetify's drawer settles a frame
   late) unless the user already touched the canvas. Snap to a 20 px grid on
   `nodeMoved`; Tidy up reuses `layoutWorkflowNodes`. Keyboard: arrows pan
   (Shift for larger steps), `+`/`-` zoom, `0` fit, `1` = 100 %; the canvas is
   focused explicitly on pointerdown because Drawflow's fixed mode calls
   `preventDefault()` on mousedown and the run view never got focus otherwise.
4. **Quick add.** A "+" handle on the output port, a floating "+" button and
   right-click on empty canvas open `WorkflowQuickAddMenu` (steps + template
   search). A node added from a handle is placed to the right and connected
   with `on_success` after the cycle guard.

Accepted Drawflow limitations: no multi-select / marquee, no minimap, no note
resize.

## What shipped

- `web/src/components/WorkflowGraph.vue` — rewritten around the store; public
  API: `addNode`, `addNodeAtCenter`, `syncNode`, `setCondition`,
  `removeSelectedNode`, `removeEdge`, `selectNode`, `clearSelection`,
  `fitView`, `zoomIn/zoomOut`, `tidyUp`, `reload`, `getViewport/applyViewport`.
  Props: `nodes`, `edges`, `templates`, `editable`, `nodeRuns`,
  `nodeProblems`, `canResolveApprovals`. Events: `change`, `node-selected`,
  `connection-selected`, `blocked`, `node-click`, `resolve-approval`.
- `web/src/components/workflow/` — `WorkflowNodeCard`, `WorkflowEdgeLabel`,
  `WorkflowCanvasControls`, `WorkflowQuickAddMenu`, `WorkflowNodeProperties`,
  `WorkflowEdgeProperties`, `_palette.scss` (turned into `--wf-*` CSS
  variables on the graph root, dark override via `WorkflowGraph--dark`).
- `web/src/lib/workflowViewport.js` (fit, zoomAt, snap, bounds),
  `web/src/lib/workflowHistory.js` (undo/redo, 50 steps, key-based
  coalescing of property edits), `workflowGraph.js` (+ `statusKind`,
  `edgeRunState`, `formatDurationLong`), `workflowLayout.js` (320 × 120 grid).
- `WorkflowEditor.vue`: toolbar with undo/redo, problems chip, dirty dot on
  Save; palette items add on click; properties drawer; `beforeRouteLeave` +
  `beforeunload` guard.
- `WorkflowRun.vue`: same canvas read-only; status icon + duration on cards,
  Approve/Reject inside the approval card, task log on click; the workflow and
  templates are fetched once so polling never rebuilds the canvas.
- Vuetify's global `background-repeat: no-repeat` reset hid the dot grid; the
  canvas sets `background-repeat: repeat` explicitly.

## Tests and evidence

- `npm run test:unit`: 400 passing (`workflowViewport.spec.js`,
  `workflowHistory.spec.js`, `workflow-node-card.spec.js`, extended
  `workflowGraph.spec.js` / `workflowLayout.spec.js`).
- Playwright checklist against a local stand
  (`AGENTS/tasks/images/workflow-editor/interact-wf.cjs`): quick-add, palette
  click, drag connect, cycle guard, pill condition change, snap, Delete,
  undo/redo round trip, problem badge + disabled Save, leave guard — 16/16.
  `approve-wf.cjs`: approve in card, delay countdown ticks, viewport survives
  polling, task log opens. `keys-wf.cjs`: keyboard pan/zoom/fit/reset in the
  editor and the run view, no effect while typing in the properties panel.
- Screenshots (light/dark, editor/run/legacy/empty) in
  `AGENTS/tasks/images/workflow-editor/`.

## Docs

`docs/docs/user-guide/workflows.md`: rewritten "Creating a workflow", new
"Editor controls" section (navigation matrix mouse / trackpad / buttons /
keyboard plus an editing table), "Approval nodes", "Running and monitoring"; four new
WebP screenshots under `docs/static/assets/workflow-*.webp`; ten locale copies
updated to the same structure (`node scripts/check-docs.mjs`).

## Not done / follow-ups

- Keyboard access to cards (tabindex, Enter opens properties); canvas-level
  keyboard navigation is done.
- Properties panel as a bottom sheet on narrow screens.
- Multi-select, minimap, note resize — need a canvas library change.
