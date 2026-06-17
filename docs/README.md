# Developer documentation

Internal guides for contributors and operators. User-facing product docs live at [docs.semaphoreui.com](https://docs.semaphoreui.com).

| Guide | Audience | Covers |
|-------|----------|--------|
| [Configuration](configuration.md) | Developers, operators | `config.json` / `config.yaml`, env vars, JSON Schema |
| [Runners and tags](runners-and-tags.md) | Developers, operators | Remote runners, tag routing, webhooks, fleet timeouts |
| [Runner executors](runner-executors.md) | Operators | Local, Docker, and Kubernetes task execution on runners |
| [Workflows](workflows.md) | Developers (Pro) | DAG workflows, approvals, artifacts, API overview |
| [Tasks API pagination](tasks-api-pagination.md) | Developers, API consumers | Keyset pagination for project task history |
| [Cluster dashboard](cluster-dashboard.md) | Operators (HA) | Admin cluster API, task state inspection, recovery |

Implementation plans for upcoming work are under [`AGENTS/plans/`](../AGENTS/plans/).
