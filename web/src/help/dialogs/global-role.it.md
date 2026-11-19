---
form: src/components/EditGlobalRoleForm.vue
---

# Edit Global Role

Un ruolo globale concede permessi sull'intera istanza di Semaphore agli utenti a cui è assegnato; lo
assegni nella finestra di dialogo di un utente. Non dà alcun accesso all'interno dei progetti: questo
deriva solo dai ruoli che un utente ha come membro di un progetto. Un amministratore ha ogni
permesso globale, qualunque siano i ruoli assegnati. I ruoli globali sono una funzionalità Pro.

## Nome {#globalRole-name}

Il nome con cui il ruolo compare nella scheda Ruoli globali della pagina Utenti e nella finestra di
dialogo di un utente. Il campo è obbligatorio. Puoi rinominare un ruolo in qualsiasi momento.

## Slug {#globalRole-slug}

L'identificatore del ruolo. Semaphore lo compila a partire dal nome mentre digiti il nome, finché
non modifichi tu stesso lo slug. Accetta da 1 a 100 lettere, cifre, trattini bassi e trattini, ad
esempio `user-admins`.

Lo slug deve essere unico tra i ruoli globali; un ruolo di progetto può avere lo stesso slug. Non
può essere cambiato una volta salvato il ruolo: il campo è bloccato quando modifichi un ruolo.

## Global permissions {#globalRole-permissionsTitle}

Cosa può fare in tutta l'istanza un utente che ha il ruolo. Un nuovo ruolo non ne ha nessuno
selezionato. I permessi di tutti i ruoli globali di un utente si sommano.

- **Can manage users**: creare utenti, modificarli ed eliminarli, e vederne i dettagli nell'elenco
  degli utenti. Un tale utente non può rendere un utente amministratore o utente Pro, e non può
  modificare o eliminare un amministratore o un utente che ha un permesso globale che lui stesso non
  ha.
- **Can manage global roles**: creare, modificare ed eliminare i ruoli globali e i ruoli di progetto
  condivisi, e decidere quali utenti hanno quale ruolo globale. Concedilo solo a persone a cui
  affideresti l'intera istanza.
- **Can manage system settings**: modificare le impostazioni dell'istanza, le applicazioni, gli
  esecutori globali e le mappature dei gruppi; vedere il cluster e le informazioni sul server;
  vedere e rimuovere i compiti di tutti i progetti; pulire la cache del server.
- **Can read the audit log**: leggere il registro di audit dell'istanza ed esportarlo.
