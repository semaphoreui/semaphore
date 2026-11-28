---
form: src/views/Cluster.vue
---

# Clear tasks from Redis

Os nós de um cluster partilham no Redis o estado do conjunto de tarefas: que tarefas aguardam,
quais estão em execução e que nó é dono de cada uma delas. Esta caixa de diálogo elimina do Redis
os grupos de registos que assinalar, para recuperar um cluster que está bloqueado, por exemplo com
uma tarefa terminada que continua a aparecer como em execução. Nada é eliminado da base de dados:
as tarefas mantêm o seu estado, o seu log e o seu histórico.

Seja o que for que assinale, cada nó também descarta as tarefas que mantém em memória e volta a
lê-las do Redis e da base de dados. Apenas **Queue** está assinalado quando a caixa de diálogo
abre; assinale apenas os grupos que estão bloqueados. O resultado mostra quantos registos foram
removidos.

## Queue {#clusterClearDialog-queue}

As tarefas que aguardam que um nó as inicie. As tarefas limpas não são iniciadas: ficam na base de
dados com o estado Waiting. Pare-as ou volte a executar os respetivos modelos.

## Running {#clusterClearDialog-running}

As tarefas que o cluster conta como em execução. A limpeza não as para: uma tarefa continua a ser
executada no seu executor. Mas o cluster perde-lhe o rasto: a tarefa deixa de contar para os
limites de tarefas paralelas do servidor e do seu executor, pelo que podem iniciar-se mais tarefas
do que os limites permitem, e uma tarefa que um executor ainda não assumiu deixa de lhe ser
entregue.

## Active by project {#clusterClearDialog-active}

As tarefas não terminadas de cada projeto. Elas impõem o limite de tarefas paralelas de um projeto
e impedem que uma segunda tarefa de um modelo se inicie enquanto outra está em execução, quando o
modelo não permite tarefas paralelas. Limpe este grupo quando as tarefas de um projeto aguardam
indefinidamente atrás de tarefas que já terminaram. As tarefas em execução nesse momento deixam de
contar, pelo que novas tarefas do projeto podem iniciar-se ao lado delas.

## Aliases {#clusterClearDialog-aliases}

Os nomes aleatórios que as tarefas de Terraform, OpenTofu e Terragrunt recebem durante a sua
existência, associados à tarefa. Uma tarefa em execução chega ao seu estado do Terraform no
servidor através deste nome: depois da limpeza, uma tarefa em execução destas aplicações deixa de
poder ser encontrada através dele.

## Claims {#clusterClearDialog-claims}

Que nó é dono de cada tarefa. Um nó renova as reivindicações das suas tarefas a cada minuto; uma
reivindicação que não é renovada expira ao fim de cinco minutos. Uma tarefa sem reivindicação
conta como órfã: uma tarefa que ainda não chegou a um executor é devolvida à fila para outro nó, e
uma tarefa que um executor já executa é deixada a terminar aí.

## Runtime fields {#clusterClearDialog-runtimeFields}

Os dados de cada tarefa que a base de dados não guarda: o utilizador que a iniciou, a versão com
que foi iniciada, o seu alias, o seu executor e o seu estado, e o projeto de cada tarefa. Sem
eles, um nó que não iniciou uma tarefa mostra menos sobre ela, e o processo de limpeza de tarefas
órfãs ignora uma tarefa em execução, pelo que um registo bloqueado deixa de ser removido por si
só.
