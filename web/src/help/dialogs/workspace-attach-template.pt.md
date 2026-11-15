---
form: src/components/TemplateSelectForm.vue
---

# Choose workspace to attach

Associa este workspace a um modelo, para que apareça entre os workspaces do modelo com o seu
estado e os seus aliases. Um workspace pertence a um único modelo de cada vez: depois de
associado, sai da página Inventário e passa a ser gerido no modelo. Associá-lo não o torna o
workspace predefinido do modelo: para isso, abra o modelo e prima **Make default**.

## Template {#inventoryAttach-templateId}

O modelo ao qual associar o workspace. O campo é obrigatório. A lista mostra os modelos do
projeto criados para a mesma aplicação que o workspace, Terraform ou OpenTofu.

O workspace mantém o seu nome, a sua chave e o seu estado. Para o mover mais tarde para outro
modelo, prima **Detach** no modelo a que está associado: volta à página Inventário. O workspace
predefinido de um modelo não pode ser desassociado.
