---
form: src/components/RoleMappingForm.vue
---

# New mapping

Une correspondance attribue un rôle dans un projet à chaque utilisateur qui se connecte via un
fournisseur LDAP ou OpenID Connect et qui y appartient à un groupe. Semaphore lit les groupes de
l'utilisateur à chaque connexion via le fournisseur et attribue ou retire les rôles correspondants
en conséquence, de sorte qu'un utilisateur qui quitte le groupe perd le rôle à sa connexion
suivante. Un rôle attribué par correspondance fait de l'utilisateur un membre du projet. Une
correspondance ne peut pas être modifiée : supprimez-la et ajoutez-en une nouvelle.

## Provider {#roleMapping-provider}

L'annuaire LDAP ou le fournisseur OpenID Connect dont vient le groupe. La liste contient les
fournisseurs définis dans la configuration du serveur. Le champ est obligatoire.

Les groupes ne sont lus que lorsque le fournisseur est configuré pour cela : un fournisseur LDAP
nécessite une base de recherche des groupes (`group_search_dn`), un fournisseur OpenID Connect
indique la revendication contenant les groupes (`groups_claim`). Une correspondance d'un
fournisseur retiré ensuite de la configuration reste dans la liste, marquée
**Provider not configured**.

## Group {#roleMapping-group}

Le nom du groupe tel que le fournisseur le transmet. Le champ est obligatoire.

- LDAP : la valeur de l'attribut de nom de groupe de l'entrée du groupe, `cn` par défaut, par
  exemple `devops`. Un DN complet tel que `cn=devops,ou=groups,dc=example,dc=com` est réduit à son
  `cn`.
- OpenID Connect : l'une des valeurs de la revendication des groupes, telle que le fournisseur
  l'envoie.

## Project {#roleMapping-projectId}

Le projet dans lequel le rôle est attribué. Le champ est obligatoire. Un administrateur voit tous
les projets ; tout autre utilisateur ne voit que les projets dont il est membre.

## Rôle {#roleMapping-role}

Le rôle que les membres du groupe obtiennent dans le projet. Le champ est obligatoire et devient
disponible une fois un projet sélectionné. La liste contient les rôles de ce projet, les rôles
partagés et les rôles intégrés. Lorsque vous choisissez un autre projet, un rôle qui n'y existe pas
est effacé.

Un rôle **Owner** attribué par une correspondance ne compte pas lorsque Semaphore vous empêche de
retirer le rôle au dernier propriétaire d'un projet, car il peut disparaître à la connexion
suivante. Conservez au moins un propriétaire ajouté manuellement.
