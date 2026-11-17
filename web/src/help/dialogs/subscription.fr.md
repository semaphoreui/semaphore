---
form: src/components/SubscriptionForm.vue
---

# Subscription & Billing

La clé d'abonnement déverrouille les fonctionnalités de Semaphore PRO et Enterprise sur ce
serveur. La boîte de dialogue active une clé et affiche l'offre de la clé active : son titulaire,
son statut, sa date d'expiration et le nombre d'utilisateurs Pro, d'exécuteurs et d'autres éléments
qu'elle autorise et qui sont utilisés. Elle est ouverte par un administrateur depuis le menu
utilisateur de l'édition Pro. La boîte de dialogue s'intitule **Upgrade to Semaphore PRO** tant
qu'aucun abonnement n'est actif.

## Enter your PRO or EE key {#subscription-key}

La clé d'abonnement, telle que vous l'avez reçue avec votre abonnement PRO ou Enterprise.
Collez-la ici, ou chargez-la depuis un fichier avec le bouton de téléversement, puis appuyez sur
**Activate New Key**. Le champ est obligatoire pour l'activation.

Le champ est en lecture seule lorsque la clé est définie dans la configuration du serveur
(`subscription.key` ou `subscription.key_file`, ou les variables `SEMAPHORE_SUBSCRIPTION_KEY`
et `SEMAPHORE_SUBSCRIPTION_KEY_FILE`) : la clé ne peut alors être modifiée qu'à cet endroit.

## Téléverser un fichier de clé {#subscription-upload}

Affiché lorsqu'aucun abonnement n'a encore été activé.

Lit la clé depuis un fichier de votre ordinateur (`.txt`, `.key`, `.pem` ou `.lic`) dans le champ
ci-dessus. La clé n'est pas activée tant que vous n'appuyez pas sur **Activate New Key**.

## Actions sur la clé {#subscription-actionsMenu}

Affiché lorsqu'un abonnement est actif ou a expiré.

Un menu d'actions sur la clé du serveur :

- **Reload** : récupère à nouveau l'état de l'abonnement, par exemple après son renouvellement.
- **Upload** : lit une nouvelle clé depuis un fichier dans le champ ci-dessus, comme le fait le
  bouton de téléversement.
- **Reset** : supprime immédiatement la clé du serveur, sans confirmation.

## Activate New Key {#subscription-activate}

Active sur ce serveur la clé saisie dans le champ. Lorsque la clé est valide, l'offre de
l'abonnement s'affiche en dessous. Lorsque votre compte n'est pas encore un utilisateur Pro et que
l'offre dispose d'une place d'utilisateur Pro libre, la boîte de dialogue demande s'il faut faire de
votre compte un utilisateur Pro.

Une clé déjà activée est signalée par un avertissement. Le bouton est désactivé lorsque la clé est
définie dans la configuration du serveur.

## Buy Pro {#subscription-buy}

Ouvre la page du portail Semaphore où l'on achète un abonnement PRO, dans un nouvel onglet du
navigateur.

## Get 30-day free trial {#subscription-trial}

Affiché tant qu'aucun abonnement n'est actif.

Ouvre la page du portail Semaphore qui délivre une clé pour un essai gratuit de 30 jours, dans un
nouvel onglet du navigateur. Collez la clé obtenue dans le champ ci-dessus.
