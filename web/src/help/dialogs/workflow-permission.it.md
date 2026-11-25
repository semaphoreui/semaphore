---
form: src/components/EditTemplatePermissionForm.vue
---

# Workflow permissions

Concede a un ruolo permessi aggiuntivi su questo singolo workflow. Un membro che ha il ruolo li
ottiene in aggiunta ai permessi che il ruolo ha nell'intero progetto: una concessione qui aggiunge
soltanto, non toglie mai un permesso. Un membro con più ruoli ottiene ogni permesso concesso a uno
qualsiasi di essi.

## Ruolo {#templatePerm-roleSlug}

Il ruolo che riceve i permessi. L'elenco mostra i ruoli di questo progetto, i ruoli condivisi e i
ruoli integrati. Il campo è obbligatorio, e un ruolo può avere una sola concessione su un workflow:
il server ne rifiuta una seconda.

Il ruolo di una concessione salvata non può essere cambiato. Il server mantiene il ruolo e salva
solo i permessi, qualunque cosa tu scelga qui. Per dare i permessi a un altro ruolo, rimuovi questa
concessione dall'elenco e aggiungine una nuova.

## Permissions {#templatePerm-permissionsTitle}

Cosa può fare con questo workflow un membro che ha il ruolo:

- **Can see this workflow**: il workflow è nell'elenco dei workflow, con le sue esecuzioni.
- **Can run this workflow**: avviare esecuzioni del workflow.
- **Can edit this workflow**: modificare ed eliminare il workflow e gestire questi permessi.
- **Can approve this workflow**: approvare o rifiutare i passaggi di approvazione delle sue
  esecuzioni.

Eseguire, modificare o approvare include vedere il workflow: quando selezioni uno di questi, anche
**Can see this workflow** viene selezionato e non può essere deselezionato. Lascia tutte le caselle
deselezionate per salvare una concessione che non dà nulla.
