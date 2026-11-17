---
form: src/components/IntegrationForm.vue, src/components/TaskParamsForm.vue
---

# Editar Integração

Uma integração inicia um modelo de tarefa, ou um workflow, quando outro serviço chama um URL de
webhook do Semaphore, como um evento push do GitHub. O URL é um alias, `/api/integrations/` e um
código aleatório: os aliases do projeto encontram-se na página Integrações, os aliases de uma
integração na sua própria página. Essa página contém também os correspondentes e os valores que
a integração extrai do pedido.

## Name {#integrationForm-name}

O nome com que a integração aparece na página Integrações. As tarefas que a integração inicia são
atribuídas a ela no registo de auditoria. O campo é obrigatório.

## Runs {#integrationForm-targetToggle}

Apresentado quando o servidor suporta workflows.

O que a integração inicia: **Template** inicia uma tarefa de um modelo de tarefa, **Workflow**
inicia uma execução de um workflow. Mudar para **Workflow** descarta os parâmetros da tarefa
abaixo.

## Workflow to run {#integrationForm-workflowId}

Apresentado quando **Runs** é **Workflow**.

O workflow que um pedido inicia. Cada valor que a integração extrai do pedido é transmitido ao
workflow como resposta à sua variável de pesquisa com o mesmo nome. As integrações com workflows
requerem a edição Pro: sem ela, o campo fica desativado e o servidor recusa a alteração.

## Task Template to run {#integrationForm-templateId}

Apresentado quando **Runs** é **Template**.

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

## Auth method {#integrationForm-authMethod}

Como o Semaphore verifica que um pedido provém do serviço esperado. Um pedido que falha a
verificação não inicia nada, e o webhook responde-lhe da mesma forma que a qualquer outro pedido.

- **None**: sem verificação. Qualquer pessoa que conheça o URL pode iniciar a tarefa.
- **GitHub Webhooks**: o cabeçalho `X-Hub-Signature-256` tem de conter a assinatura HMAC-SHA256
  do corpo feita com o segredo, tal como o GitHub a envia.
- **Bitbucket Webhooks**: a mesma verificação no cabeçalho `X-Hub-Signature`, tal como o
  Bitbucket o envia.
- **Token**: o cabeçalho indicado em **Auth header** tem de conter o próprio segredo.
- **HMAC** e **HMAC (SHA-512)**: o cabeçalho indicado em **Auth header** tem de conter o
  HMAC-SHA256 ou o HMAC-SHA512 do corpo feito com o segredo, em hexadecimal minúsculo, sem um
  prefixo como `sha256=`.
- **BasicAuth**: o pedido tem de levar, no seu cabeçalho `Authorization`, o nome de utilizador e
  a palavra-passe da chave indicada em **Segredo de autenticação**.

## Auth header {#integrationForm-authHeader}

Apresentado quando **Auth method** é **Token**, **HMAC** ou **HMAC (SHA-512)**.

O nome do cabeçalho do pedido que leva o token ou a assinatura, por exemplo `X-Webhook-Token`.
As maiúsculas e minúsculas do nome não importam. Um pedido sem esse cabeçalho é recusado.

## Segredo de autenticação {#integrationForm-authSecretId}

Apresentado quando **Auth method** não é **None**.

A chave do Armazenamento de Chaves com que o pedido é verificado. Para **BasicAuth** só serve uma
chave do tipo **Login com senha**. Para os outros métodos serve também uma chave do tipo
**Texto**: a verificação utiliza o seu valor, ou a palavra-passe de uma chave **Login com senha**.

O campo é necessário para todos os métodos exceto **None**: sem uma chave, ou com um segredo
vazio, todos os pedidos são recusados.
