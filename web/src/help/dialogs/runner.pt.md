---
form: src/components/RunnerForm.vue
---

# Editar Executor

Um executor é um processo `semaphore runner` separado que executa tarefas na sua própria máquina
em vez de no servidor Semaphore. Uma tarefa vai para um executor quando o servidor está
configurado para executar tarefas em executores, ou quando o modelo de tarefa ou o seu inventário
tem uma tag de executor. O servidor escolhe então um executor que esteja habilitado, registado,
online e abaixo do seu limite de tarefas paralelas: primeiro os executores do projeto, depois os
executores globais, por ordem aleatória dentro de cada grupo. Quando todos os executores
correspondentes estão offline ou ocupados, a tarefa aguarda na fila; quando nenhum executor
corresponde, a tarefa falha.

## Nome {#runner-name}

Obrigatório. O nome com que o executor aparece na página Executores, por exemplo o host em que é
executado. Não altera as tarefas que o executor recebe.

## Tags {#runner-tags}

Disponível no Semaphore Pro; nas outras edições o campo está bloqueado.

As tags que selecionam as tarefas deste executor. Uma tarefa com uma tag de executor vai apenas
para os executores que têm essa tag; a tag do modelo de tarefa tem prioridade sobre a tag do
inventário. Escreva uma tag e prima Enter, ou escolha uma utilizada por outros executores. As
tags vazias e repetidas são descartadas ao guardar.

Um executor de projeto precisa de pelo menos uma tag, a menos que **Is default** esteja
assinalado.

## Is default {#runner-isDefault}

Uma tarefa sem tag de executor vai apenas para os executores marcados como predefinidos.
Desmarque a caixa num executor que deva receber apenas tarefas com uma das suas tags.

A caixa está assinalada num novo executor global e desmarcada num novo executor de projeto.

## Register {#runner-registered}

Apresentado quando cria um executor.

- Assinalada (a predefinição): o servidor cria o executor com o seu token. Depois de **Criar**, a
  caixa de diálogo apresenta o token e os comandos que iniciam o executor com ele.
- Desmarcada: o executor é criado sem token. Mais tarde, clique em **Regenerate registration
  token** (o ícone de duas setas) na linha do executor para obter um token de registo de
  utilização única, válido durante uma hora, e registe o executor com
  `semaphore runner register`.

Um executor não registado não recebe tarefas.

## Webhook {#runner-webhook}

Opcional. Um URL que o servidor chama quando atribui uma tarefa a este executor, por exemplo para
iniciar o executor a pedido. O servidor envia um pedido `POST` com um corpo JSON:

```json
{"action": "start", "project_id": 1, "task_id": 42, "template_id": 7, "runner_id": 3}
```

O URL tem de responder com o estado 200 ou 204 no prazo de 30 segundos; caso contrário, a tarefa
falha. Um executor com webhook é sempre considerado online, mesmo que não tenha consultado o
servidor recentemente. Deixe o campo vazio para um executor que está sempre em funcionamento.

## Número máximo de tarefas paralelas {#runner-maxParallelTasks}

O número de tarefas que o executor pode executar em simultâneo. Quando o limite é atingido, a
tarefa seguinte vai para outro executor correspondente ou aguarda na fila. Deixe o campo vazio
ou defina `0` para não ter limite.

## Habilitado {#runner-active}

Só um executor habilitado recebe novas tarefas. Desmarque a caixa para retirar o executor de
serviço sem o eliminar; o interruptor na lista de executores faz o mesmo.

A caixa está desmarcada num novo executor: assinale-a, ou habilite o executor mais tarde na
lista.
