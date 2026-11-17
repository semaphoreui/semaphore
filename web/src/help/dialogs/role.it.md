---
form: src/components/EditRoleForm.vue
---

# Edit Role

Un ruolo è un insieme di permessi con un nome all'interno di un progetto. Assegni i ruoli ai membri
di un progetto, e un membro può fare ciò che uno qualsiasi dei suoi ruoli consente. Aperta dalla
pagina Roles di un progetto, la finestra di dialogo modifica solo un ruolo di quel progetto; aperta
dalla pagina Roles al di fuori di un progetto, modifica un ruolo condiviso che ogni progetto può
usare. I ruoli integrati Owner, Manager, Task Runner e Guest non possono essere modificati.

## Nome {#role-name}

Il nome con cui il ruolo compare nella pagina Roles e ovunque si scelga un ruolo, ad esempio quando
aggiungi un membro a un progetto. Il campo è obbligatorio. Puoi rinominare un ruolo in qualsiasi
momento.

## Slug {#role-slug}

L'identificatore del ruolo. Semaphore lo compila a partire dal nome mentre digiti il nome, finché
non modifichi tu stesso lo slug. Accetta da 1 a 100 lettere, cifre, trattini bassi e trattini, ad
esempio `release-managers`.

Lo slug non può essere cambiato una volta salvato il ruolo: il campo è bloccato quando modifichi un
ruolo.

- Uno slug deve essere unico, e il controllo ignora maiuscole e minuscole. Lo slug di un ruolo
  condiviso deve essere diverso da ogni altro ruolo; lo slug di un ruolo di progetto deve essere
  diverso dai ruoli condivisi e dagli altri ruoli del suo progetto.
- `owner`, `manager`, `task_runner`, `guest` e `all` sono riservati, in maiuscolo o in minuscolo.

## Permissions {#role-permissionsTitle}

Cosa possono fare nel progetto i membri che hanno il ruolo. Un nuovo ruolo non ne ha nessuno
selezionato. I permessi di tutti i ruoli di un membro si sommano; un amministratore di Semaphore può
fare tutto in ogni progetto, qualunque siano i ruoli.

- **Can read project**: vedere il progetto: i suoi modelli di compito, i compiti e i loro log, gli
  inventari, le chiavi, i repository, i gruppi di variabili, i programmi e i membri. Ogni lettura
  nel progetto lo richiede, e nessun altro permesso lo include, quindi selezionalo per quasi ogni
  ruolo. Senza di esso un membro vede solo i modelli di compito e i workflow che concedono al ruolo
  un accesso proprio, con i loro compiti.
- **Can run project tasks**: avviare compiti dai modelli di compito, fermarli, e confermare o
  rifiutare un compito che attende una conferma.
- **Can update project**: modificare le impostazioni del progetto ed eliminarlo.
- **Can manage project resources**: creare, modificare ed eliminare modelli di compito, inventari,
  chiavi, repository, gruppi di variabili, programmi, viste, integrazioni, archivi di segreti,
  esecutori del progetto e i ruoli del progetto; eliminare compiti; impostare quali ruoli possono
  usare un modello di compito.
- **Can manage project users**: aggiungere membri al progetto, cambiarne i ruoli e rimuoverli.
- **Can run workflows**: avviare un workflow e fermarne le esecuzioni.
- **Can edit workflows**: creare, modificare ed eliminare workflow, e impostare quali ruoli possono
  usare un workflow.
- **Can approve workflows**: approvare o rifiutare un passaggio di un'esecuzione di workflow che
  attende un'approvazione.
