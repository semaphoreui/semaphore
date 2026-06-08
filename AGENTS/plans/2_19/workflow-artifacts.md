# Workflow artifacts (set_stats parity)

Status: implemented (initial cut, runner-local execution).

## Why

Until now Semaphore could not pass arbitrary key/value pairs from one task
template to another inside the same workflow. The only value that ever
flowed across nodes was the `version` of an upstream Build template (exposed
to a Deploy template as `incoming_version` / `SEMAPHORE_TASK_INCOMING_VERSION`).

AWX solves this with `set_stats` and "workflow artifacts": a producer task
calls `ansible.builtin.set_stats: data: {...}` and the AWX controller merges
the resulting dictionary into the `extra_vars` of every downstream node in
the same workflow run. Semaphore now does the same.

## How it works

1. Every task that runs locally is given an environment variable
   `SEMAPHORE_ARTIFACTS_FILE` pointing at a per-task JSON file path.
2. The task may write a JSON object to that file. Two supported recipes:
   - **Ansible**: just use `ansible.builtin.set_stats:`. Semaphore registers
     a built-in aggregate callback plugin (`semaphore_artifacts`) that
     captures the play-level stats and writes the file for you.
   - **Anything else** (Bash, Python, Terraform, ...): write the JSON object
     yourself, e.g. `echo '{"deployed_version":"1.2.3"}' > "$SEMAPHORE_ARTIFACTS_FILE"`.
3. After the task exits, Semaphore reads the file, validates it (must be a
   JSON object with simple keys, `^[A-Za-z_][A-Za-z0-9_]*$`, max 256 KB),
   and persists it on the task row (`task.artifacts`).
4. When a downstream node in the same `WorkflowRun` starts, Semaphore merges
   artifacts from every finished upstream task and:
   - injects them at the top level of Ansible `extra_vars` (so they are
     available as plain Jinja vars, just like in AWX),
   - exposes the same map under the namespaced key
     `semaphore_workflow_artifacts` for callers that prefer explicitness,
   - exports each scalar as `SEMAPHORE_WF_<UPPER_KEY>` for shell/Terraform
     templates.

Later tasks override earlier ones, mirroring AWX's merge semantics.

## Quick recipes

### Ansible producer

```yaml
- hosts: all
  tasks:
    - name: Capture deployed version for downstream templates
      ansible.builtin.set_stats:
        data:
          deployed_version: "1.2.3"
          ready: true
```

### Ansible consumer

```yaml
- hosts: all
  tasks:
    - name: Use upstream value
      debug:
        msg: "Promoting {{ deployed_version }}"
```

### Shell / Bash producer

```bash
cat > "$SEMAPHORE_ARTIFACTS_FILE" <<'JSON'
{ "deployed_version": "1.2.3" }
JSON
```

### Shell / Bash consumer

```bash
echo "Promoting ${SEMAPHORE_WF_DEPLOYED_VERSION}"
```

## Reserved keys

The following keys cannot be used as artifact names; they will be rejected
at validation time to prevent collisions with Semaphore-injected variables:

- `semaphore_vars`
- `semaphore_workflow_artifacts`
- `task_details`
- `incoming_version`

## Limits

- Maximum payload size: 256 KB.
- Top-level value must be a JSON object.
- Keys must match `^[A-Za-z_][A-Za-z0-9_]*$`.
- Only scalar values are projected as shell environment variables; arrays
  and nested objects remain available via Ansible extra vars.

## API

- `task.artifacts` is now part of every task JSON response.
- `GET /api/project/{project_id}/workflows/{workflow_id}/runs/{run_id}/artifacts`
  returns the merged artifact map for a run (useful for UIs and debugging).

## Out of scope (follow-ups)

- Remote runners (`useRemoteRunner: true`) do not yet stream artifacts back
  to the server; this requires extending the runner→server protocol.
- UI panels for displaying artifacts on the task and run views.
- Automatic capture of `terraform output -json` as a sub-key.
