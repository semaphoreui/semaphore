---
form: src/components/RepositoryForm.vue
---

# Edit Repository

A repository is where the playbooks, Terraform code or scripts of a task template come from: a
Git repository, or a directory on the machine that runs the tasks. Before every task Semaphore
clones the Git repository, or updates the copy it cloned before, checks out the branch, and runs
the task in that copy.

## Name {#repository-name}

The name the repository is listed under on the Repositories page and in the repository selector
of a task template and of an inventory. The field is required.

## URL or path {#repository-gitUrl}

Where the code is. The field is required. The chips under it show the kind of address you typed,
and a click on a chip rewrites the address to that kind.

- **ssh**: `git@github.com:org/repo.git` or `ssh://git@github.com/org/repo.git`. Git logs in
  with the user in the address and the **SSH Key** chosen in **Access Key**.
- **https** and **http**: `https://github.com/org/repo.git`. A private repository needs a
  **Login with password** key; a public one needs **None**. Avoid `http`: the login and the
  password then travel unencrypted.
- **git**: `git://host/org/repo.git`, the read-only Git protocol, which has no login.
- **file**: `file:///srv/git/repo.git`, a Git repository on the disk of the machine that runs the
  task, cloned like a remote one.
- **abs. path**: an absolute path such as `/srv/playbooks` or `D:\playbooks`. Semaphore does not
  clone it: the task runs right in that directory, which must exist on the machine that runs the
  task (the Semaphore server or a runner). The branch does not apply to it.

Changing the address removes the copies the Semaphore server cloned from the old one.

## Branch / Tag {#repository-gitBranch}

The branch Semaphore clones and keeps up to date, for example `main`. The field is required,
except for an absolute path, for which it is locked.

A task template can name a branch of its own, which wins over this one, and a template can let
its tasks pick a branch when they are started. A tag works too, but a tag can not be pulled, so
Semaphore clones the repository anew before every task.

## Single checkout for all templates {#repository-monorepo}

Off by default: every task template gets its own copy of the repository. Tick the box to clone
the repository once and share that copy between all the templates that use it, for a repository
too large to keep a copy per template.

The shared copy stays on the branch of this repository: templates using it can not have a branch
of their own, and their tasks can not pick another branch or commit. Locked for an absolute path.

## Access Key {#repository-sshKeyId}

The key from the Key Store that Git logs in to the repository with. The field is required. It
offers keys of three types, the one to choose depends on the URL:

- **SSH Key** for an `ssh` URL. Semaphore loads the private key into an SSH agent for the length
  of the clone or the update. The user is the one in the URL, as `git` in
  `git@github.com:org/repo.git`.
- **Login with password** for an `https` URL: the login and the password, or the access token
  of the Git server. With an empty login the password is sent alone, as the user name, which is
  what some servers expect of a token. The same credentials are given to `ansible-galaxy` for
  roles and collections on the same server.
- **None** for a public repository, a `git` or `file` URL, or an absolute path.

A key that does not match the URL gives Git no credentials, and a private repository then fails
to clone.
