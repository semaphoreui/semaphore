---
form: src/components/TerraformAliasForm.vue
---

# Modifica Key

Un alias è un indirizzo del backend HTTP che Semaphore fornisce per questo workspace: Terraform o
OpenTofu, puntati su di esso, leggono, scrivono, bloccano e sbloccano lo stato del workspace in
Semaphore. Semaphore crea l'indirizzo quando crei l'alias; in questa finestra di dialogo scegli solo
la chiave che lo protegge. Gli alias sono una funzionalità di Semaphore PRO.

## Auth key {#tfAlias-authKeyId}

La chiave che protegge l'alias. L'elenco offre solo chiavi di tipo **Accesso con password** dal
Negozio chiavi del progetto; quando l'elenco è vuoto, creane prima una lì.

Un client dell'alias invia il nome utente e la password di questa chiave. Dopo il salvataggio, copia
l'indirizzo dall'elenco degli alias e inseriscilo nel backend `http` della tua configurazione,
insieme al nome utente e alla password della chiave:

```hcl
terraform {
  backend "http" {
    address        = "ALIAS_URL"
    lock_address   = "ALIAS_URL"
    unlock_address = "ALIAS_URL"
    username       = "LOGIN"
    password       = "PASSWORD"
  }
}
```

L'indirizzo risponde ai metodi predefiniti del backend `http`, quindi un solo indirizzo serve lo
stato, il blocco e lo sblocco. Scegliere un'altra chiave per un alias esistente ne mantiene
l'indirizzo: i client hanno allora bisogno del nome utente e della password della nuova chiave.
