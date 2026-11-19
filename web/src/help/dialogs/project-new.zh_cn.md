---
form: src/components/ProjectForm.vue
---

# 新项目

项目包含一项工作的任务模板、库存、仓库、变量组和密钥，以及参与这项工作的人员。您将成为您所创建项目的所有者。
只有管理员可以创建项目，除非服务器管理员允许每个用户创建项目。以下每一项设置以后都可以在项目的设置页面中更改。

## 项目名称 {#newProject-name}

项目在侧边栏顶部的项目菜单中显示的名称。此字段为必填项。

## 最大并行任务数 {#newProject-maxParallelTasks}

此项目可以同时运行的任务数量。超出限制后启动的任务会在队列中等待，直到某个正在运行的任务结束。

将此字段留空或设为 `0`，表示项目本身不设限制。无论哪种情况，由管理员设置的整个服务器的限制都适用。

## Telegram 聊天 ID {#newProject-tg}

可选。接收此项目警报的 Telegram 聊天。如果此字段为空，警报将发送到为整个服务器设置的聊天。

只有当管理员已为服务器启用 Telegram 警报，并且此项目允许警报时，此字段才会生效。

## 允许此项目的警报 {#newProject-alert}

为此项目的任务启用警报。警报的发送渠道由管理员为整个服务器设置：电子邮件、Telegram、Slack、Rocket.Chat、
Microsoft Teams、DingTalk 或 Gotify。

- 任务失败时会发送电子邮件，收件人是能够查看该任务、并且在自己的账户中启用了警报的项目成员。
- 任务成功、失败或等待确认时，会向即时通讯工具发送消息。

## Demo {#newProject-demo}

在新项目中填入用于试用 Semaphore 的示例：公共仓库
`https://github.com/semaphoreui/semaphore-demo.git`、库存、一个保管库密码密钥，以及适用于 Ansible、OpenTofu、
Terragrunt、Bash 和 PowerShell 的任务模板，这些模板分布在 **Build**、**Deploy** 和 **Tools** 视图中。

如果不需要示例，请关闭此开关：项目只会获得密钥 **None** 和视图 **All**。这些示例是项目中的普通条目：
可以随意编辑或删除它们。
