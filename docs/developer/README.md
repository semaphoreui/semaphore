# Developer documentation

In-repo notes for contributors and operators who work directly with the Semaphore
source tree. For end-user guides see [docs.semaphoreui.com](https://docs.semaphoreui.com).

| Topic | Covers |
| --- | --- |
| [Runner auth and task JWT](runner-auth-and-jwt.md) | Runner bearer tokens, registration, online/offline status, fleet timeouts, per-task JWT issuance |
| [Runner executors](runner-executors.md) | `local`, `docker`, and `k8s` executor strategies, runner config, OSS vs Pro |
| [Docker images](docker-images.md) | `server`, `runner`, `job`, and `helper` images — purpose, contents, publishing |
| [Workflows](workflows.md) | Pro workflow templates, runs, approvals, per-node task params, artifacts, API surface |
| [Secret storages](secret-storages.md) | Vault, OpenBao, Enterprise backends, sync, readonly mode, API routes |
| [Parallel tasks git lock](parallel-tasks-git-lock.md) | `KeyLock` serialization for shared repository directories |
| [Survey and variable types](survey-and-variable-types.md) | Survey `int`/`enum` vars, typed variable-group editor, `TaskParams` |
| [Authentication security](auth-security.md) | Session cookies, CSRF middleware, password-change verification |
| [HA cluster dashboard](ha-cluster-dashboard.md) | Admin cluster routes, task-state snapshot/clear, runner name in task payloads |
| [Input validation](input-validation.md) | Playbook path checks, git URL injection prevention, branch override gate, custom roles, access key updates |

When you change a public API or operational workflow, update the matching page here
and keep `api-docs.yml` in sync (see [CONTRIBUTING.md](../../CONTRIBUTING.md)).
