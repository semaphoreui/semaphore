# Agent setup — initializing the agent context layer

Instructions for the agent that sets up `AGENTS/` in a project. All paths in this file are relative
to the project root.

## Rules

- Require the template's `AGENTS/` folder copied into the project root by the user. A project that
  already had its own `AGENTS/` is the user's to move or rename first — never merge into it.
- Run from the project root.
- Edit only the context files: everything under `AGENTS/` except `AGENTS/agents/`. Outside them,
  touch only what a section below names: the root symlinks and the harness memory symlink.
- Exclude `AGENTS/agents/` from every search — its files name MCPs and placeholders on purpose.
- Run the sections in order.
- Continue past a failed section — warn the user which functionality will not work.
- End with the final report, the project study offer and the restart advice, in that order.

## Talking to the user

- Write in the user's language.
- Skip preambles, recaps and closing pleasantries — start with the point.
- Put one fact, question or action per bullet.
- Give a reason only when it is not obvious — one short clause after a dash.
- State facts plainly — no "might", "probably", "it seems".
- Skip praise and marketing words.
- Skip explaining what a tool, MCP or symlink is — name what was done or is needed.
- Ask one question at a time and wait for the answer.
- Use no emoji and no capitals for emphasis.

## Root instruction files

`AGENTS/agent-primary.md` is the single source. Each harness reads it through a relative root
symlink — portable when the project is moved or cloned:

| Root file | Read by |
| --- | --- |
| `CLAUDE.md` | Claude Code |
| `AGENTS.md` | OpenAI Codex and other `AGENTS.md`-convention harnesses |

1. Inspect each root file:
   - absent → `ln -s AGENTS/agent-primary.md <file>`;
   - a symlink resolving to `AGENTS/agent-primary.md` → leave it;
   - a regular file or a symlink elsewhere → don't touch it. Ask the user whether to delete, rename
     or back it up; act only on an explicit answer, then create the symlink.
2. Confirm with `readlink -f CLAUDE.md AGENTS.md`; print `AGENTS/agent-primary.md`.

Never edit the root files afterwards — edit `AGENTS/agent-primary.md`.

## MCP checks

The context files must use each MCP name exactly as this harness registers it.

| | Research | Tasks |
| --- | --- | --- |
| Server | <https://github.com/TurkovBogdan/urb-research> | <https://github.com/TurkovBogdan/urb-workbench> |
| Template name | `urb-research` | `urb-workbench` |
| Identify by tools | `research_list`, `research_get`, `area_create`, `query_search_run`, `source_review` | `workspaces_list`, `workspace_use`, `task_create`, `stage_close`, `note_add` |
| Read-only probe | `research_list` | `workspaces_list` |
| Files that name it | `AGENTS/agent-primary.md`, `AGENTS/research/INDEX.md` | `AGENTS/agent-primary.md`, `AGENTS/tasks/INDEX.md` |
| If missing | Research in this layer requires it | Strongly recommend installing it |

For each server:

1. Find it by its tool set, not by name; confirm with the read-only probe.
   - found once → note its registered name;
   - found under several names, e.g. stable and dev → ask the user which one the project uses;
   - absent or failing to connect → tell the user, give the repository link, leave the files
     unchanged, skip the rest for this server.
2. Search: `grep -rn --exclude-dir=agents "<template name>" AGENTS/`.
3. Replace every mismatch with the registered name — plain mentions and inside `mcp__<server>__…`
   identifiers. Leave `github.com/...` links. Re-run the search: no stale name remains.

## Tasks workspace

Run only if urb-workbench was found. The workspace is the user's choice, never yours.

1. Show the `workspaces_list` output (code, title, description); ask which workspace the project
   uses. Wait for an explicit `WORKSPACE@` code.
   - Never pick one yourself, even when a title matches the project.
   - Never create one — no MCP tool can. Never fall back to an existing one.
2. No suitable workspace → ask the user to create one in the urb-workbench interface and give its
   code. To open the interface: `workspace_use` any existing workspace (per-connection, harmless),
   then `interface_open` with its code. Browser fails → pass the address from the error to the user.
   No workspace exists at all → ask the user to open the interface themselves.
3. Check the given code appears in `workspaces_list`. No code given → leave the placeholders, warn
   that tasks will not work until a workspace is set, move on.
4. Run `workspace_use` with the code; it must succeed.
5. Replace every `WORKSPACE@…` placeholder with the code, the workspace title next to it:
   - `AGENTS/tasks/INDEX.md` — the Workspace table and the session-start call;
   - `AGENTS/agent-primary.md` — the Tasks section: the workspace line and the session-start call.
6. Confirm with `grep -rn --exclude-dir=agents "WORKSPACE@" AGENTS/`: one code, no placeholder.

Never pin the workspace in the MCP client config (`MCP_WORKSPACE`) — the agent binds it every
session.

## Local database tool

- Ask one question: should the agent get a tool for quick queries to the project's local database?
- Search, read or build nothing before the answer.
- Explicit yes → follow `AGENTS/agents/setup/db-query/SETUP.md`.
- Anything else → skip.

## Memory migration

Move the agent's memory into `AGENTS/memory/` — out of the harness's private storage, into the
portable layer.

**Codex** keeps only global memory — skip; write project memory straight into `AGENTS/memory/`.

**Claude** keeps it in `~/.claude/projects/<encoded-path>/memory/`: the absolute project path with
every non-alphanumeric character replaced by `-` (`/home/user/my_app` → `-home-user-my-app`).
Confirm by listing `~/.claude/projects/`.

1. Locate the directory. Absent or empty → step 4.
2. Merge — don't copy — into `AGENTS/memory/` per the rules at the top of `AGENTS/memory/MEMORY.md`:
   - write each short fact into `MEMORY.md` as one or two lines, under Rules or Gotchas;
   - group facts that cluster around one topic into `AGENTS/memory/<topic>.md`, with a row in the
     Topic files table;
   - move a fact that needs sections and prose to a doc in `AGENTS/docs/` or `AGENTS/work/`;
   - drop harness front matter — the format is plain lines and topic files;
   - keep the template header; drop a section's example line once it has a real entry.
3. Build a map: every harness file and index entry → the file in `AGENTS/` that holds it. Fix the
   merge until every fact has a destination.
4. Delete the harness directory only after step 3 passes — a partial migration is data loss. Create
   a symlink at its path to the project's `AGENTS/memory/`; create the parent if missing.
5. Confirm the symlink lists exactly the files in `AGENTS/memory/`.

## Final report

List, section by section:

- created or changed — files, symlinks, recorded names and codes;
- skipped, and why;
- what will not work until the user acts, with install links;
- the memory map, if one was made.

## Project study

The project sections at the top of `AGENTS/agent-primary.md` — "About the project" through
"Testing" — still hold template placeholders. Each placeholder says what its section must contain.

1. Offer the study, in the user's language, in these words or close to them: "The project sections of the agent
   instructions are still empty. I can launch subagents to study the project in depth — stack,
   structure, environments, how it runs, builds and is tested — and fill them in. It takes a
   while and a fair amount of tokens. Start?"
2. Anything but an explicit yes → skip to Restart; name the empty sections as the user's to fill.
3. Launch subagents in parallel, one per question group; no subagent tool → study each group
   yourself, one after another:

   | Subagent | Fills |
   | --- | --- |
   | Project and structure | About the project, Structure |
   | Environments | Inventory: dev / prod / test |
   | Run and build | Running and building |
   | Tests | Testing |

4. Give each subagent its section's placeholder text as the spec, plus these rules:
   - study as deep as the project allows: code, configs, compose and CI files, scripts, the
     project's own docs — not only the README; leave `AGENTS/` out;
   - read only — never run a build, deploy, migration, test or any command that changes state;
   - state only what a file shows; name the file behind every fact;
   - give exact commands, paths and names, copied from the project, not paraphrased;
   - never copy a secret — name the file and the key that hold it;
   - list what could not be established as open questions — never guess;
   - return findings, don't edit files.
5. Write the sections yourself from the findings:
   - replace each placeholder whole;
   - keep them short — the file loads every session; link details that already exist in the
     project's own docs; put new detail in a `AGENTS/work/` zone per "Docs and work" in
     `AGENTS/agent-primary.md`;
   - write every path and link from the project root;
   - one topic in one place — nothing already in memory, docs or work.
6. Show the user the filled sections and the open questions; apply their corrections and answers.

## Restart

Tell the user to restart the agent or IDE — root files, the memory symlink and new MCPs load only in
a new session.
