---
form: src/views/Tokens.vue
---

# New Token

Un token API consente a uno script o a un altro servizio di chiamare l'API di Semaphore a tuo nome,
con gli stessi diritti del tuo account. Invialo nell'intestazione `Authorization` di una richiesta:
`Bearer` e il token. Il token viene mostrato una sola volta, subito dopo la sua creazione: copialo
in quel momento.

## Token Name {#token-name}

Opzionale. Un nome che distingue il token nell'elenco dei token, ad esempio il nome dello script che
lo usa. Non fa parte del token e non cambia ciò che il token può fare.

## Expires {#token-expires}

Quando il token smette di funzionare. Scegli un periodo calcolato da adesso, **Never** per un token
che funziona finché non viene eliminato, oppure **Custom** per scegliere tu stesso il momento.

Un token che non scade mai funziona finché non lo elimini dall'elenco. Preferisci un periodo per un
token dato a uno script o a un altro servizio.

## Expiration date {#token-expiresAt}

Visibile quando **Expires** è **Custom**.

La data e l'ora in cui il token smette di funzionare, nel fuso orario del tuo browser. Il momento
deve essere nel futuro: il server rifiuta un token che sarebbe già scaduto.
