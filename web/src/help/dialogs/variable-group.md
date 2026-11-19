---
form: src/components/EnvironmentForm.vue
skip: varGroupSync-syncInterval
---

# Edit Variable Group

A variable group holds the extra variables, the environment variables and the secrets that
Semaphore passes to every task of a task template that uses the group. Ansible gets the extra
variables as `--extra-vars`, Terraform and OpenTofu as `-var`; the environment variables are set
for the process of the task. A template may use several groups: when two of them define the same
name, the group that comes later in the template wins.

## Group Name {#varGroup-name}

The name the group is listed under on the Variable Groups page and in the variable group
selector of a task template. The field is required.

## Secret storage {#varGroup-secretStorageId}

Shown when the group is new and the server supports secret storages. Optional.

The secret storage the secrets of this group are kept in, for example a HashiCorp Vault. Leave
the field empty to keep the secrets in the database of Semaphore, encrypted.

The storage is chosen once, when the group is created: an existing group keeps its storage, and
the field is no longer shown. A storage marked **Read only** does not take new secrets: saving a
secret into it fails.

## Secret key prefix {#varGroup-secretStorageKeyPrefix}

Shown when the group is new and the server supports secret storages. Available once a
**Secret storage** is chosen.

The start of the path every secret of this group gets in the storage. Semaphore adds 10 random
characters to it: with the prefix `semaphore/prod/` a secret is stored as `semaphore/prod/`
followed by those characters. The prefix is taken as it is, so end it with `/` when the secrets
should sit under it as a folder.

The prefix is set once, when the group is created. The **Secrets** tab shows it as the source
path pattern.

## Table, JSON or YAML {#varGroup-extraVarsMode}

How you edit the extra variables:

- **Table**: one row per variable, each with a type. Add a row with the **+** button next to the
  switch.
- **JSON**: all the variables as one JSON object.
- **YAML**: all the variables as one YAML mapping.

The dialog opens in **Table**. A switch converts what you typed; when it can not be read, the
dialog stays in the mode you were in and shows the error. Whatever the mode, the variables are
saved as one object of names and values.

## Extra variables in JSON {#varGroup-extraVarsJson}

Shown when the extra variables are edited as **JSON**.

The extra variables as one JSON object. A name may not be empty; a value may be a string, a
number, a list or an object:

```json
{
  "app_version": "1.4.2",
  "replicas": 3,
  "regions": ["eu-west-1", "us-east-1"]
}
```

## Extra variables in YAML {#varGroup-extraVarsYaml}

Shown when the extra variables are edited as **YAML**.

The extra variables as one YAML mapping. The text is read with the types JSON knows: an unquoted
date such as `2024-01-01` stays a string, and `.inf` or `.nan` are refused. A list or a scalar at
the top level is refused as well.

```yaml
app_version: 1.4.2
replicas: 3
regions:
  - eu-west-1
  - us-east-1
```

## Variable name {#varGroup-varName}

Shown when the extra variables are edited as a **Table**, in every row.

The name of the variable as the playbook or the Terraform code sees it, for example
`app_version`.

## Variable type {#varGroup-varType}

Shown when the extra variables are edited as a **Table**, in every row.

How the value of the row is read:

- **String**: the text as it is.
- **Number**: a number, for example `42`.
- **List**: a JSON array, for example `["a", "b"]`.
- **Dict**: a JSON object, for example `{"key": "value"}`.

A value that does not match its type stops the save with the name of the variable.

## Variable value {#varGroup-varValue}

Shown when the extra variables are edited as a **Table**, in every row.

The value of the variable, in the format of its type. For a **List** or a **Dict**, the button
at the end of the field opens a larger editor.

Semaphore passes the values to Ansible as data: a `{{ … }}` expression inside a value is not
evaluated, the playbook gets the text as it is.

## Environment variable name {#varGroup-envName}

On the **Variables** tab, under **Environment variables**, in every row. Add a row with the
**+** button of the section.

The name of a variable set in the environment of the process that runs the task, for example
`AWS_REGION`.

## Environment variable value {#varGroup-envValue}

On the **Variables** tab, under **Environment variables**, in every row.

The value of the variable, as text. It is stored as plain text and shown to everyone who opens
the group: put passwords and tokens in the **Secrets** tab instead.

## Sync keys enabled {#varGroup-syncEnabled}

Shown on the **Secrets** tab of a group that has a secret storage.

Turn it on to import secrets from the storage into this group, from the paths set under
**Sync paths**. The import runs when you press the sync button of the group on the Variable
Groups page, and on its own at the interval set in **Sync paths**. Turned off, nothing is
imported.

## Sync paths {#varGroup-syncPaths}

Shown when **Sync keys enabled** is on.

Opens the dialog with the paths of the storage to import from and the interval of the automatic
import. The number on the button is the number of paths. The sync button of the group on the
Variable Groups page stays disabled until the group has at least one path.

The dialog has only **Close**: what you change in it is saved with the group, by **Save** of
this dialog.

## Secret variable name {#varGroup-secretVarName}

On the **Secrets** tab, under **Extra variables**, in every row. Add a row with the **+** button
of the section.

The name of an extra variable whose value is a secret. The task gets it as it gets the extra
variables of the **Variables** tab; a secret wins over an extra variable of the same name.
Ansible gets the secrets in a file encrypted with Ansible Vault, Terraform and OpenTofu as
`-var`. As the dialog warns, a secret passed this way may still appear in plain text in the log
of Ansible.

## Secret variable value {#varGroup-secretVarValue}

On the **Secrets** tab, under **Extra variables**, in every row.

The value of the secret. Semaphore encrypts it, or writes it into the secret storage of the
group, and never shows it again: when you open the group later, the field is empty. Leave it
empty to keep the stored value, type a new one to replace it. Deleting the row deletes the secret
when you save the group.

## Secret environment variable name {#varGroup-secretEnvName}

On the **Secrets** tab, under **Environment variables**, in every row. Add a row with the **+**
button of the section.

The name of an environment variable whose value is a secret, for example `DB_PASSWORD`. It is
set in the environment of the process that runs the task, as the environment variables of the
**Variables** tab are.

## Secret environment variable value {#varGroup-secretEnvValue}

On the **Secrets** tab, under **Environment variables**, in every row.

The value of the secret. As for a secret extra variable, it is never shown again: leave the field
empty to keep the stored value, type a new one to replace it.
