# Graphical workflow editor

Status: implemented (initial cut). Built with Drawflow (decision D4).

## What shipped

- **Backend positions.** `WorkflowNode.PositionX/PositionY` (`int`); the
  `position_x` / `position_y` columns are part of the (unreleased) workflow
  tables in `db/sql/migrations/v2.18.15.sql`, and `writeWorkflowGraph` persists
  them. Positions ride along in the node `INSERT`, so they survive the
  delete-and-reinsert.
- **Shared renderer** `web/src/components/WorkflowGraph.vue` — wraps Drawflow,
  treats the canvas as the source of truth and emits `{nodes, edges}` on every
  change. Used editable in the editor and read-only (status overlay) in the run
  view.
- **Full-page editor** `web/src/views/project/WorkflowEditor.vue` at routes
  `/workflows/new` and `/workflows/:workflowId/edit`; palette drag-and-drop,
  node property panel, edge condition selector, live self-edge/cycle guards,
  a problems panel mirroring `ValidateWorkflowTemplate`, and topological
  auto-layout for legacy (position-less) workflows.
- **Node kinds.** `task` (runs a template), `approval` (gates the run), and
  `note` — a pure annotation that holds free-form text (`WorkflowNode.Note`,
  column `note`). Note nodes do not execute and are excluded from the run graph:
  they have no canvas ports, can not be connected by edges, never count as a
  root, and are skipped by `ValidateWorkflowTemplate`, `WorkflowRootNode`, and
  the `WorkflowService` runner.
- **Full-screen run view** `web/src/views/project/WorkflowRun.vue` — the node/edge
  table and the artifact cards were removed; the page is now a full-viewport
  (`100vh`) graph. The currently-active node is highlighted with a Concourse-style
  animation (pulsing glow + moving diagonal stripes for running/waiting; amber
  pulse for pending approvals). Approve/Reject buttons render on the canvas itself
  (overlay bar pinned to the bottom of the schema). Zoom in/out/reset controls
  live in the toolbar (primary zoom UX); Drawflow's `fixed` mode also gives drag
  to pan and Ctrl+wheel to zoom. Status is repainted in place on each 5s poll so
  pan/zoom is preserved.
- **Shared layout helper** `web/src/lib/workflowLayout.js` — the layered
  topological layout is used by both the editor (seeds + persists positions on
  save) and `WorkflowGraph` (lays out any consumer whose stored positions are all
  zero, so the run view of a legacy workflow no longer piles nodes at the origin).
- `Workflows.vue` now routes to the editor (dialog path removed); i18n keys added;
  `WorkflowForm.vue` retired.

## Pro feature gating

Workflows are a Pro feature, gated with the same mechanism as
`terraform_backend` / `project_runners`: a controller interface in
`pro_interfaces`, a real implementation in `pro_impl`, a no-op stub in `pro`
(swapped at build time via `replace github.com/semaphoreui/semaphore/pro =>
./pro` vs `./pro_impl`), plus a feature flag.

- **`pro_interfaces`** — `WorkflowController` (the 11 HTTP handlers) and
  `WorkflowTaskPool` (the narrow subset of `*services/tasks.TaskPool` the
  controller needs — declared here so the pro modules depend only on
  `pro_interfaces` + `db`, avoiding a `services/tasks` import / cycle). Added
  `Workflows bool` to `Features`.
- **`pro_impl/api/projects/workflows.go`** — the real controller (the request
  logic moved out of `api/projects/workflows.go`); it delegates orchestration to
  the `WorkflowService`. `pro_impl/pkg/features` sets `Workflows:
  planDetails.IsPro()`.
- **`pro_impl/services/server/workflow_svc.go`** — the real `WorkflowService`:
  the orchestration engine extracted from `TaskPool` (start/progress runs,
  approvals, artifact merge). It depends only on `db.Store` and a
  `WorkflowTaskEnqueuer` (the pool's `AddTask`), holds its own mutex, and is a
  self-contained entity rather than methods on the open task pool.
- **`pro/api/projects/workflows.go`** and **`pro/services/server/workflow_svc.go`**
  — the open-source stubs: list endpoints return `[]`, the rest `404`; the
  service methods are safe no-ops. `pro/pkg/features` already returns an empty
  `Features{}` (so `Workflows` is `false` in open builds).
- **`api/router.go`** constructs `workflowController :=
  proProjects.NewWorkflowController(workflowService)` and registers its methods.
  `api/projects/workflows.go` keeps only the two context-loader middlewares
  (`WorkflowsMiddleware`, `WorkflowRunsMiddleware`), which use the open
  `db.Store`.
- **Wiring (`cli/cmd/root.go`)** — resolves the pool↔service cycle: create the
  pool, then `workflowService := proServer.NewWorkflowService(store, &taskPool)`,
  then `taskPool.SetWorkflowService(workflowService)`; the service is also passed
  to `api.Route(...)` for the controller.
- **Frontend** — `App.vue` hides the Workflows nav item unless
  `features.workflows` is set.

**What stays in the open module, and why.** Only thin glue remains open:
`TaskPool` keeps two delegators (`HandleWorkflowTaskCompletion`,
`GetWorkflowRunArtifacts`) that forward to the injected service, because the open
task-execution lifecycle (`TaskRunner`) calls them on every finished task. They
are safe no-ops when the stub service is wired (open builds). The workflow
`db.Store` methods (`WorkflowManager`) also stay open — they are plain CRUD over
the open schema, consumed by both the service and the context-loader middlewares.
The orchestration engine itself now lives entirely in `pro_impl`
(`workflow_runner.go` was removed from `services/tasks`).

**Migration note (fixed in passing).** The `position_x`/`position_y` columns had
been folded into the unreleased `v2.18.15.sql`, but `position_y` was written with
a Cyrillic `у`, so the Latin `position_y` column the Go code expects was never
created. Corrected to ASCII; the redundant standalone `v2.18.16` migration and
its `Migration.go` entry were removed.

## Deviations from the design below

- **Positions are `int`, not `float64`/`double`.** Postgres has no bare `double`
  type and the migration dialect transformer (`db/sql/migration.go`) doesn't
  rewrite it; `int` pixel coordinates are dialect-safe and match the schema.
- **`UpdateWorkflow` API contract kept at `204`** (backend change 5 not applied).
  Instead the editor re-fetches via the existing `GET` after every save to
  rebind the reassigned node ids — a self-contained client concern that avoids
  changing the documented API / swagger / Dredd contract.
- **Edge conditions** render as color-coded connection paths + a legend + a
  click-to-edit selector, rather than mid-path text labels (labels-on-path
  remain a follow-up).
- **Pre-existing fix:** `db/workflow_test.go` imported the removed `db/bolt`
  package (dead since the bolt removal), which prevented the `db` test package
  from building. Switched it to `sql.CreateTestStore()` and made the fixtures
  FK-compatible, restoring 5 workflow-validation tests + adding a positions
  regression test.

---

Status (original design): planned.

## Why

Workflows are stored as a DAG of `WorkflowNode`s connected by `WorkflowEdge`s
(see `db/Workflow.go`), but the only editor today — `web/src/components/WorkflowForm.vue` —
renders the graph as two flat lists ("Nodes" and "Edges") inside a 700px-wide
modal dialog. Users build edges by picking *From*/*To* node IDs out of
dropdowns. For anything beyond a trivial linear pipeline this is hard to read,
error-prone (you reason about `#3 → #5` instead of seeing the arrow), and gives
no sense of the graph shape, branches, or convergence points.

We want a canvas-based editor: drag nodes from a palette, position them freely,
draw connections between them with the mouse, label each connection with its
condition (`on_success` / `on_failure` / `always`), and see the DAG the way it
actually runs. The same renderer should also visualize a **run** (currently a
flat `v-data-table` in `web/src/views/project/WorkflowRun.vue` that never draws
the edges at all).

## Current architecture (verified)

### Data model — `db/Workflow.go`

- `WorkflowTemplate { ID, ProjectID, Name, Description, Nodes []WorkflowNode, Edges []WorkflowEdge }`.
  `Nodes`/`Edges` are tagged `db:"-"` — they are **not** columns on the
  template; they live in child tables.
- `WorkflowNode { ID, WorkflowTemplateID, TemplateID, Kind, ConvergenceMode, ApprovalTimeout, ApprovalMessage, InventoryID, EnvironmentID, Limit }`.
  **There are no x/y position fields.** This is the core gap.
- `WorkflowEdge { ID, WorkflowTemplateID, SourceNodeID, DestinationNodeID, Condition }`.
- `ValidateWorkflowTemplate(d, workflow)` enforces (and runs before every write):
  non-empty name; ≥1 node; unique node ids; valid kind/convergence/condition
  enums; task nodes require `template_id` and forbid approval fields; approval
  nodes forbid task fields and require `approval_timeout > 0` when set; template
  must exist and belong to the project; no zero/self/dangling edges; **exactly
  one root** (zero-in-degree node); and the graph **must be a DAG** (three-color
  DFS cycle check).

### Persistence — `db/sql/workflow.go` (SQL is the only backend)

- **BoltDB has been removed.** `db/factory/store.go` panics
  `"Bolt is not supported starting from version 2.19"` for `DbDriverBolt`, and
  `db/bolt/workflow.go` no longer compiles (`undefined: BoltDb`) — it is orphaned
  dead code. No Bolt work is needed; ideally delete `db/bolt/workflow.go` +
  `db/bolt/workflow_test.go` as cleanup.
- Tables created brand-new on this branch in
  `db/sql/migrations/v2.18.15.sql`: `project__workflow_template`,
  `project__workflow_node`, `project__workflow_edge`, `project__workflow_run`,
  `project__workflow_approval`. **The workflow feature is unreleased.**
- **Write path is full delete-and-reinsert.** `CreateWorkflowTemplate` and
  `UpdateWorkflowTemplate` both call `writeWorkflowGraph`, which
  `delete`s all edges then all nodes for the template, re-inserts every node
  (new autoincrement id), builds `nodeIDMap[clientID] → newDbID`, then re-inserts
  edges remapped through that map.
  **Consequence: node DB IDs change on every save.** Client-supplied ids
  (positive ints from `nextNodeId()`, or the backend's negative `-(idx+1)`
  placeholders for `id==0`) are transient — used only to remap edges within one
  write, then discarded.

### Frontend — Vue 2.6 + Vuetify 2.6 (vue-cli 5 / webpack)

- `WorkflowForm.vue` is mounted **only** inside an `EditDialog` modal from
  `web/src/views/project/Workflows.vue` (via the `ItemListPageBase` mixin's
  `editItem()` → `editDialog = true`). Save is driven by the dialog footer
  flipping `needSave`, watched by the `ItemFormBase` mixin.
- `ItemFormBase` (`web/src/components/ItemFormBase.js`) load/save lifecycle:
  `getNewItem()` → `{name, description, nodes:[], edges:[]}`; `afterLoadData()`
  normalizes nodes; `save()` **requires `this.$refs.form.validate()` to be
  truthy**, then `POST` (`getItemsUrl()`) / `PUT` (`getSingleItemUrl()`) with body
  `{...this.item, project_id}`. `beforeSave()` is the hook to serialize canvas
  state into `nodes[]`/`edges[]`.
- Routes (`web/src/router/index.js`): list `/project/:projectId/workflows`
  (`Workflows.vue`), run `/project/:projectId/workflows/:workflowId/runs/:runId`
  (`WorkflowRun.vue`). **No edit route exists** — editing is dialog-only.
- **No graph/diagram library is installed** (deps: vuetify, vue, vue-router,
  chart.js, vue-codemirror, vuedraggable, vue-virtual-scroll-list, cron-parser,
  dayjs). The `d3`/`dagre` hits in the tree are transitive swagger-ui assets only.
- i18n: all current workflow strings already exist in `web/src/lang/en.js`
  (lines ~497–540): `workflowNodes`, `workflowEdges`, `workflowNodeKind*`,
  `workflowConvergence*`, `workflowCondition*`, `workflowApproval*`, etc.

## Design decisions

### D1 — Persist node positions as per-node columns

Add `position_x` / `position_y` to `project__workflow_node` and
`PositionX float64` / `PositionY float64` to `db.WorkflowNode`.

**Why per-node columns, not a JSON layout blob on the template:** node DB ids are
reassigned on every save (see above), so a template-level `layout` blob keyed by
node id would be invalidated on every `Update` unless it were rewritten through
the same `nodeIDMap` remap that edges use. Per-node columns sidestep this
entirely — the position rides along in the **same `INSERT` as the node**, so it
round-trips for free. Positions are also intrinsically 1:1 with a node, and
`ON DELETE CASCADE` drops them automatically when a node is deleted.

(A small template-level `layout` JSON column may still be added later for
*canvas-level* state — zoom/pan viewport — that is not tied to any node. Not
required for v1.)

### D2 — Editor is a full-page route, not the dialog

A canvas needs the whole viewport; a 700px dialog cannot host pan/zoom + palette
+ a node property panel. Mount a new full-page `WorkflowEditor.vue` and add
routes `/project/:projectId/workflows/new` and
`/project/:projectId/workflows/:workflowId/edit`, mirroring the existing
full-page `WorkflowRun.vue`. `Workflows.vue` switches its "New"/pencil/name
actions from `editItem()` (open dialog) to `$router.push(...)`. The
`WorkflowForm.vue` dialog can be retired once the page reaches parity (keep its
node/edge helper logic, migrate it into the editor / a mixin).

### D3 — One shared renderer for editor and run views

Build `web/src/components/WorkflowGraph.vue` (props: `nodes`, `edges`,
`editable`, optional `nodeStatuses`). In `editable` mode it emits
add/move/connect/delete events; in read-only mode it overlays per-node run status
colors (reuse `WorkflowRun.vue`'s `statusColor()`) and approval buttons. Embed it
both in `WorkflowEditor.vue` and in `WorkflowRun.vue` (which today draws no edges
at all — this closes that gap for free).

### D4 — Graph library: **Drawflow** (recommended), custom SVG as fallback

This is the main thing to confirm before implementation. Evaluation:

| Option | Vue 2 | License | Size (gz) | Verdict |
|---|---|---|---|---|
| **Drawflow** | native (`new Drawflow(el, Vue, this)`) | MIT | ~7–8 KB | **Recommended** |
| Custom SVG/HTML | n/a (own code) | — | ~0 | Fallback (full control, highest build cost) |
| Rete.js v2 (`rete-vue-plugin/vue2`) | yes | MIT | ~100 KB+ | Heavier alternative if we outgrow Drawflow |
| jsPlumb community | wrapper, weak | MIT/GPL2 dual | large | Archived; no maintained free Vue 2 binding — avoid |
| Cytoscape.js | wrapper | MIT | ~112 KB | Analysis engine, canvas-rendered; rich node bodies awkward — avoid |
| GoJS | wrapper | **commercial** | large | Disqualified (proprietary, ~$4k/seat) |

`@vue-flow/core` (the obvious modern choice) is **Vue 3 only** and unusable here.

**Drawflow** is MIT, purpose-built for drag-and-drop flow editors, tiny, and
renders real Vue 2 components as node bodies (`editor.registerNode('taskNode', TaskNode)`),
so task vs approval nodes can be Vuetify cards. It lacks two things we must add in
the wrapper: DAG enforcement (cycle check on connect) and edge labels — both
modest, well-bounded work. If the team prefers zero new runtime deps, the
**custom SVG** path (absolutely-positioned Vuetify node cards + SVG `<path>`
edges, drag via mousedown, reusing the already-present `vuedraggable`) is the
documented fallback and keeps the renderer trivially shareable with `WorkflowRun`.

### D5 — Edge condition as a clickable edge property

Model the condition the way the data model already does — a property on the
`WorkflowEdge`. After the user draws a connection it defaults to `on_success`;
clicking the edge opens a small menu to switch to `on_failure` / `always`, and
the edge renders a colored label (green / red / grey, matching
`workflowConditionOnSuccess/OnFailure/Always`). This maps 1:1 to `Condition` and
allows two differently-conditioned edges between the same pair if ever needed.
(Alternative: typed output ports per condition on the source node — rejected as
less flexible and harder to relabel.)

## Backend changes

1. **`db/Workflow.go`** — add to `WorkflowNode`:
   ```go
   PositionX float64 `db:"position_x" json:"position_x" backup:"position_x"`
   PositionY float64 `db:"position_y" json:"position_y" backup:"position_y"`
   ```
   The `backup:` tags keep layout in project export/import.
   `ValidateWorkflowTemplate` needs **no change** — positions don't participate
   in validation.

2. **`db/sql/migrations/v2.18.16.sql`** (new):
   ```sql
   alter table `project__workflow_node` add column `position_x` double not null default 0;
   alter table `project__workflow_node` add column `position_y` double not null default 0;
   ```
   **`db/sql/migrations/v2.18.16.err.sql`** (new, rollback):
   ```sql
   alter table `project__workflow_node` drop column `position_y`;
   alter table `project__workflow_node` drop column `position_x`;
   ```
   Plain backtick/SQLite syntax transforms cleanly to MySQL/Postgres via
   `db/sql/migration.go` `prepareMigration()` — no `{{if .Sqlite}}` dialect
   branches needed for a numeric `add column`.

   _Alternative:_ since the workflow tables are brand-new and unreleased on this
   branch, the two columns could be folded directly into the `create table
   project__workflow_node` in `v2.18.15.sql`. A separate additive migration is
   cleaner/safer and is the recommendation.

3. **`db/Migration.go`** — append `{Version: "2.18.16"},` to the `commonScripts`
   slice, right after the `{Version: "2.18.15"}` entry (currently line ~130).

4. **`db/sql/workflow.go` `writeWorkflowGraph`** — add the two columns to the
   node `INSERT` (column list + args: `node.PositionX, node.PositionY`). The read
   path uses `select *`, so `getWorkflowNodes` needs no change.

5. **`api/projects/workflows.go` `UpdateWorkflow` — return the saved template.**
   Today it responds `204 No Content`. Because node ids are reassigned on every
   save, after an update the canvas holds **stale ids**; continuing to edit and
   re-save could remap edges against ids the server no longer has. Change
   `UpdateWorkflow` to re-read and return the template as `200` + body (as
   `AddWorkflow` already does on create), and have the editor rebind node
   ids+positions from the response. This is the single biggest integration risk;
   addressing it server-side keeps the client simple. (`UpdateWorkflowTemplate`
   in `db/sql/workflow.go` currently returns only `error` — either re-`GetWorkflowTemplate`
   in the handler, or change the store method to return the template.)

## Frontend changes

1. **`web/src/components/WorkflowGraph.vue`** (new) — shared canvas renderer.
   Props `nodes`, `edges`, `editable`, `nodeStatuses?`. Wraps Drawflow (or the
   custom SVG renderer). Editor mode emits `node-added`, `node-moved`
   (`{id, x, y}`), `connection-created` (`{source, target}`),
   `connection-removed`, `node-removed`, `selection-changed`. Read-only mode
   colors nodes by `nodeStatuses` and renders edge condition labels.

2. **`web/src/views/project/WorkflowEditor.vue`** (new) — full-page editor.
   - Mixes in `ItemFormBase` to reuse `getNewItem` / `afterLoadData` /
     `getItemsUrl` / `getSingleItemUrl` / `save`. Override URLs to
     `/api/project/{projectId}/workflows[/{itemId}]`.
   - Layout: left palette (Task node / Approval node), center `WorkflowGraph`
     (`editable`), right node-properties panel (template picker, convergence,
     `ArgsPicker` for `limit`, approval timeout/message — reuse
     `WorkflowForm.vue:68–135` field logic + `onNodeKindChanged` gating), top
     toolbar (name/description, zoom/fit/auto-arrange, Save).
   - Keep a `<v-form ref="form">` wrapper (even if it only validates `name`) so
     `ItemFormBase.save()`'s `this.$refs.form.validate()` guard passes; the
     toolbar Save sets `needSave`.
   - `beforeSave()` serializes canvas state into `item.nodes[]` (with
     `position_x`/`position_y`) and `item.edges[]`.
   - On save success, rebind ids from the returned template (see backend change 5),
     or `$router.replace` from `/new` to `/{id}/edit`.
   - `afterLoadData()`: seed positions for nodes that have none (legacy/all-zero)
     via a simple layered auto-layout (topological columns by edge depth) so
     existing workflows open as a sensible graph rather than a pile at (0,0).

3. **`web/src/router/index.js`** — import `WorkflowEditor` (near lines 34–35) and
   add:
   ```js
   { path: '/project/:projectId/workflows/new', component: WorkflowEditor },
   { path: '/project/:projectId/workflows/:workflowId/edit', component: WorkflowEditor },
   ```
   (`/edit` suffix avoids colliding with the existing `/runs/:runId` child.)

4. **`web/src/views/project/Workflows.vue`** — repoint the "New Workflow" button,
   the name link, and the pencil action from `editItem(...)` to
   `$router.push('/project/{projectId}/workflows/new' | '.../{id}/edit')`. Drop the
   `EditDialog` + `WorkflowForm` usage for editing (keep `YesNoDialog` delete and
   `runWorkflow`).

5. **`web/src/views/project/WorkflowRun.vue`** — embed
   `<WorkflowGraph :nodes="workflow.nodes" :edges="workflow.edges" :node-statuses="..." :editable="false" />`
   above (or replacing) the `v-data-table`, finally drawing the DAG with live
   status colors. Build `nodeStatuses` from `details.nodes` (`nodeStatusRaw`).

6. **`web/src/lang/en.js`** — add canvas-specific keys (none exist yet):
   - Palette/toolbar: `workflowEditorPalette`, `workflowPaletteTaskNode`,
     `workflowPaletteApprovalNode`, `workflowDragToCanvasHint`,
     `workflowToolbarZoomIn`, `workflowToolbarZoomOut`, `workflowToolbarFit`,
     `workflowToolbarAutoLayout`, `workflowToolbarDelete`.
   - Inline UI: `workflowRootBadge`, `workflowNodeIncomplete`,
     `workflowConnectHint`, `workflowEdgeDelete`.
   - Live-guard toasts: `workflowSelfEdgeBlocked`, `workflowCycleBlocked`,
     `workflowDuplicateEdgeBlocked`, `workflowSecondRootHint`.
   - Problems panel (friendly translations of backend errors):
     `workflowProblemsPanelTitle`, `workflowErrorNoNodes`, `workflowErrorNoRoot`,
     `workflowErrorMultipleRoots`, `workflowErrorHasCycle`,
     `workflowErrorTaskNeedsTemplate`, `workflowErrorApprovalTimeoutPositive`,
     `workflowValidationPassed`.
   Reuse existing enum/field labels (`workflowConditionOn*`, `workflowNodeKind*`,
     `workflowConvergence*`, `workflowApprovalTimeout/Message`, `workflowNodeLimit`).

7. **`web/package.json`** — add `"drawflow"` (if D4 = Drawflow). None if custom SVG.

## Client-side validation (mirror of `ValidateWorkflowTemplate`)

**Enforce live** (block the action on the canvas):
- **No self-edge** — reject a connection dropped back on the same node.
- **No cycle** — before committing a new edge, run the same three-color DFS on the
  prospective graph; reject the drop if it would create a back-edge. (Most
  valuable guard — cycles are hard to spot visually.)
- **Node-kind field gating** — show template/limit/inventory/environment only for
  task nodes, approval timeout/message only for approval nodes (already done in
  `WorkflowForm.vue` + `onNodeKindChanged`); makes the "node forbids X fields"
  rules unviolatable.
- **Cascade-delete edges** when a node is deleted (`WorkflowForm.vue:299–306`
  already does this for the list editor — keep it).
- Endpoint integrity is structural: the canvas can only connect existing handles.

**Validate on save** (global/incomplete states; surface in a "Problems" panel +
node badges, block Save, and always show the backend's response in the existing
`formError` banner as source of truth):
- name non-empty; ≥1 node.
- **exactly one root** — fluctuates while building (each new unconnected node is
  temporarily a root), so don't block live; show a persistent "roots: N (must be
  1)" indicator + a root badge on each zero-in-degree node.
- task node requires `template_id` (warning badge on incomplete nodes).
- approval timeout > 0; template exists / belongs to project (backstop — picker
  only lists this project's templates).

## Phased implementation

- **Phase 0** — Confirm D4 (Drawflow vs custom SVG).
- **Phase 1 — Backend positions.** `WorkflowNode` fields, migration
  `v2.18.16.{sql,err.sql}`, register in `Migration.go`, `writeWorkflowGraph`
  INSERT, and `UpdateWorkflow` returning the saved template. Ship independently of
  the UI (positions just default to 0).
- **Phase 2 — Read-only renderer.** `WorkflowGraph.vue` in read-only mode; wire
  into `WorkflowRun.vue` to draw the DAG with status colors. Lowest-risk first
  user-visible win.
- **Phase 3 — Editor page.** `WorkflowEditor.vue` + routes + palette + property
  panel + edge drawing/labels + live guards (self-edge, cycle) + auto-layout +
  save/rebind. Repoint `Workflows.vue` actions.
- **Phase 4 — Polish & retire.** Problems panel, i18n, zoom/fit/auto-arrange,
  remove `WorkflowForm.vue` + `EditDialog` editing path, delete dead
  `db/bolt/workflow*.go`.

## Testing

- **Go** (`go test ./db/sql/ ...`): extend workflow store tests to assert
  `position_x`/`position_y` round-trip through `Create`/`Update` (and survive the
  delete-and-reinsert), and that `UpdateWorkflow` returns the re-read template
  with new node ids + preserved positions. Confirm migration applies on
  sqlite/mysql/postgres (existing migration test harness).
- **Validation** unchanged — add a regression asserting positions don't affect
  `ValidateWorkflowTemplate`.
- **Frontend** (vue-cli mocha, `npm run test:unit` in `web/`): component test for
  `WorkflowGraph.vue` cycle/self-edge rejection and the client-side
  single-root/incomplete-node detection.

## Risks & open questions

- **Node-id reassignment on save (highest risk).** Addressed by backend change 5
  (return saved template) + client rebind. Without it, an edit→edit→save sequence
  can remap edges against stale ids. Also note: editing a template reassigns node
  ids, so historical runs' `task.workflow_node_id` become dangling — a pre-existing
  property of the delete-and-reinsert design, not introduced here, but worth
  flagging (the run view already tolerates missing nodes).
- **Library lock-in.** Drawflow is single-maintainer / feature-stable (last
  release 0.0.60, Sep 2024). API surface is small; the wrapper isolates it and we
  can fork or swap to the custom SVG renderer behind the same `WorkflowGraph.vue`
  props if needed.
- **Edge labels in Drawflow** aren't native — implemented as an overlay layer
  positioned at connection-path midpoints (recompute on `nodeMoved`).

## Out of scope (follow-ups)

- Template-level canvas state (zoom/pan viewport, edge bend points) via a
  `layout` JSON column on `project__workflow_template`.
- Auto-arrange beyond a simple layered layout (e.g. dagre).
- Undo/redo history, multi-select drag, copy/paste of subgraphs.
- Minimap for large workflows.
- **Run-view per-node details panel.** Removing the run-view table also removed
  where per-task status detail, the task link, and per-task / merged artifacts
  (`set_stats`) were shown. A click-a-node side panel on the run view should
  bring these back (task link, status, artifacts) without reintroducing the
  table.
