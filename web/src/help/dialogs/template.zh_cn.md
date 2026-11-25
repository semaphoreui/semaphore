---
form: src/components/TemplateForm.vue, src/components/SurveyVars.vue, src/components/TemplateVaults.vue
skip: surveyVarDialog-name, surveyVarDialog-title, surveyVarDialog-description, surveyVarDialog-type, surveyVarDialog-target, surveyVarDialog-valueName, surveyVarDialog-valueValue, surveyVarDialog-defaultValue, surveyVarDialog-required, templateVaultDialog-name, templateVaultDialog-type, templateVaultDialog-vaultKeyId, templateVaultDialog-script
---

# 编辑模板 'Ansible Playbook'

任务模板说明 Semaphore 运行什么以及如何运行：仓库及其中的文件、库存和变量，以及用户在启动任务时可以更改哪些选项。每个任务都从模板启动。
对话框的字段取决于模板的应用（Ansible、Terraform、OpenTofu、Terragrunt、Bash、Python、PowerShell 等）：只有部分应用才有的字段会加以说明。
更改对点击**保存**之后启动的任务生效。

## 模板类型 {#templateForm-typeTabs}

顶部的标签页设置模板的种类：

- **任务**：运行代码。大多数模板属于这一种。
- **构建**：生成工件。每个构建任务都会获得一个版本，从**起始版本**开始计数。
- **部署**：部署由构建生成的工件。启动它时，您需要选择要部署其版本的构建任务。

构建或部署任务会在其环境中获得变量 `SEMAPHORE_TASK_TYPE`、`SEMAPHORE_TASK_TARGET_VERSION`（构建）和
`SEMAPHORE_TASK_INCOMING_VERSION`（部署）。Ansible 会在额外变量 `semaphore_vars.task_details` 中获得相同的值。

## 起始版本 {#templateForm-startVersion}

当类型为**构建**时显示。必填。

第一个构建任务的版本，例如 `1.0.0` 或 `v0.1`。之后的每次构建都会将上一个版本的最后一个数字加一，并保留其周围的文本和补零：
`1.0.0`、`1.0.1`、`1.0.2`。将此处的数字提高到最后一次构建之上即可跳到更高的版本，例如 `2.0.0`；不含数字的版本永远不会变化。

## 构建模板 {#templateForm-buildTemplateId}

当类型为**部署**时显示。必填。

其任务生成此模板所部署工件的模板。列表包含项目的构建模板，以及位于其下方的部署模板：一个部署可以跟在另一个部署之后，
此时它会获得那个部署从其构建获得的版本。

## 自动运行 {#templateForm-autorun}

当类型为**部署**时显示。

每当**构建模板**的任务成功时，自动以该任务生成的版本启动此模板的一个任务。失败的构建不会启动任何内容。

## 名称 {#templateForm-name}

模板在任务模板页面中显示的名称。此字段为必填项。

## 仓库 {#templateForm-repositoryId}

项目中保存要运行的代码的仓库。此字段为必填项。在每个任务开始之前，Semaphore 都会更新其仓库副本并从中运行代码。

字段右侧的标签显示任务检出的分支。点击它可以打开**分支**字段。

## 分支 {#templateForm-gitBranch}

在点击**仓库**右侧的分支标签后显示。

此模板的任务检出的分支或标签，用来替代仓库中设置的分支。将此字段留空则使用仓库的分支。当 Semaphore 能够读取仓库的分支时，
此字段会列出这些分支。

当仓库为 monorepo 时，此字段处于禁用状态：所有模板共享该仓库的同一个检出副本，位于仓库的分支上。

## 剧本文件路径 {#templateForm-playbook}

要运行的内容，以从仓库根目录开始的路径表示，例如 `deploy/site.yml`。该路径不能是绝对路径，也不能指向仓库之外。
列表中提供仓库中的文件；您也可以输入不在列表中的路径。字段右侧的按钮会重新读取列表。

标签和含义随应用而变化：

- **Ansible**：剧本。此项为必填；列表中提供 `.yml` 和 `.yaml` 文件。
- **Terraform**、**OpenTofu**、**Terragrunt**：**Subdirectory path**，即包含代码的目录。留空则在仓库的根目录中运行代码。
  输入 `/` 可以列出目录中的内容。
- **Bash**、**Python**、**PowerShell** 和其他应用：**Script Filename**，即要运行的脚本。此项为必填。脚本在仓库的根目录中运行。

## Different working directory {#templateForm-workingDirectoryToggle}

对于 Ansible 模板显示。

在仓库根目录以外的目录中运行 `ansible-playbook`。勾选它即可显示 **Working directory** 字段。取消勾选则重新在根目录中运行。

## Working directory {#templateForm-workingDirectory}

当勾选了 **Different working directory** 时显示。必填。

`ansible-playbook` 和 `ansible-galaxy` 运行所在的目录，以从仓库根目录开始的路径表示，例如 `ansible`。当项目的
`ansible.cfg` 位于子目录中时使用它，因为 Ansible 从当前目录读取该文件。剧本路径仍然相对于仓库的根目录。该路径不能是绝对路径，
也不能指向仓库之外。

## 库存 {#templateForm-inventoryId}

对于 Ansible、Terraform、OpenTofu 和 Terragrunt 模板显示。

- **Ansible**：剧本运行所针对的库存，以及主机的凭据。此字段为必填项。
- **Terraform**、**OpenTofu**、**Terragrunt**：**Workspace**，即代码运行所在的工作区。Semaphore 会在 plan
  之前选择它，并在它不存在时创建它。保存时没有工作区的新模板会获得一个名为 `default` 的工作区。

列表包含项目中为模板的应用创建的库存。

## 变量组 {#templateForm-environmentIds}

可选。此模板的每个任务都会获得其变量、环境变量和秘密值的变量组。您可以选择多个：它们按字段中的顺序合并，当两个变量组设置了相同的名称时，
后面的变量组优先于前面的。

变量传递给代码的方式取决于应用：对于 Ansible 是额外变量，对于 Terraform 和 OpenTofu 是 `-var`，对于脚本是
`name=value` 参数。环境变量对每种应用都会设置。

## 视图 {#templateForm-viewId}

可选。模板所在的视图：视图是任务模板页面中的标签页，用于对项目的模板进行分组。将此字段留空则不将模板放入任何视图。

## 运行器标签 {#templateForm-runnerTag}

当项目运行器可用时显示。

此模板的任务只在带有此标签的运行器上运行，即使服务器本身可以运行它们。此标签会覆盖库存的运行器标签。将此字段留空则使用库存的标签，
或者不使用标签。

## Executor image {#templateForm-executorImage}

运行器为此模板的任务启动的容器镜像，例如 `semaphoreui/job:latest`。将此字段留空则使用运行器配置中的镜像。

只有 Docker 和 Kubernetes 执行器会在容器中运行任务，因此只有当其中之一可用时此字段才会启用。不在容器中运行的任务会忽略它。

## 调查变量 {#surveyVars}

此模板的任务启动时向用户询问的值。点击 **+** 添加变量，点击标签可以更改变量，点击标签上的叉号可以删除变量；拖动标签可以设置提问的顺序。

答案会像变量组的变量一样，或者作为环境变量（取决于每个变量的设置）传递给代码。它们优先于设置了相同名称的变量组。

## Allow parallel tasks {#templateForm-allowParallelTasks}

允许此模板的任务同时运行。不勾选时，模板的任务会在队列中等待，直到正在运行的任务结束。无论哪种情况，项目和服务器的并行任务上限都适用。

## Issue JWT to task runner {#templateForm-jwtEnabled}

当管理员已为服务器启用 JWT 签发时显示。

在环境变量 `SEMAPHORE_JWT` 中为此模板的每个任务提供一个由 Semaphore 签名的 JSON Web Token。代码可以向其他服务出示该令牌，
以证明自己是哪个任务：令牌中包含任务、项目、模板以及启动任务的用户。该服务使用 Semaphore 服务器
`/.well-known/jwks.json` 中的公钥检查签名。

## JWT audience {#templateForm-jwtAudience}

当勾选了 **Issue JWT to task runner** 时显示。可选。

令牌所面向的服务，写入其 `aud` 声明，例如 `https://vault.example.com`。每输入一个值后按 Enter。最多 32
个值；将此字段留空则生成不带受众的令牌。

## JWT TTL {#templateForm-jwtTtl}

当勾选了 **Issue JWT to task runner** 时显示。可选。

令牌从任务开始起保持有效的时长，以 `30s`、`15m` 或 `1h` 这样的时长表示。将此字段留空则使用服务器的默认值，除非管理员设置了其他值，
否则为一小时。该值不能超过服务器的最大值，除非另行设置，否则为 24 小时。

## 我想通过 cron 仅为某个仓库的新提交运行任务 {#templateForm-cronVisible}

当仓库中出现新提交时启动此模板的一个任务。Semaphore 会按所选的间隔向仓库查询其最新提交，并且只有当提交自上次检查以来发生变化时才启动任务。

勾选此框并保存会创建检查所用的计划；取消勾选并保存会删除该计划。模板的其他计划会保留。

## 仓库 {#templateForm-cronRepositoryId}

当勾选了**我想通过 cron 仅为某个仓库的新提交运行任务**时显示。必填。

要监视新提交的仓库。Semaphore 监视的是该仓库中设置的分支，而不是模板的**分支**。

## 检查间隔 {#templateForm-cronFormat}

当勾选了**我想通过 cron 仅为某个仓库的新提交运行任务**时显示。

Semaphore 检查仓库是否有新提交的频率：从每分钟一次到每天一次。默认值为每分钟一次。

## 抑制成功警报 {#templateForm-suppressSuccessAlerts}

阻止此模板中成功的任务发送警报。失败任务的警报仍会发送。只有当项目启用了警报时才生效。

## 抑制错误警报 {#templateForm-suppressErrorAlerts}

阻止此模板中失败的任务发送警报。成功任务的警报仍会发送。只有当项目启用了警报时才生效。

## CLI args {#templateForm-args}

添加到每个任务命令中的参数。每个标签是一个参数，按原样传递，不经过 shell：将 `--forks=10` 放在一个标签中，或者将
`--forks` 和 `10` 放在两个标签中。拖动标签可以更改其顺序。列表右侧的按钮会以 JSON 数组的形式打开参数：

```json
["--forks=10", "-e", "env=staging"]
```

- **Ansible**：添加到 `ansible-playbook` 中，位于剧本之前。
- **Terraform**、**OpenTofu**、**Terragrunt**：添加到 `plan` 和 `apply` 中，不添加到 `init` 中。
- 脚本：添加在脚本名称之后。

## CLI 参数 {#templateForm-allowOverrideArgsInTask}

位于**提示**下。允许用户在启动任务时添加自己的参数。这些参数添加在模板的参数之后。

## 分支 {#templateForm-allowOverrideBranchInTask}

位于**提示**下。允许用户在启动任务时选择另一个分支、标签或提交，以替代模板的分支。不勾选时，任务始终检出模板的分支。

当仓库为 monorepo 时，此框处于禁用状态。

## 库存 {#templateForm-allowOverrideInventory}

位于**提示**下。对于 Ansible、Terraform、OpenTofu 和 Terragrunt 模板显示。

允许用户在启动任务时选择另一个库存。只有 Ansible 模板会使用在那里选择的库存：即使勾选了此框，Terraform、OpenTofu 或
Terragrunt 任务也会在模板的工作区中运行。

## 限制 {#templateForm-limit}

对于 Ansible 模板显示。

从库存中选出的、运行剧本的主机或组，以 `--limit` 传递。每个标签是一个模式，例如 `web` 或 `db01`；将列表留空则在每台主机上运行。

## 标签 {#templateForm-tags}

对于 Ansible 模板显示。

只运行带有这些标签的任务和角色，以 `--tags` 传递。将列表留空则运行所有内容。

## 跳过标签 {#templateForm-skipTags}

对于 Ansible 模板显示。

跳过带有这些标签的任务和角色，以 `--skip-tags` 传递。

## 保管库 {#templateVaults}

对于 Ansible 模板显示。

剧本解密其文件所需的 Ansible Vault 密码。点击 **+** 添加一个，点击标签可以更改它，点击标签上的叉号可以删除它。
每个保管库都有一个名称，即其 vault ID，并从密钥存储中的密钥或仓库中的客户端脚本提供密码。

## 自动批准 {#templateForm-autoApprove}

对于 Terraform、OpenTofu 和 Terragrunt 模板显示。

不经询问就应用 plan。不勾选时，任务会在 plan 之后停止，并等待用户确认或拒绝；没有更改的 plan 会直接结束任务，没有可应用的内容。

## Override backend settings {#templateForm-overrideBackend}

当 Semaphore 能够保存 Terraform 状态时，对于 Terraform、OpenTofu 和 Terragrunt 模板显示。

将代码的状态保存在 Semaphore 中。在 `init` 之前，Semaphore 会在代码目录中写入一个设置 `http` 后端的文件，并在任务结束后删除它：

```hcl
terraform {
  backend "http" {
  }
}
```

## Backend filename {#templateForm-backendFilename}

与 **Override backend settings** 一起显示；勾选该框时启用。可选。

Semaphore 写入后端设置的文件名，例如 `semaphore_backend.tf`。默认值为 `backend.tf`。名称必须以 `.tf`
结尾，并且只能包含字母、数字、`_`、`-` 和 `.`。请选择一个您的代码中没有的文件名：Semaphore 会覆盖该文件，并在任务结束后删除它。

## 限制 {#templateForm-allowOverrideLimit}

位于 **Ansible 提示**下。允许用户在启动任务时设置要运行的主机，以替代模板的**限制**。

## 标签 {#templateForm-allowOverrideTags}

位于 **Ansible 提示**下。允许用户在启动任务时设置标签，以替代模板的**标签**。

## 跳过标签 {#templateForm-allowOverrideSkipTags}

位于 **Ansible 提示**下。允许用户在启动任务时设置要跳过的标签，以替代模板的**跳过标签**。

## 调试 {#templateForm-allowDebug}

位于 **Ansible 提示**下。允许用户在启动任务时开启详细输出，从 `-v` 到 `-vvvvvv`。

## 干运行 {#templateForm-showDryRun}

位于 **Ansible 提示**下。默认勾选。在启动任务时提供**干运行**选项，它以 `--check` 运行剧本：Ansible
报告它会更改什么，但不做任何更改。取消勾选此框则移除该选项。

## 差异 {#templateForm-showDiff}

位于 **Ansible 提示**下。默认勾选。在启动任务时提供**差异**选项，它以 `--diff` 运行剧本以显示对文件所做的更改。
取消勾选此框则移除该选项。

## 自动批准 {#templateForm-allowAutoApprove}

位于 Terraform、OpenTofu 和 Terragrunt 的提示下。允许用户在启动任务时勾选 **Auto Approve**，以便在无需确认的情况下应用
plan。任务对话框无论如何都会提供该选项，但没有勾选此框时 Semaphore 会忽略它：除非勾选了模板的**自动批准**，否则任务会在
plan 之后等待确认。

## Galaxy install options {#templateForm-galaxy}

对于 Ansible 模板显示。点击它可以显示 Galaxy 安装的选项；标签会统计与默认值不同的选项数量。

在剧本运行之前，Semaphore 会使用 `ansible-galaxy` 安装仓库所需的角色和集合。它会在剧本旁边以及仓库的根目录中查找
`requirements.yml`、`roles/requirements.yml` 和 `collections/requirements.yml`，并且只有当某个文件自上一个任务以来发生了变化时才运行安装。

## Skip Galaxy install {#templateForm-skipGalaxyInstall}

位于 **Galaxy install options** 下。跳过 Galaxy 安装：剧本使用之前安装的角色和集合运行。当依赖以其他方式安装，或者安装很慢且不需要时使用它。

## Role install args {#templateForm-galaxyRoleArgs}

位于 **Galaxy install options** 下。添加到 `ansible-galaxy role install` 中的参数，每个标签一个，例如
`--ignore-certs` 或 `--server=https://galaxy.example.com`。Semaphore 只接受可以安全传递的安装标志：未知的标志、
`--token`、`--api-key`、`-p` 或 `--roles-path` 会导致模板无法保存。

## Collection install args {#templateForm-galaxyCollectionArgs}

位于 **Galaxy install options** 下。添加到 `ansible-galaxy collection install` 中的参数，每个标签一个，例如
`--pre` 或 `--upgrade`。适用与 **Role install args** 相同的规则：不接受 `--token`、`--api-key`、`-p` 和
`--collections-path`。

## Skip Galaxy install {#templateForm-allowOverrideSkipGalaxyInstall}

位于**提示**中的 **Galaxy install options** 下。允许用户在启动任务时选择是否跳过 Galaxy 安装。任务的选择优先于模板的
**Skip Galaxy install**。
