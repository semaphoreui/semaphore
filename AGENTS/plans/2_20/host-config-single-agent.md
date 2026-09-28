# Implementation Plan — One SSH Agent per Task for Host Config Mappings

- **Pull request:** https://github.com/semaphoreui/semaphore/pull/4251
- **Branch:** `sem-272-host-config-page`

## 1. Goal

The current implementation creates one SSH agent and one Unix socket for every SSH host mapping. This is correct, but the number of agents and sockets grows with the number of mappings and parallel tasks.

Change the implementation so that each task has:

- one isolated SSH agent;
- one Unix socket;
- all SSH credentials used by the task loaded into that agent;
- one public-key file per unique SSH credential;
- deterministic selection of the correct credential for every host or repository mapping;
- no private keys written to disk.

For example, ten parallel tasks with ten SSH mappings should use approximately ten agents and ten sockets instead of one hundred of each.

## 2. User-facing URL constraint

Do not modify user-managed files such as:

- `requirements.yml`;
- repository URLs;
- inventory files;
- submodule configuration.

Runtime aliases are allowed only in temporary Git and SSH configuration.

For example:

```text
requirements.yml:
https://github.com/company-a/role.git

temporary Git rewrite:
https://github.com/company-a/ -> git@semaphore-mapping-42:company-a/
```

The original URL must remain unchanged.

## 3. Task-scoped temporary directory

Create a unique directory for every task:

```text
/tmp/semaphore/<project>/task-<random>/
├── ssh-agent.sock
├── ssh-config.conf
└── public-keys/
    ├── key-17.pub
    ├── key-21.pub
    └── key-35.pub
```

Requirements:

- directory mode `0700`;
- no sharing between tasks;
- safe for parallel tasks of the same project;
- accessible to the user running Git and Ansible;
- removed completely during cleanup;
- generated files written atomically.

## 4. One SSH agent per task

Replace the per-mapping agent list with one task-level agent:

```go
type HostConfigInstallation struct {
    TempDir        string
    ConfigFile     string
    Agent          *Agent
    PublicKeysDir  string
    PublicKeyFiles map[int]string

    gitParams []string
}
```

`PublicKeyFiles` must be indexed by `SSHKeyID`. If multiple mappings use the same credential, they must share one agent identity and one public-key file.

Installation steps:

1. collect all SSH credentials referenced by mappings;
2. deduplicate them by `SSHKeyID`;
3. create one agent containing all unique keys;
4. start one Unix socket;
5. derive and save one public key per unique credential;
6. generate the temporary SSH configuration.

Login/password HTTPS mappings do not need an SSH agent.

## 5. Derive and save public keys

The agent already parses private keys before adding them to the keyring. Extend the same parsing pass:

1. parse the private key and passphrase;
2. add the parsed private key to the in-memory keyring;
3. call `ssh.NewPublicKey(parsedKey)`;
4. serialize it with `ssh.MarshalAuthorizedKey`;
5. keep the public key in memory;
6. write only the public key to the task directory.

Private keys must never be written to disk.

Public-key files should be written safely:

- mode `0600`;
- write to a temporary file;
- call `Sync`;
- close the file;
- atomically rename it to `<key-id>.pub`.

Although public keys are not secrets, mode `0600` avoids exposing credential relationships and keeps temporary SSH files consistently protected.

## 6. Select credentials through SSH configuration

All mappings use the same socket, but each mapping selects its own public key:

```sshconfig
Host github.com
  User git
  IdentityAgent /tmp/.../ssh-agent.sock
  IdentityFile /tmp/.../public-keys/key-17.pub
  IdentitiesOnly yes
```

For a URL mapping using an internal runtime alias:

```sshconfig
Host semaphore-mapping-42
  HostName github.com
  User git
  IdentityAgent /tmp/.../ssh-agent.sock
  IdentityFile /tmp/.../public-keys/key-21.pub
  IdentitiesOnly yes
```

`IdentitiesOnly yes` is required. It prevents SSH from offering unrelated keys from the shared agent.

The `IdentityFile` points to a public key. The private key remains inside the agent and is never read from disk.

## 7. Git URL rewrites

Keep the temporary `GIT_CONFIG_PARAMETERS` approach.

### HTTPS URLs

For:

```text
https://github.com/company-a/
```

generate a temporary rewrite similar to:

```text
url.git@semaphore-mapping-42:company-a/.insteadOf=https://github.com/company-a/
```

### SSH and scp-style URLs

Also support URLs such as:

```text
git@github.com:company-a/
```

and rewrite them temporarily to:

```text
git@semaphore-mapping-42:company-a/
```

The original files remain unchanged. Git applies the rewrite only for the current task.

Use Git's longest-prefix behavior so a repository-specific mapping overrides a broader organization or host mapping. Prefixes intended to cover a group should end with `/`.

## 8. Git environment

Git processes started by the task must receive:

```text
SSH_AUTH_SOCK=/tmp/.../ssh-agent.sock
GIT_SSH_COMMAND=ssh -F /tmp/.../ssh-config.conf
GIT_CONFIG_PARAMETERS=<temporary URL rewrites>
```

This must work for:

- the main repository;
- submodules;
- inventory repositories;
- scheduled repository polling;
- `ansible-galaxy`;
- Git commands started from playbooks;
- Terraform, OpenTofu, and Terragrunt module downloads.

Child processes must inherit the same environment.

## 9. Ansible integration

Ansible must use the same generated SSH configuration:

```text
--ssh-common-args "-F /tmp/.../ssh-config.conf"
```

For example:

```sshconfig
Host production-db
  IdentityAgent /tmp/.../ssh-agent.sock
  IdentityFile /tmp/.../public-keys/key-35.pub
  IdentitiesOnly yes
```

Verify both paths:

1. Ansible connects to inventory hosts using the correct credential.
2. Git commands started by the playbook inherit the task's socket and temporary Git rewrites.

The same environment should be added to requirements installation, Terraform initialization, and other child processes.

## 10. Cleanup and failure handling

Cleanup must be reliable and idempotent after:

- successful completion;
- cancellation;
- preparation failure;
- agent startup failure;
- SSH configuration failure;
- child-process failure.

Cleanup order:

1. stop accepting new agent connections;
2. close the agent;
3. remove the socket;
4. remove `ssh-config.conf`;
5. remove public-key files;
6. remove the task directory;
7. clear in-memory credential and path references.

Make `Agent.Close()` safe to call more than once, for example with `sync.Once`.

## 11. Docker and Kubernetes executors

The open-source repository contains provider stubs while the working Docker and Kubernetes implementations are supplied separately.

The implementation must verify that:

- mappings and decrypted credentials are passed to the individual task executor;
- every Docker container or Kubernetes Pod gets its own agent, socket, public-key files, and SSH config;
- task resources are not shared between parallel tasks;
- the task environment contains `SSH_AUTH_SOCK`;
- Git receives `GIT_SSH_COMMAND` and `GIT_CONFIG_PARAMETERS`;
- Ansible receives the generated SSH configuration;
- all temporary resources are removed when the task finishes.

Prefer creating the agent inside the task container or Pod. Do not mount a runner-host Unix socket into a Kubernetes Pod without an explicit isolation design.

## 12. Tests

### SSH agent and public keys

Test that:

- one agent accepts multiple private keys;
- public keys are derived during the same parsing pass;
- passphrase-protected keys work;
- duplicate `SSHKeyID` values produce one agent identity and one `.pub` file;
- private keys are never written to the task directory;
- `Agent.Close()` is idempotent.

### SSH configuration

Test that:

- all mappings use the same socket;
- different mappings select different public-key files;
- `IdentitiesOnly yes` is present;
- two mappings for the same real host can use different credentials;
- `ssh -G` accepts the generated configuration;
- unmapped hosts retain existing behavior.

### Git rewrites

Test:

- multiple HTTPS mappings for the same host;
- multiple SSH/scp-style mappings for the same host;
- longest-prefix matching;
- repository-specific mappings overriding organization mappings;
- unchanged `requirements.yml`;
- submodules;
- inventory repositories;
- login/password HTTPS mappings without an SSH agent.

### Ansible

Test that:

- `--ssh-common-args` points to the generated config;
- the correct key is selected for each inventory host;
- Git environment variables reach commands started from the playbook;
- login/password-only mappings do not create an SSH agent.

### Parallel execution and cleanup

Verify that ten parallel tasks:

- have ten different task directories;
- have ten different sockets;
- cannot access each other's files;
- clean up resources after success;
- clean up resources after failure.

## 13. Expected result

For one task:

```text
1 SSH agent
1 Unix socket
N public-key files
1 generated SSH config
temporary Git URL rewrites
```

For ten parallel tasks:

```text
10 agents
10 sockets
```

This reduces resource usage while preserving task isolation and deterministic credential selection.