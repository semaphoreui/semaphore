# Host config: per-host / per-URL credentials for a task, and the one-agent-per-task proposal

A repository has one key; submodules, `requirements.yml` roles, Terraform modules, inventory
repositories and inventory hosts on other servers need their own. **Host config** maps an SSH
host or a git URL prefix to a Key Store credential; the mapping is applied to every git and SSH
connection of the task through a generated `ssh_config` and temporary git URL rewrites, without
touching user files. A follow-up plan proposed collapsing the per-mapping SSH agents into one
agent per task.

Status (2026-09-27):
- Host config feature — implemented: PR #4251 (`sem-272-host-config-page`, merged into develop
  2026-09-25, `232bfb242`, unreleased); migration `v2.20.7` (`project__host_config`);
  `db/HostConfig.go`, `pkg/ssh/host_config.go`, `api/projects/host_configs.go`,
  `web/src/views/project/HostConfig.vue`; user guide `docs/docs/user-guide/host-config.md`.
- One SSH agent per task — **not implemented**: `HostConfigInstallation` holds `Agents []*Agent`,
  one agent and one socket per SSH mapping (`StartSSHAgent` per mapping).

## Decisions and why (as implemented)

- **Mappings are project-level and user files stay untouched.** URL mappings become
  `url.<replacement>.insteadOf` entries carried in `GIT_CONFIG_PARAMETERS` (not a command line,
  so credentials stay out of the process list; git logs only the pre-rewrite URL). Host mappings
  become `Host` blocks in a generated config passed as `GIT_SSH_COMMAND=ssh -F …` and as
  Ansible `--ssh-common-args`. Git's longest-prefix match makes a repo mapping override an org
  mapping for free.
- **Each URL mapping gets its own alias `semaphore-mapping-<id>`** so several mappings on one
  real host can carry different credentials.
- **The generated config `Include`s the administrator's `ssh.config_path` after a `Match all`**
  so a mapping wins over the admin's block for the same host, and everything else keeps the
  admin config.
- **One agent per mapping, keys never on disk.** Code comment in `InstallHostConfigs`: a single
  agent holding every key would let ssh offer the wrong one first and sshd closes the connection
  after `MaxAuthTries`. Private keys live only in the in-memory keyring; sockets are `0600` and
  chowned to the process user.
- **Strict allowlists on everything that lands in config text**: host regex, URL must be
  http/https without userinfo and without `=`/quotes/whitespace (git splits an entry at the first
  `=`), ssh login regex (a newline would inject a `ProxyCommand`). Login/password mappings only
  over https and need no agent — the credential goes into the rewritten URL with `=` escaped.
- **Loaded wherever git runs for the project**: task prepare (`LocalExecutor.installHostConfigs`,
  torn down in `destroyKeys`), inventory repo, `ansible-galaxy`, Terraform init, schedule polling
  (`SchedulePool`), branch listing (`api/projects/repository.go`) via
  `db_lib.InstallProjectHostConfigs`. Remote runners receive the mappings with decrypted keys in
  the poll payload (`api/runners/runners.go`, decrypted at dispatch because the key may have been
  rotated while the task waited).

## The one-agent-per-task proposal (open problem)

Motivation: N mappings × M parallel tasks = N×M agents and sockets; ten tasks with ten mappings
should need ten agents, not a hundred. Agreed approach: one task-scoped `0700` directory with one
socket, one config and a `public-keys/<key-id>.pub` per unique `SSHKeyID`; every mapping block
gets `IdentityAgent` + `IdentityFile <pub>` + `IdentitiesOnly yes`, which is what answers the
"wrong key first / MaxAuthTries" objection in the current code comment — ssh offers only the
listed public key and the private key still never leaves the agent. Public keys are derived in
the same parse pass the agent already does (`ssh.NewPublicKey` → `MarshalAuthorizedKey`), written
atomically (temp file, sync, rename, `0600`). Cleanup must be idempotent after success, cancel,
prepare failure or agent failure, in the order: stop accepting, close agent, remove socket,
config, pub files, directory.

Constraints and open questions:
- Keep `GIT_CONFIG_PARAMETERS` rewrites; the plan also wants scp-style mappings
  (`git@github.com:org/`) — today `validateURL` accepts only http/https, so that is a validation
  change with its own injection review.
- Docker/K8s executors: the plan requires an agent *inside* the container/Pod, never a mounted
  runner-host socket. Today `pro_impl/services/tasks/docker/provider.go` and `k8s/provider.go`
  discard the `hostConfigs` argument (`_ []db.HostConfig`) — host config does not work there at
  all yet, which is a bigger gap than the agent count.
- `Agent.Close` is not idempotent (`close(a.done)` twice panics); `Destroy` nils the slice so
  it is safe at installation level only. The plan asks for `sync.Once`.
- Temp layout today is `project_<id>/ssh-config-<rand>.conf` + `ssh-agent-<key>-<rand>.sock`
  in the project tmp dir, not a per-task directory.

Not implemented; to be created as a workbench task when picked up.

## Open, broken or deferred (feature as shipped)

- ⚠️ 2026-09-27: Docker and Kubernetes executors ignore host configs (see above).
- URL-mapping aliases always use ssh port 22 (`Hostname()` drops the http port); a mapping would
  need its own ssh port field for non-standard servers.
- Agents are started for every mapping whether or not the task reaches that host (lazy start
  noted as a future option).
- Local side branch `sem-272-host-config-page-refactor` (3 commits ahead of develop, unmerged):
  wraps access-key/host-config writes in DB transactions (`db/sql/transaction.go`, `db/Store.go`),
  no feature change.
- `pro_impl` was adapted for the new `ExecutorProvider` signature on `main`
  (`fix/sem-272-incompatible-interfaces`, PR #13 there); the `pro_impl` branch
  `sem-272-host-config-page` also exists locally.

## Key files and references

- `db/HostConfig.go` (model, validation, `SSHAlias`), `db/sql/host_config.go`,
  `db/sql/migrations/v2.20.7.sql`.
- `pkg/ssh/host_config.go` (`InstallHostConfigs`, blocks, rewrites, `Destroy`),
  `pkg/ssh/agent.go` (`Agent`, `StartSSHAgent`, `GetGitEnvWithHostConfigs`).
- `db_lib/HostConfigLoader.go`, `db_lib/GitRepository.go` (`HostConfigs` field).
- `services/tasks/local_executor.go` (`installHostConfigs`, `hostConfigEnv`, ansible args),
  `services/tasks/local_executor_inventory.go` (`destroyKeys`), `services/tasks/TaskRunner.go`
  (`loadHostConfigs`), `services/tasks/executor.go` (`ExecutorProvider.NewExecutor`).
- `api/projects/host_configs.go`, `api/runners/runners.go` (payload to runner),
  `web/src/views/project/HostConfig.vue`, `web/src/components/HostConfigForm.vue`.
- Docs: `docs/docs/user-guide/host-config.md`.
- Sources: `AGENTS/plans/2_20/host-config-single-agent.md`.
  PR https://github.com/semaphoreui/semaphore/pull/4251; branches `sem-272-host-config-page`,
  `sem-272-host-config-page-refactor` (local).
