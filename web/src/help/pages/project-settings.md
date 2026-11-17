---
form: src/views/project/Settings.vue, src/components/ProjectForm.vue
skip: newProject-demo
---

# Project Settings

The settings of the project as a whole: its name, how many of its tasks may run at once and
whether it sends alerts. The Danger Zone below the form holds the actions that reach the whole
project: its backup, its cache and its removal. Changes of the form take effect after **Save**.

## Project Name {#newProject-name}

The name the project is listed under in the project menu at the top of the side bar. The field
is required.

## Max number of parallel tasks {#newProject-maxParallelTasks}

How many tasks of this project may run at the same time. A task started above the limit waits
in the queue until one of the running tasks ends.

Leave the field empty or set it to `0` for no limit of its own. The limit of the whole server,
set by its administrator, applies in either case.

## Telegram Chat ID {#newProject-tg}

Optional. The Telegram chat that receives the alerts of this project. When the field is empty,
the alerts go to the chat set for the whole server.

The field has an effect only when the administrator has turned Telegram alerts on for the
server and alerts are allowed for this project.

## Allow alerts for this project {#newProject-alert}

Turns alerts on for the tasks of this project. Where the alerts go is set for the whole server
by its administrator: e-mail, Telegram, Slack, Rocket.Chat, Microsoft Teams, DingTalk or Gotify.

- An e-mail is sent when a task fails, to the members of the project who may see the task and
  have alerts turned on in their account.
- A message to the messengers is sent when a task succeeds, fails or waits for a confirmation.

A task template can keep its successful or its failed tasks from sending alerts.

## Test Alerts {#settings-testAlerts}

Sends a test alert through every channel the server has turned on, so that you can check that
the alerts of this project arrive.

The button works only when **Allow alerts for this project** is ticked and saved. Tick it and
press **Save** first.

## Backup Project {#settings-exportProject}

Downloads a JSON file with the configuration of the project: task templates, inventories,
repositories, variable groups, keys, schedules, integrations and the other things set up in it.
The file can be turned back into a project with **Restore Project** in the project menu at the
top of the side bar.

The file leaves out the secrets of the keys and the secrets on the **Secrets** tab of the
variable groups: they have to be entered again after a restore. Everything else is copied as it
is, also a password typed straight into another field: the address of a repository with a
password in it, the text of a static inventory, the values of ordinary variables and environment
variables. Check the file for such values before you share it. The history of tasks is not a part
of the backup.

## Clear cache {#settings-clearCache}

Deletes the files Semaphore keeps on the server for this project, such as its copies of the
repositories. They are downloaded again when the next task starts.

Use it when a task fails on a stale or broken copy of a repository. The action can not be
undone, and it does not touch the settings of the project.

## Delete Project {#settings-deleteProject}

Deletes the project with everything in it: task templates, the history of tasks, inventories,
repositories, variable groups, keys and schedules. Semaphore asks for a confirmation first.

A deleted project can not be brought back. Download a backup before you delete a project you
may need again.
