---
form: src/components/TaskForm.vue, src/components/TaskParamsAnsibleForm.vue, src/components/TaskParamsTerraformForm.vue
---

# 新任务

启动模板的一个任务：运行一次其剧本、Terraform 代码或脚本。模板决定此对话框显示哪些字段，下面的每个部分都说明了其字段何时出现。
底部的按钮根据模板的类型显示为**运行**、**构建**或**部署**。

当模板带有调查时，调查的问题是位于**消息**之后的字段。标有 `*` 的字段为必填项，Integer 字段只接受数字，默认答案已经填好。
Semaphore 将答案作为变量传递给任务：对于 Ansible 是额外变量，对于 Terraform 是 `-var`，对于脚本是 `name=value`
参数，对于设置为环境变量的问题则是环境变量。答案会替换模板变量组中的同名变量。秘密答案在输入时会被隐藏，并且不会随任务一起保存。

## 提交 {#newTask-commit}

当您重新运行一个较早的任务、且该任务运行时的提交已知时显示。

任务会检出仓库的这个提交，而不是分支的最新提交。关闭此框则运行最新的提交。

只有当模板允许覆盖分支时才会使用该提交，并且对于标记为 monorepo 的仓库永远不会使用。否则任务运行最新的提交。

## 构建版本 {#newTask-buildTaskId}

当模板的类型为**部署**时显示。

要部署的构建：构建模板最近 20 个成功任务中的一个，按版本和消息列出。默认选中最新的一个。此字段为必填项。

构建的版本会传递给任务：在 Ansible 中为 `semaphore_vars.task_details.incoming_version`，在脚本中为环境变量
`SEMAPHORE_TASK_DETAILS_INCOMING_VERSION`。当您无权查看构建模板时，列表为空。

## 消息 {#newTask-message}

可选。说明任务为何运行的备注，例如 `Hotfix for the login page`。它会与任务一起显示在任务历史中，不会改变运行中的任何内容。

剧本可以通过 `semaphore_vars.task_details.message` 读取它，脚本可以通过环境变量
`SEMAPHORE_TASK_DETAILS_MESSAGE` 读取它。

## 分支 {#newTask-gitBranch}

当模板允许覆盖分支时显示，但仓库为 monorepo 时除外。

任务检出的仓库分支或标签。此字段初始为模板的分支。如果为空，任务使用模板的分支；如果模板没有分支，则使用仓库的分支。
不是有效 Git 分支名称的名称会被拒绝。

## 库存 {#newTask-inventoryId}

当模板允许在任务中选择库存、并且您有权查看项目时显示。对于 Terraform、OpenTofu 或 Terragrunt 模板，此字段的标签为
**Workspace**，但服务器只对 Ansible 模板使用此选择：Terraform、OpenTofu 或 Terragrunt 任务会在模板的工作区中运行，无论您在此处选择什么。

任务运行所针对的库存。此字段初始为模板的库存，并列出项目中为模板的应用创建的库存。如果所选库存在任务启动时已不存在，
任务会针对模板的库存运行。

## 限制 {#taskParams-limit}

对于允许覆盖限制的 Ansible 模板显示。

剧本在其上运行的库存主机或组，每个标签一个模式，例如 `web` 或 `10.0.0.11`。Semaphore 将它们用逗号连接后作为
`--limit` 传递。列表初始为模板的限制。删除所有标签则在库存的所有主机上运行。

## 标签 {#taskParams-tags}

对于允许覆盖标签的 Ansible 模板显示。

只运行剧本中带有这些标签之一的任务和角色：Semaphore 将它们作为 `--tags` 传递，每个标签一个。列表初始为模板的标签。
删除所有标签则运行整个剧本。

## 跳过标签 {#taskParams-skipTags}

对于允许覆盖跳过标签的 Ansible 模板显示。

剧本中带有这些标签之一的任务和角色不会运行：Semaphore 将它们作为 `--skip-tags` 传递，每个标签一个。列表初始为模板的跳过标签。

## 调试 {#taskParams-debug}

对于允许调试的 Ansible 模板显示。

以详细级别标志运行 `ansible-playbook`，使任务日志显示更多细节。标签会显示该标志：`-v` 重复的次数由**调试级别**决定。

## 调试级别 {#taskParams-debugLevel}

与**调试**一起显示，在勾选**调试**时启用。

日志的详细程度，从 1（`-v`）到 6（`-vvvvvv`）。默认值为 4（`-vvvv`），它还会显示与主机连接的详细信息。

## 干运行 {#taskParams-dryRun}

对于 Ansible 模板显示，除非模板隐藏了它。

以 `--check` 运行剧本：Ansible 报告它会在主机上更改什么，但不做任何更改。

## 差异 {#taskParams-diff}

对于 Ansible 模板显示，除非模板隐藏了它。

以 `--diff` 运行剧本：日志会显示对主机上的文件和模板所做的更改。同时勾选**干运行**，即可在不进行更改的情况下查看这些更改。

## Skip Galaxy install {#taskParams-skipGalaxyInstall}

对于允许覆盖 Galaxy 安装的 Ansible 模板显示。

在剧本运行之前，如果自上次安装以来某个文件发生了变化，Semaphore 会安装仓库中 `requirements.yml` 文件所列出的角色和集合。
勾选此框可以为此任务跳过这一步。

无论模板如何设置，此框初始都不勾选，不勾选时会安装这些依赖。

## Plan {#taskParams-plan}

对于 Terraform、OpenTofu 或 Terragrunt 模板显示。

运行 `plan` 并在此结束任务：不应用任何内容。可用于先审查更改。不勾选时，如果 plan 有更改，任务会继续应用它。

## Destroy {#taskParams-destroy}

对于 Terraform、OpenTofu 或 Terragrunt 模板显示。

向 `plan` 和 `apply` 添加 `-destroy`：任务会规划并随后应用对工作区状态中每个资源的删除。同时勾选 **Plan**，即可在不销毁的情况下查看将要销毁的内容。

## Auto Approve {#taskParams-autoApprove}

对于 Terraform、OpenTofu 或 Terragrunt 模板显示。

plan 一生成就立即应用。不勾选时，任务会在 plan 之后停止，并等待有人在任务页面中确认或拒绝它。

只有当模板允许自动批准时，此框才会生效。设置为自动批准的模板总是不经询问就应用，无论此框如何设置。

## Upgrade {#taskParams-upgrade}

对于 Terraform、OpenTofu 或 Terragrunt 模板显示。

以 `-upgrade` 运行 `init`：提供程序和模块会升级到其版本约束所允许的最新版本。

## Reconfigure {#taskParams-reconfigure}

对于 Terraform、OpenTofu 或 Terragrunt 模板显示。

以 `-reconfigure` 运行 `init`：重新设置后端，并忽略已保存的后端配置。不勾选时，Semaphore 运行
`init -migrate-state`，它会将现有状态迁移到更改后的后端。

## CLI args {#newTask-args}

当模板允许覆盖 CLI 参数时显示。

任务命令行的额外参数，每个标签一个：`--forks=10` 是一个标签，`--forks` 和 `10` 是两个。点击标签可以编辑它，拖动可以更改顺序，
使用最后一个标签可以添加参数。

列表初始为模板的参数。Semaphore 总是传递模板的参数，并在其后添加此列表，因此一旦您更改了列表，列表中保留的模板参数会被传递两次。
请删除它们，只保留您添加的参数。
