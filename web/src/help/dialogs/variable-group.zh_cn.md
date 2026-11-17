---
form: src/components/EnvironmentForm.vue
skip: varGroupSync-syncInterval
---

# 编辑变量组

变量组包含 Semaphore 传递给使用该组的任务模板的每个任务的额外变量、环境变量和秘密值。Ansible 以 `--extra-vars`
的形式获得额外变量，Terraform 和 OpenTofu 以 `-var` 的形式获得；环境变量会设置到任务的进程中。一个模板可以使用多个变量组：
当其中两个变量组定义了相同的名称时，在模板中排在后面的变量组优先。

## 组名 {#varGroup-name}

变量组在变量组页面以及任务模板的变量组选择器中显示的名称。此字段为必填项。

## Secret storage {#varGroup-secretStorageId}

当变量组是新建的且服务器支持秘密存储时显示。可选。

保存此变量组秘密值的秘密存储，例如 HashiCorp Vault。将此字段留空，则秘密值会加密保存在 Semaphore 的数据库中。

存储只能在创建变量组时选择一次：已有的变量组会保留其存储，并且不再显示此字段。标记为 **Read only**
的存储不接受新的秘密值：向其中保存秘密值会失败。

## Secret key prefix {#varGroup-secretStorageKeyPrefix}

当变量组是新建的且服务器支持秘密存储时显示。在选择了 **Secret storage** 之后可用。

此变量组的每个秘密值在存储中的路径开头。Semaphore 会在其后添加 10 个随机字符：使用前缀 `semaphore/prod/`
时，秘密值保存为 `semaphore/prod/` 加上这些字符。前缀按原样使用，因此如果秘密值应当像文件夹一样位于其下，请以
`/` 结尾。

前缀只能在创建变量组时设置一次。**Secrets** 标签页会将其显示为来源路径模式。

## Table、JSON 或 YAML {#varGroup-extraVarsMode}

编辑额外变量的方式：

- **Table**：每个变量一行，每行都有一个类型。使用开关旁边的 **+** 按钮添加一行。
- **JSON**：所有变量作为一个 JSON 对象。
- **YAML**：所有变量作为一个 YAML 映射。

对话框以 **Table** 模式打开。切换模式会转换您输入的内容；如果内容无法读取，对话框会保持在原来的模式并显示错误。
无论使用哪种模式，变量都会保存为一个由名称和值组成的对象。

## JSON 格式的额外变量 {#varGroup-extraVarsJson}

当以 **JSON** 编辑额外变量时显示。

作为一个 JSON 对象的额外变量。名称不能为空；值可以是字符串、数字、列表或对象：

```json
{
  "app_version": "1.4.2",
  "replicas": 3,
  "regions": ["eu-west-1", "us-east-1"]
}
```

## YAML 格式的额外变量 {#varGroup-extraVarsYaml}

当以 **YAML** 编辑额外变量时显示。

作为一个 YAML 映射的额外变量。文本按 JSON 所知的类型读取：不带引号的日期（例如 `2024-01-01`）仍然是字符串，`.inf`
或 `.nan` 会被拒绝。顶层的列表或标量同样会被拒绝。

```yaml
app_version: 1.4.2
replicas: 3
regions:
  - eu-west-1
  - us-east-1
```

## 变量名称 {#varGroup-varName}

当以 **Table** 编辑额外变量时，在每一行中显示。

剧本或 Terraform 代码所看到的变量名称，例如 `app_version`。

## 变量类型 {#varGroup-varType}

当以 **Table** 编辑额外变量时，在每一行中显示。

该行的值的读取方式：

- **String**：按原样使用的文本。
- **Number**：一个数字，例如 `42`。
- **List**：一个 JSON 数组，例如 `["a", "b"]`。
- **Dict**：一个 JSON 对象，例如 `{"key": "value"}`。

与其类型不符的值会阻止保存，并显示该变量的名称。

## 变量值 {#varGroup-varValue}

当以 **Table** 编辑额外变量时，在每一行中显示。

变量的值，采用其类型的格式。对于 **List** 或 **Dict**，字段末尾的按钮会打开一个更大的编辑器。

Semaphore 将这些值作为数据传递给 Ansible：值中的 `{{ … }}` 表达式不会被求值，剧本会按原样获得该文本。

## 环境变量名称 {#varGroup-envName}

在 **Variables** 标签页的**环境变量**下，每一行中显示。使用该部分的 **+** 按钮添加一行。

在运行任务的进程的环境中设置的变量名称，例如 `AWS_REGION`。

## 环境变量值 {#varGroup-envValue}

在 **Variables** 标签页的**环境变量**下，每一行中显示。

变量的值，为文本。它以明文保存，并对每个打开该变量组的人可见：请将密码和令牌放在 **Secrets** 标签页中。

## Sync keys enabled {#varGroup-syncEnabled}

在带有秘密存储的变量组的 **Secrets** 标签页中显示。

开启后，会从 **Sync paths** 下设置的路径，将秘密值从存储导入到此变量组中。当您在变量组页面中点击该变量组的同步按钮时，
以及按照 **Sync paths** 中设置的间隔自动进行时，都会运行导入。关闭时，不会导入任何内容。

## Sync paths {#varGroup-syncPaths}

当 **Sync keys enabled** 开启时显示。

打开一个对话框，其中包含要从中导入的存储路径以及自动导入的间隔。按钮上的数字是路径的数量。在变量组至少有一个路径之前，
变量组页面中该变量组的同步按钮保持禁用状态。

该对话框只有**关闭**按钮：您在其中所做的更改会随变量组一起，通过此对话框的**保存**进行保存。

## 秘密变量名称 {#varGroup-secretVarName}

在 **Secrets** 标签页的**额外变量**下，每一行中显示。使用该部分的 **+** 按钮添加一行。

值为秘密值的额外变量的名称。任务获得它的方式与获得 **Variables** 标签页中额外变量的方式相同；秘密值优先于同名的额外变量。
Ansible 在一个使用 Ansible Vault 加密的文件中获得这些秘密值，Terraform 和 OpenTofu 则以 `-var` 的形式获得。
正如对话框所警告的，以这种方式传递的秘密值仍可能以明文形式出现在 Ansible 的日志中。

## 秘密变量值 {#varGroup-secretVarValue}

在 **Secrets** 标签页的**额外变量**下，每一行中显示。

秘密值本身。Semaphore 会将其加密，或者写入变量组的秘密存储，并且永远不会再次显示它：以后打开变量组时，此字段为空。
留空则保留已保存的值，输入新值则替换它。删除该行会在您保存变量组时删除该秘密值。

## 秘密环境变量名称 {#varGroup-secretEnvName}

在 **Secrets** 标签页的**环境变量**下，每一行中显示。使用该部分的 **+** 按钮添加一行。

值为秘密值的环境变量的名称，例如 `DB_PASSWORD`。它会像 **Variables** 标签页中的环境变量一样，设置到运行任务的进程的环境中。

## 秘密环境变量值 {#varGroup-secretEnvValue}

在 **Secrets** 标签页的**环境变量**下，每一行中显示。

秘密值本身。与秘密额外变量一样，它永远不会再次显示：将此字段留空则保留已保存的值，输入新值则替换它。
