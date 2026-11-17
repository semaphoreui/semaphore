---
form: src/components/TaskForm.vue, src/components/TaskParamsAnsibleForm.vue, src/components/TaskParamsTerraformForm.vue
---

# New Task

Starts a task of the template: one run of its playbook, Terraform code or script. The template
decides which fields this dialog shows, and each section below says when its field is there. The
button at the bottom is **Run**, **Build** or **Deploy**, after the type of the template.

When the template has a survey, its questions are fields after **Message**. A field marked `*` is
required, an Integer field takes digits only, and the default answers are filled in. Semaphore
passes the answers to the task as variables: extra vars for Ansible, `-var` for Terraform,
`name=value` arguments for a script, or environment variables for a question set so. An answer
replaces a variable of the same name in the variable groups of the template. A secret answer is
hidden as you type and is not kept with the task.

## Commit {#newTask-commit}

Shown when you run an earlier task again, and that task ran on a known commit.

The task checks out this commit of the repository instead of the newest commit of the branch.
Close the box to run the newest commit.

The commit is used only when the template allows overriding the branch, and never for a
repository marked as a monorepo. Otherwise the task runs the newest commit.

## Build Version {#newTask-buildTaskId}

Shown when the type of the template is **Deploy**.

The build to deploy: one of the last 20 successful tasks of the build template, listed by version
and message. The newest one is selected. The field is required.

The version of the build reaches the task: as `semaphore_vars.task_details.incoming_version` in
Ansible, as the environment variable `SEMAPHORE_TASK_DETAILS_INCOMING_VERSION` in a script. The
list is empty when you may not see the build template.

## Message {#newTask-message}

Optional. A note on why the task runs, for example `Hotfix for the login page`. It is shown with
the task in the task history and changes nothing in the run.

The playbook can read it as `semaphore_vars.task_details.message`, and a script as the
environment variable `SEMAPHORE_TASK_DETAILS_MESSAGE`.

## Branch / Tag {#newTask-gitBranch}

Shown when the template allows overriding the branch, unless the repository is a monorepo.

The branch or tag of the repository the task checks out. The field starts with the branch of the
template. When it is empty, the task uses the branch of the template, or the branch of the
repository when the template has none. A name that is not a valid Git branch name is refused.

## Inventory {#newTask-inventoryId}

Shown when the template allows choosing the inventory in a task and you may see the project. For
a Terraform, OpenTofu or Terragrunt template the field is labelled **Workspace**, but the server
uses the choice for Ansible templates only: a Terraform, OpenTofu or Terragrunt task runs in the
workspace of the template, whatever you pick here.

The inventory the task runs against. The field starts with the inventory of the template and
lists the inventories of the project made for the app of the template. When the chosen inventory
is gone by the time the task starts, the task runs against the inventory of the template.

## Limit {#taskParams-limit}

Shown for an Ansible template that allows overriding the limit.

The hosts or groups of the inventory the playbook runs on, one pattern per chip, for example
`web` or `10.0.0.11`. Semaphore passes them as `--limit` joined by commas. The list starts with
the limit of the template. Remove every chip to run on all the hosts of the inventory.

## Tags {#taskParams-tags}

Shown for an Ansible template that allows overriding the tags.

Only the tasks and roles of the playbook with one of these tags run: Semaphore passes them as
`--tags`, one tag per chip. The list starts with the tags of the template. Remove every chip to
run the whole playbook.

## Skip tags {#taskParams-skipTags}

Shown for an Ansible template that allows overriding the skipped tags.

The tasks and roles of the playbook with one of these tags do not run: Semaphore passes them as
`--skip-tags`, one tag per chip. The list starts with the skipped tags of the template.

## Debug {#taskParams-debug}

Shown for an Ansible template that allows debugging.

Runs `ansible-playbook` with a verbosity flag, so the log of the task shows more detail. The label
shows the flag: `-v` as many times as **Debug level** says.

## Debug level {#taskParams-debugLevel}

Shown with **Debug**, and enabled when **Debug** is ticked.

How verbose the log is, from 1 (`-v`) to 6 (`-vvvvvv`). The default is 4 (`-vvvv`), which also
shows the details of the connections to the hosts.

## Dry Run {#taskParams-dryRun}

Shown for an Ansible template, unless the template hides it.

Runs the playbook with `--check`: Ansible reports what it would change on the hosts and changes
nothing.

## Diff {#taskParams-diff}

Shown for an Ansible template, unless the template hides it.

Runs the playbook with `--diff`: the log shows the changes made to files and templates on the
hosts. Tick **Dry Run** too to see the changes without making them.

## Skip Galaxy install {#taskParams-skipGalaxyInstall}

Shown for an Ansible template that allows overriding the Galaxy install.

Before the playbook runs, Semaphore installs the roles and collections listed in the
`requirements.yml` files of the repository, when a file has changed since the last install. Tick
the box to skip that step for this task.

The box starts unticked whatever the template sets, and an unticked box installs the
requirements.

## Plan {#taskParams-plan}

Shown for a Terraform, OpenTofu or Terragrunt template.

Runs `plan` and ends the task there: nothing is applied. Use it to review the changes first.
Without it, the task goes on to apply the plan when it has changes.

## Destroy {#taskParams-destroy}

Shown for a Terraform, OpenTofu or Terragrunt template.

Adds `-destroy` to `plan` and `apply`: the task plans, and then applies, the removal of every
resource in the state of the workspace. Tick **Plan** too to see what would be destroyed without
destroying it.

## Auto Approve {#taskParams-autoApprove}

Shown for a Terraform, OpenTofu or Terragrunt template.

Applies the plan as soon as it is made. Without it, the task stops after the plan and waits for
someone to confirm or reject it on the page of the task.

The box has an effect only when the template allows auto approve. A template set to approve
always applies without asking, whatever the box says.

## Upgrade {#taskParams-upgrade}

Shown for a Terraform, OpenTofu or Terragrunt template.

Runs `init` with `-upgrade`: the providers and modules are upgraded to the newest versions their
version constraints allow.

## Reconfigure {#taskParams-reconfigure}

Shown for a Terraform, OpenTofu or Terragrunt template.

Runs `init` with `-reconfigure`: the backend is set up anew, and the saved configuration of the
backend is ignored. Without it, Semaphore runs `init -migrate-state`, which moves the existing
state to a changed backend.

## CLI args {#newTask-args}

Shown when the template allows overriding the CLI arguments.

Extra arguments for the command line of the task, one per chip: `--forks=10` is one chip,
`--forks` and `10` are two. Click a chip to edit it, drag it to change the order, and use the
last chip to add one.

The list starts with the arguments of the template. Semaphore always passes the arguments of the
template and adds this list after them, so once you change the list, the arguments of the
template left in it are passed twice. Remove them and keep only the arguments you add.
