# Load scenarios — what was estimated and what it implies

Both source studies (2026-06-04) were static read-throughs; no profile, benchmark or slow-query log
was captured. Every number below is a derived estimate, not a measurement. Finding IDs refer to
`bottlenecks.md`; verdicts there are as of 2026-09-27.

## Scenario A — hundreds of tasks queued or running at once

Where the cost concentrates, in order of estimated weight:

1. Scheduling: one goroutine, O(N) queue re-scan per event, a `GetProject` per candidate (TP-1, TP-2).
   A burst of N completions was estimated at ~N × O(N) passes with a DB round-trip each.
2. Per-task polling: 300 remote tasks ⇒ ~300 `GetTask` scans/s and, in HA, ~300 DB reads/s (TP-3).
   Removed on 2026-06 by event-driven finalisation — the biggest structural change since the study.
3. Output fan-out: cost ≈ lines × users × connections through one hub goroutine, on the reader
   goroutine, with back-pressure into the subprocess (OUT-1, OUT-2).
4. Steady write stream: one `TouchRunner` per runner poll and one `TouchSession` per UI request
   (RUN-2, API-7) — serialised on SQLite's single writer.
5. Amplifiers: unlimited default parallelism (TP-7) with an unbounded MySQL/Postgres pool (DB-8) —
   memory per task ≈ 20 MB of scanner buffers + two 100k channels (OUT-7).

Runner-side cadence (verified 2026-09-27): progress `PUT` fixed at 1 s, job `GET` at
`runner.check_interval_seconds` (default 1 s); each request costs an unindexed `runner.token` lookup.

## Scenario B — one static inventory of ~5 000 hosts

Conclusion of the inventory trace: Semaphore never parses, splits or iterates hosts in Go. There are
no per-host loops, key installs or syscalls; per-host fan-out belongs to Ansible (`forks`). The cost
is carrying and re-serialising the blob, and — dominantly — the output volume such a run produces.

| Execution path | Cost per run (estimated) | Scales with |
| --- | --- | --- |
| Remote runner | inventory (~0.5–2 MB) JSON-encoded on every poll while the job is starting (INV-1) | bytes × polls — worst; was × runners and RSA-encrypted until 2026-06 |
| Local static inventory | one `os.WriteFile` + one `[]byte` copy (INV-3) | bytes × tasks |
| Local file inventory | none in Go; Ansible reads the repo file | nothing |
| Dynamic / Terraform | state is a separate blob, not host-expanded | nothing |

A 5 000-host run emits 10⁵+ stdout lines, so on the local path the output pipeline (OUT-1, OUT-5,
OUT-6, OUT-7) — not the inventory — is the dominant cost; the JSON output endpoint would load and
marshal all of those lines per viewer.

## Suggested measurements (never run)

- `pprof` CPU + goroutine profiles of the server under ~200 concurrent chatty tasks; expect
  `sendToWs`/`json.Marshal`, `handleQueue`/`GetProject` to dominate.
- Slow-query log on MySQL/Postgres; expect `SELECT * FROM runner WHERE token=?`, the `GetTaskStats`
  `GROUP BY`, and the `clearTasks` `ORDER BY created` to appear; watch open connections vs
  `max_connections`.
- One 5 000-host task on a remote runner: size of each `GET /api/internal/runners` response while the
  task is starting, and CPU per poll.
- Heap profile with many concurrent tasks: scanner buffers and channels (OUT-7), runner-side
  `logRecords` growth when the server is slow (OUT-6).

## Remediation order the sources agreed on

1. Cheap: the eight missing indexes (DB-0), bounded DB pool (DB-8), debounced `TouchRunner`/
   `TouchSession` (RUN-2, API-7), delete the ws `fmt.Println` (OUT-8), per-row fallback on batch
   insert failure (OUT-4).
2. Output path: marshal once + async lossy fan-out (OUT-1/OUT-2), paginate the JSON output endpoint
   (OUT-5), cap runner buffers (OUT-6), lazy scanner buffer (OUT-7).
3. Scheduling: O(1) lookups and per-runner counters in the state store (TP-3..TP-6), cached parallel
   limits out of `blocks()` (TP-2), real admission control instead of `9999` (TP-7), build the runner
   job once (RUN-3/INV-1).
4. Hygiene: batch the N+1s (DB-1, DB-2, DB-7), retention off the write path (DB-4), cached stats
   (DB-10), alerts on a background worker (TP-8).
