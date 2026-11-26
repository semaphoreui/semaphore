---
form: src/components/ProjectForm.vue
---

# New Project

A project holds the task templates, inventories, repositories, variable groups and keys of one
piece of work, and the people who work on it. You become the owner of the project you create.
Only an administrator can create a project, unless the administrator of the server allows it to
every user. Every setting below can be changed later on the Settings page of the project.

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

## Demo {#newProject-demo}

Fills the new project with examples to try Semaphore on: the public repository
`https://github.com/semaphoreui/semaphore-demo.git`, inventories, a vault password key and
task templates for Ansible, OpenTofu, Terragrunt, Bash and PowerShell, sorted into the views
**Build**, **Deploy** and **Tools**.

Leave the switch off for a project without examples: it gets only the key **None** and the
view **All**. The examples are ordinary items of the project: edit
or delete them as you like.
