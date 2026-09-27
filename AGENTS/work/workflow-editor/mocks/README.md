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

## Scripts

| Script | Does | Needs |
| --- | --- | --- |
| `seed-wf.sh` | Seeds five templates (one with an XSS name), the "Release" workflow with positions, the position-less "Nightly checks" (with a `<script>` note), starts one Release run; writes ids to `/tmp/semaphore-stand/wf-ids.txt`. Idempotent for templates. | Running stand, repository 1 and inventory 1 in project 1. |
| `shoot-wf.cjs` | Editor/run/legacy/empty screenshots in light and dark; output `OUT` (default `/tmp/semaphore-stand/shots-wf`), `RUN` selects the run id. | `wf-ids.txt`. |
| `interact-wf.cjs` | 16 PASS/FAIL checks on the editor: quick-add from the port, palette click, drag connect, cycle guard, pill condition change, snap, Delete, undo/redo round trip, problem badge + disabled Save, leave guard. Exits 1 on any FAIL or page error. | `wf-ids.txt`. Mutates the Release workflow in the browser but does not save. |
| `approve-wf.cjs` | Approve inside the card, delay countdown ticks, viewport survives polling, task log opens on click. | `/tmp/semaphore-stand/wf3.txt` with `<workflow id> <run id>` of a run waiting at an approval — seeded by hand, not by `seed-wf.sh`. |
| `keys-wf.cjs` | Keyboard pan/zoom/fit/reset in the editor and the run view; asserts no effect while typing in the properties panel. | `wf-ids.txt`, run 2 of Release. |
| `shoot-nodes.cjs` | Crops one card per node kind (task hovered with the "+", approval, delay, note) with edge pills hidden and zoom 1; source of `docs/static/assets/workflow-node-*.webp`. | `WF` env = id of the "Docs node kinds" workflow on the stand (default 4). |
| `make-hotkeys-svg.py <out.svg>` | Generates the keyboard cheat-sheet `docs/static/assets/workflow-hotkeys.svg`; edit `COLUMNS` and re-run when a shortcut changes. | Python 3 only. |
