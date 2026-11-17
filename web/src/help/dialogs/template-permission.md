---
form: src/components/EditTemplatePermissionForm.vue
---

# Template permissions

Gives a role extra permissions on this one template. A member who holds the role gets them on
top of the permissions the role has in the whole project: a grant here only adds, it never takes
a permission away. A member with several roles gets every permission any of them is granted.

## Role {#templatePerm-roleSlug}

The role that gets the permissions. The list shows the roles of this project, the shared roles
and the built-in roles. The field is required, and a role can have only one grant on a template:
the server refuses a second one.

The role of a saved grant can not be changed. The server keeps the role and saves only the
permissions, whatever you pick here. To give the permissions to another role, remove this grant
from the list and add a new one.

## Permissions {#templatePerm-permissionsTitle}

What a member who holds the role may do with this template:

- **Can see this template**: the template is in the list of templates, with its tasks and
  their logs.
- **Can run this template**: start tasks from the template, and stop, confirm or reject its
  tasks.
- **Can edit this template**: change and delete the template, stop all its tasks, and manage
  these permissions.

Running or editing includes seeing the template: when you check either of them, **Can see this
template** is checked too and can not be cleared. Leave every box clear to save a grant that
gives nothing.
