# Workflow editor — direction

## Editor redesign (2026-09-26)

Task `TASK@88179c9bc1` (7 stages) in the workbench; research `RESEARCH@5bf4fc406f` (n8n/Kestra →
Vue Flow + dagre; Airflow 3/Dify → React Flow + elkjs; Vue Flow is Vue 3 only; Drawflow
unmaintained since 2024-09) and the wider market survey `RESEARCH@e6d5aa256f`.

Decision (Denis): no custom canvas and no library swap — Drawflow stays; only the visual layer and
UX are redone n8n-style (nodes mounted as Vue components with `parent: this`, edge overlays, own
wheel/zoom/fit, quick-add «+»). Standing guardrails: no new npm dependency, no Drawflow patching,
no API/DB change and no second Vue 3 runtime without asking first; no autosave; the palette stays.
Decisions, status and open items: `editor-architecture.md`; the approved mock, screenshots and
shooting scripts: `mocks/README.md`.

## Workflows-as-Code (2026-09-27)

Triggered by an Enterprise customer asking for "YAML workflows like Kestra". Research
`RESEARCH@bc356e4b9b` (group `GROUP@455e5d3336`), digest in
the research body itself — do not re-research Kestra syntax.

Agreed direction: YAML is a second projection of the existing DAG model, engine untouched. Format
`apiVersion: semaphore/v1, kind: Workflow`, nodes with slug `id`, `template` by name,
`depends_on: [{node, on}]` on the child, `converge`, no coordinates/DB ids/secrets. Phases:
(1) spec package + JSON Schema + validate/source/apply API + Source tab + revisions + CLI;
(2) managed-by-Git sync + triggers for workflows; (3) inputs/retry/sub-workflow/Kestra converter.
Enterprise wants Git-as-truth, CI validation, revision audit, secrets outside the file,
UI⇄text equivalence; Kestra pain points to exploit: JVM footprint, EE paywall on approvals/audit/SSO.

Known gaps: template names are not unique per project (backup uses `makeUniqueNames`),
Schedule/Integration have no workflow target yet, CLI has no API mode.
