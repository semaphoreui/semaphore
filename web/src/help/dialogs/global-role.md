---
form: src/components/EditGlobalRoleForm.vue
---

# Edit Global Role

A global role grants permissions over the whole instance of Semaphore to the users it is assigned
to; you assign it in the dialog of a user. It gives no access inside projects: that comes only
from the roles a user holds as a member of a project. An administrator holds every global
permission whatever roles are assigned. Global roles are a Pro feature.

## Name {#globalRole-name}

The name the role is listed under on the Global roles tab of the Users page and in the dialog of a
user. The field is required. You can rename a role at any time.

## Slug {#globalRole-slug}

The identifier of the role. Semaphore fills it in from the name while you type the name, until
you edit the slug yourself. It takes 1 to 100 letters, digits, underscores and hyphens, for
example `user-admins`.

The slug must be unique among the global roles; a project role may have the same slug. It can not
be changed once the role is saved: the field is locked when you edit a role.

## Global permissions {#globalRole-permissionsTitle}

What a user who holds the role may do across the whole instance. A new role has none ticked. The
permissions of all the global roles of a user add up.

- **Can manage users**: create users, change and delete them, and see their details in the list
  of users. Such a user can not make a user an administrator or a Pro user, and can not change
  or delete an administrator or a user who holds a global permission they do not hold
  themselves.
- **Can manage global roles**: create, change and delete the global roles and the shared project
  roles, and decide which users hold which global role. Give it only to people you trust with the
  whole instance.
- **Can manage system settings**: change the settings of the instance, the applications, the
  global runners and the group mappings; see the cluster and the information about the server;
  see and remove the tasks of all projects; clear the cache of the server.
- **Can read the audit log**: read the audit log of the instance and export it.
