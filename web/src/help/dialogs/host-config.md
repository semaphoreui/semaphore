---
form: src/components/HostConfigForm.vue
---

# Edit mapping

A mapping tells Semaphore which credential of the Key Store to use for an SSH host or a Git
URL. A repository carries one key, but a task may also reach other servers: submodules, roles
and collections of `requirements.yml`, Terraform modules, inventory repositories. The mappings
of the project apply to every Git and SSH connection of its tasks, and to the Git commands
Semaphore runs for the project, such as the check of a repository for new commits. No file of
your repository changes.

## Type {#hostConfigForm-type}

What the mapping matches:

- **Host**: an SSH host name. Every SSH connection of a task to that host uses the key of the
  mapping: Git over SSH, and Ansible connecting to an inventory host of that name. The mapping
  wins over the SSH configuration of the server for the same host.
- **URL**: a Git URL or the beginning of one. Git connects to every repository under that URL
  with the credential of the mapping.

Switching to **Host** clears a **Credential** that a host mapping can not use.

## Host or URL {#hostConfigForm-host}

The field is required. One mapping of a type may exist for each host or URL.

For **Host**: the host name exactly as Git or Ansible connect to it, for example `github.com` or
`10.0.0.21`. Letters, digits, dots, hyphens and underscores only: no user, no port, no
wildcards.

For **URL**: an address that starts with `https://` or `http://`, without a login in it and
without spaces, quotes or `=`. Give a whole repository, or a prefix that ends with `/` to cover a
group of repositories:

```
https://github.com/acme/
https://github.com/acme/infra.git
```

When several mappings match, the longest one wins: the repository mapping above overrides the
mapping of `acme` for that one repository.

With an SSH key, a URL mapping makes Git connect over SSH to the host of the URL on port `22`,
whatever port the URL names.

## Credential {#hostConfigForm-sshKeyId}

The key from the Key Store used for the connections the mapping matches. The field is required.

- **Host** accepts an **SSH Key** only. When the key has a login, SSH connects as that user.
- **URL** accepts an **SSH Key** or a **Login with password**. An SSH key connects as its login,
  or as `git` when it has none. A login with password works with an `https://` URL only; for a
  token, put the token in the password and leave the login empty.

A login with password reaches Git through its environment, not its command line, and the task
log shows only the original URL.
