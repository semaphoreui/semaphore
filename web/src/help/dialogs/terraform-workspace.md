---
form: src/components/TerraformInventoryForm.vue
---

# New Terraform Workspace

A workspace is the inventory of a Terraform, OpenTofu or Terragrunt template: it names the
workspace a task selects after `init`, and the key `init` downloads private modules with. The
workspace you create here is attached to this template and listed above its state. Tasks of the
template run in its default workspace, the one marked with a check; **Make default** changes it.
The same dialog edits a workspace when you open it with the pencil button.

## Workspace name {#tfWorkspace-inventory}

The name of the Terraform workspace, for example `production`. The field is required.

Before a plan, Semaphore runs `workspace select -or-create=true` with this name, so the workspace
is created when it does not exist yet. When `workspace list` fails, for example because the
backend does not support workspaces, Semaphore skips the selection.

Semaphore stores the workspace as `TEMPLATE - NAME`, the name of the template and this name: you
see that full name on the Inventory page and when you attach the workspace to a template.
Renaming a workspace does not rename it in Terraform: the next task selects, or creates, a
workspace of the new name.

## SSH key for private modules {#tfWorkspace-sshKeyId}

The key `init` uses to download modules from private Git repositories over SSH. The field is
required in this dialog. The list offers the keys of the type **SSH Key**, **Login with password**
and **None** from the Key Store.

Only an **SSH Key** has an effect: Semaphore loads it into an SSH agent for the time of `init`,
and Git takes it for module sources such as `git::ssh://git@example.com/modules/vpc.git`. A key of
the type **Login with password** passes no credentials to Git. Choose **None** when the modules
need no key.
