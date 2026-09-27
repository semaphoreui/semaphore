# AGENTS.md — instructions for the agent

## About the project

One paragraph: what the project is, its stack, how it runs end to end, its single goal.

## Structure

Top-level folders by role — developed, runs it, tests it, reference — not by technology.

## Inventory: dev / prod / test

Where runtime artifacts live per environment — databases, caches, logs, build output — and what is
never committed.

## Running and building

Commands to bring up, build and deploy; the single entry point; a link to the full runbook.

## Testing

Suite layout — markers, layers, opt-in parts — and the exact commands; a link to the full doc.

## At session start

| Read | When |
| --- | --- |
| [`AGENTS/memory/MEMORY.md`](AGENTS/memory/MEMORY.md) — short rules, gotchas, index of memory topic files | Every session |
| [`AGENTS/tasks/INDEX.md`](AGENTS/tasks/INDEX.md) — workspace binding and task rules | Every session |
| Every `INDEX.md` under `AGENTS/docs/` and `AGENTS/work/`, at any depth — list the folders, don't only follow links | Every session |
| [`AGENTS/research/INDEX.md`](AGENTS/research/INDEX.md) — shortlist of `RESEARCH@` codes | Before any research |
| [`AGENTS/tools/INDEX.md`](AGENTS/tools/INDEX.md) — project scripts for the environment | Before improvising a raw command |

## Where a fact goes

- Write a short rule or gotcha straight into `AGENTS/memory/MEMORY.md`, one or two lines — the fact
  and its reason, not the story of how it was found.
- Move entries that cluster around one topic into `AGENTS/memory/<topic>.md`; replace them in
  `MEMORY.md` with one index row saying what the file holds.
- Read a memory topic file only when the task touches its topic.
- Keep `MEMORY.md` under 150 lines — it loads every session; past that, group entries into topic
  files.
- Put a fact that needs sections and prose in a doc in `docs/` or `work/`, not in memory.
- Keep each topic in one place — never in both memory and a doc.
- Follow this memory format over the harness's own — no file per fact, no front matter.
- Put this project's subsystem, decision or open problem in [`work/`](AGENTS/work/INDEX.md).
- Put an approach that would serve another project in [`docs/`](AGENTS/docs/INDEX.md).
- Put a repeatable operation on the environment in a script in `tools/` — read
  [`AGENTS/tools/CREATE.md`](AGENTS/tools/CREATE.md) first.
- Put a temporary file or a one-off report in `AGENTS/tmp/`.
- Check a fact from memory, docs or work against the code before relying on it — code wins; fix the
  stale doc in the same pass.

## Docs and work

- Keep every doc inside a zone — a folder with its own `INDEX.md`, created with the folder. Nothing
  sits loose in the root of `docs/` or `work/`.
- Nest subzones in `docs/` only, each with its own `INDEX.md`.
- Keep `work/` zones flat — a topic that outgrows its zone becomes a new zone.
- Name a `work/` zone by subject, not by code module.
- Write an `INDEX.md` in `docs/` and `work/` as an index of files only: one paragraph on what the folder covers, then a
  `When you need it | Read` table. No facts or rules.
- Give every doc and every child `INDEX.md` at least one row — one row per situation, so a file
  may get several.
- Write an index row as the reason to open the file, not a retelling.
- Open a doc only for the zones the task touches. Don't search for a file by name past the indexes.
- Link to a doc, never copy it.
- Update the zone index in the same pass as the doc.
- Flag a trap or a divergence from the code in the doc with ⚠️, dated — a stale doc is worse than
  none.
- Keep `docs/` portable — an app name, port or container name means the doc belongs in `work/`.
- Split a subject with a shared half: the approach stays in `docs/`, the implementation goes to a
  `work/` zone and links back.
- Write a `docs/` doc only for what a config cannot express and code does not show.
- Let the project win over a `docs/` convention — fix the doc in the same pass.

## Tasks

- Run, check and store every task and plan in the `workbench` MCP — rules in
  `AGENTS/tasks/INDEX.md`.
- Use workspace `WORKSPACE@…` (…) — the only one for this project.
- Bind it at session start, before any task call — the MCP client config does not pin it:

```
mcp__workbench__workspace_use(workspace_code="WORKSPACE@…")
```

## Research

- Do and store all research in the `research` MCP — the repo keeps only the code shortlist.

## Temporary files

- Never read `AGENTS/tmp/` at session start.
- Never link into `AGENTS/tmp/` from docs, memory or this file.
- Move a result worth keeping to its own home before the work is done.
