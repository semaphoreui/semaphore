---
form: src/components/InventorySelectForm.vue
---

# Choose workspace to attach

将项目中已有的工作区附加到此模板，使其连同状态和别名一起出现在模板的工作区中。一个工作区同一时间只属于一个模板。
附加不会使该工作区成为默认工作区：在您点击 **Make default** 之前，任务会继续在默认工作区中运行。

## Workspace {#workspaceSelect-inventoryId}

要附加的工作区。此字段为必填项。列表显示项目中为此模板的应用（Terraform、OpenTofu 或 Terragrunt）创建、
且未附加到任何模板的工作区；用 **Detach** 从模板分离的工作区会再次出现在这里。

工作区会保留其名称、密钥和状态。如果在此期间另一个模板已经占用了该工作区，服务器会拒绝它。
