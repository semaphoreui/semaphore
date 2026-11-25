---
form: src/components/IntegrationMatcherForm.vue
---

# Edit Matcher

匹配器是针对传入请求的一个条件。匹配器决定发送到项目别名的请求启动哪个集成，因此一个 URL 可以服务于多个集成：
只有当集成的每一个匹配器都成立时，该集成才会运行。只有当集成开启了**使用项目别名**时才会使用匹配器；开启后，
该集成只能通过项目别名访问。

通过项目别名找到、但没有任何匹配器的集成永远不会运行：请至少添加一个匹配器。

## Name {#integrationMatcherForm-name}

匹配器在集成页面中显示的名称。名称不会改变匹配器检查的内容。此字段为必填项。

## Match on {#integrationMatcherForm-matchType}

匹配器读取的请求部分：

- **Header**：请求标头的值。
- **Body**：请求的正文。

## Body Data Format {#integrationMatcherForm-bodyDataType}

当 **Match on** 为 **Body** 时显示。

正文的读取方式：

- **JSON**：正文被解析为 JSON，**Key** 是指向其中某个值的路径。
- **String**：整个正文作为文本进行比较，不使用 **Key**。

## Key {#integrationMatcherForm-key}

要从请求中读取的内容。此字段为必填项。

- 对于 **Header**：标头的名称，例如 `X-GitHub-Event`。不区分大小写。
- 对于 **JSON** 格式的 **Body**：值的路径，各部分用点号连接；数组中的元素写作 `.[0]`，例如
  `repository.full_name` 或 `commits.[0].id`。
- 对于 **String** 格式的 **Body**：不使用，但表单需要一个值；随便输入即可。

## Comparison Method {#integrationMatcherForm-method}

从请求中读取的值与 **Value** 的比较方式：

- **==**：两者相等。
- **!=**：两者不同。请求中缺少的标头或键视为不同。
- **Contains**：从请求中读取的值包含 **Value**。

## Value {#integrationMatcherForm-value}

用于比较的文本，例如 `push` 或 `refs/heads/main`。比较是精确的，并且区分大小写。JSON 正文中的数字或布尔值以其文本形式进行比较：
`42`、`true`。此字段为必填项。
