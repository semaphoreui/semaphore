---
form: src/components/KeyForm.vue
---

# Modifica Key

Una chiave del Negozio chiavi contiene una credenziale che il progetto usa senza mostrarla: la chiave
SSH o la password con cui Ansible accede agli host, le credenziali di un repository Git, la password
di un vault Ansible, il segreto di un'integrazione. Inventari, repository, modelli di compito e
integrazioni fanno riferimento a una chiave tramite il suo nome. Una volta salvato, il segreto non
viene mai rimandato al browser.

## Nome chiave {#key-name}

Il nome con cui la chiave compare nel Negozio chiavi e nei selettori di chiavi di inventari,
repository, modelli di compito e integrazioni. Il campo è obbligatorio. Scegli un nome che dica cosa
apre la chiave, ad esempio `Deploy key of the web servers`.

## Source {#key-source}

Dove viene conservato il segreto della chiave. Le schede:

- **Local**: il segreto viene digitato in questa finestra di dialogo e memorizzato nel database di
  Semaphore, cifrato quando l'amministratore ha impostato una chiave di cifratura per le chiavi di
  accesso. È il valore predefinito.
- **Storage**: il segreto è conservato in un archivio di segreti esterno del progetto, scelto in
  **Storage** qui sotto. La scheda è presente solo nell'edizione Pro. Gli archivi stessi si
  configurano nella scheda **Storages** del Negozio chiavi.
- **Env**: il segreto viene letto da una variabile di ambiente del processo del server Semaphore.
  Digita il nome della variabile nel campo sotto le schede.
- **File**: il segreto viene letto da un file sul server Semaphore. Digita il percorso assoluto del
  file nel campo sotto le schede. Il file deve trovarsi all'interno della directory dei segreti del
  server (`dirs.secrets` nella sua configurazione, `/tmp/semaphore` per impostazione predefinita);
  un percorso al di fuori di essa, o uno che contiene `..`, fallisce quando la chiave viene usata.
  Un'interruzione di riga finale del file viene scartata.

Con **Env** e **File**, Semaphore non scrive mai il segreto: legge il valore ogni volta che un
compito usa la chiave, quindi modifichi il segreto là dove si trova. Il valore deve essere adatto al
**Tipo** della chiave. Una chiave **Stringa** prende il valore così com'è. Una chiave **Chiave SSH**
o **Accesso con password** si aspetta un JSON con i campi di quel tipo, ad esempio:

```json
{"login": "deploy", "password": "s3cret"}
```

```json
{"login": "deploy", "passphrase": "", "private_key": "-----BEGIN OPENSSH PRIVATE KEY-----\n..."}
```

Quando modifichi una chiave, le schede sono bloccate finché **Sovrascrivi** non è selezionato. Una
chiave sincronizzata da un archivio di segreti mantiene la sua origine, e il suo nome e il suo tipo
non possono essere cambiati.

## Storage {#key-sourceStorageId}

Visibile quando è scelta la scheda **Storage**.

L'archivio di segreti del progetto che contiene il segreto: HashiCorp Vault, OpenBao, Devolutions
Server, AWS Secrets Manager o Azure Key Vault, come configurato nella scheda **Storages** del
Negozio chiavi.

- Un archivio che non è in sola lettura riceve il segreto che digiti in questa finestra di dialogo:
  Semaphore ve lo scrive quando salvi. Eliminare la chiave elimina anche il segreto dall'archivio.
  Una chiave conservata in un tale archivio non può essere spostata in seguito in un altro archivio.
- Un archivio in sola lettura viene soltanto letto. La finestra di dialogo non mostra allora alcun
  campo per il segreto, e Semaphore lo legge dall'archivio ogni volta che un compito usa la chiave.

## Source Key {#key-sourceStorageKey}

Visibile quando è selezionato uno **Storage**.

Dove si trova il segreto in quell'archivio, ad esempio il percorso del segreto in HashiCorp Vault.

## Tipo {#key-type}

Che tipo di credenziale contiene la chiave. Il campo è obbligatorio, e i campi sotto di esso
cambiano in base al tipo.

- **Chiave SSH**: una chiave privata, con un nome utente e una passphrase opzionali. Ansible si
  connette agli host con essa, e Git clona con essa un repository con un URL `ssh`.
- **Accesso con password**: un nome utente e una password. Ansible si connette agli host con essi o
  diventa un altro utente con essi (**Credenziali Sudo** di un inventario); Git clona con essi un
  repository `https`; la password apre un vault Ansible di un modello di compito.
- **Stringa**: un singolo valore segreto senza nome utente: un token, una chiave API, una password.
  Usala per il segreto di un'integrazione. Non può essere una chiave SSH, credenziali sudo, la
  password di un vault o la chiave di accesso di un repository.
- **Nessuno**: nessuna credenziale. Usala per un repository che non richiede accesso (un repository
  `https` pubblico, un URL `file` o un percorso locale) e per un inventario i cui host non hanno
  bisogno di credenziali o definiscono da soli la connessione.

## Nome utente {#key-loginPasswordLogin}

Visibile quando il tipo è **Accesso con password**. Opzionale.

Il nome utente della credenziale.

- Ansible si connette agli host come questo utente (`--user`); come **Credenziali Sudo** di un
  inventario, è l'utente che Ansible diventa (`--become-user`). Quando il campo è vuoto, Ansible
  usa l'utente indicato dall'inventario o dalla propria configurazione.
- Per un repository `https`, Git accede con questo utente e la password. Lascialo vuoto per un token
  che il server Git accetta da solo: la password viene allora inviata come nome utente.

## Password {#key-loginPasswordPassword}

Visibile quando il tipo è **Accesso con password**. Obbligatorio.

La password della credenziale. Semaphore la usa per rispondere alle richieste di password di
Ansible: la password SSH (`--ask-pass`) per le credenziali utente, la password di `become`
(`--ask-become-pass`) per le credenziali sudo, e la password del vault per un vault Ansible. Per un
repository `https` è la password o il token di accesso del server Git.

## Valore {#key-stringValue}

Visibile quando il tipo è **Stringa**. Obbligatorio.

Il valore segreto stesso, ad esempio il token con cui un'integrazione verifica le richieste che
riceve. Viene memorizzato così come è digitato.

## Nome utente {#key-sshLogin}

Visibile quando il tipo è **Chiave SSH**. Opzionale.

L'utente con cui Ansible si connette agli host (`--user`). Quando il campo è vuoto, Ansible usa
l'utente indicato dall'inventario o dalla propria configurazione.

Per un repository Git, inserisci invece l'utente nell'URL del repository, come in
`git@github.com:org/repo.git`.

## Passphrase {#key-sshPassphrase}

Visibile quando il tipo è **Chiave SSH**. Opzionale.

La passphrase con cui è cifrata la chiave privata. Lasciala vuota per una chiave senza passphrase.
Una passphrase errata fa fallire ogni compito che usa la chiave al caricamento della chiave. Una
chiave generata da Semaphore non ha passphrase: ciò che contiene questo campo viene allora
scartato.

## Generate SSH Key {#key-generateSshKey}

Visibile quando il tipo è **Chiave SSH**.

Semaphore crea una nuova coppia di chiavi quando salvi: una chiave privata RSA a 2048 bit, che resta
nella chiave, e la sua chiave pubblica, che viene mostrata una volta in una finestra di dialogo dopo
il salvataggio e poi sotto **Public key** ogni volta che modifichi la chiave. Aggiungi la chiave
pubblica a `~/.ssh/authorized_keys` sugli host, o come deploy key del repository Git.

Il campo della chiave privata è bloccato mentre la casella è selezionata. Su una chiave esistente,
seleziona prima **Sovrascrivi**: la vecchia coppia di chiavi viene sostituita. Una chiave il cui
segreto viene letto da **Env**, **File** o da un archivio in sola lettura non può essere generata,
perché Semaphore non ha dove scriverla.

## Chiave privata {#key-sshPrivateKey}

Visibile quando il tipo è **Chiave SSH**. Obbligatorio a meno che **Generate SSH Key** non sia
selezionato.

La chiave privata, in formato OpenSSH o PEM, comprese le sue righe `BEGIN` e `END`. Per un compito,
Semaphore la carica in un proprio agente SSH, che Ansible e Git usano per la durata del compito.

## Public key {#key-publicKey}

Visibile quando modifichi una chiave SSH generata da Semaphore, finché **Sovrascrivi** non è
selezionato.

La metà pubblica della coppia di chiavi generata. **Show** la espande, il pulsante di copia la mette
negli appunti. Aggiungila a `~/.ssh/authorized_keys` sugli host, o come deploy key del repository
Git.

È la chiave pubblica della coppia generata per ultima da Semaphore. Dopo che hai sostituito a mano
la chiave privata con **Sovrascrivi**, la chiave mostrata qui non le corrisponde più.

## Sovrascrivi {#key-overrideSecret}

Visibile quando modifichi una chiave esistente.

Semaphore non mostra mai più un segreto salvato, quindi i campi del segreto sono vuoti e bloccati.
Seleziona la casella per sostituire il segreto: l'origine, il tipo e i campi del segreto si
sbloccano, e tutto ciò che contengono viene salvato di nuovo, quindi inserisci di nuovo l'intero
segreto, non solo la parte che cambia. Senza di essa, il salvataggio modifica solo il nome della
chiave.

Salvare una chiave rimuove anche le copie che il server Semaphore conserva dei repository che la
usano: il compito successivo li clona di nuovo.
