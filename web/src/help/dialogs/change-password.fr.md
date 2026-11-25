---
form: src/components/ChangePasswordForm.vue
---

# Changer le mot de passe

Définit un nouveau mot de passe pour se connecter à Semaphore avec un nom d'utilisateur et un mot
de passe. Le nouveau mot de passe prend effet lorsque vous appuyez sur **Enregistrer**. Un
utilisateur externe, qui se connecte via LDAP ou OpenID Connect, n'a pas ici de mot de passe à
changer.

## Current password {#changePassword-currentPassword}

Affiché lorsque vous changez votre propre mot de passe.

Votre mot de passe actuel. Le champ est obligatoire : Semaphore le vérifie avant de définir le
nouveau, afin qu'une session ou un jeton API volé ne puisse pas être transformé en votre mot de
passe. Le serveur refuse la modification lorsque le mot de passe ne correspond pas.

Un administrateur ou un gestionnaire des utilisateurs qui réinitialise le mot de passe d'un autre
utilisateur ne voit pas ce champ.

## Mot de passe {#changePassword-password}

Le nouveau mot de passe. Le champ est obligatoire. Cliquez sur l'icône en forme d'œil pour afficher
ce que vous avez saisi.

Semaphore ne conserve qu'une empreinte (hash) du mot de passe.
