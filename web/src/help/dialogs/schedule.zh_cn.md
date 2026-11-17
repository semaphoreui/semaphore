---
form: src/components/ScheduleForm.vue, src/components/TaskParamsForm.vue, src/components/TaskParamsAnsibleForm.vue, src/components/TaskParamsTerraformForm.vue
---

# 编辑计划

计划会在无人点击**运行**的情况下启动任务模板的任务：按照 cron 表达式反复启动，或者在设定的时刻启动一次。任务会获得在此对话框中保存的参数。
所有时间都采用 Semaphore 服务器的时区，显示在时间字段旁边：除非管理员设置了 `schedule.timezone` 或
`SEMAPHORE_SCHEDULE_TIMEZONE`，否则为 `UTC`。

## Name {#scheduleForm-name}

计划在计划页面中显示的名称。此字段为必填项。名称不会改变计划所做的事情。

## Template {#scheduleForm-templateId}

计划启动的任务模板。此字段为必填项。任务的运行方式与从模板启动时相同，使用模板自身的库存、仓库、变量组和其他设置，
并在模板允许任务更改的地方使用下面的参数。

## 任务参数 {#scheduleForm-taskParams}

在选择了模板时显示。

计划启动的每个任务的参数。出现哪些字段取决于模板：其调查变量，以及它允许任务更改的选项，例如分支、库存、**限制**、
**标签**、**跳过标签**或 **CLI args**，以及 Ansible 或 Terraform、OpenTofu 和 Terragrunt 的选项。

每次计划触发时，服务器都会检查模板：一旦模板不再允许任务更改某个值，此处保存的该值就会被忽略。

## 消息 {#taskParams-message}

可选。随计划启动的每个任务一起保存并显示在任务历史中的备注，例如 `Nightly deploy`。它不会改变任务的运行方式。

## 调查变量 {#taskParams-survey}

当模板带有调查变量时显示。

计划任务所使用的模板调查变量的值。它们传递给任务的方式与有人手动填写调查时相同。必需的变量必须有值，计划才能保存。

对话框打开时会复制模板的默认值，并随计划一起保存。之后在模板中更改的默认值不会影响之前保存的计划。

秘密类型变量的值不会随计划一起保存：计划任务在运行时不带该值。

## 分支 {#taskParams-gitBranch}

当模板允许任务覆盖分支时显示。

计划任务检出的仓库分支或标签。此字段初始为模板的分支。只有在模板允许覆盖时，服务器才会使用该值。

## 库存 {#taskParams-inventoryId}

当模板允许任务选择库存时显示。

计划任务运行所针对的库存，用来替代模板的库存。对于 Terraform 或 OpenTofu 模板，此字段名为 **Workspace**，但服务器只对
Ansible 模板应用此选择，对于其他模板则保留模板的工作区。

## 调试 {#taskParams-debug}

对于允许调试的 Ansible 模板显示。

以更多输出运行 `ansible-playbook`：`-v` 重复的次数由**调试级别**设置。

## 调试级别 {#taskParams-debugLevel}

对于允许调试的 Ansible 模板显示；在勾选**调试**时生效。

输出的详细程度，从 `1`（`-v`）到 `6`（`-vvvvvv`）。默认值为 `4`。

## 干运行 {#taskParams-dryRun}

对于 Ansible 模板显示，除非模板隐藏了它。

以 `--check` 运行剧本：Ansible 报告它会更改什么，但不做任何更改。

## 差异 {#taskParams-diff}

对于 Ansible 模板显示，除非模板隐藏了它。

以 `--diff` 运行剧本：Ansible 会显示它对文件和模板所做的更改。

## Skip Galaxy install {#taskParams-skipGalaxyInstall}

对于允许任务更改此项的 Ansible 模板显示。

在剧本运行之前跳过安装 `requirements.yml` 文件中的角色和集合。当依赖已经安装、且任务不应访问 Ansible Galaxy 时，请勾选此项。

## Plan {#taskParams-plan}

对于 Terraform、OpenTofu 或 Terragrunt 模板显示。

只运行 `plan`。任务在 plan 之后结束，不应用任何内容。

## Destroy {#taskParams-destroy}

对于 Terraform、OpenTofu 或 Terragrunt 模板显示。

向 plan 和 apply 添加 `-destroy`：任务会销毁工作区的资源，而不是创建或更新它们。

## Auto Approve {#taskParams-autoApprove}

对于 Terraform、OpenTofu 或 Terragrunt 模板显示。

不等待确认就应用 plan。只有当模板允许自动批准时才生效；设置为自动批准每个任务的模板在任何情况下都会应用 plan。

不使用自动批准时，计划任务会在 plan 之后停止，并等待有人在任务页面中确认或拒绝它。

## Upgrade {#taskParams-upgrade}

对于 Terraform、OpenTofu 或 Terragrunt 模板显示。

以 `-upgrade` 运行 `init`：提供程序和模块会更新到配置所允许的最新版本。

## Reconfigure {#taskParams-reconfigure}

对于 Terraform、OpenTofu 或 Terragrunt 模板显示。

以 `-reconfigure` 而不是 `-migrate-state` 运行 `init`：重新设置后端，并且不迁移现有状态。

## Run at {#scheduleForm-runAt}

对于只运行一次的计划显示。

任务开始的日期和时间，采用字段末尾显示的服务器时区。新计划初始为下一个整点。该时刻必须在将来：服务器会拒绝已经过去的时间。

任务启动之后，计划会被关闭；如果勾选了 **Delete after run**，则会被删除。如果 Semaphore 在那一时刻没有运行，任务不会在之后补启动：
下次启动时，计划会在不运行的情况下被关闭或删除。

## Delete after run {#scheduleForm-deleteAfterRun}

对于只运行一次的计划显示。

在其任务启动之后删除计划。不勾选则保留计划（处于关闭状态），以便您设置新的时间并重新开启它。

## Show cron format {#scheduleForm-rawCron}

对于 cron 计划显示。

在输入 cron 表达式（开启）与通过 **Timing** 和时间复选框构建表达式（关闭）之间切换。两种方式编辑的是同一个表达式。浏览器会记住您的选择。

当表达式无法以复选框的形式显示时（表达式无效，或者使用了 `@daily` 或 `CRON_TZ=` 之类的形式），此开关会锁定为开启。

## Cron {#scheduleForm-cronFormat}

对于 cron 计划，当 **Show cron format** 开启时显示。

由五个字段组成的 cron 表达式：分钟、小时、月中的日、月、星期中的日（`0` 为星期日）。此字段为必填项，服务器会在您输入时检查表达式。
新计划初始为 `* * * * *`，即每分钟。

```
0 3 * * *       every day at 03:00
*/15 * * * *    every 15 minutes
30 8 * * 1-5    at 08:30 from Monday to Friday
0 0 1 * *       at midnight on the first day of every month
```

服务器还接受 `@hourly`、`@daily`、`@weekly`、`@monthly`、`@yearly` 和 `@every 1h30m`，以及用于指定其他时区的前缀，例如
`CRON_TZ=Europe/Berlin`。下方的预览无法显示这些表达式的下次运行时间。

## Timing {#scheduleForm-timing}

对于 cron 计划，当 **Show cron format** 关闭时显示。

任务运行的频率。下方的复选框会随之变化：

- **Hourly**：**Minutes**。
- **Daily**：**Hours** 和 **Minutes**。
- **Weekly**：**Weekdays**、**Hours** 和 **Minutes**。
- **Monthly**：**Days**、**Hours** 和 **Minutes**。
- **Yearly**：**Months**、**Days**、**Hours** 和 **Minutes**。

## Months {#scheduleForm-months}

当 **Timing** 为 **Yearly** 时显示。

任务运行的月份。勾选多个则在其中每个月份运行。一个都不勾选时，任务每个月都运行。

## Weekdays {#scheduleForm-weekdays}

当 **Timing** 为 **Weekly** 时显示。

任务运行的星期几。一个都不勾选时，任务每天都运行。

## Days {#scheduleForm-days}

当 **Timing** 为 **Monthly** 或 **Yearly** 时显示。

任务运行的月中日期。没有该日期的月份会被跳过：设置为 31 日的任务在四月不会运行。一个都不勾选时，任务每天都运行。

## Hours {#scheduleForm-hours}

当 **Timing** 为 **Daily**、**Weekly**、**Monthly** 或 **Yearly** 时显示。

任务运行的小时，采用右侧显示的时区。一个都不勾选时，任务每小时都运行。

## Minutes {#scheduleForm-minutes}

对于 cron 计划，当 **Show cron format** 关闭时显示。

任务运行的分钟，以五分钟为步长。一个都不勾选时，任务每分钟都运行：对于应每小时或每天运行一次的任务，请至少勾选一个，例如
`:00`。

## Next run time {#scheduleForm-nextRun}

按照表单中当前的计划，下一个任务将在何时启动：分别以服务器的时区和您浏览器的本地时间显示。短横线表示表达式无效，或者是预览无法读取的表达式，例如
`@daily`。

## 启用 {#scheduleForm-active}

计划是否启动任务。取消勾选则保留计划但不运行它。计划列表中的开关作用相同。新计划默认启用。
