---
form: src/components/InventorySelectForm.vue
---

# Choose workspace to attach

Anexa a este modelo um workspace que já existe no projeto, para que ele apareça entre os
workspaces do modelo com o seu estado e os seus aliases. Um workspace pertence a um modelo por
vez. Anexar não torna o workspace o padrão: as tarefas continuam sendo executadas no workspace
padrão até você clicar em **Make default**.

## Workspace {#workspaceSelect-inventoryId}

O workspace a anexar. O campo é obrigatório. A lista mostra os workspaces do projeto que são
feitos para o app deste modelo (Terraform, OpenTofu ou Terragrunt) e não estão anexados a nenhum
modelo; um workspace desanexado do seu modelo com **Detach** volta a aparecer aqui.

O workspace mantém o nome, a chave e o estado. O servidor recusa um workspace que outro modelo
tenha assumido nesse meio-tempo.
