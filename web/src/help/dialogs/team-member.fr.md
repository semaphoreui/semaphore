---
form: src/components/TeamMemberForm.vue
---

# Nouveau membre de l'équipe

Ajoute un utilisateur au projet avec un rôle. Le rôle détermine ce que le membre peut voir et faire
dans le projet ; rien d'autre de son compte ne change. Lorsque l'administrateur a activé les
invitations, l'utilisateur reçoit à la place une invitation et devient membre une fois qu'il l'a
acceptée.

## Email ou Username {#teamMember-inviteType}

Affiché lorsque les utilisateurs peuvent être invités à la fois par e-mail et par nom
d'utilisateur.

La manière de choisir la personne : **Email** pour inviter une adresse e-mail, **Username** pour
choisir un utilisateur qui a déjà un compte sur ce serveur Semaphore.

## Email {#teamMember-email}

Affiché lorsque la personne est invitée par e-mail.

L'adresse e-mail de la personne à inviter. Elle devient membre du projet une fois qu'elle a accepté
l'invitation.

## Utilisateur {#teamMember-userId}

Affiché lorsque la personne est choisie par nom d'utilisateur.

L'utilisateur à ajouter, affiché avec son nom d'utilisateur et son nom complet. Le champ est
obligatoire. La liste contient chaque utilisateur du serveur Semaphore qui n'est pas encore membre
du projet : une personne sans compte doit d'abord en obtenir un auprès d'un administrateur.

## Rôle {#teamMember-role}

Ce que le membre peut faire dans le projet. Le champ est obligatoire. Les rôles intégrés :

- **Owner** : tout, y compris les paramètres du projet, son équipe et sa suppression.
- **Manager** : exécute des tâches et des workflows, approuve des workflows, et crée, modifie et
  supprime les modèles de tâches, inventaires, clés, dépôts, groupes de variables, calendriers,
  intégrations et workflows. Un manager ne peut pas modifier les paramètres du projet ni son équipe.
- **Task Runner** : voit le projet, exécute des tâches et des workflows et approuve des workflows.
  Un task runner ne peut rien modifier dans le projet.
- **Guest** : voit le projet, ses tâches et leurs journaux, et ne modifie rien.

La liste contient aussi les rôles créés pour ce projet (**Rôles du projet**) et les rôles partagés
par tous les projets (**Shared roles**). Un tel rôle accorde les permissions définies pour lui sur
la page Rôles. Une invitation ne peut porter qu'un rôle intégré.

Un membre qui peut gérer l'équipe peut attribuer n'importe quel rôle, **Owner** compris. La boîte
de dialogue attribue un seul rôle ; ajoutez-en d'autres ensuite dans la colonne **Rôle** de la
liste de l'équipe.
