---
form: src/components/IntegrationForm.vue, src/components/TaskParamsForm.vue
---

# Editar execução

Uma execução é aquilo que uma integração inicia quando outro serviço chama um URL de webhook do
Semaphore, por exemplo num evento push do GitHub: uma tarefa de um modelo de tarefa ou um workflow.
O URL e a verificação do pedido pertencem ao endpoint da integração, não à execução. A página de uma
execução contém os seus matchers e os valores que retira do pedido.

## Nome {#integrationForm-name}

O nome com que a integração aparece na página Integrações. As tarefas que a integração inicia são
atribuídas a ela no registo de auditoria. O campo é obrigatório.

## Integração {#integrationForm-groupId}

A integração a que a execução pertence. A execução inicia-se com os pedidos que chegam ao endpoint
dessa integração e passam a sua autorização. O campo é obrigatório.

## Alvo {#integrationForm-targetToggle}

Apresentado quando o servidor suporta workflows.

O que a integração inicia: **Template** inicia uma tarefa de um modelo de tarefa, **Workflow**
inicia uma execução de um workflow. Mudar para **Workflow** descarta os parâmetros da tarefa
abaixo.

## Workflow to run {#integrationForm-workflowId}

Apresentado quando **Alvo** é **Workflow**.

O workflow que um pedido inicia. Cada valor que a integração extrai do pedido é transmitido ao
workflow como resposta à sua variável de pesquisa com o mesmo nome. As integrações com workflows
requerem a edição Pro: sem ela, o campo fica desativado e o servidor recusa a alteração.

## Modelo de tarefa a executar {#integrationForm-templateId}

Apresentado quando **Alvo** é **Template**.

O modelo de tarefa de que um pedido inicia uma tarefa. O campo é obrigatório. Depois de escolhido
um modelo, os parâmetros das suas tarefas aparecem abaixo dele.

## Parâmetros da tarefa {#integrationForm-taskParams}

Apresentado quando está escolhido um modelo de tarefa.

Os parâmetros que todas as tarefas iniciadas por esta integração recebem, como se fosse o próprio
utilizador a iniciá-las a partir do modelo. Aqui estão apenas as opções que o modelo permite que
uma tarefa substitua. Os valores extraídos do pedido prevalecem sobre estes: um valor do tipo
**Variables** sobre uma resposta com o mesmo nome, um valor do tipo **Task Params** sobre um
parâmetro com o mesmo nome.

Quando o modelo permite argumentos CLI numa tarefa, **Argumentos CLI** contém argumentos
adicionais para a linha de comandos da tarefa.

## Mensagem {#taskParams-message}

Opcional. A mensagem de todas as tarefas que a integração inicia, apresentada no histórico de
tarefas.

## Variáveis de Pesquisa {#taskParams-survey}

Apresentado quando o modelo tem variáveis de pesquisa.

As respostas ao questionário do modelo, as mesmas para todas as tarefas. Um valor do tipo
**Variables** extraído do pedido substitui a resposta com o mesmo nome.

As respostas a variáveis do tipo secret não são guardadas com a integração: transmita um segredo
através do grupo de variáveis do modelo.

## Ramo {#taskParams-gitBranch}

Apresentado quando o modelo permite substituir o ramo numa tarefa.

O ramo ou a tag do repositório que as tarefas obtêm. Deixe-o como está para utilizar o ramo do
modelo. Um valor chamado `git_branch` do tipo **Task Params**, extraído do pedido, prevalece sobre
ele; um `refs/heads/` no início é removido desse valor.

## Inventário {#taskParams-inventoryId}

Apresentado quando o modelo permite substituir o inventário numa tarefa.

O inventário, ou no Terraform o workspace, sobre o qual as tarefas são executadas em vez do
inventário do modelo.

## Opções do Ansible {#taskParams-ansible}

Apresentado quando o modelo executa um playbook Ansible.

As opções do `ansible-playbook` para todas as tarefas, cada uma apresentada quando o modelo a
permite: **Limite**, **Tags**, **Pular tags**, **Depurar**, **Execução Simulada** (`--check`),
**Diferença** e **Skip Galaxy install**.

## Opções do Terraform {#taskParams-terraform}

Apresentado quando o modelo executa Terraform, OpenTofu ou Terragrunt.

- **Plan**: a tarefa executa o plano e para; nada é aplicado.
- **Destroy**: o plano e a aplicação destroem os recursos.
- **Auto Approve**: o plano é aplicado sem confirmação, quando o modelo permite a aprovação
  automática. Sem esta opção, uma tarefa aguarda até que alguém a confirme no Semaphore.
- **Upgrade** e **Reconfigure**: o `init` é executado com `-upgrade` e `-reconfigure`.
