---
form: src/components/EditTemplatePermissionForm.vue
---

# Workflow permissions

Gives a role extra permissions on this one workflow. A member who holds the role gets them on
top of the permissions the role has in the whole project: a grant here only adds, it never takes
a permission away. A member with several roles gets every permission any of them is granted.

## Role {#templatePerm-roleSlug}

The role that gets the permissions. The list shows the roles of this project, the shared roles
and the built-in roles. The field is required, and a role can have only one grant on a workflow:
the server refuses a second one.

The role of a saved grant can not be changed. The server keeps the role and saves only the
permissions, whatever you pick here. To give the permissions to another role, remove this grant
from the list and add a new one.

## Permissions {#templatePerm-permissionsTitle}

What a member who holds the role may do with this workflow:

- **Can see this workflow**: the workflow is in the list of workflows, with its runs.
- **Can run this workflow**: start runs of the workflow.
- **Can edit this workflow**: change and delete the workflow, and manage these permissions.
- **Can approve this workflow**: approve or reject the approval steps of its runs.

Running, editing or approving includes seeing the workflow: when you check any of them, **Can
see this workflow** is checked too and can not be cleared. Leave every box clear to save a grant
that gives nothing.
