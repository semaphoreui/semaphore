---
form: src/components/TemplateSelectForm.vue
---

# Choose workspace to attach

Attaches this workspace to a template, so it appears among the workspaces of the template with
its state and aliases. A workspace belongs to one template at a time: once attached, it leaves
the Inventory page and is managed in the template. Attaching does not make it the default
workspace of the template: open the template and press **Make default** for that.

## Template {#inventoryAttach-templateId}

The template to attach the workspace to. The field is required. The list shows the templates of
the project made for the same app as the workspace, Terraform or OpenTofu.

The workspace keeps its name, key and state. To move it to another template later, press
**Detach** in the template it is attached to: it comes back to the Inventory page. The default
workspace of a template can not be detached.
