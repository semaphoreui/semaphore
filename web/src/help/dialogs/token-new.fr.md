---
form: src/views/Tokens.vue
---

# New Token

Un jeton API permet à un script ou à un autre service d'appeler l'API de Semaphore en votre nom,
avec les mêmes droits que votre compte. Envoyez-le dans l'en-tête `Authorization` d'une requête :
`Bearer` suivi du jeton. Le jeton n'est affiché qu'une fois, juste après sa création : copiez-le à
ce moment-là.

## Token Name {#token-name}

Optionnel. Un nom qui distingue le jeton dans la liste des jetons, par exemple le nom du script qui
l'utilise. Il ne fait pas partie du jeton et ne change pas ce que le jeton peut faire.

## Expires {#token-expires}

Le moment où le jeton cesse de fonctionner. Choisissez une durée comptée à partir de maintenant,
**Never** pour un jeton qui fonctionne jusqu'à sa suppression, ou **Custom** pour choisir vous-même
le moment.

Un jeton qui n'expire jamais fonctionne jusqu'à ce que vous le supprimiez de la liste. Préférez une
durée pour un jeton confié à un script ou à un autre service.

## Expiration date {#token-expiresAt}

Affiché lorsque **Expires** vaut **Custom**.

La date et l'heure auxquelles le jeton cesse de fonctionner, dans le fuseau horaire de votre
navigateur. Le moment doit être dans le futur : le serveur refuse un jeton qui serait déjà expiré.
