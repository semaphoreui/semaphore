---
form: src/components/RunnerForm.vue
---

# Edit Runner

A runner is a separate `semaphore runner` process that runs tasks on its own machine instead of on
the Semaphore server. A task goes to a runner when the server is set to run tasks on runners, or
when the task template or its inventory has a runner tag. The server then picks a runner that is
enabled, registered, online and not at its limit of parallel tasks: runners of the project first,
then global runners, in random order within each group. When every matching runner is offline or
busy, the task waits in the queue; when no runner matches at all, the task fails.

## Name {#runner-name}

Required. The name the runner is listed under on the Runners page, for example the host it runs
on. It does not change which tasks the runner gets.

## Tags {#runner-tags}

Available in Semaphore Pro; in other editions the field is locked.

The tags that pick the tasks of this runner. A task with a runner tag goes only to the runners
that carry that tag; the tag of the task template wins over the tag of the inventory. Type a tag
and press Enter, or pick one used by other runners. Blank and repeated tags are dropped on save.

A project runner needs at least one tag unless **Is default** is ticked.

## Is default {#runner-isDefault}

A task without a runner tag goes only to the runners marked as default. Clear the box for a
runner that should take only tasks with one of its tags.

The box is ticked for a new global runner and cleared for a new project runner.

## Register {#runner-registered}

Shown when you create a runner.

- Ticked (the default): the server creates the runner with its token. After **Create** the
  dialog shows the token and the commands that start the runner with it.
- Cleared: the runner is created without any token. Later, click **Regenerate registration
  token** (the icon of two arrows) in the row of the runner to get a one-time registration
  token, valid for one hour, and register the runner with `semaphore runner register`.

An unregistered runner gets no tasks.

## Webhook {#runner-webhook}

Optional. A URL the server calls when it gives a task to this runner, for example to start the
runner on demand. The server sends a `POST` request with a JSON body:

```json
{"action": "start", "project_id": 1, "task_id": 42, "template_id": 7, "runner_id": 3}
```

The URL must answer with the status 200 or 204 within 30 seconds; otherwise the task fails. A
runner with a webhook is always treated as online, even when it has not polled the server
recently. Leave the field empty for a runner that runs all the time.

## Max number of parallel tasks {#runner-maxParallelTasks}

The number of tasks the runner may run at once. When the limit is reached, the next task goes to
another matching runner or waits in the queue. Leave the field empty or set `0` for no limit.

## Enabled {#runner-active}

Only an enabled runner gets new tasks. Clear the box to take the runner out of service without
deleting it; the switch in the list of runners does the same.

The box is cleared for a new runner: tick it, or enable the runner later in the list.
