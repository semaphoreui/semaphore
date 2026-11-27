---
form: src/components/AppForm.vue
---

# Edit App

应用是 Semaphore 用来运行任务的工具：Ansible、Terraform、OpenTofu、Terragrunt、Bash、Python、PowerShell，
或者您自己的命令。每个任务模板都为一个应用而创建，此对话框设置服务器如何启动该应用，以及应用在界面中如何显示。

这些设置保存在数据库中，并在保存它们的服务器上立即生效；集群的其他节点会在重新启动时读取它们。您清空的字段会保留其旧值，
直到服务器重新启动。

## ID {#appForm-id}

必填。保存应用所使用的键；任务模板通过此键引用其应用。内置应用为 `ansible`、`terraform`、`tofu`、`terragrunt`、
`pulumi`、`bash`、`python` 和 `powershell`。任何其他 ID 都会添加一个您自己的应用：Semaphore 会像运行脚本一样运行它，
使用 **Path** 中指定的命令，或者在 **Path** 为空时使用与 ID 同名的命令。

请使用不含点号的 ID。更改已有应用的 ID 不会重命名它：设置会以新 ID 保存为一个新应用，而旧应用仍然保留。

## Icon {#appForm-icon}

可选。Material Design Icons 图标集中某个图标的名称，不带 `mdi-` 前缀，例如 `language-python`。
留空则保留应用的内置图标；此时您自己的应用会显示一个问号。

## 名称 {#appForm-title}

可选。应用在应用列表以及**新模板**菜单中显示的名称。留空则保留内置名称，例如 `Ansible Playbook`；
此时您自己的应用没有名称，因此请为它指定一个名称。

## Path {#appForm-path}

可选。Semaphore 用来替代默认命令运行的命令，可以是在服务器的 `PATH` 中查找的名称，也可以是完整路径，例如
`/usr/local/bin/terraform-1.9`。默认命令为 `terraform`、`tofu`、`terragrunt`、`bash`、`python3` 和
`powershell`；您自己的应用运行与其 ID 同名的命令。

Ansible 会忽略此字段：Semaphore 始终运行 `ansible-playbook` 和 `ansible-galaxy`。

## Priority {#appForm-priority}

应用在应用列表以及**新模板**菜单中的位置：数字越大的应用越靠前。内置应用从 Ansible 的 1000 开始，到 Python
的 500 为止。优先级为 `0` 或更小的内置应用会在服务器启动时恢复其内置优先级。

## Args {#appForm-args}

可选。紧跟在命令之后放到命令行上的参数，每个参数一个标签。点击 **+** 添加参数，点击标签可以更改它，拖动标签可以更改其顺序。

- Terraform、OpenTofu 和 Terragrunt：这些参数位于 Semaphore 为每条命令添加的参数（例如 `init` 和 `apply`）之前。
- 脚本应用：这些参数位于模板的脚本之前，并替换默认参数：PowerShell 默认使用 `-File` 运行，因此请将 `-File`
  放在列表的最后。
- Ansible 会忽略此字段。

## Active {#appForm-active}

处于活动状态的应用会出现在**新模板**菜单中。取消勾选此框可以从该菜单中隐藏应用：已经用它创建的模板会保留其应用，并且仍然可以运行。

新建应用时此框默认不勾选：勾选它即可使应用可用。
