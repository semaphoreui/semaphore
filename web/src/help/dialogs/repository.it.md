---
form: src/components/RepositoryForm.vue
---

# Modifica Repository

Un repository è il luogo da cui provengono i playbook, il codice Terraform o gli script di un
modello di compito: un repository Git, o una directory sulla macchina che esegue i compiti. Prima di
ogni compito Semaphore clona il repository Git, o aggiorna la copia clonata in precedenza, esegue il
checkout del ramo ed esegue il compito in quella copia.

## Nome {#repository-name}

Il nome con cui il repository compare nella pagina Repository e nel selettore di repository di un
modello di compito e di un inventario. Il campo è obbligatorio.

## URL o percorso {#repository-gitUrl}

Dove si trova il codice. Il campo è obbligatorio. I chip sotto il campo mostrano il tipo di
indirizzo che hai digitato, e un clic su un chip riscrive l'indirizzo in quel tipo.

- **ssh**: `git@github.com:org/repo.git` o `ssh://git@github.com/org/repo.git`. Git accede con
  l'utente indicato nell'indirizzo e con la **Chiave SSH** scelta in **Chiave di accesso**.
- **https** e **http**: `https://github.com/org/repo.git`. Un repository privato ha bisogno di una
  chiave **Accesso con password**; uno pubblico ha bisogno di **Nessuno**. Evita `http`: il nome
  utente e la password viaggiano allora non cifrati.
- **git**: `git://host/org/repo.git`, il protocollo Git in sola lettura, che non ha accesso con
  credenziali.
- **file**: `file:///srv/git/repo.git`, un repository Git sul disco della macchina che esegue il
  compito, clonato come uno remoto.
- **percorso ass.**: un percorso assoluto come `/srv/playbooks` o `D:\playbooks`. Semaphore non lo
  clona: il compito viene eseguito direttamente in quella directory, che deve esistere sulla
  macchina che esegue il compito (il server Semaphore o un esecutore). Il ramo non si applica.

Cambiare l'indirizzo rimuove le copie che il server Semaphore aveva clonato da quello vecchio.

## Ramo {#repository-gitBranch}

Il ramo che Semaphore clona e mantiene aggiornato, ad esempio `main`. Il campo è obbligatorio,
tranne che per un percorso assoluto, per il quale è bloccato.

Un modello di compito può indicare un proprio ramo, che prevale su questo, e un modello può
consentire ai suoi compiti di scegliere un ramo al loro avvio. Funziona anche un tag, ma un tag non
può essere aggiornato con pull, quindi Semaphore clona di nuovo il repository prima di ogni compito.

## Single checkout for all templates {#repository-monorepo}

Disattivato per impostazione predefinita: ogni modello di compito riceve una propria copia del
repository. Seleziona la casella per clonare il repository una sola volta e condividere quella copia
tra tutti i modelli che lo usano, per un repository troppo grande per tenerne una copia per modello.

La copia condivisa resta sul ramo di questo repository: i modelli che la usano non possono avere un
proprio ramo, e i loro compiti non possono scegliere un altro ramo o commit. Bloccato per un
percorso assoluto.

## Chiave di accesso {#repository-sshKeyId}

La chiave del Negozio chiavi con cui Git accede al repository. Il campo è obbligatorio. Offre chiavi
di tre tipi; quella da scegliere dipende dall'URL:

- **Chiave SSH** per un URL `ssh`. Semaphore carica la chiave privata in un agente SSH per la durata
  della clonazione o dell'aggiornamento. L'utente è quello nell'URL, come `git` in
  `git@github.com:org/repo.git`.
- **Accesso con password** per un URL `https`: il nome utente e la password, o il token di accesso
  del server Git. Con un nome utente vuoto la password viene inviata da sola, come nome utente, che
  è ciò che alcuni server si aspettano da un token. Le stesse credenziali vengono date ad
  `ansible-galaxy` per i ruoli e le collezioni sullo stesso server.
- **Nessuno** per un repository pubblico, un URL `git` o `file`, o un percorso assoluto.

Una chiave che non corrisponde all'URL non dà a Git alcuna credenziale, e la clonazione di un
repository privato allora fallisce.
