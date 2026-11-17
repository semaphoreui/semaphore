---
form: src/components/TemplateSelectForm.vue
---

# Choose workspace to attach

将此工作区附加到一个模板，使其连同状态和别名一起出现在该模板的工作区中。一个工作区同一时间只属于一个模板：
附加之后，它会离开库存页面，改为在模板中管理。附加不会使它成为模板的默认工作区：为此请打开模板并点击
**Make default**。

## Template {#inventoryAttach-templateId}

要附加此工作区的模板。此字段为必填项。列表显示项目中与此工作区为同一应用（Terraform 或 OpenTofu）创建的模板。

工作区会保留其名称、密钥和状态。以后要将它移到另一个模板，请在它所附加的模板中点击 **Detach**：它会回到库存页面。
模板的默认工作区无法分离。
