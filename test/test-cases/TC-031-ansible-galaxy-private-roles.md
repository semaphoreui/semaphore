# TC-031 — Ansible Galaxy install with private git roles (multi-key)

| Field        | Value                            |
|--------------|----------------------------------|
| Area         | Ansible / Galaxy                 |
| Priority     | High                             |
| Type         | Functional / Integration         |
| Automatable  | Partial (manual git setup)       |
| Linear       | QA-1                             |

## Objective

Verify that `ansible-galaxy install -r requirements.yml` resolves git-based
roles and collections from **multiple private repositories**, each protected by
a **different deploy key**, and that Semaphore uses the correct SSH credentials
for each source during the Galaxy install step of a task.

## Background (from QA-1 thread)

Semaphore stores SSH keys in the database and loads them into an in-memory SSH
agent — keys are **not** written to the filesystem. For repository clone/pull
this works because each repo has one bound access key. Galaxy is harder: a single
`requirements.yml` may reference several private git hosts, each needing its own
key.

The proposed Semaphore-side approach (under development) is to generate
per-task SSH host aliases and inject `GIT_SSH_COMMAND` / `SSH_AUTH_SOCK` before
invoking `ansible-galaxy`, reusing the same mechanism for `ansible-playbook`
git-based content. The open design question is **how Semaphore learns the
host → key mapping** (user-supplied configuration vs. inferred from Key Store).

## Preconditions

* Semaphore instance with Ansible runner (local or remote).
* Ability to create **four private GitHub repositories** (or GitLab equivalents).
* Shell access on the runner host for the baseline (Phase A) test.
* Project `Infra QA` (or equivalent) with Key Store access.

## Test data — private role repositories

Create four minimal Ansible roles, one per private repo. Each repo gets its own
**read-only deploy key** (unique key pair; do not reuse keys across repos).

| Repo alias   | Example git URL                                      | Role name in Galaxy | Deploy key name (Key Store) |
|--------------|------------------------------------------------------|---------------------|-----------------------------|
| `role-alpha` | `git@github.com:<org>/ansible-role-alpha.git`        | `alpha`             | `galaxy-alpha-key`          |
| `role-beta`  | `git@github.com:<org>/ansible-role-beta.git`         | `beta`              | `galaxy-beta-key`             |
| `role-gamma` | `git@github.com:<org>/ansible-role-gamma.git`        | `gamma`             | `galaxy-gamma-key`            |
| `role-delta` | `git@github.com:<org>/ansible-role-delta.git`        | `delta`             | `galaxy-delta-key`            |

Each role repo should contain at minimum:

```
meta/main.yml          # galaxy_info + dependencies (empty)
tasks/main.yml         # debug task proving the role was installed
```

Register each public key as a **deploy key** on the matching GitHub repo (read
access only).

### Playbook repository

Use the existing `playbooks` repository (or a dedicated test repo) containing:

**`requirements.yml`**

```yaml
---
roles:
  - name: alpha
    scm: git
    src: git@github.com:<org>/ansible-role-alpha.git
    version: main

  - name: beta
    scm: git
    src: git@github.com:<org>/ansible-role-beta.git
    version: main

  - name: gamma
    scm: git
    src: git@github.com:<org>/ansible-role-gamma.git
    version: main

  - name: delta
    scm: git
    src: git@github.com:<org>/ansible-role-delta.git
    version: main
```

**`site.yml`** (smoke playbook that references installed roles):

```yaml
---
- hosts: localhost
  connection: local
  gather_facts: false
  roles:
    - alpha
    - beta
    - gamma
    - delta
```

## Phase A — Baseline: manual `ansible-galaxy` on runner (outside Semaphore)

Validates the test fixtures before testing Semaphore integration.

1. On the runner host, generate or copy all four private keys to a temp
   directory (e.g. `/tmp/galaxy-test-keys/`). **Do not commit these keys.**
2. Start `ssh-agent` and add all four keys:

   ```bash
   eval "$(ssh-agent -s)"
   ssh-add /tmp/galaxy-test-keys/alpha
   ssh-add /tmp/galaxy-test-keys/beta
   ssh-add /tmp/galaxy-test-keys/gamma
   ssh-add /tmp/galaxy-test-keys/delta
   ```

3. Clone the playbook repo (or copy `requirements.yml` into a clean directory).
4. Run:

   ```bash
   ansible-galaxy role install -r requirements.yml --force
   ```

5. Confirm all four roles appear under `~/.ansible/roles/` (or the configured
   roles path) and `ansible-galaxy list` shows `alpha`, `beta`, `gamma`, `delta`.

### Phase A — Expected results

* Galaxy install completes with exit code 0.
* No `Permission denied (publickey)` or `Host key verification failed` errors.
* All four roles are present locally.

If Phase A fails, fix deploy keys / URLs before proceeding to Phase B.

## Phase B — Semaphore task with Galaxy install

1. **Key Store**: create four SSH keys (`galaxy-alpha-key` … `galaxy-delta-key`)
   with the same private key bodies used in Phase A.
2. **Repository**: ensure the playbook repo (with `requirements.yml` and
   `site.yml`) is registered in Semaphore. Bind the repo's own deploy key for
   clone/pull (may be a fifth key, or one of the four if the playbook repo is
   also private).
3. **Template**: create an Ansible task template pointing at `site.yml`, with
   Galaxy install **enabled** (`skip_galaxy_install` = false / unset).
4. Configure host → key mapping per the implementation under test (e.g. template
   settings, variable group, or global `ssh.config_path`). *Skip this step if
   the feature is not yet merged — document as BLOCKED.*
5. **Run** the template.
6. Inspect the task log for the Galaxy install section.

### Phase B — Expected results (target behaviour)

* Task log shows `ansible-galaxy role install -r …/requirements.yml --force`
  (or collection equivalent) succeeding.
* No authentication errors for any of the four git sources.
* Playbook run reaches role execution; debug output from all four roles
  appears.
* Task status = **success**, exit code 0.

### Phase B — Known gaps on `develop` (as of QA-1 review)

Document as **FAIL / known limitation** if observed without the feature branch:

| Symptom | Root cause |
|---------|------------|
| `Permission denied (publickey)` during Galaxy step | `AnsibleApp.InstallRequirements` does not inject `SSH_AUTH_SOCK` / `GIT_SSH_COMMAND`; `args.Installer` is passed but unused (unlike `TerraformApp.init`). |
| Only repo clone works, Galaxy fails | Repository SSH agent is ephemeral (created/destroyed per git command in `CmdGitClient`); not available during Galaxy. |
| Inventory SSH key ignored for Galaxy | `SSH_AUTH_SOCK` from inventory key is appended **after** `prepareRun()` (which runs Galaxy), only for `ansible-playbook`. |
| Wrong key used when multiple keys in agent | No per-host mapping; OpenSSH offers keys in agent order — may work only when each deploy key is unique to one repo. |

Reference implementation to mirror: `TerraformApp.init()` in
`db_lib/TerraformApp.go` (uses `keyInstaller.Install` + `GetGitEnv()`).

## Phase C — Negative test (wrong key)

1. Replace one deploy key in Key Store with a key **not** registered on the
   corresponding GitHub repo.
2. Re-run the template (delete cached Galaxy hash under the template internal
   dir if needed, or touch `requirements.yml`).
3. Confirm the task fails during Galaxy install with an explicit
   `Permission denied (publickey)` for the affected role only.

## Postconditions

* Document actual vs. expected results in the QA-1 Linear issue.
* Remove temporary key files from the runner host.
* Optionally delete the four test role repos and keys if they were created only
  for this test.

## Automation notes

Full CI automation requires four private repos with deploy keys stored as CI
secrets (`TEST_GALAXY_ROLE_*` env vars). Phase A can be scripted in a
GitHub Actions job; Phase B requires a running Semaphore instance and is better
suited to manual or MCP e2e testing until host → key mapping is exposed in the
API/UI.
