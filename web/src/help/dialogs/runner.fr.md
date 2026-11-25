---
form: src/components/RunnerForm.vue
---

# Modifier l'exécuteur

Un exécuteur est un processus `semaphore runner` distinct qui exécute des tâches sur sa propre
machine au lieu du serveur Semaphore. Une tâche est envoyée à un exécuteur lorsque le serveur est
configuré pour exécuter les tâches sur des exécuteurs, ou lorsque le modèle de tâche ou son
inventaire a une étiquette d'exécuteur. Le serveur choisit alors un exécuteur activé, enregistré,
en ligne et qui n'a pas atteint sa limite de tâches parallèles : d'abord les exécuteurs du projet,
puis les exécuteurs globaux, dans un ordre aléatoire au sein de chaque groupe. Lorsque tous les
exécuteurs correspondants sont hors ligne ou occupés, la tâche attend dans la file d'attente ;
lorsqu'aucun exécuteur ne correspond, la tâche échoue.

## Nom {#runner-name}

Obligatoire. Le nom sous lequel l'exécuteur apparaît sur la page Exécuteurs, par exemple l'hôte sur
lequel il s'exécute. Il ne change pas les tâches que l'exécuteur reçoit.

## Tags {#runner-tags}

Disponible dans Semaphore Pro ; dans les autres éditions, le champ est verrouillé.

Les étiquettes qui sélectionnent les tâches de cet exécuteur. Une tâche avec une étiquette
d'exécuteur ne va qu'aux exécuteurs qui portent cette étiquette ; l'étiquette du modèle de tâche
l'emporte sur celle de l'inventaire. Saisissez une étiquette et appuyez sur Entrée, ou choisissez-en
une utilisée par d'autres exécuteurs. Les étiquettes vides et en double sont supprimées à
l'enregistrement.

Un exécuteur de projet nécessite au moins une étiquette, sauf si **Is default** est coché.

## Is default {#runner-isDefault}

Une tâche sans étiquette d'exécuteur ne va qu'aux exécuteurs marqués par défaut. Décochez la case
pour un exécuteur qui ne doit prendre que les tâches portant l'une de ses étiquettes.

La case est cochée pour un nouvel exécuteur global et décochée pour un nouvel exécuteur de projet.

## Register {#runner-registered}

Affiché lorsque vous créez un exécuteur.

- Coché (par défaut) : le serveur crée l'exécuteur avec son jeton. Après **Créer**, la boîte de
  dialogue affiche le jeton et les commandes qui démarrent l'exécuteur avec celui-ci.
- Décoché : l'exécuteur est créé sans aucun jeton. Plus tard, cliquez sur
  **Regenerate registration token** (l'icône à deux flèches) dans la ligne de l'exécuteur pour
  obtenir un jeton d'enregistrement à usage unique, valable une heure, et enregistrez l'exécuteur
  avec `semaphore runner register`.

Un exécuteur non enregistré ne reçoit aucune tâche.

## Webhook {#runner-webhook}

Optionnel. Une URL que le serveur appelle lorsqu'il confie une tâche à cet exécuteur, par exemple
pour démarrer l'exécuteur à la demande. Le serveur envoie une requête `POST` avec un corps JSON :

```json
{"action": "start", "project_id": 1, "task_id": 42, "template_id": 7, "runner_id": 3}
```

L'URL doit répondre avec le statut 200 ou 204 dans les 30 secondes ; sinon la tâche échoue. Un
exécuteur avec un webhook est toujours considéré comme en ligne, même s'il n'a pas interrogé le
serveur récemment. Laissez le champ vide pour un exécuteur qui fonctionne en permanence.

## Nombre maximum de tâches parallèles {#runner-maxParallelTasks}

Le nombre de tâches que l'exécuteur peut exécuter simultanément. Lorsque la limite est atteinte, la
tâche suivante va à un autre exécuteur correspondant ou attend dans la file d'attente. Laissez le
champ vide ou indiquez `0` pour ne fixer aucune limite.

## Activé {#runner-active}

Seul un exécuteur activé reçoit de nouvelles tâches. Décochez la case pour mettre l'exécuteur hors
service sans le supprimer ; l'interrupteur dans la liste des exécuteurs fait de même.

La case est décochée pour un nouvel exécuteur : cochez-la, ou activez l'exécuteur plus tard dans la
liste.
