---
form: src/components/EditRoleForm.vue
---

# Edit Role

A role is a named set of permissions inside a project. You give roles to the members of a project,
and a member may do what any of their roles allows. Opened from the Roles page of a project, the
dialog edits a role of that project only; opened from the Roles page outside a project, it edits a
shared role that every project can use. The built-in roles Owner, Manager, Task Runner and Guest
can not be changed.

## Name {#role-name}

The name the role is listed under on the Roles page and wherever a role is picked, for example
when you add a member to a project. The field is required. You can rename a role at any time.

## Slug {#role-slug}

The identifier of the role. Semaphore fills it in from the name while you type the name, until
you edit the slug yourself. It takes 1 to 100 letters, digits, underscores and hyphens, for
example `release-managers`.

The slug can not be changed once the role is saved: the field is locked when you edit a role.

- A slug must be unique, and the check ignores the case of letters. The slug of a shared role
  must differ from every other role; the slug of a project role must differ from the shared
  roles and from the other roles of its project.
- `owner`, `manager`, `task_runner`, `guest` and `all` are reserved, in any case.

## Permissions {#role-permissionsTitle}

What the members who hold the role may do in the project. A new role has none ticked. The
permissions of all the roles a member holds add up; an administrator of Semaphore may do
everything in every project whatever the roles.

- **Can read project**: see the project: its task templates, tasks and their logs, inventories,
  keys, repositories, variable groups, schedules and members. Every read in the project asks for
  it, and no other permission includes it, so tick it for almost every role. Without it a member
  sees only the task templates and workflows that grant the role access of their own, with their
  tasks.
- **Can run project tasks**: start tasks from task templates, stop them, and confirm or reject a
  task that waits for a confirmation.
- **Can update project**: change the settings of the project and delete it.
- **Can manage project resources**: create, change and delete task templates, inventories, keys,
  repositories, variable groups, schedules, views, integrations, secret storages, project runners
  and the roles of the project; delete tasks; set which roles may use a task template.
- **Can manage project users**: add members to the project, change their roles and remove them.
- **Can run workflows**: start a workflow and stop its runs.
- **Can edit workflows**: create, change and delete workflows, and set which roles may use a
  workflow.
- **Can approve workflows**: approve or reject a step of a workflow run that waits for an
  approval.
