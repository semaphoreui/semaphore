---
form: src/views/Cluster.vue
---

# Clear tasks from Redis

Les nœuds d'un cluster partagent l'état du pool de tâches dans Redis : quelles tâches attendent,
lesquelles s'exécutent et quel nœud possède chacune d'elles. Cette boîte de dialogue supprime de
Redis les groupes d'enregistrements que vous cochez, pour rétablir un cluster bloqué, par exemple
avec une tâche terminée qui apparaît toujours comme en cours d'exécution. Rien n'est supprimé de la
base de données : les tâches conservent leur statut, leur journal et leur historique.

Quoi que vous cochiez, chaque nœud abandonne aussi les tâches qu'il garde en mémoire et les relit
depuis Redis et la base de données. Seul **Queue** est coché à l'ouverture de la boîte de dialogue ;
ne cochez que les groupes qui sont bloqués. Le résultat indique combien d'enregistrements ont été
supprimés.

## Queue {#clusterClearDialog-queue}

Les tâches qui attendent qu'un nœud les démarre. Les tâches effacées ne sont pas démarrées : elles
restent dans la base de données avec le statut Waiting. Arrêtez-les, ou exécutez à nouveau leurs
modèles.

## Running {#clusterClearDialog-running}

Les tâches que le cluster compte comme en cours d'exécution. L'effacement ne les arrête pas : une
tâche continue de s'exécuter sur son exécuteur. Mais le cluster en perd la trace : la tâche ne
compte plus dans les limites de tâches parallèles du serveur et de son exécuteur, de sorte que plus
de tâches peuvent démarrer que les limites ne le permettent, et une tâche qu'un exécuteur n'a pas
encore prise ne lui est plus remise.

## Active by project {#clusterClearDialog-active}

Les tâches non terminées de chaque projet. Elles font respecter la limite de tâches parallèles d'un
projet et empêchent une deuxième tâche d'un modèle de démarrer pendant qu'une autre s'exécute,
lorsque le modèle n'autorise pas les tâches parallèles. Effacez ce groupe lorsque les tâches d'un
projet attendent indéfiniment derrière des tâches qui sont terminées. Les tâches qui s'exécutent à
ce moment ne comptent plus, de sorte que de nouvelles tâches du projet peuvent démarrer à côté
d'elles.

## Aliases {#clusterClearDialog-aliases}

Les noms aléatoires que les tâches Terraform, OpenTofu et Terragrunt reçoivent pour leur durée de
vie, associés à la tâche. Une tâche en cours atteint son état Terraform sur le serveur par ce nom :
après l'effacement, une tâche en cours de ces applications ne peut plus être retrouvée par ce nom.

## Claims {#clusterClearDialog-claims}

Quel nœud possède chaque tâche. Un nœud renouvelle les revendications de ses tâches chaque minute ;
une revendication qui n'est pas renouvelée expire au bout de cinq minutes. Une tâche sans
revendication est considérée comme orpheline : une tâche qui n'a pas encore atteint un exécuteur est
remise dans la file d'attente pour un autre nœud, et une tâche qu'un exécuteur exécute déjà est
laissée se terminer sur celui-ci.

## Runtime fields {#clusterClearDialog-runtimeFields}

Les détails de chaque tâche que la base de données ne contient pas : l'utilisateur qui l'a
démarrée, la version avec laquelle elle a été démarrée, son alias, son exécuteur et son statut,
ainsi que le projet de chaque tâche. Sans eux, un nœud qui n'a pas démarré une tâche affiche moins
d'informations à son sujet, et le nettoyeur des tâches orphelines ignore une tâche en cours, de
sorte qu'un enregistrement bloqué n'est plus supprimé de lui-même.
