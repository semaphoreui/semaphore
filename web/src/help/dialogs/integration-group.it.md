---
form: src/components/IntegrationGroupForm.vue, src/components/IntegrationAuthForm.vue
---

# Modifica integrazione

Un’integrazione è un insieme di esecuzioni che partono da un endpoint. Quando una richiesta a un URL
di quell’endpoint supera l’autorizzazione, parte ogni esecuzione dell’integrazione i cui matcher
accettano la richiesta.

## Nome {#integrationGroupForm-name}

Il nome con cui l’integrazione compare nella pagina «Integrazioni». Il campo è obbligatorio.

## Endpoint {#integrationGroupForm-endpointId}

L’endpoint i cui URL avviano le esecuzioni di questa integrazione. Più integrazioni possono usare lo
stesso endpoint. Il campo è obbligatorio.

## Sovrascrivi autorizzazione {#integrationGroupForm-authOverride}

Lascia l’opzione disattivata per controllare le richieste come fa l’endpoint. Attivala per
controllare le richieste di questa integrazione con un metodo e un segreto propri, impostati nei
campi che compaiono sotto; le impostazioni dell’endpoint non si applicano più a questa integrazione.

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
