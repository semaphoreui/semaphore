---
form: src/components/IntegrationForm.vue, src/components/TaskParamsForm.vue
---

# 编辑集成

当另一个服务调用 Semaphore 的 webhook URL（例如 GitHub 的 push 事件）时，集成会启动一个任务模板或一个工作流。该 URL
是一个别名，由 `/api/integrations/` 和一段随机代码组成：项目的别名在集成页面中，单个集成的别名在其自己的页面中。
该页面还包含集成的匹配器以及集成从请求中提取的值。

## Name {#integrationForm-name}

集成在集成页面中显示的名称。集成启动的任务在审计日志中会归属于它。此字段为必填项。

## 运行 {#integrationForm-targetToggle}

当服务器支持工作流时显示。

集成启动的内容：**Template** 启动任务模板的一个任务，**Workflow** 启动工作流的一次运行。切换到 **Workflow**
会丢弃下面的任务参数。

## Workflow to run {#integrationForm-workflowId}

当**运行**为 **Workflow** 时显示。

请求启动的工作流。集成从请求中提取的每个值都会作为同名调查变量的答案传递给工作流。工作流集成需要 Pro 版本：没有 Pro
版本时此字段处于禁用状态，服务器会拒绝此更改。

## Task Template to run {#integrationForm-templateId}

当**运行**为 **Template** 时显示。

请求为其启动任务的任务模板。此字段为必填项。选择模板之后，其任务的参数会显示在下方。

## 任务参数 {#integrationForm-taskParams}

在选择了任务模板时显示。

此集成启动的每个任务所获得的参数，就像您亲自从模板启动任务一样。这里只包含模板允许任务覆盖的选项。从请求中提取的值会覆盖这些参数：
类型为 **Variables** 的值覆盖同名的答案，类型为 **Task Params** 的值覆盖同名的参数。

当模板允许在任务中使用 CLI 参数时，**CLI args** 包含任务命令行的额外参数。

## 消息 {#taskParams-message}

可选。集成启动的每个任务的消息，显示在任务历史中。

## 调查变量 {#taskParams-survey}

当模板带有调查变量时显示。

模板调查的答案，对每个任务都相同。从请求中提取的类型为 **Variables** 的值会替换同名的答案。

秘密类型变量的答案不会随集成一起保存：请改为通过模板的变量组传递秘密值。

## 分支 {#taskParams-gitBranch}

当模板允许在任务中覆盖分支时显示。

任务检出的仓库分支或标签。保持原样则使用模板的分支。从请求中提取的、名为 `git_branch` 且类型为 **Task Params**
的值会覆盖它；该值开头的 `refs/heads/` 会被去除。

## 库存 {#taskParams-inventoryId}

当模板允许在任务中覆盖库存时显示。

任务运行所针对的库存（对于 Terraform 则是工作区），用来替代模板的库存。

## Ansible 选项 {#taskParams-ansible}

当模板运行 Ansible 剧本时显示。

每个任务的 `ansible-playbook` 选项，每个选项在模板允许时才会显示：**限制**、**标签**、**跳过标签**、**调试**、
**干运行**（`--check`）、**差异**和 **Skip Galaxy install**。

## Terraform 选项 {#taskParams-terraform}

当模板运行 Terraform、OpenTofu 或 Terragrunt 时显示。

- **Plan**：任务运行 plan 后停止，不应用任何内容。
- **Destroy**：plan 和 apply 会销毁资源。
- **Auto Approve**：当模板允许自动批准时，plan 会在无需确认的情况下被应用。不勾选时，任务会等待有人在 Semaphore
  中确认它。
- **Upgrade** 和 **Reconfigure**：`init` 以 `-upgrade` 和 `-reconfigure` 运行。

## Auth method {#integrationForm-authMethod}

Semaphore 如何检查请求是否来自您所期望的服务。未通过检查的请求不会启动任何内容，webhook 对它的响应与对其他任何请求的响应相同。

- **None**：不检查。任何知道该 URL 的人都可以启动任务。
- **GitHub Webhooks**：`X-Hub-Signature-256` 标头必须包含使用秘密值生成的正文 HMAC-SHA256 签名，与 GitHub
  发送的一致。
- **Bitbucket Webhooks**：对 `X-Hub-Signature` 标头进行相同的检查，与 Bitbucket 发送的一致。
- **Token**：**Auth header** 中指定的标头必须包含秘密值本身。
- **HMAC** 和 **HMAC (SHA-512)**：**Auth header** 中指定的标头必须包含使用秘密值生成的正文 HMAC-SHA256 或
  HMAC-SHA512，采用小写十六进制，且不带 `sha256=` 之类的前缀。
- **BasicAuth**：请求必须在其 `Authorization` 标头中携带**认证密钥**中密钥的登录名和密码。

## Auth header {#integrationForm-authHeader}

当 **Auth method** 为 **Token**、**HMAC** 或 **HMAC (SHA-512)** 时显示。

携带令牌或签名的请求标头的名称，例如 `X-Webhook-Token`。名称不区分大小写。不带该标头的请求会被拒绝。

## 认证密钥 {#integrationForm-authSecretId}

当 **Auth method** 不为 **None** 时显示。

用来检查请求的、来自密钥存储的密钥。对于 **BasicAuth**，只有类型为**使用密码登录**的密钥适用。对于其他方法，类型为**字符串**的密钥也适用：
检查使用其值，或者**使用密码登录**密钥的密码。

除 **None** 之外的每种方法都需要此字段：没有密钥或秘密值为空时，每个请求都会被拒绝。
