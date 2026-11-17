---
form: src/views/project/Team.vue
skip: team-rowRolesSelect
---

# Équipe

Les membres du projet et les rôles qu'ils détiennent. Un rôle détermine ce qu'un membre peut faire
dans le projet : lancer des tâches, modifier les modèles et les autres ressources, gérer l'équipe.
Un membre peut avoir plusieurs rôles et obtient chaque permission que l'un d'eux accorde.

- **Nouveau membre de l'équipe** ajoute un utilisateur au projet avec un rôle. Le bouton est
  visible des membres autorisés à gérer l'équipe.
- La colonne **Rôle** change les rôles d'un membre immédiatement, sans enregistrement. Un membre
  qui ne peut pas gérer l'équipe voit les rôles sans pouvoir les changer.
- Un rôle marqué d'un cadenas vient d'un groupe de l'utilisateur dans LDAP ou OIDC, que
  l'administrateur a associé à ce rôle. Il ne peut pas être retiré ici : il disparaît quand
  l'utilisateur quitte le groupe ou que l'association est supprimée.
- Le projet doit garder au moins un Owner. Semaphore refuse de retirer son dernier Owner ou de lui
  enlever le rôle Owner, sauf si c'est un administrateur de Semaphore qui le fait.
- **Quitter le projet** vous retire du projet. Le bouton n'apparaît que si l'administrateur autorise
  les membres à quitter les projets ; le dernier Owner ne peut pas non plus quitter le projet.
