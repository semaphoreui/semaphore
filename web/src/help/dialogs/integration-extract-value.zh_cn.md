---
form: src/components/IntegrationExtractValueForm.vue
---

# Edit Extract Value

提取值从传入请求中取出一部分内容（例如 push 事件的分支），并将其传递给集成启动的任务。集成的每一个值都会从每个启动任务的请求中取出。

## Name {#integrationValueForm-name}

值在集成页面中显示的名称。名称不会改变取出的内容或其去向。此字段为必填项。

## Source of the Value {#integrationValueForm-valueSource}

取出值的请求部分：

- **Header**：请求标头的值。
- **Body**：请求的正文。

## Data Type of Body {#integrationValueForm-bodyDataType}

当 **Source of the Value** 为 **Body** 时显示。

正文的读取方式：

- **JSON**：正文被解析为 JSON，**Key** 是指向其中某个值的路径。值会保留其 JSON 类型：数字、列表或对象会原样传递给任务。
- **String**：整个正文作为文本取出，不使用 **Key**。

## Key {#integrationValueForm-key}

要从请求中取出的内容。此字段为必填项。

- 对于 **Header**：标头的名称，例如 `X-GitHub-Event`。不区分大小写。请求中缺少的标头会得到一个空值。
- 对于 **JSON** 格式的 **Body**：值的路径，各部分用点号连接；数组中的元素写作 `.[0]`，例如
  `repository.full_name` 或 `commits.[0].id`。正文中缺少的路径不会得到任何值，任务会保留集成中设置的值。
- 对于 **String** 格式的 **Body**：不使用，但表单需要一个值；随便输入即可。

## Variable Usage {#integrationValueForm-variableType}

值在模板的任务中的去向：

- **Variables**：任务的额外变量，与调查变量的答案的传递方式相同。它会替换同名的答案或变量组中的同名变量。
- **Task Params**：任务的参数，与任务对话框中选项的保存方式相同。当模板允许覆盖分支时，名称 `git_branch`
  会设置任务的分支；开头的 `refs/heads/` 会被去除。

运行工作流的集成会忽略此用途：每个值都会成为 **Variable** 中指定的工作流调查变量的答案。

## Variable {#integrationValueForm-variable}

传递值时使用的名称，例如对于剧本以 `{{ branch }}` 读取的额外变量，名称为 `branch`。此字段为必填项。
