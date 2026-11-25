---
form: src/components/TerraformInventoryForm.vue
---

# Novo Terraform Workspace

Um workspace é o inventário de um modelo de Terraform, OpenTofu ou Terragrunt: ele indica o
workspace que uma tarefa seleciona depois do `init` e a chave com que o `init` baixa módulos
privados. O workspace que você cria aqui é anexado a este modelo e listado acima do estado dele.
As tarefas do modelo são executadas no workspace padrão dele, o marcado com um visto; **Make
default** o altera. A mesma caixa de diálogo edita um workspace quando você a abre com o botão de
lápis.

## Workspace name {#tfWorkspace-inventory}

O nome do workspace do Terraform, por exemplo `production`. O campo é obrigatório.

Antes de um plan, o Semaphore executa `workspace select -or-create=true` com este nome, então o
workspace é criado quando ainda não existe. Quando `workspace list` falha, por exemplo porque o
backend não oferece suporte a workspaces, o Semaphore pula a seleção.

O Semaphore armazena o workspace como `TEMPLATE - NAME`, o nome do modelo e este nome: você vê esse
nome completo na página Inventário e quando anexa o workspace a um modelo. Renomear um workspace
não o renomeia no Terraform: a próxima tarefa seleciona, ou cria, um workspace com o novo nome.

## SSH key for private modules {#tfWorkspace-sshKeyId}

A chave que o `init` usa para baixar módulos de repositórios Git privados por SSH. O campo é
obrigatório nesta caixa de diálogo. A lista oferece as chaves dos tipos **Chave SSH**, **Login com
senha** e **Nenhum** do Armazenamento de Chaves.

Apenas uma **Chave SSH** tem efeito: o Semaphore a carrega em um agente SSH durante o `init`, e o
Git a usa para fontes de módulos como `git::ssh://git@example.com/modules/vpc.git`. Uma chave do
tipo **Login com senha** não passa credenciais ao Git. Escolha **Nenhum** quando os módulos não
precisam de chave.
