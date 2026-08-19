---
name: semaphore-cli-docs
description: >
  Generate or update the CLI documentation under `docs/docs/admin-guide/cli/` from the Cobra command code in
  `cli/`. Use this skill whenever the user asks to "update the CLI docs", "actualize/sync the CLI documentation",
  "document the CLI", "the CLI docs are out of date", "regenerate the runner/user/vault/project/migrate docs", or
  after adding or changing a Cobra command, subcommand, flag, alias, or required-argument check anywhere under
  `cli/cmd/`. Covers every command group (`users`, `projects`, `vaults`, `runner`, `migrate`) plus the operational
  commands (`server`, `setup`, `version`) and the global flags on the root command. Trigger when the user adds a
  new `*Cmd` Cobra command, a new `PersistentFlags()` call, or renames/removes a CLI flag.
---

# Semaphore — CLI Documentation Generator

This skill produces and maintains the CLI reference pages under
`docs/docs/admin-guide/cli/` and the CLI index `docs/docs/admin-guide/cli.md`. Every page is **derived from the
Cobra command code** in `cli/cmd/`. If a command, subcommand, or flag isn't in the code, it must not be in the
docs; if a flag's name, default, or required-ness changed, the docs must change with it.

> `docs/` is a **git submodule** (the documentation site). Editing files under `docs/docs/...` and
> `docs/sidebars.js` is correct; those changes are committed to the submodule, not the main repo.

The docs are the human-facing contract. The Go code is the source of truth. When they disagree, the code wins —
**but** read the command's `Run` body, not just its help strings, because help text drifts (see Pitfalls).

## Inputs (source of truth)

Read these in full before writing. Each command group maps to exactly one page.

| Source files (`cli/cmd/`) | Doc page |
|---------------------------|----------|
| `root.go` | `docs/docs/admin-guide/cli.md` — global flags, command overview |
| `server.go`, `setup.go`, `version.go` | `docs/docs/admin-guide/cli.md` — operational commands |
| `user.go`, `user_add.go`, `user_change.go`, `user_delete.go`, `user_get.go`, `user_list.go`, `user_token.go`, `user_totp.go` | `docs/docs/admin-guide/cli/users.md` |
| `project.go`, `project_export.go`, `project_import.go` | `docs/docs/admin-guide/cli/projects.md` |
| `vault.go`, `vault_check.go`, `vault_rekey.go` | `docs/docs/admin-guide/cli/vaults.md` |
| `runner.go`, `runner_register.go`, `runner_setup.go`, `runner_start.go`, `runner_unregister.go` | `docs/docs/admin-guide/cli/runners.md` |
| `migrate.go` | `docs/docs/admin-guide/cli/migrations.md` |

`*_test.go` files (e.g. `runner_register_test.go`, `user_token_test.go`) encode exact behavior — read them to
document precise rules (e.g. registration-token resolution order, token expiry formatting).

`syslog.go` / `syslog_windows.go` are internal helpers, **not** commands — never document them.

Re-discover the command tree before each run; don't trust this table if files were added/removed:

```bash
ls cli/cmd/*.go | grep -v _test
grep -rn "cobra.Command{" cli/cmd/ | grep -v _test          # every command/subcommand
grep -rn "AddCommand\|PersistentFlags()\." cli/cmd/          # wiring + flags
```

## How to read a Cobra command

Extract these facts mechanically from each `*Cmd = &cobra.Command{...}` and its `init()`:

- **Command name** — the `Use:` field (first word). Subcommands are wired with `parent.AddCommand(child)`.
- **Aliases** — the `Aliases:` slice. Document them (e.g. `users`/`user`, `projects`/`project`,
  `vaults`/`vault`, `server`/`service`). A command whose `Run` only calls `cmd.Help()` is a pure group — it has
  no behavior of its own, just subcommands.
- **Flags** — each `cmd.PersistentFlags().<Type>Var(&arg, "name", default, "usage")` becomes one table row:
  flag = `--name`, description from the usage string, and note the default when it's meaningful (e.g.
  `--enabled` defaults to `true`, `--project-id` defaults to `0` = "global"). `StringSliceVar` → "comma-separated
  or repeat the flag".
- **Required / mutually-exclusive args** — read the `Run` body. Patterns like
  `if args.x == "" { fmt.Println("Argument --x required"); ok = false }` mean **required**; checks that reject
  setting both of two flags mean **exactly one of**. Document these rules explicitly ("Exactly one of `--a` or
  `--b` is required").
- **Output** — what the command prints on success (IDs, a table, nothing). Mirror the real `fmt.Printf` format
  in the doc, including example output blocks for commands that print structured results (e.g. `vault check`).

### Global flags (from `root.go`)

Documented once in `cli.md`; they apply to every command. Currently: `--config`, `--no-config`, `--log-level`
(`DEBUG|INFO|WARN|ERROR|FATAL|PANIC`, env `SEMAPHORE_LOG_LEVEL`), `--debug-filter` (DEBUG-only namespace filter,
env `SEMAPHORE_DEBUG_FILTER`). Re-read `Execute()` in `root.go` for the live list.

## Page structure (match the existing style)

Every command-group page follows the shape established in `users.md` and `vaults.md`:

1. `# <Title>` then a one-sentence purpose.
2. A fenced `semaphore <group> --help` block.
3. An alias note as a `>` blockquote when the group has an alias (`> \`vault\` is an alias for \`vaults\`.`).
4. A subcommand table (`| Command | Purpose |`) with in-page anchor links when the page is long.
5. One `##` section per subcommand: a short description, a copy-pasteable example, then a
   `| Flag | Description |` table. State required/exactly-one rules in prose under the table.

Conventions:

- Internal links are absolute from the route base `/`: **`/admin-guide/cli/<page>`** (NOT `/cli/<page>` — that
  base is wrong and 404s). Cross-links to concept pages use the same form, e.g. `/admin-guide/runners`,
  `/admin-guide/security/encryption`.
- Admonitions use Docusaurus syntax: `:::tip`, `:::warning` … `:::`.
- Keep prose task-oriented and match the tone/heading depth of neighboring pages.

## Index page (`cli.md`) and sidebar

- `cli.md` carries: intro + `semaphore help`, a command-group table linking to each page, the **global options**
  table, and short sections for `version`, `setup`, `server`, `runner`, `migrate`.
- Every command-group page must be registered in **`docs/sidebars.js`** under the `label: 'CLI'` category's
  `items` array, as `'admin-guide/cli/<page>'`. Adding a page without registering it hides it from navigation.

## Workflow

### 1. Enumerate the live command tree
Run the discovery greps above. Build the set of command groups and subcommands that actually exist. Note any
group/subcommand/flag that has no corresponding doc section, and any doc section with no corresponding code.

### 2. Extract facts per command
For each command, pull name, aliases, flags (name/default/usage), required/exclusive rules (from `Run`), and
success output. Cross-check `*_test.go` for exact behavior.

### 3. Update each page
Rewrite or patch each page so it matches the extracted facts and the page structure above. Add sections for new
subcommands; remove sections for deleted ones; fix flag tables. Leave a page untouched only after confirming it
already matches the code.

### 4. Update the index and sidebar
Refresh `cli.md` (command table, global options, operational commands) and ensure every page is listed in
`docs/sidebars.js`.

### 5. Build to catch broken links
From the `docs/` directory:

```bash
cd docs && npm run build 2>&1 | tail -40
```

The build must end with `[SUCCESS] Generated static files`. **Broken-link** errors in pages you touched must be
fixed. Pre-existing **broken-anchor** warnings in unrelated pages (e.g. `admin-guide/configuration`,
`admin-guide/installation_manually`) are not yours to fix — note them and move on. If `node_modules` is absent,
run `npm install` first (or tell the user the build couldn't run).

## Pitfalls

- **Dead flags.** A flag registered in `init()` but never read in the command's `Run` does nothing — do not
  document it as functional. Example: after BoltDB removal, `migrate` still registers `--err-log-size`,
  `--skip-task-output`, `--merge-existing-users`, but `Run` only uses `--apply-to` / `--undo-to`. Verify each
  flag is actually consumed before documenting it.
- **Removed features → version-gate, don't silently delete.** When a command/flag is removed (e.g.
  `migrate --from-boltdb`; BoltDB is "not supported starting from 2.19" per `cli/setup/setup.go`), keep the
  guidance but mark the version boundary with a `:::warning`, so users on older versions are still served. Note
  when other parts of the repo (e.g. `deployment/docker/server/server-wrapper`) still reference a removed flag —
  that's a stale call site worth flagging to the user, but out of scope for the docs change.
- **Stale help strings.** A command's `Short`/`Long` text can lag the code. Verify config-key names against the
  actual structs in `util/` (e.g. `vault rekey`'s `--old-key` help says `access_key`, but the real config field
  is `secret_key` / `secret_key_file` in `util/config.go`). Document the struct, not the help string.
- **Group commands have no flags of their own.** `users`, `projects`, `vaults`, `runner`, `token`, `totp` just
  print help and dispatch to subcommands. Don't invent flags for them.
- **Don't document the internal `server` runtime.** `server.go`/`root.go` wire a lot of services; the doc only
  needs "start the server, `service` is an alias, respects global flags."

## When to ask or refuse

- If a command's behavior is genuinely ambiguous from the code (e.g. a flag is read but its effect spans
  packages you can't quickly trace), summarize what you found and ask the user rather than guessing.
- If asked to document a flag or command that **doesn't exist in `cli/cmd/`**, refuse and explain: CLI docs are
  generated from the Cobra code — add the command/flag first, then regenerate.
- Before deleting a whole documented section (a removed command group), propose it explicitly; prefer
  version-gating over silent removal when older releases still ship the feature.
