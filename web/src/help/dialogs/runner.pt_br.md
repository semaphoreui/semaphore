---
form: src/components/RunnerForm.vue
---

# Editar Executor

Um executor é um processo `semaphore runner` separado que executa tarefas na sua própria máquina,
em vez de no servidor Semaphore. Uma tarefa vai para um executor quando o servidor está
configurado para executar tarefas em executores, ou quando o modelo de tarefa ou o inventário dele
tem uma tag de executor. O servidor então escolhe um executor que esteja habilitado, registrado,
online e abaixo do seu limite de tarefas paralelas: primeiro os executores do projeto, depois os
executores globais, em ordem aleatória dentro de cada grupo. Quando todos os executores
correspondentes estão offline ou ocupados, a tarefa aguarda na fila; quando nenhum executor
corresponde, a tarefa falha.

## Nome {#runner-name}

Obrigatório. O nome com que o executor aparece na página Executores, por exemplo o host em que ele
é executado. Ele não altera quais tarefas o executor recebe.

## Tags {#runner-tags}

Disponível no Semaphore Pro; nas outras edições o campo fica bloqueado.

As tags que escolhem as tarefas deste executor. Uma tarefa com uma tag de executor vai apenas para
os executores que têm essa tag; a tag do modelo de tarefa tem precedência sobre a tag do
inventário. Digite uma tag e pressione Enter, ou escolha uma usada por outros executores. Tags em
branco e repetidas são descartadas ao salvar.

Um executor de projeto precisa de pelo menos uma tag, a menos que **Is default** esteja marcado.

## Is default {#runner-isDefault}

Uma tarefa sem tag de executor vai apenas para os executores marcados como padrão. Desmarque a
caixa para um executor que deve receber apenas tarefas com uma das suas tags.

A caixa vem marcada para um novo executor global e desmarcada para um novo executor de projeto.

## Register {#runner-registered}

Exibido quando você cria um executor.

- Marcado (o padrão): o servidor cria o executor com o seu token. Depois de **Criar**, a caixa de
  diálogo mostra o token e os comandos que iniciam o executor com ele.
- Desmarcado: o executor é criado sem nenhum token. Depois, clique em **Regenerate registration
  token** (o ícone de duas setas) na linha do executor para obter um token de registro de uso
  único, válido por uma hora, e registre o executor com `semaphore runner register`.

Um executor não registrado não recebe tarefas.

## Webhook {#runner-webhook}

Opcional. Uma URL que o servidor chama quando entrega uma tarefa a este executor, por exemplo para
iniciar o executor sob demanda. O servidor envia uma requisição `POST` com um corpo JSON:

```json
{"action": "start", "project_id": 1, "task_id": 42, "template_id": 7, "runner_id": 3}
```

A URL precisa responder com o status 200 ou 204 em até 30 segundos; caso contrário, a tarefa
falha. Um executor com webhook é sempre tratado como online, mesmo quando não consultou o servidor
recentemente. Deixe o campo vazio para um executor que funciona o tempo todo.

## Número máximo de tarefas paralelas {#runner-maxParallelTasks}

O número de tarefas que o executor pode executar ao mesmo tempo. Quando o limite é atingido, a
próxima tarefa vai para outro executor correspondente ou aguarda na fila. Deixe o campo vazio ou
defina `0` para não ter limite.

## Habilitado {#runner-active}

Apenas um executor habilitado recebe novas tarefas. Desmarque a caixa para tirar o executor de
serviço sem excluí-lo; o interruptor na lista de executores faz o mesmo.

A caixa vem desmarcada para um novo executor: marque-a, ou habilite o executor depois na lista.
