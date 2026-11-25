---
form: src/components/EditTemplatePermissionForm.vue
---

# Template permissions

Accorde à un rôle des permissions supplémentaires sur ce seul modèle. Un membre qui détient le rôle
les reçoit en plus des permissions dont le rôle dispose dans tout le projet : une attribution ici
ne fait qu'ajouter, elle ne retire jamais de permission. Un membre qui a plusieurs rôles reçoit
toutes les permissions accordées à l'un d'eux.

## Rôle {#templatePerm-roleSlug}

Le rôle qui reçoit les permissions. La liste affiche les rôles de ce projet, les rôles partagés et
les rôles intégrés. Le champ est obligatoire, et un rôle ne peut avoir qu'une seule attribution sur
un modèle : le serveur refuse une deuxième.

Le rôle d'une attribution enregistrée ne peut pas être modifié. Le serveur conserve le rôle et
n'enregistre que les permissions, quel que soit votre choix ici. Pour accorder les permissions à un
autre rôle, supprimez cette attribution de la liste et ajoutez-en une nouvelle.

## Permissions {#templatePerm-permissionsTitle}

Ce qu'un membre qui détient le rôle peut faire avec ce modèle :

- **Can see this template** : le modèle figure dans la liste des modèles, avec ses tâches et leurs
  journaux.
- **Can run this template** : lancer des tâches à partir du modèle, et arrêter, confirmer ou
  rejeter ses tâches.
- **Can edit this template** : modifier et supprimer le modèle, arrêter toutes ses tâches et gérer
  ces permissions.

Exécuter ou modifier inclut le droit de voir le modèle : lorsque vous cochez l'une de ces cases,
**Can see this template** est aussi coché et ne peut pas être décoché. Laissez toutes les cases
décochées pour enregistrer une attribution qui n'accorde rien.
