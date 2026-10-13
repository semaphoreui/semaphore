# mocks — approved mock, screenshots and the scripts that shoot them

Design artefacts of the editor redesign (`TASK@88179c9bc1`); decisions are in
`../editor-architecture.md`, not here. Every script targets the local stand
`http://localhost:3100` (admin/admin123, project 1) and Playwright loaded through `NODE_PATH`
with `channel: 'chrome'` — stand and Playwright recipes are in `AGENTS/memory/ui-stands.md`.

## Files

| File | What it is |
| --- | --- |
| `2026-09-26-workflow-node-mock.html` | Approved static mock: editor canvas, run view, card states, empty state; `#dark` in the URL flips the theme. Open from the repo checkout — MDI icons load from `web/node_modules`. Its "Tab opens quick-add" hint is stale (see ⚠️ in `editor-architecture.md`). |
| `2026-09-26-workflow-node-mock-{light,dark}.png` | Renders of the mock as approved. |
| `workflow-editor/editor-{light,dark}.png`, `editor-selected-*.png`, `editor-quickadd-light.png`, `editor-legacy-*.png`, `editor-empty-light.png`, `run-{light,dark}.png` | Screenshots of the shipped UI at 1440 × 900 @2x, shot by `shoot-wf.cjs`. "Legacy" is a workflow saved without positions (auto layout). |
| `workflow-editor/editor-interactions.png` | End state of the `interact-wf.cjs` checklist. |
| `workflow-editor/run-approval-pending.png`, `run-delay-countdown.png` | Run view with a pending approval and the delay countdown, from `approve-wf.cjs`. |
| `workflow-editor/editor-edge-explicit-{light,dark}.png`, `editor-edge-byname-{light,dark}.png`, `run-outputs-{light,dark}.png`, `task-outputs-{light,dark}.png` | Workflow inputs (2026-10-09, `TASK@1845d9d4d7` stage 7): the edge panel with "Map inputs explicitly" and the mapping table (keys suggested from the last run), the by_name panel with the destination's survey variables (a check on those the source produced), the run view with "N outputs" on producer cards and the swap icon on explicit pills, the task dialog's Outputs panel. Shot by `shoot-inputs.cjs` on the "Artifacts QA" project. |
| `workflow-editor/editor-revision-active-run.png`, `run-after-edit.png` | Revisions (2026-09-27): the editor with the "rev. 3" and "1 active run" chips, and run #8 still rendered on its revision 1 after two later saves. Shot with `AGENTS/tools/wf-stand/shoot-run.cjs`. |

## Scripts

| Script | Does | Needs |
| --- | --- | --- |
| `seed-wf.sh` | Seeds five templates (one with an XSS name), the "Release" workflow with positions, the position-less "Nightly checks" (with a `<script>` note), starts one Release run; writes ids to `/tmp/semaphore-stand/wf-ids.txt`. Idempotent for templates. | Running stand, repository 1 and inventory 1 in project 1. |
| `shoot-wf.cjs` | Editor/run/legacy/empty screenshots in light and dark; output `OUT` (default `/tmp/semaphore-stand/shots-wf`), `RUN` selects the run id. | `wf-ids.txt`. |
| `interact-wf.cjs` | 16 PASS/FAIL checks on the editor: quick-add from the port, palette click, drag connect, cycle guard, pill condition change, snap, Delete, undo/redo round trip, problem badge + disabled Save, leave guard. Exits 1 on any FAIL or page error. | `wf-ids.txt`. Mutates the Release workflow in the browser but does not save. |
| `approve-wf.cjs` | Approve inside the card, delay countdown ticks, viewport survives polling, task log opens on click. | `/tmp/semaphore-stand/wf3.txt` with `<workflow id> <run id>` of a run waiting at an approval — seeded by hand, not by `seed-wf.sh`. |
| `keys-wf.cjs` | Keyboard pan/zoom/fit/reset in the editor and the run view; asserts no effect while typing in the properties panel. | `wf-ids.txt`, run 2 of Release. |
| `shoot-inputs.cjs` | Shoots the four inputs screenshots above and runs 7 PASS/FAIL checks (checkbox state, mapping rows, hint line, by_name chips with a matched output, output count on a card, Outputs panel in the task dialog). `--dark` for the dark theme; `WF` (default 9, explicit workflow), `WF_BYNAME` (default 1), `OUT` (default `/tmp/semaphore-stand/shots-inputs`). Edges are selected by a dispatched `mousedown` on the path — a real click lands on the condition pill. | `AGENTS/tools/wf-stand/seed-outputs.sh` seeded, a finished run of the current revision of both workflows (`run.sh 9`, `run.sh 1`) for the key hints. |
| `interact-inputs.cjs` | 10 PASS/FAIL checks of the inputs round trip on workflow `WF` (default 12): select the explicit edge, not dirty before editing, type a key, dirty, explicit icon on the pill, Save → API shows the pair, uncheck → by_name chips, Save → fields gone, original graph restored by PUT. Exits 1 on any FAIL or page error. | Running stand with the seeded project; mutates workflow 12 and restores it. |
| `shoot-nodes.cjs` | Crops one card per node kind (task hovered with the "+", approval, delay, note) with edge pills hidden and zoom 1; source of `docs/static/assets/workflow-node-*.webp`. | `WF` env = id of the "Docs node kinds" workflow on the stand (default 4). |
| `make-hotkeys-svg.py <out.svg>` | Generates the keyboard cheat-sheet `docs/static/assets/workflow-hotkeys.svg`; edit `COLUMNS` and re-run when a shortcut changes. | Python 3 only. |
