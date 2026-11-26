---
form: src/components/EditGlobalRoleForm.vue
---

# Edit Global Role

Un rôle global accorde des permissions sur l'ensemble de l'instance de Semaphore aux utilisateurs
auxquels il est attribué ; vous l'attribuez dans la boîte de dialogue d'un utilisateur. Il ne donne
aucun accès à l'intérieur des projets : celui-ci ne vient que des rôles qu'un utilisateur détient
en tant que membre d'un projet. Un administrateur détient toutes les permissions globales, quels que
soient les rôles attribués. Les rôles globaux sont une fonctionnalité Pro.

## Nom {#globalRole-name}

Le nom sous lequel le rôle apparaît dans l'onglet Rôles globaux de la page Utilisateurs et dans la
boîte de dialogue d'un utilisateur. Le champ est obligatoire. Vous pouvez renommer un rôle à tout
moment.

## Slug {#globalRole-slug}

L'identifiant du rôle. Semaphore le remplit à partir du nom pendant que vous saisissez le nom,
jusqu'à ce que vous modifiiez vous-même le slug. Il comporte de 1 à 100 lettres, chiffres, traits
de soulignement et traits d'union, par exemple `user-admins`.

Le slug doit être unique parmi les rôles globaux ; un rôle de projet peut avoir le même slug. Il ne
peut plus être modifié une fois le rôle enregistré : le champ est verrouillé lorsque vous modifiez
un rôle.

## Global permissions {#globalRole-permissionsTitle}

Ce qu'un utilisateur qui détient le rôle peut faire sur l'ensemble de l'instance. Un nouveau rôle
n'en a aucune de cochée. Les permissions de tous les rôles globaux d'un utilisateur s'additionnent.

- **Can manage users** : créer des utilisateurs, les modifier et les supprimer, et voir leurs
  détails dans la liste des utilisateurs. Un tel utilisateur ne peut pas faire d'un utilisateur un
  administrateur ou un utilisateur Pro, et ne peut pas modifier ni supprimer un administrateur ou un
  utilisateur qui détient une permission globale qu'il ne détient pas lui-même.
- **Can manage global roles** : créer, modifier et supprimer les rôles globaux et les rôles de
  projet partagés, et décider quels utilisateurs détiennent quel rôle global. Ne l'accordez qu'aux
  personnes à qui vous confiez l'instance entière.
- **Can manage system settings** : modifier les paramètres de l'instance, les applications, les
  exécuteurs globaux et les correspondances de groupes ; voir le cluster et les informations sur le
  serveur ; voir et supprimer les tâches de tous les projets ; effacer le cache du serveur.
- **Can read the audit log** : lire le journal d'audit de l'instance et l'exporter.
