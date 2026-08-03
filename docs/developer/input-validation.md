# Input validation and security gates

Semaphore validates several user-supplied fields at the API and execution layers
to prevent path traversal, git option injection, privilege escalation, and
cross-object tampering. This page documents the checks added or tightened in
recent releases so contributors know where to extend them.

## Playbook and script paths

**Code:** `db/playbook_path.go`, called from `db.Template.Validate`,
`db.Task.Validate`, and `services/tasks/local_executor.go`.

Template and task `playbook` values (also used for shell scripts, Terraform
subdirectories, and similar app paths) must be **relative** and must stay inside
the repository checkout bound to the template.

| Input | Result |
| --- | --- |
| `playbooks/site.yml` | Allowed |
| `./deploy/main.tf` | Allowed (cleaned to a relative path) |
| `/etc/passwd` | Rejected — absolute path |
| `C:\Windows\System32\evil.ps1` | Rejected — Windows drive-letter path |
| `../../etc/passwd` | Rejected — escapes repository via `..` |
| `subdir\..\..\secret.yml` | Rejected — backslashes normalized before checks |

Backslashes are treated as path separators so Windows-style paths cannot bypass
the relative-path rule.

Empty playbook is allowed at validation time for Terraform templates (inventory
and app type govern whether a path is required at run time).

## Repository git URLs

**Code:** `db/git_url.go`, called from `db.Repository.Validate`.

The git client passes the repository URL as a positional argument to the `git`
binary. A URL that begins with `-` would be parsed as a git option and could
enable arbitrary command execution (git option injection).

| Input | Result |
| --- | --- |
| `https://github.com/org/repo.git` | Allowed |
| `git@github.com:org/repo.git` | Allowed |
| `--upload-pack=/path/to/script` | Rejected |
| ` -c core.sshCommand=evil` | Rejected (leading whitespace trimmed, then `-` check) |

Legitimate schemes (`https://`, `ssh://`, `git://`, `file://`, scp-like
`user@host:path`, and local paths) never start with `-`, so rejecting dash-prefixed
values is safe.

## Git branch override at task launch

**Code:** `services/tasks/local_executor.go` — `resolveGitBranch`.

Branch resolution order:

1. Repository default branch
2. Template `git_branch` (if set) — always wins over the repository default
3. Task `git_branch` — **only** when `allow_override_branch_in_task` is `true` on the template

When the template flag is `false` (default), a task runner cannot redirect a
branch-pinned template to an arbitrary repository branch. This closes a privilege
gap where users with run-only access could otherwise check out sensitive branches.

Template field: `allow_override_branch_in_task` (UI: **Branch** override checkbox
in template advanced options). Exposed in `api-docs.yml` on the template schema.

## Custom roles and built-in permission precedence

**Code:** `db/Role.Validate`, `api/projects/project.go` (`ProjectMiddleware`).

### Creating custom roles

`ValidateRole` rejects:

- Empty `name` or `slug`
- Slugs that match built-in roles (`owner`, `manager`, `task_runner`, `guest`, …)

Reusing a built-in slug would let a custom role shadow the built-in definition
and escalate permissions for every member assigned that slug.

### Resolving permissions per request

`ProjectMiddleware` loads custom role permissions from the database **only when
the user's role slug is not a built-in** (`!roleSlug.IsValid()`). Built-in roles
always use the in-code `rolePermissions` map as the source of truth, even if a
stale custom row with the same slug exists in the database.

## Access key updates

**Code:** `api/projects/keys.go` — `UpdateKey`.

`PUT /api/project/{project_id}/keys/{key_id}` validates that the JSON body is
consistent with the URL-resolved key:

| Check | HTTP status | Error |
| --- | --- | --- |
| Body `id` ≠ URL `key_id` | 400 | `Access key id in URL and in body must be the same` |
| Body `project_id` missing or ≠ key's project | 400 | `You can not move access key to other project` |

These checks run before any store update, preventing IDOR-style moves of keys
across projects or keys.

Synchronized keys still cannot change `name` or `type` (unchanged behaviour).

## Repository branch before playbook browse

**Code:** `api/projects/repository.go` — `GetRepositoryPlaybooks`.

When the UI browses playbooks on a remote repository, the optional `?branch=`
query parameter is validated with `ValidateGitBranch` **before** any clone or pull.
Invalid branch names are rejected without touching git.

Each distinct branch gets its own scratch checkout directory
(`repository_{id}_browse_{branchHash}`) because single-branch clones cannot
switch to unfetched refs. This prevents browse failures when users switch
branches in the template editor.

## Task commit hash

**Code:** `pkg/git/git_branch.go` — `ValidateCommitHash`, called from
`db.Task.ValidateNewTask`.

Optional `commit_hash` on a new task must be a plain hex git object name (7–64
characters). This prevents values like `--upload-pack=…` from being passed to
`git checkout` when a task pins a specific commit.

| Input | Result |
| --- | --- |
| `a1b2c3d` | Allowed (abbreviated SHA-1) |
| `deadbeef…` (40 hex chars) | Allowed (full SHA-1) |
| `main` | Rejected — not hex |
| `--upload-pack=evil` | Rejected |

Empty commit hash is allowed (branch-based checkout proceeds as usual).

## Integration sub-resource updates

**Code:** `api/projects/integration_matcher.go`,
`api/projects/integration_extract_value.go`.

Updates to integration matchers and extract values validate that URL path IDs
match the JSON body and that nested resources belong to the integration resolved
by `IntegrationMiddleware` (which loads the integration by `(project_id,
integration_id)` from the URL).

| Check | HTTP status | Error |
| --- | --- | --- |
| Body `id` ≠ URL resource ID | 400 | `Matcher ID in body and URL must be the same` / `Value ID in body and URL must be the same` |
| Body `integration_id` ≠ URL integration | 400 | `Integration ID in body and URL must be the same` (matchers) or empty body (extract values) |

Child resources cannot be reparented to another integration through a PUT body.

## Template vault updates

**Code:** `db/sql/template_vault.go` — `UpdateTemplateVaults`.

When a template's vault list is saved, existing vault rows are updated with a
`WHERE` clause scoped to `(project_id, template_id)`. A vault ID from another
project or template cannot be overwritten or reparented — the update is a no-op
for foreign IDs. `project_id` and `template_id` are not writable through the
update SET clause.

## Project backup secrets

**Code:** `services/project/backup_marshal_test.go`, `db.Runner` backup tags.

Project backup JSON never includes runner authentication material:

| Excluded field | Reason |
| --- | --- |
| `token` | Runner bearer secret |
| `registration_token` | One-time registration secret |
| `registration_token_expires_at` | Registration metadata tied to the secret |

Non-secret runner fields (`name`, `webhook`, tags, etc.) are still exported.
After restore, runners must be re-registered or tokens re-assigned manually.

## Related validation elsewhere

These are not new but often confused with the rules above:

- **Git branch names** — `ValidateGitBranch` in `pkg/git/git_branch.go` rejects
  invalid ref syntax. Applied on repository, template, task, and repository-browse
  paths.
- **Template arguments** — must be valid JSON when set (`db.Template.Validate`).
- **JWT template params** — `TemplateJWTParams.Validate` enforces audience size
  and TTL caps when per-task JWT is enabled.

When adding a new user-controlled field that reaches shell, git, or the filesystem,
add validation in `db/` (or the service layer) and document the constraint here.
