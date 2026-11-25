---
form: src/components/InventorySelectForm.vue
---

# Choose workspace to attach

Attaches a workspace that already exists in the project to this template, so it appears among
the workspaces of the template with its state and aliases. A workspace belongs to one template at
a time. Attaching does not make the workspace the default one: tasks keep running in the default
workspace until you press **Make default**.

## Workspace {#workspaceSelect-inventoryId}

The workspace to attach. The field is required. The list shows the workspaces of the project that
are made for the app of this template (Terraform, OpenTofu or Terragrunt) and are not attached to
any template; a workspace detached from its template with **Detach** shows up here again.

The workspace keeps its name, key and state. The server refuses a workspace that another template
has taken in the meantime.
