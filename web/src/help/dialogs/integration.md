---
form: src/components/IntegrationForm.vue, src/components/TaskParamsForm.vue
---

# Edit Integration

An integration starts a task template, or a workflow, when another service calls a webhook URL
of Semaphore, such as a push event of GitHub. The URL is an alias, `/api/integrations/` and a
random code: the aliases of the project are on the Integrations page, the aliases of one
integration on its own page. That page also holds the matchers and the values the integration
takes from the request.

## Name {#integrationForm-name}

The name the integration is listed under on the Integrations page. Tasks the integration starts
are attributed to it in the audit log. The field is required.

## Runs {#integrationForm-targetToggle}

Shown when the server supports workflows.

What the integration starts: **Template** starts a task of a task template, **Workflow** starts a
run of a workflow. Switching to **Workflow** drops the task parameters below.

## Workflow to run {#integrationForm-workflowId}

Shown when **Runs** is **Workflow**.

The workflow a request starts. Every value the integration takes from the request is passed to
the workflow as the answer to its survey variable of the same name. Workflow integrations need
the Pro edition: without it the field is disabled and the server refuses the change.

## Task Template to run {#integrationForm-templateId}

Shown when **Runs** is **Template**.

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

## Auth method {#integrationForm-authMethod}

How Semaphore checks that a request comes from the service you expect. A request that fails the
check starts nothing, and the webhook answers it the same as any other request.

- **None**: no check. Anyone who knows the URL can start the task.
- **GitHub Webhooks**: the `X-Hub-Signature-256` header must hold the HMAC-SHA256 signature of
  the body made with the secret, as GitHub sends it.
- **Bitbucket Webhooks**: the same check on the `X-Hub-Signature` header, as Bitbucket sends it.
- **Token**: the header in **Auth header** must hold the secret itself.
- **HMAC** and **HMAC (SHA-512)**: the header in **Auth header** must hold the HMAC-SHA256 or
  HMAC-SHA512 of the body made with the secret, in lowercase hexadecimal, without a prefix
  such as `sha256=`.
- **BasicAuth**: the request must carry the login and the password of the key in
  **Auth secret** in its `Authorization` header.

## Auth header {#integrationForm-authHeader}

Shown when **Auth method** is **Token**, **HMAC** or **HMAC (SHA-512)**.

The name of the request header that carries the token or the signature, for example
`X-Webhook-Token`. The case of the name does not matter. A request without that header is
refused.

## Auth secret {#integrationForm-authSecretId}

Shown when **Auth method** is not **None**.

The key from the Key Store the request is checked against. For **BasicAuth** only a key of the
type **Login with password** fits. For the other methods a key of the type **String** fits too:
the check uses its value, or the password of a **Login with password** key.

The field is needed by every method but **None**: without a key, or with an empty secret, every
request is refused.
