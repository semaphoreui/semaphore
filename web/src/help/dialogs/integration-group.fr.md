---
form: src/components/IntegrationGroupForm.vue, src/components/IntegrationAuthForm.vue
---

# Modifier l’intégration

Une intégration est un ensemble d’exécutions lancées depuis un point de terminaison. Lorsqu’une
requête vers une URL de ce point de terminaison passe l’autorisation, chaque exécution de
l’intégration dont les matchers acceptent la requête se lance.

## Nom {#integrationGroupForm-name}

Le nom sous lequel l’intégration figure sur la page « Intégrations ». Le champ est obligatoire.

## Point de terminaison {#integrationGroupForm-endpointId}

Le point de terminaison dont les URL lancent les exécutions de cette intégration. Plusieurs
intégrations peuvent utiliser le même point de terminaison. Le champ est obligatoire.

## Remplacer l’autorisation {#integrationGroupForm-authOverride}

Laissez l’option désactivée pour vérifier les requêtes comme le fait le point de terminaison.
Activez-la pour vérifier les requêtes de cette intégration avec une méthode et un secret qui lui
sont propres, définis dans les champs qui apparaissent en dessous ; les paramètres du point de
terminaison ne s’appliquent alors plus à cette intégration.

## Méthode d’authentification {#integrationAuthForm-authMethod}

La manière dont Semaphore vérifie qu'une requête provient du service attendu. Une requête qui
échoue à la vérification ne lance rien, et le webhook lui répond comme à n'importe quelle autre
requête.

- **Aucun** : aucune vérification. Quiconque connaît l'URL peut lancer la tâche.
- **GitHub Webhooks** : l'en-tête `X-Hub-Signature-256` doit contenir la signature HMAC-SHA256 du
  corps calculée avec le secret, telle que GitHub l'envoie.
- **Bitbucket Webhooks** : la même vérification sur l'en-tête `X-Hub-Signature`, tel que Bitbucket
  l'envoie.
- **Token** : l'en-tête indiqué dans **En-tête d’authentification** doit contenir le secret lui-même.
- **HMAC** et **HMAC (SHA-512)** : l'en-tête indiqué dans **En-tête d’authentification** doit contenir le
  HMAC-SHA256 ou le HMAC-SHA512 du corps calculé avec le secret, en hexadécimal minuscule, sans
  préfixe tel que `sha256=`.
- **BasicAuth** : la requête doit porter l'identifiant et le mot de passe de la clé choisie dans
  **Secret d'authentification** dans son en-tête `Authorization`.

## En-tête d’authentification {#integrationAuthForm-authHeader}

Affiché lorsque **Méthode d’authentification** vaut **Token**, **HMAC** ou **HMAC (SHA-512)**.

Le nom de l'en-tête de la requête qui porte le jeton ou la signature, par exemple
`X-Webhook-Token`. La casse du nom n'a pas d'importance. Une requête sans cet en-tête est refusée.

## Secret d'authentification {#integrationAuthForm-authSecretId}

Affiché lorsque **Méthode d’authentification** ne vaut pas **Aucun**.

La clé du Magasin de clés à laquelle la requête est comparée. Pour **BasicAuth**, seule une clé de
type **Connexion avec mot de passe** convient. Pour les autres méthodes, une clé de type **Chaîne**
convient aussi : la vérification utilise sa valeur, ou le mot de passe d'une clé **Connexion avec
mot de passe**.

Le champ est nécessaire pour toutes les méthodes sauf **Aucun** : sans clé, ou avec un secret vide,
toute requête est refusée.
