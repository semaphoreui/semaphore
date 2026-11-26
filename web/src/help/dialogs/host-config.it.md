---
form: src/components/HostConfigForm.vue
---

# Edit mapping

Una mappatura indica a Semaphore quale credenziale del Negozio chiavi usare per un host SSH o un URL
Git. Un repository ha una sola chiave, ma un compito può raggiungere anche altri server:
sottomoduli, ruoli e collezioni di `requirements.yml`, moduli Terraform, repository di inventario.
Le mappature del progetto si applicano a ogni connessione Git e SSH dei suoi compiti, e ai comandi
Git che Semaphore esegue per il progetto, come il controllo di un repository alla ricerca di nuovi
commit. Nessun file del tuo repository cambia.

## Tipo {#hostConfigForm-type}

Cosa corrisponde alla mappatura:

- **Host**: un nome host SSH. Ogni connessione SSH di un compito a quell'host usa la chiave della
  mappatura: Git tramite SSH, e Ansible che si connette a un host dell'inventario con quel nome. La
  mappatura prevale sulla configurazione SSH del server per lo stesso host.
- **URL**: un URL Git o il suo inizio. Git si connette a ogni repository sotto quell'URL con la
  credenziale della mappatura.

Passare a **Host** cancella una **Credential** che una mappatura di host non può usare.

## Host or URL {#hostConfigForm-host}

Il campo è obbligatorio. Può esistere una sola mappatura di un tipo per ogni host o URL.

Per **Host**: il nome host esattamente come Git o Ansible vi si connettono, ad esempio `github.com`
o `10.0.0.21`. Solo lettere, cifre, punti, trattini e trattini bassi: nessun utente, nessuna porta,
nessun carattere jolly.

Per **URL**: un indirizzo che inizia con `https://` o `http://`, senza credenziali di accesso e
senza spazi, virgolette o `=`. Indica un intero repository, o un prefisso che termina con `/` per
coprire un gruppo di repository:

```
https://github.com/acme/
https://github.com/acme/infra.git
```

Quando corrispondono più mappature, vince la più lunga: la mappatura del repository sopra
sovrascrive la mappatura di `acme` per quel solo repository.

Con una chiave SSH, una mappatura di URL fa connettere Git tramite SSH all'host dell'URL sulla porta
`22`, qualunque porta indichi l'URL.

## Credential {#hostConfigForm-sshKeyId}

La chiave del Negozio chiavi usata per le connessioni che corrispondono alla mappatura. Il campo è
obbligatorio.

- **Host** accetta solo una **Chiave SSH**. Quando la chiave ha un nome utente, SSH si connette come
  quell'utente.
- **URL** accetta una **Chiave SSH** o un **Accesso con password**. Una chiave SSH si connette con
  il suo nome utente, o come `git` quando non ne ha uno. Un accesso con password funziona solo con
  un URL `https://`; per un token, inserisci il token nella password e lascia vuoto il nome utente.

Un accesso con password arriva a Git tramite il suo ambiente, non la sua riga di comando, e il log
del compito mostra solo l'URL originale.
