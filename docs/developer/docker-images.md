# Docker images

Semaphore publishes several container images from this repository. They serve
different roles in deployment and in the Docker/Kubernetes runner executors.

## Image overview

| Image | Dockerfile | Purpose |
| --- | --- | --- |
| `semaphoreui/semaphore` | `deployment/docker/server/Dockerfile` | Semaphore **server** (API + web UI) |
| `semaphoreui/runner` | `deployment/docker/runner/Dockerfile` | Semaphore **runner** process + toolchain |
| `semaphoreui/job` | `deployment/docker/job/Dockerfile` | **Task build** container for Docker/K8s executors |
| `semaphoreui/helper` | `deployment/docker/helper/Dockerfile` | **Git-clone init** container (K8s executor) |

`server` and `runner` images compile the Go binary via `task build` and bundle
Ansible, Terraform, OpenTofu, and Terragrunt for local-style execution on the
container host.

`job` and `helper` are slim Debian-based images without the Semaphore binary.
They are used exclusively by the Pro Docker and Kubernetes executors when a
runner dispatches work into isolated containers/Pods.

## Job image (`semaphoreui/job`)

Built from `deployment/docker/job/Dockerfile` (Debian 13 slim).

**Includes:** Ansible (pinned `ansible-core` + `ansible` package), OpenTofu,
Terraform, Terragrunt, Git, OpenSSH client, Python 3, `curl`, `wget`, `unzip`.

**Runs as:** `nobody` user, working directory `/workspace`.

**Used for:**

- Docker executor `executor.docker.image` (default `semaphoreui/job:latest`)
- Docker executor `executor.docker.helper_image` (git clone before the build step)
- K8s executor `executor.k8s.image` when you want the full toolchain in the Pod

CI publishes this image on every `develop` build (`deploy-job` job in
`.github/workflows/dev.yml`).

## Helper image (`semaphoreui/helper`)

Built from `deployment/docker/helper/Dockerfile` (Debian 13 slim).

**Includes:** Git, OpenSSH client, Python 3, `curl`, `sshpass` — no Ansible or
IaC binaries.

**Runs as:** `nobody` user, working directory `/workspace`.

**Used for:**

- K8s executor `executor.k8s.helper_image` — init container that clones the
  task repository before the main build container starts
- Lighter alternative when only repository checkout is needed

CI publishes this image on every `develop` build (`deploy-helper` job in
`.github/workflows/dev.yml`).

## Server and runner images

See [deployment/docker/README.md](../../deployment/docker/README.md) for build
and test tasks (`task docker:build`, `task docker:test`).

Environment variables for customising image names during local builds:

- `DOCKER_ORG` (default `semaphoreui`)
- `DOCKER_SERVER` (default `semaphore`)
- `DOCKER_RUNNER` (default `runner`)
- `DOCKER_CMD` (default `docker`)

## Choosing images for executors

| Executor | Build container (`image`) | Init / clone (`helper_image`) |
| --- | --- | --- |
| Docker | `semaphoreui/job:latest` | `semaphoreui/job:latest` (same image) |
| Kubernetes | `semaphoreui/job:latest` (recommended) | `semaphoreui/helper:latest` (recommended) |

Override paths via `runner.executor.docker.*` or `runner.executor.k8s.*` in the
runner config, or the corresponding `SEMAPHORE_RUNNER_DOCKER_*` /
`SEMAPHORE_RUNNER_K8S_*` environment variables.

Per-template overrides (`executor_image` on the template) take precedence over
the runner default for Docker and Kubernetes executors. See
[runner-executors.md](runner-executors.md#per-template-image-override).

## Related code

- Runner defaults: `util/config.go` (`RunnerDockerConfig`, `RunnerK8sConfig`)
- Executor factory: `services/runners/executor_factory.go`
- CI publish: `.github/workflows/dev.yml` (`deploy-job`, `deploy-helper`)
