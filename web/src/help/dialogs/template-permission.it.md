---
form: src/components/EditTemplatePermissionForm.vue
---

# Template permissions

Concede a un ruolo permessi aggiuntivi su questo singolo modello. Un membro che ha il ruolo li
ottiene in aggiunta ai permessi che il ruolo ha nell'intero progetto: una concessione qui aggiunge
soltanto, non toglie mai un permesso. Un membro con più ruoli ottiene ogni permesso concesso a uno
qualsiasi di essi.

## Ruolo {#templatePerm-roleSlug}

Il ruolo che riceve i permessi. L'elenco mostra i ruoli di questo progetto, i ruoli condivisi e i
ruoli integrati. Il campo è obbligatorio, e un ruolo può avere una sola concessione su un modello:
il server ne rifiuta una seconda.

Il ruolo di una concessione salvata non può essere cambiato. Il server mantiene il ruolo e salva
solo i permessi, qualunque cosa tu scelga qui. Per dare i permessi a un altro ruolo, rimuovi questa
concessione dall'elenco e aggiungine una nuova.

## Permissions {#templatePerm-permissionsTitle}

Cosa può fare con questo modello un membro che ha il ruolo:

- **Can see this template**: il modello è nell'elenco dei modelli, con i suoi compiti e i loro
  log.
- **Can run this template**: avviare compiti dal modello, e fermare, confermare o rifiutare i suoi
  compiti.
- **Can edit this template**: modificare ed eliminare il modello, fermare tutti i suoi compiti e
  gestire questi permessi.

Eseguire o modificare include vedere il modello: quando selezioni uno dei due, anche **Can see this
template** viene selezionato e non può essere deselezionato. Lascia tutte le caselle deselezionate
per salvare una concessione che non dà nulla.
