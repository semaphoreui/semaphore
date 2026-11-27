---
form: src/components/EditRoleForm.vue
---

# Edit Role

Un rôle est un ensemble nommé de permissions à l'intérieur d'un projet. Vous attribuez des rôles
aux membres d'un projet, et un membre peut faire ce que l'un quelconque de ses rôles autorise.
Ouverte depuis la page Rôles d'un projet, la boîte de dialogue modifie un rôle de ce projet
uniquement ; ouverte depuis la page Rôles hors d'un projet, elle modifie un rôle partagé que chaque
projet peut utiliser. Les rôles intégrés Owner, Manager, Task Runner et Guest ne peuvent pas être
modifiés.

## Nom {#role-name}

Le nom sous lequel le rôle apparaît sur la page Rôles et partout où un rôle est choisi, par exemple
lorsque vous ajoutez un membre à un projet. Le champ est obligatoire. Vous pouvez renommer un rôle à
tout moment.

## Slug {#role-slug}

L'identifiant du rôle. Semaphore le remplit à partir du nom pendant que vous saisissez le nom,
jusqu'à ce que vous modifiiez vous-même le slug. Il comporte de 1 à 100 lettres, chiffres, traits
de soulignement et traits d'union, par exemple `release-managers`.

Le slug ne peut plus être modifié une fois le rôle enregistré : le champ est verrouillé lorsque vous
modifiez un rôle.

- Un slug doit être unique, et la vérification ignore la casse des lettres. Le slug d'un rôle
  partagé doit différer de tous les autres rôles ; le slug d'un rôle de projet doit différer des
  rôles partagés et des autres rôles de son projet.
- `owner`, `manager`, `task_runner`, `guest` et `all` sont réservés, quelle que soit la casse.

## Permissions {#role-permissionsTitle}

Ce que les membres qui détiennent le rôle peuvent faire dans le projet. Un nouveau rôle n'en a
aucune de cochée. Les permissions de tous les rôles qu'un membre détient s'additionnent ; un
administrateur de Semaphore peut tout faire dans chaque projet, quels que soient les rôles.

- **Can read project** : voir le projet : ses modèles de tâches, ses tâches et leurs journaux, ses
  inventaires, clés, dépôts, groupes de variables, calendriers et membres. Chaque lecture dans le
  projet la requiert, et aucune autre permission ne l'inclut ; cochez-la donc pour presque chaque
  rôle. Sans elle, un membre ne voit que les modèles de tâches et les workflows qui accordent
  eux-mêmes l'accès au rôle, avec leurs tâches.
- **Can run project tasks** : démarrer des tâches à partir des modèles de tâches, les arrêter, et
  confirmer ou rejeter une tâche qui attend une confirmation.
- **Can update project** : modifier les paramètres du projet et le supprimer.
- **Can manage project resources** : créer, modifier et supprimer les modèles de tâches,
  inventaires, clés, dépôts, groupes de variables, calendriers, vues, intégrations, stockages de
  secrets, exécuteurs de projet et les rôles du projet ; supprimer des tâches ; définir quels rôles
  peuvent utiliser un modèle de tâche.
- **Can manage project users** : ajouter des membres au projet, modifier leurs rôles et les
  retirer.
- **Can run workflows** : démarrer un workflow et arrêter ses exécutions.
- **Can edit workflows** : créer, modifier et supprimer des workflows, et définir quels rôles
  peuvent utiliser un workflow.
- **Can approve workflows** : approuver ou rejeter une étape d'une exécution de workflow qui attend
  une approbation.
