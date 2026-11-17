---
form: src/components/TerraformInventoryForm.vue
---

# Novo Terraform Workspace

Um workspace é o inventário de um modelo Terraform, OpenTofu ou Terragrunt: indica o workspace
que uma tarefa seleciona depois de `init` e a chave com que `init` transfere os módulos privados.
O workspace que cria aqui fica associado a este modelo e aparece listado acima do seu estado. As
tarefas do modelo são executadas no seu workspace predefinido, o que está marcado com um visto;
**Make default** altera-o. A mesma caixa de diálogo edita um workspace quando a abre com o botão
do lápis.

## Workspace name {#tfWorkspace-inventory}

O nome do workspace do Terraform, por exemplo `production`. O campo é obrigatório.

Antes de um plano, o Semaphore executa `workspace select -or-create=true` com este nome, pelo que
o workspace é criado quando ainda não existe. Quando `workspace list` falha, por exemplo porque o
backend não suporta workspaces, o Semaphore salta a seleção.

O Semaphore guarda o workspace como `TEMPLATE - NAME`, o nome do modelo seguido deste nome: vê
esse nome completo na página Inventário e quando associa o workspace a um modelo. Mudar o nome de
um workspace não lhe muda o nome no Terraform: a tarefa seguinte seleciona, ou cria, um workspace
com o novo nome.

## SSH key for private modules {#tfWorkspace-sshKeyId}

A chave que `init` utiliza para transferir módulos de repositórios Git privados por SSH. O campo
é obrigatório nesta caixa de diálogo. A lista oferece as chaves do tipo **Chave SSH**,
**Login com senha** e **Nenhum** do Armazenamento de Chaves.

Só uma **Chave SSH** tem efeito: o Semaphore carrega-a num agente SSH durante `init`, e o Git
utiliza-a para origens de módulos como `git::ssh://git@example.com/modules/vpc.git`. Uma chave do
tipo **Login com senha** não transmite credenciais ao Git. Escolha **Nenhum** quando os módulos
não precisam de chave.
