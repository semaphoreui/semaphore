---
form: src/components/EndpointForm.vue, src/components/IntegrationAuthForm.vue
---

# Modifica endpoint

Un endpoint è un punto di ingresso per i webhook: gli URL chiamati dagli altri servizi e il modo in
cui le loro richieste vengono controllate. Le integrazioni si collegano a un endpoint e avviano le
proprie esecuzioni con le richieste che esso accetta. Gli URL di un endpoint, i suoi alias, si
aggiungono nella scheda «Endpoint».

## Nome {#endpointForm-name}

Il nome con cui l’endpoint compare nella scheda «Endpoint» e nel modulo di un’integrazione. Il campo
è obbligatorio.

## Descrizione {#endpointForm-description}

Facoltativo. Una nota su a cosa serve l’endpoint, mostrata accanto al suo nome.

## Metodo di autenticazione {#integrationAuthForm-authMethod}

Come Semaphore verifica che una richiesta provenga dal servizio che ti aspetti. Una richiesta che
non supera il controllo non avvia nulla, e il webhook le risponde come a qualsiasi altra richiesta.

- **Nessuno**: nessun controllo. Chiunque conosca l'URL può avviare il compito.
- **GitHub Webhooks**: l'intestazione `X-Hub-Signature-256` deve contenere la firma HMAC-SHA256 del
  corpo creata con il segreto, così come la invia GitHub.
- **Bitbucket Webhooks**: lo stesso controllo sull'intestazione `X-Hub-Signature`, così come la
  invia Bitbucket.
- **Token**: l'intestazione indicata in **Header di autenticazione** deve contenere il segreto stesso.
- **HMAC** e **HMAC (SHA-512)**: l'intestazione indicata in **Header di autenticazione** deve contenere l'HMAC-SHA256
  o l'HMAC-SHA512 del corpo creato con il segreto, in esadecimale minuscolo, senza un prefisso
  come `sha256=`.
- **BasicAuth**: la richiesta deve portare il nome utente e la password della chiave indicata in
  **Segreto di autenticazione** nella sua intestazione `Authorization`.

## Header di autenticazione {#integrationAuthForm-authHeader}

Visibile quando **Metodo di autenticazione** è **Token**, **HMAC** o **HMAC (SHA-512)**.

Il nome dell'intestazione della richiesta che porta il token o la firma, ad esempio
`X-Webhook-Token`. Maiuscole e minuscole nel nome non contano. Una richiesta senza quell'intestazione
viene rifiutata.

## Segreto di autenticazione {#integrationAuthForm-authSecretId}

Visibile quando **Metodo di autenticazione** non è **Nessuno**.

La chiave del Negozio chiavi con cui viene verificata la richiesta. Per **BasicAuth** va bene solo
una chiave di tipo **Accesso con password**. Per gli altri metodi va bene anche una chiave di tipo
**Stringa**: il controllo usa il suo valore, o la password di una chiave **Accesso con password**.

Il campo è necessario per ogni metodo tranne **Nessuno**: senza una chiave, o con un segreto vuoto, ogni
richiesta viene rifiutata.
