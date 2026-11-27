---
form: src/components/TemplateSelectForm.vue
---

# Choose workspace to attach

Anexa este workspace a um modelo, para que ele apareça entre os workspaces do modelo com o seu
estado e os seus aliases. Um workspace pertence a um modelo por vez: depois de anexado, ele sai
da página Inventário e passa a ser gerenciado no modelo. Anexar não o torna o workspace padrão do
modelo: para isso, abra o modelo e clique em **Make default**.

## Template {#inventoryAttach-templateId}

O modelo ao qual anexar o workspace. O campo é obrigatório. A lista mostra os modelos do projeto
feitos para o mesmo app do workspace, Terraform ou OpenTofu.

O workspace mantém o nome, a chave e o estado. Para movê-lo para outro modelo depois, clique em
**Detach** no modelo ao qual ele está anexado: ele volta para a página Inventário. O workspace
padrão de um modelo não pode ser desanexado.
