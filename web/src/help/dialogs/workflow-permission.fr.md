---
form: src/components/EditTemplatePermissionForm.vue
---

# Workflow permissions

Accorde à un rôle des permissions supplémentaires sur ce seul workflow. Un membre qui détient le
rôle les reçoit en plus des permissions dont le rôle dispose dans tout le projet : une attribution
ici ne fait qu'ajouter, elle ne retire jamais de permission. Un membre qui a plusieurs rôles reçoit
toutes les permissions accordées à l'un d'eux.

## Rôle {#templatePerm-roleSlug}

Le rôle qui reçoit les permissions. La liste affiche les rôles de ce projet, les rôles partagés et
les rôles intégrés. Le champ est obligatoire, et un rôle ne peut avoir qu'une seule attribution sur
un workflow : le serveur refuse une deuxième.

Le rôle d'une attribution enregistrée ne peut pas être modifié. Le serveur conserve le rôle et
n'enregistre que les permissions, quel que soit votre choix ici. Pour accorder les permissions à un
autre rôle, supprimez cette attribution de la liste et ajoutez-en une nouvelle.

## Permissions {#templatePerm-permissionsTitle}

Ce qu'un membre qui détient le rôle peut faire avec ce workflow :

- **Can see this workflow** : le workflow figure dans la liste des workflows, avec ses exécutions.
- **Can run this workflow** : lancer des exécutions du workflow.
- **Can edit this workflow** : modifier et supprimer le workflow, et gérer ces permissions.
- **Can approve this workflow** : approuver ou rejeter les étapes d'approbation de ses exécutions.

Exécuter, modifier ou approuver inclut le droit de voir le workflow : lorsque vous cochez l'une de
ces cases, **Can see this workflow** est aussi coché et ne peut pas être décoché. Laissez toutes les
cases décochées pour enregistrer une attribution qui n'accorde rien.
