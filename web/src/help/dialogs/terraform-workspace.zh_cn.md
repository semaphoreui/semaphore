---
form: src/components/TerraformInventoryForm.vue
---

# 新建 Terraform Workspace

工作区是 Terraform、OpenTofu 或 Terragrunt 模板的库存：它指定任务在 `init` 之后选择的工作区，以及
`init` 下载私有模块时使用的密钥。在此创建的工作区会附加到此模板，并列在其状态上方。模板的任务在其默认工作区中运行，
即带有对勾标记的那个；**Make default** 可以更改它。用铅笔按钮打开同一个对话框时，可以编辑工作区。

## Workspace name {#tfWorkspace-inventory}

Terraform 工作区的名称，例如 `production`。此字段为必填项。

在执行 plan 之前，Semaphore 会使用此名称运行 `workspace select -or-create=true`，因此当工作区尚不存在时会创建它。
如果 `workspace list` 失败，例如因为后端不支持工作区，Semaphore 会跳过选择这一步。

Semaphore 以 `TEMPLATE - NAME` 的形式保存工作区，即模板的名称加上此名称：您在库存页面以及将工作区附加到模板时看到的就是这个完整名称。
重命名工作区不会在 Terraform 中重命名它：下一个任务会选择或创建一个使用新名称的工作区。

## SSH key for private modules {#tfWorkspace-sshKeyId}

`init` 通过 SSH 从私有 Git 仓库下载模块时使用的密钥。在此对话框中此字段为必填项。列表中提供密钥存储中类型为
**SSH 密钥**、**使用密码登录**和**无**的密钥。

只有 **SSH 密钥**才会起作用：Semaphore 会在 `init` 期间将其加载到 SSH 代理中，Git 会将其用于
`git::ssh://git@example.com/modules/vpc.git` 这样的模块源。类型为**使用密码登录**的密钥不会向 Git
传递任何凭据。当模块不需要密钥时，请选择**无**。
