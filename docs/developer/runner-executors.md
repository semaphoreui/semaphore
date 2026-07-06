# Runner executors

## Purpose

A **runner** pulls tasks from the Semaphore server and executes them. The
`runner.executor` block in the runner config (`util.RunnerConfig`, schema key
`runner` in `config.schema.yaml`) selects **how** each task is isolated:

| `executor.type` | Behaviour | Build |
| --- | --- | --- |
| `local` (default) | Subprocesses on the runner host; keys and repos materialised locally | Open source |
| `docker` | Ephemeral container per task against a Docker daemon | Pro (`pro_impl`) |
| `k8s` | Ephemeral Pod per task in a Kubernetes namespace | Pro (`pro_impl`) |

The factory in `services/runners/executor_factory.go` picks a
`tasks.ExecutorProvider` at runner startup. Access-key hydration is shared across
all strategies so Ansible vault passwords, SSH keys, and inventory repo keys
behave the same regardless of executor type.

In the open-source tree, `pro/services/tasks/docker` and
`pro/services/tasks/k8s` are stubs that return an error at startup — the runner
logs a clear message and refuses to start when `type` is `docker` or `k8s`
without the proprietary `pro_impl` module.

## Configuration

Runner-process settings live under the top-level `runner` key (env prefix
`SEMAPHORE_RUNNER_*`). Server-side fleet timeouts (`offline_timeout_sec`, etc.)
use the separate `runners` key (`SEMAPHORE_RUNNERS_*`) — do not confuse the two.

### Local executor (default)

No extra block is required. Example runner snippet:

```json
{
  "runner": {
    "enabled": true,
    "registration_token_file": "/etc/semaphore/runner-token",
    "executor": {
      "type": "local"
    }
  }
}
```

### Docker executor (Pro)

Each task runs in a short-lived container. Defaults (`util.RunnerDockerConfig`):

| Field | Default | Env var |
| --- | --- | --- |
| `image` | `semaphoreui/job:latest` | `SEMAPHORE_RUNNER_DOCKER_IMAGE` |
| `helper_image` | `semaphoreui/job:latest` | `SEMAPHORE_RUNNER_DOCKER_HELPER_IMAGE` |
| `network` | `bridge` | `SEMAPHORE_RUNNER_DOCKER_NETWORK` |
| `pull_policy` | `if-not-present` | `SEMAPHORE_RUNNER_DOCKER_PULL_POLICY` |
| `poll_interval_seconds` | `2` | `SEMAPHORE_RUNNER_DOCKER_POLL_INTERVAL_SECONDS` |
| `cleanup_grace_seconds` | `30` | `SEMAPHORE_RUNNER_DOCKER_CLEANUP_GRACE_SECONDS` |

```json
{
  "runner": {
    "enabled": true,
    "executor": {
      "type": "docker",
      "docker": {
        "host": "unix:///var/run/docker.sock",
        "image": "semaphoreui/job:latest",
        "helper_image": "semaphoreui/job:latest",
        "memory_limit": "4g",
        "cpu_limit": 2
      }
    }
  }
}
```

`host` accepts `unix://`, `tcp://`, or `npipe://` URLs. When empty, the standard
`DOCKER_HOST` environment and platform default socket are used. Set `privileged`
only when a template truly requires it — it disables container isolation.

See [Docker images](docker-images.md) for what the `job` image contains.

### Kubernetes executor (Pro)

Each task runs in an ephemeral Pod. Defaults (`util.RunnerK8sConfig`):

| Field | Default | Env var |
| --- | --- | --- |
| `namespace` | `semaphore` | `SEMAPHORE_RUNNER_K8S_NAMESPACE` |
| `image` | `alpine:latest` | `SEMAPHORE_RUNNER_K8S_IMAGE` |
| `helper_image` | `alpine/git:latest` | `SEMAPHORE_RUNNER_K8S_HELPER_IMAGE` |
| `service_account` | `default` | `SEMAPHORE_RUNNER_K8S_SERVICE_ACCOUNT` |
| `poll_interval_seconds` | `3` | `SEMAPHORE_RUNNER_K8S_POLL_INTERVAL_SECONDS` |
| `cleanup_grace_seconds` | `30` | `SEMAPHORE_RUNNER_K8S_CLEANUP_GRACE_SECONDS` |

```json
{
  "runner": {
    "enabled": true,
    "executor": {
      "type": "k8s",
      "k8s": {
        "namespace": "semaphore",
        "image": "semaphoreui/job:latest",
        "helper_image": "semaphoreui/helper:latest",
        "service_account": "semaphore-runner"
      }
    }
  }
}
```

When `kubeconfig` is empty the runner uses in-cluster credentials (ServiceAccount
token mounted by Kubernetes). `pull_secrets` is a comma-separated list of
`imagePullSecret` names attached to each Pod.

For K8s deployments, use `semaphoreui/job` for the build container and
`semaphoreui/helper` for the git-clone init container — the helper image is
minimal (git + SSH client only).

## Connection TLS

`runner.connection` controls how the runner reaches the server:

- `server_ca_cert_file` — PEM bundle for private CAs (in addition to the system trust store).
- `skip_tls_verify` — disables verification (testing only; vulnerable to MITM).

## Related code

- Config structs: `util/config.go` (`RunnerConfig`, `ExecutorConfig`, `RunnerDockerConfig`, `RunnerK8sConfig`)
- Schema: `config.schema.yaml` (`RunnerConfig`, `ExecutorConfig`, …)
- Factory: `services/runners/executor_factory.go`, `services/runners/job_pool.go`
- Local provider: `services/tasks/local_executor_provider.go`
- Pro stubs: `pro/services/tasks/docker/config.go`, `pro/services/tasks/k8s/config.go`

For runner authentication, registration, and online/offline status, see
[Runner auth and task JWT](runner-auth-and-jwt.md). Compose examples:
[deployment/compose/README.md](../../deployment/compose/README.md#registration-modes).
