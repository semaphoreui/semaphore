---
form: src/components/IntegrationForm.vue, src/components/TaskParamsForm.vue
---

# Edit Run

A run is what an integration starts when another service calls a webhook URL of Semaphore, such as a
push event of GitHub: a task of a task template, or a workflow. The URL and the check of the request
belong to the endpoint of the integration, not to the run. The page of a run holds its matchers and
the values it takes from the request.

## Name {#integrationForm-name}

The name the integration is listed under on the Integrations page. Tasks the integration starts
are attributed to it in the audit log. The field is required.

## Integration {#integrationForm-groupId}

The integration the run belongs to. The run starts on the requests that reach the endpoint of that
integration and pass its authorization. The field is required.

## Target {#integrationForm-targetToggle}

Shown when the server supports workflows.

What the integration starts: **Template** starts a task of a task template, **Workflow** starts a
run of a workflow. Switching to **Workflow** drops the task parameters below.

## Workflow to run {#integrationForm-workflowId}

Shown when **Target** is **Workflow**.

The workflow a request starts. Every value the integration takes from the request is passed to
the workflow as the answer to its survey variable of the same name. Workflow integrations need
the Pro edition: without it the field is disabled and the server refuses the change.

## Task Template to run {#integrationForm-templateId}

Shown when **Target** is **Template**.

The task template a request starts a task of. The field is required. Once a template is picked,
the parameters of its tasks appear below it.

## Task parameters {#integrationForm-taskParams}

Shown when a task template is picked.

The parameters every task this integration starts gets, as if you started it from the template
yourself. Only the options the template lets a task override are here. Values taken from the
request override these: a value of the type **Variables** an answer of the same name, a value of
the type **Task Params** a parameter of the same name.

When the template allows CLI args in a task, **CLI args** holds extra arguments for the command
line of the task.

## Message {#taskParams-message}

Optional. The message of every task the integration starts, shown in the history of tasks.

## Survey Variables {#taskParams-survey}

Shown when the template has survey variables.

The answers to the survey of the template, the same for every task. A value of the type
**Variables** taken from the request replaces the answer of the same name.

Answers to variables of the type secret are not saved with the integration: pass a secret
through the variable group of the template instead.

## Branch / Tag {#taskParams-gitBranch}

Shown when the template allows to override the branch in a task.

The branch or tag of the repository the tasks check out. Leave it as it is to use the branch of
the template. A value named `git_branch` of the type **Task Params**, taken from the request,
overrides it; a leading `refs/heads/` is removed from that value.

## Inventory {#taskParams-inventoryId}

Shown when the template allows to override the inventory in a task.

The inventory, or for Terraform the workspace, the tasks run against instead of the one of the
template.

## Ansible options {#taskParams-ansible}

Shown when the template runs an Ansible playbook.

The options of `ansible-playbook` for every task, each one shown when the template allows it:
**Limit**, **Tags**, **Skip tags**, **Debug**, **Dry Run** (`--check`), **Diff** and **Skip Galaxy
install**.

## Terraform options {#taskParams-terraform}

Shown when the template runs Terraform, OpenTofu or Terragrunt.

- **Plan**: the task runs the plan and stops, nothing is applied.
- **Destroy**: the plan and the apply destroy the resources.
- **Auto Approve**: the plan is applied without a confirmation, when the template allows auto
  approve. Without it a task waits until someone confirms it in Semaphore.
- **Upgrade** and **Reconfigure**: `init` runs with `-upgrade` and `-reconfigure`.
