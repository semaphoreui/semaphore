# Developer documentation

In-repo notes for contributors and operators who work directly with the Semaphore
source tree. For end-user guides see [docs.semaphoreui.com](https://docs.semaphoreui.com).

| Topic | Covers |
| --- | --- |
| [Runner executors](runner-executors.md) | `local`, `docker`, and `k8s` executor strategies, runner config, OSS vs Pro |
| [Docker images](docker-images.md) | `server`, `runner`, `job`, and `helper` images — purpose, contents, publishing |
| [Workflows](workflows.md) | Pro workflow templates, runs, approvals, artifacts, API surface |
| [HA cluster dashboard](ha-cluster-dashboard.md) | Admin cluster routes, task-state snapshot/clear, runner name in task payloads |

When you change a public API or operational workflow, update the matching page here
and keep `api-docs.yml` in sync (see [CONTRIBUTING.md](../../CONTRIBUTING.md)).
