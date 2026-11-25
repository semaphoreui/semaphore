---
form: src/components/TemplateVaults.vue
---

# Vault

A vault gives Ansible the password of the files and variables encrypted with Ansible Vault in the
playbook repository. Every vault of the template becomes one `--vault-id` option of
`ansible-playbook`. The vault is saved with the template, when you save the template dialog.

## Vault ID {#templateVaultDialog-name}

Optional. The vault ID the password is for, the label the content was encrypted with
(`ansible-vault encrypt --vault-id prod@prompt`). Leave it empty for the vault ID `default`.

The ID is saved in lower case. Every vault of the template needs its own ID, and only one may be
empty.

## Type {#templateVaultDialog-type}

Where the password comes from:

- **Password**: a key from the Key Store. Semaphore types its password when Ansible asks for
  it, so it is never written to a file.
- **Client Script**: a vault password client script that prints the password; Ansible runs it.

## Vault Password {#templateVaultDialog-vaultKeyId}

Shown when the type is **Password**.

The key that holds the vault password. Only keys of the type **Login with password** are
listed; Semaphore uses their password and ignores the login. The field is required.

## Script Path {#templateVaultDialog-script}

Shown when the type is **Client Script**.

The path to the script, for example `scripts/vault-pass-client.py`. A relative path is taken from
the directory Ansible runs in: the root of the repository, or the working directory of the
template when it has one. The field is required.

The file name must end in `-client`, with or without an extension: only then does Ansible run
the file as a client script and pass it the vault ID with `--vault-id`. The script runs on the
machine that runs the task, as the user Semaphore runs as.
