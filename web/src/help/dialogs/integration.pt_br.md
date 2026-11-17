---
form: src/components/IntegrationForm.vue, src/components/TaskParamsForm.vue
---

# Editar Integração

Uma integração inicia um modelo de tarefa, ou um workflow, quando outro serviço chama uma URL de
webhook do Semaphore, como um evento de push do GitHub. A URL é um alias, `/api/integrations/` e
um código aleatório: os aliases do projeto ficam na página Integrações, os aliases de uma
integração na página dela. Essa página também contém os correspondentes e os valores que a
integração extrai da requisição.

## Name {#integrationForm-name}

O nome com que a integração aparece na página Integrações. As tarefas que a integração inicia são
atribuídas a ela no log de auditoria. O campo é obrigatório.

## Runs {#integrationForm-targetToggle}

Exibido quando o servidor oferece suporte a workflows.

O que a integração inicia: **Template** inicia uma tarefa de um modelo de tarefa, **Workflow**
inicia uma execução de um workflow. Mudar para **Workflow** descarta os parâmetros da tarefa
abaixo.

## Workflow to run {#integrationForm-workflowId}

Exibido quando **Runs** é **Workflow**.

O workflow que uma requisição inicia. Cada valor que a integração extrai da requisição é passado
ao workflow como a resposta à variável de pesquisa dele com o mesmo nome. As integrações com
workflow exigem a edição Pro: sem ela, o campo fica desativado e o servidor recusa a alteração.

## Task Template to run {#integrationForm-templateId}

Exibido quando **Runs** é **Template**.

O modelo de tarefa do qual uma requisição inicia uma tarefa. O campo é obrigatório. Depois que um
modelo é escolhido, os parâmetros das tarefas dele aparecem abaixo.

## Parâmetros da tarefa {#integrationForm-taskParams}

Exibido quando um modelo de tarefa está escolhido.

Os parâmetros que cada tarefa iniciada por esta integração recebe, como se você mesmo a iniciasse
a partir do modelo. Aqui estão apenas as opções que o modelo permite que uma tarefa substitua. Os
valores extraídos da requisição substituem estes: um valor do tipo **Variables** substitui uma
resposta com o mesmo nome, um valor do tipo **Task Params** substitui um parâmetro com o mesmo
nome.

Quando o modelo permite argumentos CLI em uma tarefa, **CLI args** contém argumentos extras para a
linha de comando da tarefa.

## Mensagem {#taskParams-message}

Opcional. A mensagem de cada tarefa que a integração inicia, exibida no histórico de tarefas.

## Variáveis de Pesquisa {#taskParams-survey}

Exibido quando o modelo tem variáveis de pesquisa.

As respostas à pesquisa do modelo, as mesmas para todas as tarefas. Um valor do tipo **Variables**
extraído da requisição substitui a resposta com o mesmo nome.

As respostas a variáveis do tipo secret não são salvas com a integração: passe um segredo pelo
grupo de variáveis do modelo.

## Ramo {#taskParams-gitBranch}

Exibido quando o modelo permite substituir o ramo em uma tarefa.

O ramo ou a tag do repositório de que as tarefas fazem o checkout. Deixe como está para usar o
ramo do modelo. Um valor chamado `git_branch` do tipo **Task Params**, extraído da requisição, o
substitui; um `refs/heads/` no início é removido desse valor.

## Inventário {#taskParams-inventoryId}

Exibido quando o modelo permite substituir o inventário em uma tarefa.

O inventário, ou para o Terraform o workspace, contra o qual as tarefas são executadas no lugar do
inventário do modelo.

## Opções do Ansible {#taskParams-ansible}

Exibido quando o modelo executa um playbook do Ansible.

As opções do `ansible-playbook` para cada tarefa, cada uma exibida quando o modelo a permite:
**Limite**, **Tags**, **Pular tags**, **Depurar**, **Execução Simulada** (`--check`),
**Diferença** e **Skip Galaxy install**.

## Opções do Terraform {#taskParams-terraform}

Exibido quando o modelo executa Terraform, OpenTofu ou Terragrunt.

- **Plan**: a tarefa executa o plano e para, nada é aplicado.
- **Destroy**: o plan e o apply destroem os recursos.
- **Auto Approve**: o plano é aplicado sem confirmação, quando o modelo permite a aprovação
  automática. Sem ele, uma tarefa aguarda até que alguém a confirme no Semaphore.
- **Upgrade** e **Reconfigure**: o `init` é executado com `-upgrade` e `-reconfigure`.

## Auth method {#integrationForm-authMethod}

Como o Semaphore verifica que uma requisição vem do serviço que você espera. Uma requisição que
falha na verificação não inicia nada, e o webhook a responde da mesma forma que qualquer outra
requisição.

- **None**: nenhuma verificação. Qualquer pessoa que conheça a URL pode iniciar a tarefa.
- **GitHub Webhooks**: o cabeçalho `X-Hub-Signature-256` precisa conter a assinatura
  HMAC-SHA256 do corpo feita com o segredo, como o GitHub a envia.
- **Bitbucket Webhooks**: a mesma verificação no cabeçalho `X-Hub-Signature`, como o Bitbucket a
  envia.
- **Token**: o cabeçalho indicado em **Auth header** precisa conter o próprio segredo.
- **HMAC** e **HMAC (SHA-512)**: o cabeçalho indicado em **Auth header** precisa conter o
  HMAC-SHA256 ou o HMAC-SHA512 do corpo feito com o segredo, em hexadecimal minúsculo, sem um
  prefixo como `sha256=`.
- **BasicAuth**: a requisição precisa levar o login e a senha da chave de
  **Segredo de autenticação** no seu cabeçalho `Authorization`.

## Auth header {#integrationForm-authHeader}

Exibido quando **Auth method** é **Token**, **HMAC** ou **HMAC (SHA-512)**.

O nome do cabeçalho da requisição que leva o token ou a assinatura, por exemplo
`X-Webhook-Token`. Maiúsculas e minúsculas no nome não importam. Uma requisição sem esse
cabeçalho é recusada.

## Segredo de autenticação {#integrationForm-authSecretId}

Exibido quando **Auth method** não é **None**.

A chave do Armazenamento de Chaves com a qual a requisição é verificada. Para **BasicAuth**, serve
apenas uma chave do tipo **Login com senha**. Para os outros métodos, uma chave do tipo **Texto**
também serve: a verificação usa o valor dela, ou a senha de uma chave **Login com senha**.

O campo é necessário para todos os métodos, exceto **None**: sem uma chave, ou com um segredo
vazio, toda requisição é recusada.
