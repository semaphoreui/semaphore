---
form: src/components/TemplateForm.vue, src/components/SurveyVars.vue, src/components/TemplateVaults.vue
skip: surveyVarDialog-name, surveyVarDialog-title, surveyVarDialog-description, surveyVarDialog-type, surveyVarDialog-target, surveyVarDialog-valueName, surveyVarDialog-valueValue, surveyVarDialog-defaultValue, surveyVarDialog-required, templateVaultDialog-name, templateVaultDialog-type, templateVaultDialog-vaultKeyId, templateVaultDialog-script
---

# Edit Template 'Ansible Playbook'

A task template says what Semaphore runs and how: the repository and the file in it, the
inventory and the variables, and which options a user may change when a task starts. Every task
is started from a template. The fields of the dialog depend on the app of the template (Ansible,
Terraform, OpenTofu, Terragrunt, Bash, Python, PowerShell and others): a field that only some
apps have says so. Changes take effect for tasks started after **Save**.

## Template type {#templateForm-typeTabs}

The tabs at the top set what kind of template it is:

- **Task**: runs the code. Most templates are of this kind.
- **Build**: makes an artifact. Every build task gets a version, counted from **Start Version**.
- **Deploy**: deploys an artifact made by a build. When you start it you pick the build task
  whose version it deploys.

A build or deploy task gets the variables `SEMAPHORE_TASK_TYPE`, `SEMAPHORE_TASK_TARGET_VERSION`
(build) and `SEMAPHORE_TASK_INCOMING_VERSION` (deploy) in its environment. Ansible gets the same
values in the extra variable `semaphore_vars.task_details`.

## Start Version {#templateForm-startVersion}

Shown when the type is **Build**. Required.

The version of the first build task, for example `1.0.0` or `v0.1`. Every next build increases the
last number of the previous version by one and keeps the text around it and its zero padding:
`1.0.0`, `1.0.1`, `1.0.2`. Raise the number here above the last build to jump ahead, for example to
`2.0.0`; a version with no number in it never changes.

## Build Template {#templateForm-buildTemplateId}

Shown when the type is **Deploy**. Required.

The template whose tasks make the artifacts this template deploys. The list holds the build
templates of the project and, below them, the deploy templates: a deploy can follow another
deploy, and it then gets the version that deploy got from its build.

## Autorun {#templateForm-autorun}

Shown when the type is **Deploy**.

Starts a task of this template by itself every time a task of the **Build Template** succeeds,
with the version that task made. A failed build starts nothing.

## Name {#templateForm-name}

The name the template is listed under on the Task Templates page. The field is required.

## Repository {#templateForm-repositoryId}

The repository of the project that holds the code to run. The field is required. Before every
task Semaphore updates its copy of the repository and runs the code from there.

The label right of the field shows the branch the tasks check out. Click it to open the
**Branch / Tag** field.

## Branch / Tag {#templateForm-gitBranch}

Shown after a click on the branch label right of **Repository**.

The branch or tag the tasks of this template check out, instead of the branch set on the
repository. Leave the field empty to use the branch of the repository. When Semaphore can read
the branches of the repository, the field lists them.

The field is disabled when the repository is a monorepo: all templates share one checkout of it,
on the branch of the repository.

## Path to playbook file {#templateForm-playbook}

What to run, as a path from the root of the repository, for example `deploy/site.yml`. The path
can not be absolute and can not lead out of the repository. The list offers the files of the
repository; you can also type a path that is not in the list. The button right of the field reads
the list again.

The label and the meaning change with the app:

- **Ansible**: the playbook. It is required; the list offers `.yml` and `.yaml` files.
- **Terraform**, **OpenTofu**, **Terragrunt**: **Subdirectory path**, the directory with the code.
  Leave it empty to run the code in the root of the repository. Type `/` to list what a directory
  holds.
- **Bash**, **Python**, **PowerShell** and other apps: **Script Filename**, the script to run. It
  is required. The script runs in the root of the repository.

## Different working directory {#templateForm-workingDirectoryToggle}

Shown for Ansible templates.

Runs `ansible-playbook` in a directory other than the root of the repository. Tick it to show
the **Working directory** field. Clear it to run in the root again.

## Working directory {#templateForm-workingDirectory}

Shown when **Different working directory** is ticked. Required.

The directory `ansible-playbook` and `ansible-galaxy` run in, as a path from the root of the
repository, for example `ansible`. Use it when the `ansible.cfg` of the project lives in a
subdirectory, since Ansible reads the file from the current directory. The playbook path stays
relative to the root of the repository. The path can not be absolute and can not lead out of the
repository.

## Inventory {#templateForm-inventoryId}

Shown for Ansible, Terraform, OpenTofu and Terragrunt templates.

- **Ansible**: the inventory the playbook runs against, with the credentials for the hosts. The
  field is required.
- **Terraform**, **OpenTofu**, **Terragrunt**: **Workspace**, the workspace the code runs in.
  Semaphore selects it before the plan and creates it when it does not exist. A new template saved
  without a workspace gets one named `default`.

The list holds the inventories of the project made for the app of the template.

## Variable Groups {#templateForm-environmentIds}

Optional. The variable groups whose variables, environment variables and secrets every task of
this template gets. You may pick several: they are merged in the order of the field, and a later
group wins over an earlier one when both set the same name.

How the variables reach the code depends on the app: extra variables for Ansible, `-var` for
Terraform and OpenTofu, `name=value` arguments for a script. Environment variables are set for
every app.

## View {#templateForm-viewId}

Optional. The view the template is listed in: views are the tabs of the Task Templates page,
which group the templates of a project. Leave the field empty to put the template in no view.

## Runner tag {#templateForm-runnerTag}

Shown when project runners are available.

Tasks of this template run only on a runner that carries this tag, even when the server could run
them itself. The tag overrides the runner tag of the inventory. Leave the field empty to use the
tag of the inventory, or none.

## Executor image {#templateForm-executorImage}

The container image the runner starts for the tasks of this template, for example
`semaphoreui/job:latest`. Leave the field empty to use the image of the runner configuration.

Only the Docker and Kubernetes executors run tasks in a container, so the field is enabled only
when one of them is available. A task that runs without a container ignores it.

## Survey Variables {#surveyVars}

The values a user is asked for when a task of this template starts. Click **+** to add a
variable, a chip to change one, the cross on a chip to remove it; drag the chips to set the
order the questions are asked in.

The answers reach the code like the variables of a variable group, or as environment variables,
as each variable sets. They win over a variable group that sets the same name.

## Allow parallel tasks {#templateForm-allowParallelTasks}

Lets tasks of this template run at the same time. Without it a task of the template waits in the
queue until the running one ends. The limits of parallel tasks of the project and of the server
apply in either case.

## Issue JWT to task runner {#templateForm-jwtEnabled}

Shown when the administrator has turned JWT issuance on for the server.

Gives every task of this template a JSON Web Token signed by Semaphore, in the environment
variable `SEMAPHORE_JWT`. The code can show the token to another service to prove which task it
is: the token names the task, the project, the template and the user who started it. The service
checks the signature with the public keys at `/.well-known/jwks.json` of the Semaphore server.

## JWT audience {#templateForm-jwtAudience}

Shown when **Issue JWT to task runner** is ticked. Optional.

The services the token is meant for, written to its `aud` claim, for example
`https://vault.example.com`. Press Enter after each value. Up to 32 values; leave the field empty
for a token without an audience.

## JWT TTL {#templateForm-jwtTtl}

Shown when **Issue JWT to task runner** is ticked. Optional.

How long the token stays valid from the start of the task, as a duration such as `30s`, `15m` or
`1h`. Leave the field empty to use the default of the server, one hour unless the administrator
set another one. The value can not exceed the maximum of the server, 24 hours unless set
otherwise.

## Auto-run task if new git commit have been found {#templateForm-cronVisible}

Starts a task of this template when a new commit appears in a repository. Semaphore asks the
repository for its last commit at the chosen interval and starts a task only when the commit
changed since the last check.

Ticking the box and saving creates the schedule of the check; clearing it and saving deletes
that schedule. The other schedules of the template are kept.

## Repository {#templateForm-cronRepositoryId}

Shown when **Auto-run task if new git commit have been found** is ticked. Required.

The repository to watch for new commits. Semaphore watches the branch set on that repository,
not the **Branch / Tag** of the template.

## checkInterval {#templateForm-cronFormat}

Shown when **Auto-run task if new git commit have been found** is ticked.

How often Semaphore checks the repository for a new commit: from every minute to once a day. The
default is every minute.

## Suppress success notifications {#templateForm-suppressSuccessAlerts}

Keeps the tasks of this template that succeed from sending alerts. The alerts of failed tasks are
still sent. It has an effect only when alerts are on for the project.

## Suppress error notifications {#templateForm-suppressErrorAlerts}

Keeps the tasks of this template that fail from sending alerts. The alerts of successful tasks
are still sent. It has an effect only when alerts are on for the project.

## CLI args {#templateForm-args}

Arguments added to the command of every task. Each chip is one argument, passed as it is,
without a shell: put `--forks=10` in one chip, or `--forks` and `10` in two. Drag the chips to
change their order. The button right of the list opens the arguments as a JSON array:

```json
["--forks=10", "-e", "env=staging"]
```

- **Ansible**: added to `ansible-playbook`, before the playbook.
- **Terraform**, **OpenTofu**, **Terragrunt**: added to `plan` and `apply`, not to `init`.
- A script: added after the script name.

## CLI args {#templateForm-allowOverrideArgsInTask}

Under **Prompts**. Lets the user add arguments of their own when a task starts. They are added
after the arguments of the template.

## Branch {#templateForm-allowOverrideBranchInTask}

Under **Prompts**. Lets the user pick another branch, tag or commit when a task starts, instead
of the branch of the template. Without it the task always checks out the branch of the template.

The box is disabled when the repository is a monorepo.

## Inventory {#templateForm-allowOverrideInventory}

Under **Prompts**. Shown for Ansible, Terraform, OpenTofu and Terragrunt templates.

Lets the user pick another inventory when a task starts. Only Ansible templates use the inventory
picked there: a Terraform, OpenTofu or Terragrunt task runs in the workspace of the template even
when the box is ticked.

## Limit {#templateForm-limit}

Shown for Ansible templates.

The hosts or groups to run the playbook on, out of the inventory, passed as `--limit`. Each chip
is one pattern, for example `web` or `db01`; leave the list empty to run on every host.

## Tags {#templateForm-tags}

Shown for Ansible templates.

Runs only the tasks and roles with these tags, passed as `--tags`. Leave the list empty to run
everything.

## Skip tags {#templateForm-skipTags}

Shown for Ansible templates.

Skips the tasks and roles with these tags, passed as `--skip-tags`.

## Vaults {#templateVaults}

Shown for Ansible templates.

The Ansible Vault passwords the playbook needs to decrypt its files. Click **+** to add one, a
chip to change it, the cross on a chip to remove it. Each vault has a name, its vault ID, and
gives the password from a key of the Key Store or from a client script of the repository.

## Auto approve {#templateForm-autoApprove}

Shown for Terraform, OpenTofu and Terragrunt templates.

Applies the plan without asking. Without it a task stops after the plan and waits until a user
confirms or rejects it; a plan without changes ends the task with nothing to apply.

## Override backend settings {#templateForm-overrideBackend}

Shown for Terraform, OpenTofu and Terragrunt templates when Semaphore can keep the Terraform
state.

Keeps the state of the code in Semaphore. Before `init`, Semaphore writes a file that sets the
`http` backend into the directory of the code, and deletes it after the task:

```hcl
terraform {
  backend "http" {
  }
}
```

## Backend filename {#templateForm-backendFilename}

Shown with **Override backend settings**; enabled when that box is ticked. Optional.

The name of the file Semaphore writes the backend into, for example `semaphore_backend.tf`. The
default is `backend.tf`. The name must end with `.tf` and hold letters, digits, `_`, `-` and `.`
only. Pick a name no file of your code has: Semaphore overwrites that file and deletes it after
the task.

## Limit {#templateForm-allowOverrideLimit}

Under **Ansible prompts**. Lets the user set the hosts to run on when a task starts, instead of
the **Limit** of the template.

## Tags {#templateForm-allowOverrideTags}

Under **Ansible prompts**. Lets the user set the tags when a task starts, instead of the **Tags**
of the template.

## Skip tags {#templateForm-allowOverrideSkipTags}

Under **Ansible prompts**. Lets the user set the tags to skip when a task starts, instead of the
**Skip tags** of the template.

## Debug {#templateForm-allowDebug}

Under **Ansible prompts**. Lets the user turn on verbose output when a task starts, from `-v` to
`-vvvvvv`.

## Dry Run {#templateForm-showDryRun}

Under **Ansible prompts**. Ticked by default. Offers the **Dry Run** option when a task starts,
which runs the playbook with `--check`: Ansible reports what it would change and changes nothing.
Clear the box to take the option away.

## Diff {#templateForm-showDiff}

Under **Ansible prompts**. Ticked by default. Offers the **Diff** option when a task starts,
which runs the playbook with `--diff` to show the changes made to files. Clear the box to take
the option away.

## Auto approve {#templateForm-allowAutoApprove}

Under the prompts of Terraform, OpenTofu and Terragrunt. Lets the user tick **Auto Approve** when
a task starts, to apply the plan without a confirmation. The task dialog offers that option in
any case, but without this box Semaphore ignores it: the task waits for a confirmation after the
plan, unless **Auto approve** of the template is ticked.

## Galaxy install options {#templateForm-galaxy}

Shown for Ansible templates. Click it to show the options of the Galaxy install; the label counts
the options that differ from the defaults.

Before the playbook runs, Semaphore installs the roles and collections the repository asks for
with `ansible-galaxy`. It looks for `requirements.yml`, `roles/requirements.yml` and
`collections/requirements.yml` next to the playbook and in the root of the repository, and runs
the install only when a file changed since the last task.

## Skip Galaxy install {#templateForm-skipGalaxyInstall}

Under **Galaxy install options**. Skips the Galaxy install: the playbook runs with the roles and
collections installed before. Use it when the requirements are installed another way or the
install is slow and not needed.

## Role install args {#templateForm-galaxyRoleArgs}

Under **Galaxy install options**. Arguments added to `ansible-galaxy role install`, one per chip,
for example `--ignore-certs` or `--server=https://galaxy.example.com`. Semaphore accepts only the
flags of the install that are safe to pass: an unknown flag, `--token`, `--api-key`, `-p` or
`--roles-path` makes the template fail to save.

## Collection install args {#templateForm-galaxyCollectionArgs}

Under **Galaxy install options**. Arguments added to `ansible-galaxy collection install`, one per
chip, for example `--pre` or `--upgrade`. The same rule applies as for **Role install args**:
`--token`, `--api-key`, `-p` and `--collections-path` are not accepted.

## Skip Galaxy install {#templateForm-allowOverrideSkipGalaxyInstall}

Under **Galaxy install options**, in **Prompts**. Lets the user choose when a task starts whether
to skip the Galaxy install. The choice of the task wins over **Skip Galaxy install** of the
template.
