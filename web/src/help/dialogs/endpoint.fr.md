---
form: src/components/EndpointForm.vue, src/components/IntegrationAuthForm.vue
---

# Modifier le point de terminaison

Un point de terminaison est un point d’entrée pour les webhooks : les URL que d’autres services
appellent et la manière dont leurs requêtes sont vérifiées. Les intégrations se rattachent à un
point de terminaison et lancent leurs exécutions sur les requêtes qu’il accepte. Les URL d’un point
de terminaison, ses alias, s’ajoutent dans l’onglet « Points de terminaison ».

## Nom {#endpointForm-name}

Le nom sous lequel le point de terminaison figure dans l’onglet « Points de terminaison » et dans le
formulaire d’une intégration. Le champ est obligatoire.

## Description {#endpointForm-description}

Facultatif. Une note sur l’usage du point de terminaison, affichée à côté de son nom.

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
