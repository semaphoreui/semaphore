---
form: src/components/UserForm.vue, src/components/GlobalRoleAssignments.vue
---

# Modifier l'utilisateur

Un utilisateur est un compte qui se connecte à Semaphore. L'onglet **Settings** contient le profil
du compte et ses options ; **Enregistrer** les écrit. L'onglet **Security** contient le mot de
passe, l'authentification à deux facteurs et les comptes de connexion liés, et l'onglet **Roles**
les rôles globaux de l'utilisateur : les deux agissent immédiatement, sans **Enregistrer**. Ce
qu'un utilisateur peut faire dans un projet se définit sur la page **Équipe** de ce projet, pas
ici.

## Nom {#user-name}

Le nom complet de l'utilisateur, affiché dans la barre latérale et dans les listes d'utilisateurs
et de membres de projet. Le champ est obligatoire.

Pour un utilisateur **Externe**, Semaphore reprend le nom du fournisseur d'identité à la
connexion, lorsque le fournisseur en transmet un ; une modification faite ici peut donc ne pas
survivre à la connexion suivante.

## Nom d'utilisateur {#user-username}

L'identifiant de connexion de l'utilisateur. Le champ est obligatoire. L'utilisateur se connecte
avec le nom d'utilisateur ou l'e-mail, et le mot de passe.

Le nom d'utilisateur d'un utilisateur **Externe** ne peut pas être modifié : le serveur refuse de
l'enregistrer.

## Email {#user-email}

L'adresse e-mail de l'utilisateur. Le champ est obligatoire. Semaphore y envoie les alertes de
tâches de l'utilisateur, et une application d'authentification l'affiche comme nom du compte. Elle
est privée : seuls l'utilisateur et ceux qui peuvent gérer les utilisateurs la voient.

Pour un utilisateur **Externe**, l'e-mail vient du fournisseur d'identité : le champ est en
lecture seule, et Semaphore met l'adresse à jour à chaque connexion.

## Mot de passe {#user-password}

Affiché lors de la création d'un nouvel utilisateur.

Le mot de passe avec lequel l'utilisateur se connecte. Le champ est obligatoire, sauf pour un
utilisateur **Externe**, qui n'a pas de mot de passe dans Semaphore. Pour changer le mot de passe
d'un utilisateur existant, utilisez **Change password** dans l'onglet **Security**.

## Envoyer des alertes {#user-alert}

Envoie un e-mail à l'utilisateur lorsqu'une tâche échoue dans un projet où l'utilisateur peut voir
la tâche. Le projet doit autoriser les alertes, et l'administrateur du serveur doit avoir activé
les alertes par e-mail. Un modèle de tâche peut empêcher ses tâches échouées d'envoyer des alertes.

L'option n'a pas d'effet sur les alertes d'un projet envoyées à Telegram, Slack et aux autres
messageries.

## Utilisateur administrateur {#user-admin}

Affiché lorsque vous êtes administrateur.

Un administrateur peut tout faire sur le serveur : gérer les utilisateurs, les rôles globaux et
les paramètres du serveur, et ouvrir chaque projet avec tous les droits.

Vous ne pouvez pas retirer l'option de votre propre compte, et le serveur conserve au moins un
administrateur : il refuse de retirer l'option au dernier.

## Pro user {#user-pro}

Affiché dans l'édition Pro. Seul un administrateur peut modifier l'option.

Attribue à l'utilisateur une place de l'abonnement Pro. L'abonnement comporte un nombre fixe
d'utilisateurs Pro : lorsqu'il est atteint, le serveur refuse d'enregistrer l'option. Sans
abonnement actif, l'option n'a pas d'effet.

## Externe {#user-external}

Affiché lorsque vous êtes administrateur. L'option ne peut être définie qu'à la création de
l'utilisateur.

Un utilisateur externe se connecte via LDAP ou OpenID Connect, et non avec un mot de passe de
Semaphore : le champ du mot de passe est désactivé, et une connexion avec mot de passe est
refusée. Le nom et l'e-mail sont repris du fournisseur d'identité à chaque connexion. Laissez la
case vide pour un utilisateur qui se connecte avec un mot de passe.

## Change password {#user-changePassword}

Affiché dans l'onglet **Security** d'un utilisateur existant qui n'est pas **Externe**, lorsque le
serveur autorise la connexion avec un mot de passe.

Ouvre une boîte de dialogue qui définit un nouveau mot de passe. Pour changer votre propre mot de
passe, vous devez indiquer le mot de passe actuel ; pour réinitialiser celui d'un autre
utilisateur, ce n'est pas nécessaire.

## Time-based one-time password {#user-totp}

Affiché dans l'onglet **Security** lorsque l'administrateur a activé l'authentification à deux
facteurs sur le serveur.

Active l'authentification à deux facteurs pour cet utilisateur. Lorsque vous activez
l'interrupteur, Semaphore crée immédiatement un secret et l'affiche sous forme de code QR :
scannez-le avec une application d'authentification. Dès lors, une connexion demande le code de
l'application après le mot de passe.

Lorsque le serveur autorise la récupération, un code de récupération s'affiche sous le code QR. Il
n'est affiché qu'une seule fois : copiez-le et conservez-le en lieu sûr. Désactiver
l'interrupteur supprime immédiatement le secret.

## Fournisseur LDAP {#user-ldapProvider}

Affiché dans l'onglet **Security** de votre propre compte lorsque le serveur dispose d'un
fournisseur LDAP auquel votre compte n'est pas encore lié.

Le serveur LDAP auquel lier votre compte, lorsqu'il y en a plusieurs.

## LDAP username {#user-ldapUsername}

Affiché avec **Fournisseur LDAP**.

Votre nom d'utilisateur sur le serveur LDAP. La liaison vous permet de vous connecter à ce compte
également via LDAP.

## LDAP password {#user-ldapPassword}

Affiché avec **Fournisseur LDAP**.

Votre mot de passe sur le serveur LDAP. Semaphore l'utilise une seule fois, pour vérifier que le
compte LDAP vous appartient, et ne le conserve pas.

## Link {#user-ldapLink}

Affiché avec **Fournisseur LDAP**.

Lie votre compte au compte LDAP ci-dessus. La liaison est refusée lorsque les identifiants sont
erronés, lorsque le compte LDAP est lié à un autre utilisateur, ou lorsque votre compte est déjà
lié à ce fournisseur. Un compte lié apparaît sous **Linked accounts** ; fermez sa puce pour le
délier.

## Global role {#global-role-select}

Affiché dans l'onglet **Roles** d'un utilisateur existant, dans l'édition Pro, lorsque vous pouvez
gérer les rôles globaux.

Un rôle global à attribuer à l'utilisateur, parmi les rôles qu'il ne détient pas encore. Un rôle
global accorde des permissions sur l'ensemble du serveur : gérer les utilisateurs, les rôles
globaux ou les paramètres du serveur, ou lire le journal d'audit. Les rôles globaux eux-mêmes se
créent dans l'onglet **Rôles globaux** de la page Utilisateurs.

## Assign {#globalRoleAssign-assign}

Affiché avec **Global role**.

Attribue immédiatement le rôle sélectionné à l'utilisateur. Les rôles que l'utilisateur détient
sont listés en dessous, chacun avec un bouton pour le retirer. **Effective global permissions**
indique ce que l'utilisateur peut faire au final, et d'où vient chaque permission.
