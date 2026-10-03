# performance — zone index

Where the Go backend and the Vue frontend stop scaling: hot queries, N+1 loads, locks, missing
indexes, unbounded reads and polling loops, each tied to a file and function and re-checked against
the code on a stated date, plus the load scenarios those findings were derived from.

| When you need it | Read |
| --- | --- |
| Touching the task pool, scheduler, state store or remote-task dispatch and wanting to know which hot spots are still open there | `bottlenecks.md` |
| Adding or reviewing a DB index, a list query or a retention/stat query on `task`, `task__output`, `runner` or `event` | `bottlenecks.md` |
| Changing task output streaming (websocket hub, log writer, runner progress) or the output API | `bottlenecks.md` |
| Changing the runner poll protocol or runner authentication | `bottlenecks.md` |
| Checking whether an old performance claim (June 2026 studies) still matches the code, or which ones were already fixed and by what commit | `bottlenecks.md` |
| Planning a load test or a profiling session and wanting the scenarios, expected hot spots and the agreed remediation order | `scenarios.md` |
| Answering "does a 5 000-host inventory hurt, and where" | `scenarios.md` |
