# Stage 1: Current process behavior

## Purpose

Reproduce Semaphore's current local Bash execution behavior before adding a task
supervisor. This prototype is standalone and does not change `LocalExecutor`.

## Initial implementation

```text
playground/
├── main.go
└── scripts/
    └── spawn-background.sh
```

The Go program runs the Bash script directly with `os/exec` and reports Bash's
exit status. The script starts a 60-second background `sleep`, prints its PID and
a manual `kill` command, then exits without waiting.

There is intentionally no supervisor or descendant cleanup.

## Run

From the repository root:

```shell
go run ./playground
```

Expected process tree:

```text
prototype
└── bash
    └── sleep
```

Bash exits while `sleep` continues running.

## Next iteration

Add a short-lived `task_supervisor` between the prototype and Bash. Keep the
prototype work in separate commits so it can be dropped before merging the
production implementation.
