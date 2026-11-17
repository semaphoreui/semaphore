---
form: src/components/ProjectForm.vue
---

# Nuovo progetto

Un progetto contiene i modelli di compito, gli inventari, i repository, i gruppi di variabili e le
chiavi di un lavoro, e le persone che ci lavorano. Diventi il proprietario del progetto che crei.
Solo un amministratore può creare un progetto, a meno che l'amministratore del server non lo
consenta a ogni utente. Ogni impostazione qui sotto può essere modificata in seguito nella pagina
Impostazioni del progetto.

## Nome progetto {#newProject-name}

Il nome con cui il progetto compare nel menu dei progetti in cima alla barra laterale. Il campo è
obbligatorio.

## Numero massimo di compiti paralleli {#newProject-maxParallelTasks}

Quanti compiti di questo progetto possono essere eseguiti allo stesso tempo. Un compito avviato
oltre il limite attende in coda finché uno dei compiti in esecuzione non termina.

Lascia il campo vuoto o impostalo a `0` per non avere un limite proprio. Il limite dell'intero
server, impostato dal suo amministratore, si applica in ogni caso.

## ID chat Telegram {#newProject-tg}

Opzionale. La chat Telegram che riceve gli avvisi di questo progetto. Quando il campo è vuoto, gli
avvisi vanno alla chat impostata per l'intero server.

Il campo ha effetto solo quando l'amministratore ha attivato gli avvisi Telegram sul server e gli
avvisi sono consentiti per questo progetto.

## Consenti avvisi per questo progetto {#newProject-alert}

Attiva gli avvisi per i compiti di questo progetto. La destinazione degli avvisi è impostata per
l'intero server dal suo amministratore: e-mail, Telegram, Slack, Rocket.Chat, Microsoft Teams,
DingTalk o Gotify.

- Un'e-mail viene inviata quando un compito fallisce, ai membri del progetto che possono vedere il
  compito e hanno attivato gli avvisi nel proprio account.
- Un messaggio ai servizi di messaggistica viene inviato quando un compito riesce, fallisce o attende
  una conferma.

## Demo {#newProject-demo}

Riempie il nuovo progetto di esempi con cui provare Semaphore: il repository pubblico
`https://github.com/semaphoreui/semaphore-demo.git`, inventari, una chiave con la password del vault
e modelli di compito per Ansible, OpenTofu, Terragrunt, Bash e PowerShell, suddivisi nelle viste
**Build**, **Deploy** e **Tools**.

Lascia l'interruttore spento per un progetto senza esempi: riceve solo la chiave **None** e la vista
**All**. Gli esempi sono normali elementi del progetto: modificali
o eliminali come preferisci.
