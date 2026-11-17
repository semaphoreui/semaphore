---
form: src/components/SurveyVars.vue
---

# 调查变量

调查变量是每次运行模板时任务对话框都会提出的一个问题。答案以变量的形式传递给任务：对于 Ansible 是额外变量，对于
Terraform 和 OpenTofu 是 `-var`，对于脚本是参数，或者是环境变量。变量随模板一起保存，即在您保存模板对话框时保存。
同一个对话框也用于编辑工作流的调查，此时不提供 **Secret** 类型。

## 名称 {#surveyVarDialog-name}

任务获得答案时使用的名称：额外变量、Terraform 变量、脚本的 `name=value` 参数或环境变量的名称。两端的空格会被去除。
此字段为必填项。

请使用调查中其他变量都没有使用的、并且工具能够接受的名称，例如 `app_version`。答案会替换模板变量组中同名的变量。

## 标题 {#surveyVarDialog-title}

任务对话框中字段的标签，例如 `Version to deploy`。它也是模板对话框中该变量标签上的文本。此字段为必填项。

## 描述 {#surveyVarDialog-description}

可选。当任务对话框中的字段获得焦点时，显示在字段下方的提示：应输入什么，或者一个示例。

## 类型 {#surveyVarDialog-type}

任务对话框要求的答案类型：

- **String**：一行文本。这是默认值。
- **Integer**：只能是数字；任务对话框会拒绝其他任何内容，包括负号。答案仍以文本形式传递给任务：请在剧本中转换它，例如使用
  `| int`。
- **Secret**：任务对话框会遮盖的文本。答案不会随任务一起保存，因此重新运行时会再次询问，并且 Ansible 会在一个用 vault
  加密的文件中获得它。秘密值没有默认值。
- **Enum**：从您在下方定义的列表中选择的一个值。
- **Text**：多行文本。
- **Select**：从您在下方定义的列表中选择的多个值。任务以列表的形式获得它们；在参数、`-var` 或环境变量中，列表会以
  JSON 形式写入，例如 `["a","b"]`。

更改类型会清除默认值，而没有列表的类型会清除这些值。

## Pass variable as {#surveyVarDialog-target}

答案传递给任务的方式：

- **Extra variable**：工具的常规方式。Ansible 以额外变量的形式获得它，Terraform 和 OpenTofu 以
  `-var name=value` 的形式获得，Bash、PowerShell 或 Python 脚本以脚本之后的 `name=value` 参数的形式获得。
- **Environment variable**：进程的环境变量 `name=value`，适用于所有类型的模板。此时变量不会以另一种方式传递。

## 键 {#surveyVarDialog-valueName}

当类型为 **Enum** 或 **Select** 时显示。

任务对话框在列表中为此值显示的文本。每个值都需要一个键，并且这些键必须互不相同：否则对话框会拒绝保存该变量。

## 值 {#surveyVarDialog-valueValue}

当类型为 **Enum** 或 **Select** 时显示。

选择此值时任务获得的内容。它可以与键不同：键 `Production` 可以传递 `prod`。

## Add Value {#surveyVarDialog-addValue}

当类型为 **Enum** 或 **Select** 时显示。

在值列表中添加一个空行。**Enum** 或 **Select** 至少需要一个值。行中的垃圾桶图标会删除该行。

## Default value {#surveyVarDialog-defaultValue}

除 **Secret** 之外的所有类型都会显示。

任务对话框打开时填入的答案；您可以在对话框中更改它。对于 **Enum**，请选择其中一个值；对于 **Select**，可以选择任意数量的值。
将此字段留空，则以空答案开始。

默认值只由任务对话框填入。以其他方式（例如通过 API）启动的任务不会获得默认值。

## 必需 {#surveyVarDialog-required}

在变量有答案之前，任务对话框不会启动任务。与默认值一样，此检查只由任务对话框进行，而不是由服务器进行。
