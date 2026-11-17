---
form: src/components/TeamMemberForm.vue
---

# Nuovo Membro del team

Aggiunge un utente al progetto con un ruolo. Il ruolo decide cosa il membro può vedere e fare nel
progetto; nient'altro del suo account cambia. Quando l'amministratore ha attivato gli inviti,
l'utente riceve invece un invito e diventa membro una volta accettato.

## Email or Username {#teamMember-inviteType}

Visibile quando gli utenti possono essere invitati sia per e-mail sia per nome utente.

Come scegli la persona: **Email** per invitare un indirizzo e-mail, **Username** per scegliere un
utente che ha già un account su questo server Semaphore.

## Email {#teamMember-email}

Visibile quando la persona viene invitata per e-mail.

L'indirizzo e-mail della persona da invitare. Diventa membro del progetto una volta accettato
l'invito.

## Utente {#teamMember-userId}

Visibile quando la persona viene scelta per nome utente.

L'utente da aggiungere, elencato con il nome utente e il nome completo. Il campo è obbligatorio.
L'elenco contiene ogni utente del server Semaphore che non è ancora membro del progetto: una persona
senza account deve prima ottenerne uno da un amministratore.

## Ruolo {#teamMember-role}

Cosa può fare il membro nel progetto. Il campo è obbligatorio. I ruoli integrati:

- **Owner**: tutto, comprese le impostazioni del progetto, il suo team e la sua eliminazione.
- **Manager**: esegue compiti e workflow, approva workflow, e crea, modifica ed elimina modelli di
  compito, inventari, chiavi, repository, gruppi di variabili, programmi, integrazioni e workflow.
  Un manager non può modificare le impostazioni del progetto o il suo team.
- **Task Runner**: vede il progetto, esegue compiti e workflow e approva workflow. Un task runner
  non può modificare nulla nel progetto.
- **Guest**: vede il progetto, i suoi compiti e i loro log, e non modifica nulla.

L'elenco contiene anche i ruoli creati per questo progetto (**Ruoli del progetto**) e i ruoli
condivisi da tutti i progetti (**Shared roles**). Un tale ruolo concede i permessi impostati per
esso nella pagina Roles. Un invito può portare solo un ruolo integrato.

Un membro che può gestire il team può assegnare qualsiasi ruolo, **Owner** compreso. La finestra di
dialogo assegna un ruolo; aggiungine altri in seguito nella colonna **Ruolo** dell'elenco del team.
