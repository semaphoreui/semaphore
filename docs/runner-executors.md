# Runner executors

A **runner** process can execute each assigned task using one of three executor strategies. The strategy is selected in the runner config block (`runner.executor.type`).

| Type | Config value | Availability | Behavior |
|------|--------------|--------------|----------|
| Local | `local` (default) | Open source | Runs the task as a subprocess on the runner host |
| Docker | `docker` | Pro build | Runs each task in an ephemeral container |
| Kubernetes | `k8s` | Pro build | Runs each task in an ephemeral Pod |

The open-source build ships stubs for Docker and Kubernetes executors. If you set `type: docker` or `type: k8s` without the Pro module, the runner logs an initialization error and refuses jobs until restarted with a valid config.

## Configuration

Executor settings live under `runner.executor` in the runner's config file (or the `runner` section of a shared config used by `semaphore runner`):

```yaml
runner:
  enabled: true
  token: "<runner-auth-token>"
  executor:
    type: local   # local | docker | k8s
```

Environment variables follow the `SEMAPHORE_RUNNER_*` prefix. Nested executor fields use `SEMAPHORE_RUNNER_DOCKER_*` or `SEMAPHORE_RUNNER_K8S_*`.

## Local executor

The default. The runner clones the repository, prepares inventory and secrets, and invokes Ansible/Terraform/etc. directly on the runner host. No extra infrastructure is required.

Use when the runner VM or container already has the required toolchains installed (or when using the `semaphoreui/job` image as the runner base).

## Docker executor

Each task runs in a short-lived container against a local or remote Docker daemon. Field shapes mirror the GitLab Docker executor for familiarity.

```yaml
runner:
  executor:
    type: docker
    docker:
      host: unix:///var/run/docker.sock
      image: semaphoreui/job:latest
      helper_image: semaphoreui/job:latest
      network: bridge
      pull_policy: if-not-present
      cpu_limit: 2
      memory_limit: 4g
      poll_interval_seconds: 2
      cleanup_grace_seconds: 30
      privileged: false
```

| Field | Default | Purpose |
|-------|---------|---------|
| `host` | platform default / `DOCKER_HOST` | Daemon URL (`unix://`, `tcp://`, `npipe://`) |
| `image` | `semaphoreui/job:latest` | Build container image |
| `helper_image` | `semaphoreui/job:latest` | Git-clone helper container |
| `network` | `bridge` | Docker network for the build container |
| `pull_policy` | `if-not-present` | `always`, `if-not-present`, or `never` |
| `cpu_limit` | none | CPU cap (`--cpus`) when > 0 |
| `memory_limit` | none | Memory cap (e.g. `2g`) |
| `privileged` | `false` | Run with `--privileged` (dangerous) |

## Kubernetes executor

Each task runs in an ephemeral Pod. Field shapes mirror the GitLab Kubernetes executor.

```yaml
runner:
  executor:
    type: k8s
    k8s:
      kubeconfig: /path/to/kubeconfig   # omit for in-cluster config
      namespace: semaphore
      image: alpine:latest
      helper_image: alpine/git:latest
      service_account: default
      pull_secrets: regcred
      poll_interval_seconds: 3
      cleanup_grace_seconds: 30
```

| Field | Default | Purpose |
|-------|---------|---------|
| `kubeconfig` | in-cluster | Path to kubeconfig file |
| `namespace` | `semaphore` | Namespace for task Pods |
| `image` | `alpine:latest` | Default build container image |
| `helper_image` | `alpine/git:latest` | Git-clone init container image |
| `service_account` | `default` | Pod service account |
| `pull_secrets` | none | Comma-separated `imagePullSecrets` |
| `poll_interval_seconds` | 3 | Pod status poll interval |
| `cleanup_grace_seconds` | 30 | Pod deletion grace period |

### Kubernetes prerequisites

1. A namespace (create if it does not exist).
2. RBAC allowing the runner service account to create/delete Pods in that namespace.
3. Network access from Pods to the Semaphore server (for log streaming and status).
4. Container images that include the toolchains your templates need (or use `semaphoreui/job`).

When `kubeconfig` is empty, the executor uses in-cluster configuration (ServiceAccount token and CA mounted by Kubernetes).

## How executors plug in

At runner startup, `services/runners/executor_factory.go` reads `runner.executor.type` and constructs an `ExecutorProvider`. The job pool uses this provider for every assigned task; switching executor type does not require changes to task routing or the Semaphore server.

```
JobPool → newExecutorProvider(config) → ExecutorProvider
       → newExecutor(job, accessKeys, provider) → Executor (runs task)
```

## Troubleshooting

| Symptom | Check |
|---------|--------|
| Runner starts but rejects all jobs | Logs for `failed to initialise executor provider`; verify `type` and Pro build for docker/k8s |
| Docker: cannot connect to daemon | `host` / `DOCKER_HOST`, socket permissions, TLS certs in `cert_path` |
| K8s: Pod stuck pending | Namespace, RBAC, image pull secrets, resource quotas |
| K8s: in-cluster auth fails | ServiceAccount, automountServiceAccountToken, cluster DNS |

## Related code

- `util/config.go` — `ExecutorConfig`, `RunnerK8sConfig`, `RunnerDockerConfig`
- `services/runners/executor_factory.go` — provider selection
- `pro/services/tasks/k8s/` — K8s provider (Pro)
- `pro/services/tasks/docker/` — Docker provider (Pro)
