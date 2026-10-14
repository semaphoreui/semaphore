# Context cleanup

Instructions for the agent auditing the `AGENTS/` documentation: repair broken references and
flag files that no longer fit a single read.

All paths in this file are relative to the project root.

## Task

1. Run the tool below over `AGENTS/`.
2. For every `LINE` finding — a dead markdown link, `[[note]]` link, or (with `--prose`) a dead
   path in backticks: find where the target moved and fix the link. If the target is truly gone,
   delete the link (and the sentence around it, if the link was its whole point) instead of
   leaving a reference to nothing.
3. For every `SIZE` finding: do not edit the file yourself. Report it to the user — splitting or
   trimming an oversized doc is a content decision, not a cleanup one.
4. Re-run the tool after fixing links to confirm the report comes back clean (aside from any
   `SIZE` findings still pending the user's call).

## The tool: `AGENTS/agents/context-cleanup/check-docks.py`

Walks the workspace's markdown files and collects what is broken into a report. Set up after
folder moves: a dangling link lives silently, and a search cannot tell it apart from a live one.

Two checks — link integrity and file size. Future ones live here too, sharing the same report.

### Usage

```
python3 AGENTS/agents/context-cleanup/check-docks.py [path] [--report FILE] [--ignore FILE] [--prose] [--exclude NAME]
```

- `path` — what to walk, default `AGENTS`.
- `--report` — where to write the report, default `AGENTS/agents/context-cleanup/report.txt`.
- `--ignore` — list of skipped folders, default `AGENTS/agents/context-cleanup/ignore.txt`.
- `--prose` — also check paths in backticks, not only `[text](target)`.
- `--exclude` — skip a folder by name, repeatable; `tmp` is always skipped.
- Exit code: `0` — clean, `1` — problems found, `2` — folder missing or report not writable.

The report goes to a file, only the summary is printed to the screen. Warning: the file is
cleared at the start of every run — the previous output does not survive the new one, even when
nothing was found this time.

`AGENTS/agents/context-cleanup/ignore.txt` holds one path from the project root per line, `#` starts
a comment; right now it holds `AGENTS/agents/setup` — the setup instructions name files the setup
creates later — and `AGENTS/agents/context-cleanup` — this file quotes sample paths on purpose;
`--prose` flags both. Skipping closes a folder to READING, not to checking: a link
from a live doc into an archived one is still verified. No file — nothing is skipped;
`--ignore /dev/null` clears the list.

A link counts as alive only if it resolves **from its own file's folder** — a path from the repo
root does not always work for the reader. An exception is made for `AGENTS/agent-primary.md` and
any other file a root symlink points at — the agent reads it through `CLAUDE.md` / `AGENTS.md` in
the root, so its links are deliberately written from the root, whether the symlinks exist yet or
not.

A note link `[[name]]` is checked the same way as a regular link — the name must match a file in
`AGENTS/memory/`. A link written inside backticks is skipped — that is a rule's sample, not an
actual link.

File size is measured in bytes against the read thresholds; anything not `NORMAL` goes into the
report.

| Status | Size | What happens on read |
| --- | --- | --- |
| `NORMAL` | ≤ 30,000 | `Read` in full and `cat` in full — every channel works |
| `WARNING` | 30,001 – 61,000 | `Read` in full; `cat` returns the path and a 2 KB preview |
| `ERROR` | > 61,000 | trimmed: from 61,000 for dense content, from 73,000 for any content |
| `HARD` | > 262,144 | `Read` refuses outright, not even one line |

The report is grouped by file: the path, then `SIZE #<STATUS>` with the size and `LINE #<number>`
with the line text for every finding. A line with two dead links is printed once.

### Examples

```
$ python3 AGENTS/agents/context-cleanup/check-docks.py
Problems found: 71. Report: …/AGENTS/agents/context-cleanup/report.txt

$ cat AGENTS/agents/context-cleanup/report.txt
AGENTS/agent-primary.md
SIZE #WARNING
32 634 B, limit 30 000

AGENTS/docs/frontend/admin-layout.md
SIZE #ERROR
151 035 B, limit 61 000

AGENTS/work/billing/checkout-chain.md
SIZE #WARNING
30 782 B, limit 30 000
LINE #317
- Source device: [...](../../docs/project/old-portal/legacy-trial.md)
LINE #318
- Stages and synthetic dates: [...](../../docs/project/migration/commands.md)

Problems found: 71
```

```
$ python3 AGENTS/agents/context-cleanup/check-docks.py AGENTS/memory --report /tmp/memory.txt
No problems found. Report: /tmp/memory.txt
```

```
$ python3 AGENTS/agents/context-cleanup/check-docks.py --prose
Problems found: 200. Report: …/AGENTS/agents/context-cleanup/report.txt
```

### Traps

- Warning: `--prose` gives false positives on samples: `AGENTS/work/<zone>/INDEX.md` in a
  template, `AGENTS/memory/topic.md` in a rule. Read the output by eye, don't fix it as a list.
- Warning: a dead link is sometimes deliberate — a board registry keeps a line for a file that's
  gone, on purpose. The call on such a line belongs to the user.
- Warning: `AGENTS/agents/context-cleanup/ignore.txt` also hides size — large files in a skipped
  folder don't show up in the default report.
  Get the full picture with `--ignore /dev/null`.
