---
form: src/components/KeyForm.vue
---

# Edit Key

A key of the Key Store holds a credential that the project uses without showing it: the SSH key
or the password Ansible logs in to the hosts with, the credentials of a Git repository, the
password of an Ansible vault, the secret of an integration. Inventories, repositories, task
templates and integrations refer to a key by its name. Once saved, the secret is never sent back
to the browser.

## Key Name {#key-name}

The name the key is listed under in the Key Store and in the key selectors of inventories,
repositories, task templates and integrations. The field is required. Pick a name that says what
the key opens, for example `Deploy key of the web servers`.

## Source {#key-source}

Where the secret of the key is kept. The tabs:

- **Local**: the secret is typed into this dialog and stored in the database of Semaphore,
  encrypted when the administrator has set an encryption key for access keys. This is the
  default.
- **Storage**: the secret is kept in an external secret storage of the project, chosen in
  **Storage** below. The tab is there in the Pro edition only. The storages themselves are set
  up on the **Storages** tab of the Key Store.
- **Env**: the secret is read from an environment variable of the Semaphore server process. Type
  the name of the variable in the field below the tabs.
- **File**: the secret is read from a file on the Semaphore server. Type the absolute path of the
  file in the field below the tabs. The file must be inside the secrets directory of the server
  (`dirs.secrets` in its configuration, `/tmp/semaphore` by default); a path outside it, or one
  with `..` in it, fails when the key is used. A trailing line break of the file is dropped.

With **Env** and **File**, Semaphore never writes the secret: it reads the value every time a
task uses the key, so you change the secret where it lives. The value must suit the **Type** of
the key. A **String** key takes the value as it is. An **SSH Key** or a **Login with password**
key expects JSON with the fields of that type, for example:

```json
{"login": "deploy", "password": "s3cret"}
```

```json
{"login": "deploy", "passphrase": "", "private_key": "-----BEGIN OPENSSH PRIVATE KEY-----\n..."}
```

When you edit a key, the tabs are locked until **Override** is ticked. A key synchronised from a
secret storage keeps its source, and its name and type can not be changed.

## Storage {#key-sourceStorageId}

Shown when the **Storage** tab is chosen.

The secret storage of the project that holds the secret: HashiCorp Vault, OpenBao, Devolutions
Server, AWS Secrets Manager or Azure Key Vault, as set up on the **Storages** tab of the Key
Store.

- A storage that is not read-only receives the secret you type in this dialog: Semaphore writes
  it there when you save. Deleting the key deletes the secret from the storage too. A key kept in
  such a storage can not be moved to another storage later.
- A read-only storage is only read. The dialog then shows no fields for the secret, and
  Semaphore reads it from the storage every time a task uses the key.

## Source Key {#key-sourceStorageKey}

Shown when a **Storage** is selected.

Where the secret is in that storage, for example the path to the secret in HashiCorp Vault.

## Type {#key-type}

What kind of credential the key holds. The field is required, and the fields below it change with
the type.

- **SSH Key**: a private key, with an optional login and passphrase. Ansible connects to the
  hosts with it, and Git clones a repository with an `ssh` URL with it.
- **Login with password**: a login and a password. Ansible connects to the hosts with them or
  becomes another user with them (**Sudo Credentials** of an inventory); Git clones an `https`
  repository with them; the password opens an Ansible vault of a task template.
- **String**: a single secret value without a login: a token, an API key, a password. Use it for
  the secret of an integration. It can not be an SSH key, sudo credentials, the password of a
  vault or the access key of a repository.
- **None**: no credential. Use it for a repository that needs no login (a public `https`
  repository, a `file` URL or a local path) and for an inventory whose hosts need no credentials
  or define the connection themselves.

## Username {#key-loginPasswordLogin}

Shown when the type is **Login with password**. Optional.

The user name of the credential.

- Ansible connects to the hosts as this user (`--user`); as **Sudo Credentials** of an
  inventory, it is the user Ansible becomes (`--become-user`). When the field is empty, Ansible
  uses the user the inventory or its own configuration names.
- For an `https` repository, Git logs in with this user and the password. Leave it empty for a
  token that the Git server takes alone: the password is then sent as the user name.

## Password {#key-loginPasswordPassword}

Shown when the type is **Login with password**. Required.

The password of the credential. Semaphore answers the password prompts of Ansible with it:
the SSH password (`--ask-pass`) for user credentials, the `become` password
(`--ask-become-pass`) for sudo credentials, and the vault password for an Ansible vault. For an
`https` repository it is the password or the access token of the Git server.

## Value {#key-stringValue}

Shown when the type is **String**. Required.

The secret value itself, for example the token an integration checks the requests it receives
against. It is stored as it is typed.

## Username {#key-sshLogin}

Shown when the type is **SSH Key**. Optional.

The user Ansible connects to the hosts as (`--user`). When the field is empty, Ansible uses the
user the inventory or its own configuration names.

For a Git repository, put the user into the repository URL instead, as in
`git@github.com:org/repo.git`.

## Passphrase {#key-sshPassphrase}

Shown when the type is **SSH Key**. Optional.

The passphrase the private key is encrypted with. Leave it empty for a key without a passphrase.
A wrong passphrase makes every task that uses the key fail when the key is loaded. A key
generated by Semaphore has no passphrase: what this field holds is discarded then.

## Generate SSH Key {#key-generateSshKey}

Shown when the type is **SSH Key**.

Semaphore creates a new key pair when you save: a 2048-bit RSA private key, which stays in the
key, and its public key, which is shown once in a dialog after saving and then under
**Public key** whenever you edit the key. Add the public key to `~/.ssh/authorized_keys` on the
hosts, or as a deploy key of the Git repository.

The private key field is locked while the box is ticked. On an existing key, tick **Override**
first: the old key pair is replaced. A key whose secret is read from **Env**, **File** or a
read-only storage can not be generated, as Semaphore has nowhere to write it.

## Private Key {#key-sshPrivateKey}

Shown when the type is **SSH Key**. Required unless **Generate SSH Key** is ticked.

The private key, in the OpenSSH or PEM format, including its `BEGIN` and `END` lines. For a task,
Semaphore loads it into an SSH agent of its own, which Ansible and Git use for the length of the
task.

## Public key {#key-publicKey}

Shown when you edit an SSH key that Semaphore generated, while **Override** is not ticked.

The public half of the generated key pair. **Show** unfolds it, the copy button puts it on the
clipboard. Add it to `~/.ssh/authorized_keys` on the hosts, or as a deploy key of the Git
repository.

It is the public key of the pair Semaphore generated last. After you replace the private key by
hand with **Override**, the key shown here no longer matches it.

## Override {#key-overrideSecret}

Shown when you edit an existing key.

Semaphore never shows a saved secret again, so the fields of the secret are empty and locked.
Tick the box to replace the secret: the source, the type and the fields of the secret open, and
everything in them is saved anew, so enter the whole secret again, not only the part that
changes. Without it, saving changes only the name of the key.

Saving a key also removes the copies the Semaphore server keeps of the repositories that use it:
the next task clones them again.
