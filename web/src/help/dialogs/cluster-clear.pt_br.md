---
form: src/views/Cluster.vue
---

# Clear tasks from Redis

Os nós de um cluster compartilham no Redis o estado do pool de tarefas: quais tarefas aguardam,
quais estão em execução e qual nó é dono de cada uma delas. Esta caixa de diálogo exclui do Redis
os grupos de registros que você marcar, para recuperar um cluster travado, por exemplo com uma
tarefa concluída que ainda aparece como em execução. Nada é excluído do banco de dados: as tarefas
mantêm o status, o log e o histórico.

Seja o que for que você marcar, cada nó também descarta as tarefas que mantém na memória e as lê
novamente do Redis e do banco de dados. Apenas **Queue** vem marcado quando a caixa de diálogo
abre; marque apenas os grupos que estão travados. O resultado mostra quantos registros foram
removidos.

## Queue {#clusterClearDialog-queue}

As tarefas que aguardam um nó para iniciá-las. As tarefas limpas não são iniciadas: elas
permanecem no banco de dados com o status Waiting. Pare-as, ou execute os modelos delas novamente.

## Running {#clusterClearDialog-running}

As tarefas que o cluster considera em execução. Limpar não as para: uma tarefa continua em
execução no seu executor. Mas o cluster perde o controle dela: a tarefa deixa de contar para os
limites de tarefas paralelas do servidor e do seu executor, então mais tarefas podem ser iniciadas
do que os limites permitem, e uma tarefa que um executor ainda não pegou não é mais entregue a ele.

## Active by project {#clusterClearDialog-active}

As tarefas não concluídas de cada projeto. Elas impõem o limite de tarefas paralelas de um projeto
e impedem que uma segunda tarefa de um modelo seja iniciada enquanto uma está em execução, quando o
modelo não permite tarefas paralelas. Limpe este grupo quando as tarefas de um projeto aguardam
indefinidamente atrás de tarefas que já terminaram. As tarefas em execução nesse momento deixam de
contar, então novas tarefas do projeto podem ser iniciadas ao lado delas.

## Aliases {#clusterClearDialog-aliases}

Os nomes aleatórios que as tarefas de Terraform, OpenTofu e Terragrunt recebem durante a sua
existência, mapeados para a tarefa. Uma tarefa em execução alcança o seu estado do Terraform no
servidor por esse nome: depois da limpeza, uma tarefa em execução desses apps não pode mais ser
encontrada por ele.

## Claims {#clusterClearDialog-claims}

Qual nó é dono de cada tarefa. Um nó renova as reivindicações das suas tarefas a cada minuto; uma
reivindicação que não é renovada expira depois de cinco minutos. Uma tarefa sem reivindicação
conta como órfã: uma tarefa que ainda não chegou a um executor é colocada de volta na fila para
outro nó, e uma tarefa que um executor já executa é deixada para terminar lá.

## Runtime fields {#clusterClearDialog-runtimeFields}

Os detalhes de cada tarefa que o banco de dados não guarda: o usuário que a iniciou, a versão com
que foi iniciada, o alias, o executor e o status dela, e o projeto de cada tarefa. Sem eles, um nó
que não iniciou uma tarefa mostra menos sobre ela, e o limpador de tarefas órfãs ignora uma tarefa
em execução, então um registro travado deixa de ser removido por conta própria.
