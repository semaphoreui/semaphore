# Stage 2: Linux supervisor on normal completion

## Goal

Add a Linux-only, short-lived supervisor process and prove that it removes the
background `sleep` after Bash exits while preserving Bash's exit status.

This remains a standalone prototype and does not change `LocalExecutor`.

## Process model

The first process is the prototype server. It starts a second process from the
same executable in a private supervisor mode:

```text
prototype server
└── task supervisor (subreaper)
    └── bash
        └── sleep
```

The prototype server only starts the supervisor, connects its output, waits for
it, and returns its exit status. Using the same executable avoids adding another
binary while still creating the separate OS process required for the subreaper
boundary.

## Supervisor lifecycle

The supervisor will:

1. Enable `PR_SET_CHILD_SUBREAPER` on itself.
2. Start Bash in a new session and process group with `Setsid`.
3. Detect Bash's exit without reaping it, keeping its PID and PGID reserved.
4. Send `SIGKILL` to Bash's process group to terminate the background `sleep`.
5. Reap Bash and save its exit status.
6. Reap descendants adopted through subreaping.
7. Exit with Bash's saved status.

The prototype server will inherit the supervisor's standard output and error
and report the supervisor's resulting status.

## Verification

Run from the repository root:

```shell
go run ./playground
```

Use the PID printed by `spawn-background.sh` to verify that the `sleep` no longer
exists after the prototype exits:

```shell
kill -0 <pid>
```

The command should fail, and the prototype should still exit with Bash's status.

## Out of scope

This stage does not include:

- Force-stop or graceful `SIGTERM` handling.
- Descendants that deliberately leave Bash's process group.
- Ansible workers.
- macOS support.
- Production integration.
