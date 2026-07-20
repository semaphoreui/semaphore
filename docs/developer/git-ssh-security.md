# Git SSH host key verification

Semaphore clones and pulls repositories over SSH through an ephemeral ssh-agent
(`pkg/ssh/agent.go`). Host-key verification is controlled by the `ssh` config
section and applied via `GIT_SSH_COMMAND` on every git operation that uses an
access key with an SSH agent.

## Why this matters

Without host-key checking, a network attacker can impersonate the git server
during clone/pull and serve malicious playbooks. Semaphore sets explicit OpenSSH
options so git does not silently trust arbitrary hosts.

## Config (`ssh` key)

| Field | Env var | Default | Effect |
| --- | --- | --- | --- |
| `strict_host_key_checking` | — | `no` | OpenSSH `StrictHostKeyChecking` mode |
| `known_hosts_file` | `SEMAPHORE_SSH_KNOWN_HOSTS_FILE` | empty | Path to a `known_hosts` file |
| `config_path` | `SEMAPHORE_SSH_PATH` (nested `ssh.config_path` or legacy `ssh_config_path`) | `~/.ssh/config` | Custom SSH config (`-F` flag) |

Legacy top-level `ssh_config_path` still works; `ssh.config_path` takes
precedence when the nested `ssh` section is set (`util/config.go` —
`GetSshConfigPath`).

### `strict_host_key_checking` values

| Value | `GIT_SSH_COMMAND` behaviour |
| --- | --- |
| `no` | `StrictHostKeyChecking=no`, `UserKnownHostsFile=/dev/null` — no verification (default, backward compatible) |
| `yes` | `StrictHostKeyChecking=yes` with `known_hosts_file` — connection fails if the host key is missing or changed |
| `accept-new` | `StrictHostKeyChecking=accept-new` with `known_hosts_file` — first seen key is pinned; later changes are rejected |

For `yes` and `accept-new`, set `known_hosts_file` to a persistent path writable
by the Semaphore process (for example under `dirs.tmp_path` or a dedicated
volume). Populate it with `ssh-keyscan` for your git hosts before enabling strict
mode.

```json
{
  "ssh": {
    "strict_host_key_checking": "accept-new",
    "known_hosts_file": "/var/lib/semaphore/known_hosts"
  }
}
```

## Where it applies

Host-key options are appended in `AccessKeyInstallation.GetGitEnv()`
(`pkg/ssh/agent.go` → `gitHostKeyCheckingOpts`). They affect:

- Repository clone and update during task runs
- Playbook browse in the template editor (when using SSH keys)

The built-in `go_git` client uses `InsecureIgnoreHostKey()` for its own transport
(`db_lib/GoGitClient.go`). Deployments that require strict host keys should use
the `cmd_git` client (`git_client: cmd_git` in config) so operations go through
OpenSSH and honour `GIT_SSH_COMMAND`.

## Operational checklist

| Goal | Action |
| --- | --- |
| Harden production git access | Set `strict_host_key_checking` to `yes` or `accept-new`; pre-seed `known_hosts_file` |
| Use internal CA or jump hosts | Set `ssh.config_path` to a custom SSH config |
| Keep legacy behaviour | Leave defaults (`no`) — suitable only on trusted networks |

## Related code

- SSH agent and `GIT_SSH_COMMAND`: `pkg/ssh/agent.go`
- Config struct: `util/config.go` — `SshConfig`, `SshStrictHostKeyChecking`
- Go-git client (no host-key check): `db_lib/GoGitClient.go`

For git URL and branch validation, see [Input validation](input-validation.md).
