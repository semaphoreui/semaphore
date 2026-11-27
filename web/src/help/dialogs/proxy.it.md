---
form: src/components/ProxyForm.vue
---

# Modifica Proxy

Un proxy è un host attraverso cui i task del progetto raggiungono server a cui non possono
collegarsi direttamente. Sceglilo nel campo **Proxy** di un repository, per scaricare il
repository attraverso di esso, o di un inventario, per collegare Ansible agli host attraverso di
esso.

## Nome {#proxyForm-name}

Obbligatorio e unico nel progetto. Repository e inventari mostrano il proxy con questo nome.

## Tipo {#proxyForm-type}

- **SSH (jump host)**: Semaphore si collega all'host via SSH e da lì apre la connessione al
  server. Solo questo tipo può richiedere un altro proxy.
- **SOCKS5**, **HTTP**, **HTTPS**: un server proxy che trasporta la connessione. Git lo usa da sé
  per un repository `https`; Semaphore vi fa passare una connessione SSH, quindi non serve
  installare altri strumenti.

## Host {#proxyForm-host}

Obbligatorio. Un nome host o un indirizzo IP, senza utente, porta né schema. Un nome host può
contenere solo lettere, cifre, punti, trattini e trattini bassi.

## Port {#proxyForm-port}

Facoltativo, da 1 a 65535. Vuoto significa `22` per un jump host SSH. Indica la porta per gli altri
tipi: senza, Git e le connessioni SSH presumono porte diverse.

## User {#proxyForm-user}

Solo SSH: il login sul jump host. Vuoto significa l'utente con cui gira Semaphore. Lettere, cifre,
punti, trattini e trattini bassi, senza un trattino all'inizio. Un proxy SOCKS5 o HTTP prende il
login dalla **Chiave di accesso**.

## Chiave di accesso {#proxyForm-sshKeyId}

Facoltativo. La chiave del negozio chiavi con cui si accede al proxy stesso, di solito non la
chiave dei server dietro di esso:

- una **Chiave SSH** per un jump host SSH;
- un **Accesso con password** per SOCKS5, HTTP e HTTPS. La password viene passata nell'ambiente,
  mai su una riga di comando.

## Requires Proxy {#proxyForm-requiresProxyId}

Solo SSH, facoltativo. Un altro proxy SSH del progetto da attraversare prima, per un jump host
raggiungibile solo da un altro. Una catena contiene al massimo 5 proxy e non può tornare a un proxy
che ne fa già parte.
