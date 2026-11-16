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
- Revisions (2026-09-27, `TASK@1b65b07db2`, migration `v2.20.8`): nodes and edges belong to a
  `project__workflow_revision` row and are never rewritten. Every PUT appends a revision in one
  transaction (`createWorkflowRevision` in `pro_impl/db/sql/workflow.go`; node ids are still
  fresh, edges remapped through `nodeIDMap`), then drops earlier revisions no run pins
  (`deleteUnreferencedWorkflowRevisions`, fenced by the FK on `run.revision_id`). A run pins
  `revision_id` at start; the engine and `GET runs/{id}` read the graph through
  `GetWorkflowRevisionGraph`, so an edit neither changes a running run nor orphans past runs'
  `task.workflow_node_id`. Runs created before the migration already lost their nodes to the old
  delete-and-reinsert and stay without node statuses. Current revision = highest `number`, no
  pointer column; `GET …/revisions` lists survivors with `has_runs`.
- `pro/db.ValidateWorkflowTemplate` is the source of truth; the editor's `problems` computed is a
  mirror and must not diverge. Note nodes are skipped by validation, root detection and the runner.
- Pro-gated: interfaces in `pro_interfaces/workflow_{ctl,svc}.go`, stubs in `pro/`, engine in
  `pro_impl/services/server/workflow_svc.go` (gitignored, local only); `TaskRunner` calls
  `HandleWorkflowTaskCompletion` on every terminal status, so runs advance without the poll.
- Engine facts verified on the stand (2026-09-27, `AGENTS/tools/wf-stand/`): a run is a loop of
  "launch every ready node, re-read, repeat" under a per-run lock; approval timeouts and delay
  expiries are resolved lazily by any progression pass (GET run, task completion, reconciler
  every 60 s, delay sweep every 5 s) — nothing sleeps in memory. `computeWorkflowRunStatus` marks
  the run `failed` when any task is not `success`, even if its `on_failure` branch ran and
  succeeded; a node whose inbound conditions cannot match is simply never launched ("Not started").
  Stop = conditional fence to `stopped` first, then force-kill of the run's tasks.

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

## Run panel (`?w=<workflow>/<run>`, TASK@244A5E2CA1, 2026-11-17)

A link to a run no longer leaves the page: it adds `w=<workflow id>/<run id>` (`?w=4/10`) to the
query of the current page, and `App.vue` shows the run in `WorkflowRunPanel.vue`, a panel fixed to
the right edge. The run view itself is `WorkflowRunView.vue`, shared by the panel and by the run
page.

- **`w` names the workflow and the run**, because every run route of the API is under its
  workflow (`/workflows/{workflow_id}/runs/{run_id}`). ⚠️ The first cut carried the run id alone
  and added a route `GET /project/{p}/workflow_runs/{run}` to find the workflow; the owner called
  the new endpoint a mistake and it was reverted the same day. The panel adds no API.
- **`w` mirrors `t`** (the task dialog): `App.vue` opens the panel from the query at load and on
  every route change, closing drops `w` only. A node click adds `t` next to `w`, so both are open;
  the dialog is a Vuetify overlay (z-index 2000+), the panel sits at 1900, hence under it.
- **Query edits are serialized** (`App.vue editQuery`): closing the dialog and the panel in a row
  (Escape twice) started two `router.replace` from the same query, and the second brought `t`
  back, reopening the dialog.
- **Links push, closing replaces**: a link is a plain `router-link` to
  `workflowRunLocation($route, workflowId, runId)` (`lib/workflowRunQuery.js`), so Back closes the
  panel and "open in a new tab" gives the page with the panel open. vue-router leaves the `/` of
  the value unencoded.
- **No scrim.** The page under the panel stays in use: a click on another run in the list swaps
  the run of the panel (the view is keyed by project, workflow and run — graph, polling and
  state start anew). Escape closes the panel unless a dialog or a menu is open above it.
- **Another project closes the panel**: `App.vue selectProject` (project menu, new and restored
  project) drops `w` from the query — a run belongs to the project it was opened in.
- **Width**: half of the window, the left edge is dragged (400 px … window width), stored in
  `localStorage` `WorkflowRunPanel__width`, a double click on the edge restores the half; full
  width and no handle on `smAndDown`.
- **The run page stays** at `/project/{p}/workflows/{w}/runs/{r}`: it is the value of
  `SEMAPHORE_WORKFLOW_URL` (`db/Task.go GetWorkflowUrl`) and of old bookmarks. No link of the
  interface leads to it any more.
- **Lists under the panel** (`Workflows.vue`, `WorkflowRuns.vue`) reload on the bus event
  `i-workflow-run`, emitted when a run is started, on the view's first successful load (the run
  may have finished since the list was read, and a finished run is never polled) and when a poll
  sees its status change — before, coming back from the run page reloaded them.

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
- Revisions UI is minimal: a "rev. N" chip and an "N active runs" chip in the editor toolbar,
  "rev. N" in the run title. No revision history page, no diff, no restore (a restore would be a
  new revision copied from an old one); YAML export of a revision is `TASK@9c1c9b0761`.
- Template names are not unique per project — matters for Workflows-as-Code (`direction.md`).

## Key files and references

- Canvas: `web/src/components/WorkflowGraph.vue`, `web/src/components/workflow/*`,
  `web/src/lib/workflowViewport.js`, `workflowHistory.js`, `workflowLayout.js`,
  `workflowGraph.js`, `workflowEditorPrefs.js`; pages `web/src/views/project/WorkflowEditor.vue`,
  `WorkflowRun.vue`, `Workflows.vue`; run view `web/src/components/WorkflowRunView.vue`, run panel
  `WorkflowRunPanel.vue`, `web/src/lib/workflowRunQuery.js`; global CSS hooks in `web/src/App.vue`
  (`html.WorkflowEditor-html`).
- Backend: `db/Workflow.go`, `db/WorkflowStore_pro.go`, `pro_impl/db/sql/workflow.go`,
  `db/sql/migrations/v2.18.15.sql`, `v2.20.8.sql` + `db/sql/migration_2_20_8.go` (revisions),
  `pro/db/Workflow.go`, `pro_interfaces/workflow_*.go`, `api/router.go` (workflow routes).
- Tests: `web/tests/unit/workflow*.spec.js`; Playwright checklists and screenshots in `mocks/`
  (`mocks/README.md`). Docs: `docs/docs/user-guide/workflows.md`, `docs/static/assets/workflow-*`.
- References: `TASK@88179c9bc1`, `RESEARCH@5bf4fc406f`, `RESEARCH@e6d5aa256f`;
  https://github.com/semaphoreui/semaphore/pull/3488 (Copilot original, open),
  https://github.com/semaphoreui/semaphore/pull/4145 (delay node),
  https://github.com/semaphoreui/semaphore/pull/4199 (workflows guide).
