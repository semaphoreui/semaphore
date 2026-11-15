---
form: src/components/RestoreProjectForm.vue
---

# Restore Project...

Creates a new project from a backup file made with **Backup Project** on the Settings page of a
project. The new project gets the task templates, inventories, repositories, variable groups,
keys, schedules, integrations and the other things set up in the backup, and you become its
owner. Only an administrator can restore a project, unless the administrator of the server lets
every user create projects.

## Project Name {#restoreProject-name}

The name of the new project. The field is required, and it replaces the name written in the
backup file.

The name must differ from the names of all projects on the server: the server refuses to restore
a project under a name that is taken already.

## Backup file {#restoreProject-file}

The JSON file of the backup. Pick the file before you press **Restore**.

A backup holds no secret values: after the restore, Semaphore tells you how many keys are empty
and need their secrets entered again. The secrets of the variable groups have to be entered
again as well. The members of the
project and the history of its tasks are not a part of the backup. A backup written by a newer
version of Semaphore is refused.
