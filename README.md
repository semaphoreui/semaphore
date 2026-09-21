<div align="center">

<img src="https://user-images.githubusercontent.com/914224/134777345-8789d9e4-ff0d-439c-b80e-ddc56b74fcee.png" width="600" alt="UniFlow UI" />

# UniFlow

**Built on [Semaphore UI](https://github.com/semaphoreui/semaphore) — adding OCI-based Execution Environments, stable multi-stage Workflows, plugins, and more community-driven features for teams automating infrastructure at scale.**

[![Dev](https://github.com/TKILLAHOME/uniflow/actions/workflows/dev.yml/badge.svg)](https://github.com/TKILLAHOME/uniflow/actions/workflows/dev.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Fork of Semaphore](https://img.shields.io/badge/fork%20of-semaphoreui%2Fsemaphore-blue)](https://github.com/semaphoreui/semaphore)

</div>

---

## What is UniFlow?

UniFlow is a community fork of the excellent [Semaphore UI](https://github.com/semaphoreui/semaphore) project, extending it with features needed by infrastructure teams running automation at scale:

- **Execution Environments** — run jobs inside OCI container images pulled from any registry. Pin your tool versions, isolate dependencies, and get reproducible runs every time.
- **Stable Workflows** — chain tasks into multi-stage pipelines with conditions and parallel execution.
- **Plugin system** — extend UniFlow with custom runner backends, inventory sources, and notifiers without touching core code.
- **Security-first** — secrets are never written to logs, containers run non-root, credentials are injected at runtime.

UniFlow supports everything Semaphore UI supports: Ansible, Terraform/OpenTofu, PowerShell, Bash, and more.

> **Credit where it's due:** UniFlow would not exist without the incredible work of the [Semaphore UI team](https://github.com/semaphoreui/semaphore). All upstream improvements continue to flow into UniFlow automatically. If a feature works for you and belongs upstream, we'll send it back as a PR.

---

## How Execution Environments work

An Execution Environment (EE) is a container image that carries everything a job needs — a specific Ansible version, Python packages, collections, roles. Instead of installing dependencies on the host, the job runs entirely inside the container.

```
Registry (Gitea, Harbor, Docker Hub, …)
    └── EE Image
            ├── Ansible pinned version
            ├── Python venv + collections
            └── System packages
                    └── UniFlow starts container
                            ├── Mounts playbook + inventory (read-only)
                            ├── Injects credentials as env vars
                            └── Streams stdout/stderr live to UI
```

---

## Quick Start

> UniFlow is in early development. The base Semaphore UI runs fully; EE and Workflow features are being built.

### Docker (Semaphore-compatible)

```bash
docker run -p 3000:3000 --name uniflow \
  -e SEMAPHORE_DB_DIALECT=sqlite \
  -e SEMAPHORE_ADMIN=admin \
  -e SEMAPHORE_ADMIN_PASSWORD=changeme \
  -e SEMAPHORE_ADMIN_NAME=Admin \
  -e SEMAPHORE_ADMIN_EMAIL=admin@localhost \
  -d ghcr.io/tkillahome/uniflow:latest
```

### Build from source

```bash
git clone https://github.com/TKILLAHOME/uniflow
cd uniflow
task build
```

Requirements: Go 1.21+, Node.js 18+, [Task](https://taskfile.dev)

---

## Roadmap

| Phase | Target | Status |
|-------|--------|--------|
| **Phase 1** — Foundation | Branch strategy, dev setup, codebase study | 🔄 In progress |
| **Phase 2** — Execution Environments | Docker/Podman runner, registry client, EE UI | 📋 Planned |
| **Phase 3** — Stable Workflows | Workflow stabilization, Podman backend, v0.1.0 release | 📋 Planned |
| **Phase 4** — Community & Extensions | NetBox inventory, ansible-builder UI, plugin API | 📋 Planned |

---

## Branch Strategy

| Branch | Purpose |
|--------|---------|
| `develop` | Main integration branch — tracks upstream Semaphore |
| `v0.x` | Version release branches |
| `feat/*` | Short-lived feature branches |
| `fix/*` | Hotfixes, cherry-pickable into version branches |

Upstream sync:
```bash
git fetch upstream
git merge upstream/develop   # clean merge — UniFlow code lives in separate packages
```

---

## Contributing

UniFlow is a community project and contributions are very welcome.

- [Contributing Guide](CONTRIBUTING.md)
- [Code of Conduct](CODE_OF_CONDUCT.md)
- Open an issue to discuss features or bugs
- PRs against `develop` please

If your contribution is general enough to benefit Semaphore UI users too, please also consider opening a PR upstream.

---

## Upstream & Compatibility

UniFlow is built **additively** on Semaphore UI:

- All UniFlow-specific code lives in new packages (`services/runner/`, `pkg/registry/`, `api/ee.go`, `web/src/ee/`)
- Existing Semaphore files are touched as little as possible
- Upstream updates merge cleanly via `git merge upstream/develop`
- The full Semaphore feature set remains intact

UniFlow will always be fully compatible with Semaphore UI's configuration format and API.

---

## License

MIT © UniFlow Contributors

Based on [Semaphore UI](https://github.com/semaphoreui/semaphore) — MIT © [Denis Gukov](https://github.com/fiftin)
