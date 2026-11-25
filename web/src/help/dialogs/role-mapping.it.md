---
form: src/components/RoleMappingForm.vue
---

# New mapping

Una mappatura assegna un ruolo in un progetto a ogni utente che accede tramite un provider LDAP o
OpenID Connect e appartiene lì a un gruppo. Semaphore legge i gruppi dell'utente a ogni accesso
tramite il provider e assegna o toglie i ruoli mappati di conseguenza, quindi un utente che lascia
il gruppo perde il ruolo al suo accesso successivo. Un ruolo mappato rende l'utente membro del
progetto. Una mappatura non può essere modificata: eliminala e aggiungine una nuova.

## Provider {#roleMapping-provider}

La directory LDAP o il provider OpenID Connect da cui proviene il gruppo. L'elenco contiene i
provider configurati nella configurazione del server. Il campo è obbligatorio.

I gruppi vengono letti solo quando il provider è configurato per farlo: un provider LDAP ha bisogno
di una base di ricerca dei gruppi (`group_search_dn`), un provider OpenID Connect indica il claim
con i gruppi (`groups_claim`). Una mappatura di un provider rimosso in seguito dalla configurazione
resta nell'elenco, contrassegnata con **Provider not configured**.

## Group {#roleMapping-group}

Il nome del gruppo così come lo fornisce il provider. Il campo è obbligatorio.

- LDAP: il valore dell'attributo del nome del gruppo nella voce del gruppo, `cn` per impostazione
  predefinita, ad esempio `devops`. Un DN completo come `cn=devops,ou=groups,dc=example,dc=com`
  viene ridotto al suo `cn`.
- OpenID Connect: uno dei valori del claim dei gruppi, così come lo invia il provider.

## Project {#roleMapping-projectId}

Il progetto in cui viene assegnato il ruolo. Il campo è obbligatorio. Un amministratore vede ogni
progetto; qualsiasi altro utente vede solo i progetti di cui è membro.

## Ruolo {#roleMapping-role}

Il ruolo che i membri del gruppo ottengono nel progetto. Il campo è obbligatorio ed è disponibile una
volta selezionato un progetto. L'elenco contiene i ruoli di quel progetto, i ruoli condivisi e i
ruoli integrati. Quando scegli un altro progetto, un ruolo che lì non esiste viene cancellato.

Un ruolo **Owner** assegnato da una mappatura non conta quando Semaphore ti impedisce di togliere il
ruolo all'ultimo proprietario di un progetto, perché può sparire all'accesso successivo. Mantieni
almeno un proprietario aggiunto a mano.
