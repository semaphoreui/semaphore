# Workflow editor — architecture and decisions

The graphical editor for workflow DAGs: a Drawflow 0.0.60 engine (pan, node drag, connection
drag, SVG paths) under an n8n-style visual layer that is entirely ours — nodes are mounted Vue
components, edges get colour, arrows and condition pills, viewport control is ours. One renderer
(`WorkflowGraph.vue`) serves the full-page editor and the read-only run view. Two generations:
the first cut (plan `graphical-workflow-editor`, shipped on develop with PR #3962/#4145) and the
redesign on `feat/refactor_workflows` (task `TASK@88179c9bc1`, 2026-09-26/27).

**Status (2026-09-27): redesign implemented, unmerged** — 14 commits ahead of develop
(`5fdfe0ae..539d98ad`), no PR opened yet (`gh pr list --head feat/refactor_workflows` is
empty). Evidence: `web/src/components/workflow/` holds the six card/pill/menu components,
`web/src/lib/workflow{Viewport,History,Layout,Graph,EditorPrefs}.js` with specs in
`web/tests/unit/`, and `grep nodeHtml|innerHTML|escapeHtml` over `WorkflowGraph.vue` +
`components/workflow/` + `views/project/Workflow*.vue` returns nothing. Docs already updated in
the `docs` submodule (`docs/docs/user-guide/workflows.md`, `workflow-hotkeys.svg`).

## Model and backend (unchanged by the redesign)

- `db.WorkflowNode` — kinds `task | approval | delay | note`, `convergence_mode all|any`,
  `position_x/position_y` (`int`), `note`, `delay_seconds`; `db.WorkflowEdge.condition`
  `on_success | on_failure | always`. All in `db/sql/migrations/v2.18.15.sql`.
- Every save is delete-and-reinsert (`writeWorkflowGraph`): node DB ids change on each PUT, edges
  are remapped through `nodeIDMap`. Consequence: past runs' `task.workflow_node_id` dangle after an
  edit; the run view tolerates missing nodes.
- `pro/db.ValidateWorkflowTemplate` is the source of truth; the editor's `problems` computed is a
  mirror and must not diverge. Note nodes are skipped by validation, root detection and the runner.
- Pro-gated: interfaces in `pro_interfaces/workflow_{ctl,svc}.go`, stubs in `pro/`, engine in
  `pro_impl/services/server/workflow_svc.go` (gitignored, local only); `TaskRunner` calls
  `HandleWorkflowTaskCompletion` on every terminal status, so runs advance without the poll.

## First cut — decisions D1..D5 and the deviations

- **D1 positions as per-node columns**, not a template-level layout JSON: ids are reassigned on
  every save, so a blob keyed by node id would need the same remap as edges. Deviation: `int`,
  not `float64` (Postgres has no bare `double`; the migration transformer does not rewrite it),
  folded into `v2.18.15.sql` instead of a separate migration.
- **D2 full-page route** (`/workflows/new`, `/workflows/:id/edit`), not the 700 px dialog;
  `WorkflowForm.vue` retired.
- **D3 one shared renderer** for editor and run view; the run view previously drew no edges.
- **D4 Drawflow** over custom SVG (fallback), Rete.js 2 (heavy), jsPlumb (archived binding),
  Cytoscape (canvas-rendered, rich nodes awkward), GoJS (commercial); `@vue-flow/core` is Vue 3
  only and the app is Vue 2.7 + Vuetify 2.
- **D5 condition is an edge property** picked after drawing (default `on_success`), not typed
  output ports per condition — easier to relabel, allows two edges between the same pair.
- Deviation: `UpdateWorkflow` stayed `204`; the editor re-fetches via GET after save to rebind
  ids, keeping the swagger/Dredd contract untouched.
- Also first cut: auto-layout only when all positions are zero; Stop endpoint; run versioning.

## Redesign — what changed and why

Why: nodes were HTML strings with `escapeHtml`, no fit view ("Reset zoom" only set zoom 1), no
undo, no edge labels, hard-coded colours, broken dark theme; users compared it to n8n/Kestra.
`RESEARCH@5bf4fc406f`/`RESEARCH@e6d5aa256f`: the standard is xyflow + auto-layout, Vue 3 only.
Denis rejected a custom canvas, a library swap and a second Vue 3 runtime (2026-09-26).

- **R1 Node = Vue component.** `addNode()` gets an empty mount point; `WorkflowNodeCard` is
  mounted with `parent: this` (Vuetify/i18n inherited) and a reactive `store` shared by all
  cards. No `v-html`/`innerHTML` anywhere (criterion 7 of the brief; hostile template name and
  `<script>` note in `seed-wf.sh` render as text).
- **R2 Edge overlay** inside Drawflow's `precanvas` (pans/zooms with the graph, re-attached after
  `clear()`): `refreshEdges()` colours each `svg.connection`, sets `marker-end`, places the
  `WorkflowEdgeLabel` pill at `getPointAtLength(len/2)`; run view animates active edges.
- **R3 Own viewport.** Capture-phase wheel: plain wheel pans, Ctrl/Cmd or pinch zooms at the
  cursor; `transform-origin: 0 0`; fit view from the DOM bounding box, kept by a
  `ResizeObserver` for 1.5 s after build unless the user touched the canvas; snap 20 px; Tidy up
  = `layoutWorkflowNodes` (320 × 120 grid, no dagre). Wheel facts: `canvas-wheel-and-back-swipe.md`.
- **R4 Quick add** from the "+" on the output port, a floating "+", or right-click on empty
  canvas (`WorkflowQuickAddMenu`, kinds + template search); a node added from a port is placed to
  the right and wired `on_success` after the cycle guard. The palette stays (quick add
  complements, does not replace it).
- **R5 Explicit Save**, no autosave: dirty dot, `beforeRouteLeave` + `beforeunload` guard, Save
  disabled while `problems` is non-empty. Undo/redo 50 steps with key-based coalescing of
  property edits (`workflowHistory.js`).
- **R6 Theme** from `_palette.scss` → `--wf-*` CSS variables on the graph root, dark override
  class; the canvas sets `background-repeat: repeat` because Vuetify's global reset hid the grid.
- **R7 Keyboard** (`onKeyDown` in `WorkflowGraph.vue`, editor `onWindowKeyDown`): arrows pan
  (Shift = big steps), `+`/`-` zoom, `0` fit, `1` = 100 %, Esc deselects/closes; Ctrl/Cmd+Z,
  Shift+Z, Y in the editor, ignored while typing; cards have `tabindex=0`, Enter/Space activate.
  The canvas is focused on pointerdown because Drawflow's fixed mode `preventDefault()`s
  mousedown. Cheat-sheet SVG generated by `mocks/workflow-editor/make-hotkeys-svg.py`.
- **R8 Collapsible palette** with state in `localStorage`; `workflowEditorPrefs.js` lets
  `App.vue` render the nav drawer mini on editor routes before mount (no shrink animation).
- **R9 Pan from the empty container**: Drawflow reads `classList[0]`, so `parent-drawflow` is
  forced to be the first class (commit `0d601c12`).
- Guardrails from the brief are in `direction.md`; none of the "ask first" items (new dep,
  Drawflow patch, viewport column) has been raised. Permissions unchanged
  (`manageProjectResources` edit, `runProjectTasks` approve/stop). Accepted Drawflow limits:
  no multi-select/marquee, no minimap, no note resize.

## Open, deferred, divergent

- ⚠️ 2026-09-27: the approved mock says "Tab also opens this menu" and the empty state says
  "press Tab"; shipped Tab moves focus between cards and the empty-state hint says right-click
  (`workflowEmptyHint`). The mock is stale on this point, the code is the decision.
- Properties panel as a bottom sheet on narrow screens — not done (no `v-bottom-sheet` in
  `WorkflowEditor.vue`).
- Run-view per-node details panel (task link, status, per-task artifacts) — removed with the
  table in the first cut and never rebuilt; only the task-log dialog opens on click.
- Template-level viewport/bend-point storage, dagre-quality layout, copy/paste of subgraphs,
  multi-select, minimap — deferred; the last three need a canvas library change.
- Stand workflow 1 shows the unsaved-changes dialog after merely selecting a node
  (`AGENTS/memory/ui-stands.md`) — unexplained.
- Template names are not unique per project — matters for Workflows-as-Code (`direction.md`).

## Key files and references

- Canvas: `web/src/components/WorkflowGraph.vue`, `web/src/components/workflow/*`,
  `web/src/lib/workflowViewport.js`, `workflowHistory.js`, `workflowLayout.js`,
  `workflowGraph.js`, `workflowEditorPrefs.js`; pages `web/src/views/project/WorkflowEditor.vue`,
  `WorkflowRun.vue`, `Workflows.vue`; global CSS hooks in `web/src/App.vue`
  (`html.WorkflowEditor-html`).
- Backend: `db/Workflow.go`, `db/sql/workflow.go`, `db/sql/migrations/v2.18.15.sql`,
  `pro/db/Workflow.go`, `pro_interfaces/workflow_*.go`, `api/router.go` (workflow routes).
- Tests: `web/tests/unit/workflow*.spec.js`; Playwright checklists and screenshots in `mocks/`
  (`mocks/README.md`). Docs: `docs/docs/user-guide/workflows.md`, `docs/static/assets/workflow-*`.
- References: `TASK@88179c9bc1`, `RESEARCH@5bf4fc406f`, `RESEARCH@e6d5aa256f`;
  https://github.com/semaphoreui/semaphore/pull/3488 (Copilot original, open),
  https://github.com/semaphoreui/semaphore/pull/4145 (delay node),
  https://github.com/semaphoreui/semaphore/pull/4199 (workflows guide).
