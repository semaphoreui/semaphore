---
form: src/components/InventorySelectForm.vue
---

# Choose workspace to attach

Associa a este modelo um workspace que já existe no projeto, para que apareça entre os workspaces
do modelo com o seu estado e os seus aliases. Um workspace pertence a um único modelo de cada
vez. Associá-lo não o torna o workspace predefinido: as tarefas continuam a ser executadas no
workspace predefinido até premir **Make default**.

## Workspace {#workspaceSelect-inventoryId}

O workspace a associar. O campo é obrigatório. A lista mostra os workspaces do projeto que foram
criados para a aplicação deste modelo (Terraform, OpenTofu ou Terragrunt) e que não estão
associados a nenhum modelo; um workspace desassociado do seu modelo com **Detach** volta a
aparecer aqui.

O workspace mantém o seu nome, a sua chave e o seu estado. O servidor recusa um workspace que
outro modelo tenha entretanto ocupado.
