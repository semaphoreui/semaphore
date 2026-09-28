# runners-and-execution — zone index

Remote runners and task execution: how the server tracks runner liveness and recovers tasks from
dead runners, what the UI shows about runners and which runner ran a task, how runner credentials
are stored, how parallel tasks share a repository checkout, and how per-host / per-URL
credentials (host config) reach git, ssh and Ansible. Each doc records the status against the
code, the decisions and their reasons, and what is still open.

| When you need it | Read |
| --- | --- |
| A task hangs in `starting`/`running` after a runner died, restarted or lost its network | `runner-liveness-and-lost-tasks.md` |
| Changing `runners.offline_timeout_sec` / `task_fail_timeout_sec` / `reconcile_interval_sec` or runner selection in `RemoteJob` | `runner-liveness-and-lost-tasks.md` |
| Touching the HA orphan cleaner or `OwnedRunningRange` and the claim partition | `runner-liveness-and-lost-tasks.md` |
| Adding runner-reported metadata (version, os/arch, uptime) or a column to the runners table | `runner-fleet-visibility.md` |
| Asked why the task page shows (or does not show) which runner ran a task | `runner-fleet-visibility.md` |
| Checking a release-notes claim about runner version/platform or token hashing | `runner-fleet-visibility.md`, `runner-token-storage.md` |
| Hashing the runner auth token, changing `RunnerMiddleware` or `GetRunnerByToken`, adding token rotation | `runner-token-storage.md` |
| Reproducing or extending how the one-time registration token is hashed | `runner-token-storage.md` |
| Git corruption or races between parallel tasks of one template; touching `KeyLock` or `updateAndCheckoutRepository` | `repo-git-lock.md` |
| Adding a new git operation to task prepare that must be serialized per directory | `repo-git-lock.md` |
| Working on host config mappings, the generated `ssh_config`, `GIT_CONFIG_PARAMETERS` rewrites or the SSH agent | `host-config-credentials.md` |
| Picking up the one-agent-per-task refactor, or host config for Docker/K8s executors | `host-config-credentials.md` |
| Returning to PR #4251 or the local `sem-272-host-config-page-refactor` branch | `host-config-credentials.md` |
